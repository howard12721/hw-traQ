package main

import (
	"context"
	"database/sql"
)

const gazerSchemaMigration = "20260706_gazer_schema"

type gazerSetting struct {
	UserID      string       `json:"-"`
	Entries     []gazerEntry `json:"entries"`
	Enabled     bool         `json:"enabled"`
	AccessToken string       `json:"-"`
}

type gazerEntry struct {
	ID          int64  `json:"id,omitempty"`
	Pattern     string `json:"pattern"`
	IncludeSelf bool   `json:"includeSelf"`
	IncludeBots bool   `json:"includeBots"`
}

type gazerStore struct {
	db *sql.DB
}

func newGazerStore(db *sql.DB) *gazerStore {
	return &gazerStore{db: db}
}

func (s *gazerStore) migrate(ctx context.Context) error {
	if err := s.ensureMigrationTable(ctx); err != nil {
		return err
	}
	return s.runMigrationOnce(ctx, gazerSchemaMigration, s.createSchema)
}

func (s *gazerStore) ensureMigrationTable(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS hw_traq_schema_migrations (
  name VARCHAR(128) NOT NULL PRIMARY KEY,
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`)
	return err
}

func (s *gazerStore) createSchema(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS gazer_users (
  user_id VARCHAR(36) NOT NULL PRIMARY KEY,
  access_token TEXT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`); err != nil {
		return err
	}

	if _, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS gazer_entries (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id VARCHAR(36) NOT NULL,
  pattern TEXT NOT NULL,
  include_self BOOLEAN NOT NULL DEFAULT FALSE,
  include_bots BOOLEAN NOT NULL DEFAULT FALSE,
  position INT NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_gazer_entries_user_id_position (user_id, position)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`); err != nil {
		return err
	}
	return nil
}

func (s *gazerStore) runMigrationOnce(ctx context.Context, name string, fn func(context.Context) error) error {
	applied, err := s.isMigrationApplied(ctx, name)
	if err != nil {
		return err
	}
	if applied {
		return nil
	}
	if err := fn(ctx); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO hw_traq_schema_migrations (name)
VALUES (?)
ON DUPLICATE KEY UPDATE name = VALUES(name)`, name)
	return err
}

func (s *gazerStore) isMigrationApplied(ctx context.Context, name string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM hw_traq_schema_migrations
WHERE name = ?`, name).Scan(&count)
	return count > 0, err
}

func (s *gazerStore) get(ctx context.Context, userID string) (gazerSetting, error) {
	setting := gazerSetting{UserID: userID}

	err := s.db.QueryRowContext(ctx, `
SELECT COALESCE(access_token, '')
FROM gazer_users
WHERE user_id = ?`, userID).Scan(&setting.AccessToken)
	switch {
	case err == nil:
	case err == sql.ErrNoRows:
	default:
		return setting, err
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT id, pattern, include_self, include_bots
FROM gazer_entries
WHERE user_id = ?
ORDER BY position, id`, userID)
	if err != nil {
		return setting, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry gazerEntry
		if err := rows.Scan(&entry.ID, &entry.Pattern, &entry.IncludeSelf, &entry.IncludeBots); err != nil {
			return setting, err
		}
		setting.Entries = append(setting.Entries, entry)
	}
	if err := rows.Err(); err != nil {
		return setting, err
	}

	setting.Enabled = len(setting.Entries) > 0
	return setting, nil
}

func (s *gazerStore) upsert(ctx context.Context, setting gazerSetting) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO gazer_users (user_id)
VALUES (?)
ON DUPLICATE KEY UPDATE user_id = VALUES(user_id)`, setting.UserID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM gazer_entries WHERE user_id = ?`, setting.UserID); err != nil {
		return err
	}
	for i, entry := range setting.Entries {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO gazer_entries (user_id, pattern, include_self, include_bots, position)
VALUES (?, ?, ?, ?, ?)`,
			setting.UserID,
			entry.Pattern,
			entry.IncludeSelf,
			entry.IncludeBots,
			i,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *gazerStore) updateAccessToken(ctx context.Context, userID, accessToken string) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO gazer_users (user_id, access_token)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE
  access_token = VALUES(access_token)`,
		userID,
		accessToken,
	)
	return err
}

func (s *gazerStore) listRestorable(ctx context.Context) ([]gazerSetting, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT user_id
FROM gazer_users
WHERE access_token IS NOT NULL AND access_token <> ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []gazerSetting
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		setting, err := s.get(ctx, userID)
		if err != nil {
			return nil, err
		}
		if setting.Enabled {
			settings = append(settings, setting)
		}
	}
	return settings, rows.Err()
}
