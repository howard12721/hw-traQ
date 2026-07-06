package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
)

const gazerSchemaMigration = "20260706_gazer_schema_v5"
const gazerNotificationMessageUniqueIndex = "uniq_gazer_notifications_user_message"
const gazerNotificationPatternUniqueIndex = "uniq_gazer_notifications_user_message_pattern"

type gazerSetting struct {
	UserID      string       `json:"-"`
	Entries     []gazerEntry `json:"entries"`
	Enabled     bool         `json:"enabled"`
	AccessToken string       `json:"-"`
}

type gazerEntry struct {
	ID          int64  `json:"id,omitempty"`
	Pattern     string `json:"pattern"`
	DisplayName string `json:"displayName"`
	IncludeSelf bool   `json:"includeSelf"`
	IncludeBots bool   `json:"includeBots"`
}

type gazerNotification struct {
	ID          int64  `json:"id"`
	UserID      string `json:"-"`
	MessageID   string `json:"messageId"`
	ChannelID   string `json:"channelId"`
	AuthorID    string `json:"authorId"`
	Content     string `json:"content"`
	Pattern     string `json:"pattern"`
	DisplayName string `json:"displayName"`
	PatternHash string `json:"-"`
	CreatedAt   string `json:"createdAt"`
	NotifiedAt  string `json:"notifiedAt"`
	Read        bool   `json:"read"`
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
  display_name TEXT NOT NULL,
  include_self BOOLEAN NOT NULL DEFAULT FALSE,
  include_bots BOOLEAN NOT NULL DEFAULT FALSE,
  position INT NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_gazer_entries_user_id_position (user_id, position)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`); err != nil {
		return err
	}

	if _, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS gazer_notifications (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id VARCHAR(36) NOT NULL,
  message_id VARCHAR(36) NOT NULL,
  channel_id VARCHAR(36) NOT NULL,
  author_id VARCHAR(36) NOT NULL,
  content MEDIUMTEXT NOT NULL,
  pattern TEXT NOT NULL,
  display_name TEXT NOT NULL,
  pattern_hash VARCHAR(64) NOT NULL,
  message_created_at VARCHAR(64) NOT NULL,
  notified_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  read_at TIMESTAMP NULL DEFAULT NULL,
  INDEX idx_gazer_notifications_user_id_id (user_id, id),
  INDEX idx_gazer_notifications_user_id_read_at (user_id, read_at),
  UNIQUE KEY uniq_gazer_notifications_user_message (user_id, message_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`); err != nil {
		return err
	}
	if err := s.ensureGazerDisplayNames(ctx); err != nil {
		return err
	}
	if err := s.ensureNotificationDedupe(ctx); err != nil {
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
	setting := gazerSetting{
		UserID:  userID,
		Entries: []gazerEntry{},
	}

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
SELECT id, pattern, display_name, include_self, include_bots
FROM gazer_entries
WHERE user_id = ?
ORDER BY position, id`, userID)
	if err != nil {
		return setting, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry gazerEntry
		if err := rows.Scan(&entry.ID, &entry.Pattern, &entry.DisplayName, &entry.IncludeSelf, &entry.IncludeBots); err != nil {
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
INSERT INTO gazer_entries (user_id, pattern, display_name, include_self, include_bots, position)
VALUES (?, ?, ?, ?, ?, ?)`,
			setting.UserID,
			entry.Pattern,
			entry.DisplayName,
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

func (s *gazerStore) createNotification(ctx context.Context, notification gazerNotification) (bool, error) {
	notification.PatternHash = gazerPatternHash(notification.Pattern)
	_, err := s.db.ExecContext(ctx, `
INSERT INTO gazer_notifications (
  user_id,
  message_id,
  channel_id,
  author_id,
  content,
  pattern,
  display_name,
  pattern_hash,
  message_created_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		notification.UserID,
		notification.MessageID,
		notification.ChannelID,
		notification.AuthorID,
		notification.Content,
		notification.Pattern,
		notification.DisplayName,
		notification.PatternHash,
		notification.CreatedAt,
	)
	if isDuplicateEntry(err) {
		return false, nil
	}
	return err == nil, err
}

func (s *gazerStore) listNotifications(ctx context.Context, userID string, limit int) ([]gazerNotification, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT
  id,
  user_id,
  message_id,
  channel_id,
  author_id,
  content,
  pattern,
  display_name,
  message_created_at,
  notified_at,
  read_at IS NOT NULL
FROM gazer_notifications
WHERE user_id = ?
ORDER BY id DESC
LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := []gazerNotification{}
	for rows.Next() {
		var notification gazerNotification
		var notifiedAt time.Time
		if err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.MessageID,
			&notification.ChannelID,
			&notification.AuthorID,
			&notification.Content,
			&notification.Pattern,
			&notification.DisplayName,
			&notification.CreatedAt,
			&notifiedAt,
			&notification.Read,
		); err != nil {
			return nil, err
		}
		notification.NotifiedAt = notifiedAt.Format(time.RFC3339Nano)
		notifications = append(notifications, notification)
	}
	return notifications, rows.Err()
}

func (s *gazerStore) markNotificationsRead(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE gazer_notifications
SET read_at = CURRENT_TIMESTAMP
WHERE user_id = ? AND read_at IS NULL`, userID)
	return err
}

func (s *gazerStore) ensureGazerDisplayNames(ctx context.Context) error {
	hasEntryDisplayName, err := s.columnExists(ctx, "gazer_entries", "display_name")
	if err != nil {
		return err
	}
	if !hasEntryDisplayName {
		if _, err := s.db.ExecContext(ctx, `
ALTER TABLE gazer_entries
ADD COLUMN display_name TEXT NULL AFTER pattern`); err != nil {
			return err
		}
	}
	if _, err := s.db.ExecContext(ctx, `
UPDATE gazer_entries
SET display_name = pattern
WHERE display_name IS NULL OR display_name = ''`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `
ALTER TABLE gazer_entries
MODIFY display_name TEXT NOT NULL`); err != nil {
		return err
	}

	hasNotificationDisplayName, err := s.columnExists(ctx, "gazer_notifications", "display_name")
	if err != nil {
		return err
	}
	if !hasNotificationDisplayName {
		if _, err := s.db.ExecContext(ctx, `
ALTER TABLE gazer_notifications
ADD COLUMN display_name TEXT NULL AFTER pattern`); err != nil {
			return err
		}
	}
	if _, err := s.db.ExecContext(ctx, `
UPDATE gazer_notifications
SET display_name = pattern
WHERE display_name IS NULL OR display_name = ''`); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
ALTER TABLE gazer_notifications
MODIFY display_name TEXT NOT NULL`)
	return err
}

func (s *gazerStore) ensureNotificationDedupe(ctx context.Context) error {
	hasPatternHash, err := s.columnExists(ctx, "gazer_notifications", "pattern_hash")
	if err != nil {
		return err
	}
	if !hasPatternHash {
		if _, err := s.db.ExecContext(ctx, `
ALTER TABLE gazer_notifications
ADD COLUMN pattern_hash VARCHAR(64) NOT NULL DEFAULT '' AFTER pattern`); err != nil {
			return err
		}
	}
	if _, err := s.db.ExecContext(ctx, `
UPDATE gazer_notifications
SET pattern_hash = LOWER(SHA2(pattern, 256))
WHERE pattern_hash = ''`); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `
DELETE n1
FROM gazer_notifications n1
JOIN gazer_notifications n2
  ON n1.user_id = n2.user_id
  AND n1.message_id = n2.message_id
  AND n1.id > n2.id`); err != nil {
		return err
	}

	hasPatternIndex, err := s.indexExists(ctx, "gazer_notifications", gazerNotificationPatternUniqueIndex)
	if err != nil {
		return err
	}
	if hasPatternIndex {
		if _, err := s.db.ExecContext(ctx, `
DROP INDEX uniq_gazer_notifications_user_message_pattern
ON gazer_notifications`); err != nil {
			return err
		}
	}

	hasMessageIndex, err := s.indexExists(ctx, "gazer_notifications", gazerNotificationMessageUniqueIndex)
	if err != nil {
		return err
	}
	if hasMessageIndex {
		return nil
	}
	_, err = s.db.ExecContext(ctx, `
CREATE UNIQUE INDEX uniq_gazer_notifications_user_message
ON gazer_notifications (user_id, message_id)`)
	return err
}

func (s *gazerStore) columnExists(ctx context.Context, tableName, columnName string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = ?
  AND COLUMN_NAME = ?`, tableName, columnName).Scan(&count)
	return count > 0, err
}

func (s *gazerStore) indexExists(ctx context.Context, tableName, indexName string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = ?
  AND INDEX_NAME = ?`, tableName, indexName).Scan(&count)
	return count > 0, err
}

func gazerPatternHash(pattern string) string {
	sum := sha256.Sum256([]byte(pattern))
	return hex.EncodeToString(sum[:])
}

func isDuplicateEntry(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
