package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const scheduledMessageSchemaMigration = "20260707_scheduled_messages_schema_v1"

type scheduledMessage struct {
	ID             string         `json:"id"`
	UserID         string         `json:"-"`
	ChannelID      string         `json:"channelId"`
	Content        string         `json:"content"`
	ScheduledAt    time.Time      `json:"scheduledAt"`
	CreatedAt      time.Time      `json:"createdAt"`
	RetryAt        *time.Time     `json:"retryAt,omitempty"`
	LastError      string         `json:"lastError,omitempty"`
	FailedAttempts int            `json:"failedAttempts,omitempty"`
	Credential     traQCredential `json:"-"`
}

type scheduledMessageStore struct {
	db *sql.DB
}

func newScheduledMessageStore(db *sql.DB) *scheduledMessageStore {
	return &scheduledMessageStore{db: db}
}

func (s *scheduledMessageStore) migrate(ctx context.Context) error {
	if err := s.ensureMigrationTable(ctx); err != nil {
		return err
	}
	return s.runMigrationOnce(ctx, scheduledMessageSchemaMigration, s.createSchema)
}

func (s *scheduledMessageStore) ensureMigrationTable(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS hw_traq_schema_migrations (
  name VARCHAR(128) NOT NULL PRIMARY KEY,
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`)
	return err
}

func (s *scheduledMessageStore) createSchema(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS scheduled_messages (
  id VARCHAR(36) NOT NULL PRIMARY KEY,
  user_id VARCHAR(36) NOT NULL,
  channel_id VARCHAR(36) NOT NULL,
  content MEDIUMTEXT NOT NULL,
  scheduled_at DATETIME(6) NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  sent_at DATETIME(6) NULL DEFAULT NULL,
  canceled_at DATETIME(6) NULL DEFAULT NULL,
  failed_attempts INT NOT NULL DEFAULT 0,
  retry_at DATETIME(6) NULL DEFAULT NULL,
  last_error TEXT NULL,
  auth_authorization TEXT NULL,
  auth_cookies MEDIUMTEXT NOT NULL,
  locked_by VARCHAR(128) NULL,
  locked_until DATETIME(6) NULL DEFAULT NULL,
  INDEX idx_scheduled_messages_user_pending (user_id, sent_at, canceled_at, scheduled_at),
  INDEX idx_scheduled_messages_due (sent_at, canceled_at, scheduled_at, retry_at, locked_until)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin`)
	return err
}

func (s *scheduledMessageStore) runMigrationOnce(ctx context.Context, name string, fn func(context.Context) error) error {
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

func (s *scheduledMessageStore) isMigrationApplied(ctx context.Context, name string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM hw_traq_schema_migrations
WHERE name = ?`, name).Scan(&count)
	return count > 0, err
}

func (s *scheduledMessageStore) create(ctx context.Context, message scheduledMessage) (scheduledMessage, error) {
	id, err := createScheduledMessageID()
	if err != nil {
		return scheduledMessage{}, err
	}
	message.ID = id
	message.CreatedAt = time.Now().UTC()

	authCookies, err := encodeCredentialCookies(message.Credential)
	if err != nil {
		return scheduledMessage{}, err
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO scheduled_messages (
  id,
  user_id,
  channel_id,
  content,
  scheduled_at,
  created_at,
  auth_authorization,
  auth_cookies
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		message.ID,
		message.UserID,
		message.ChannelID,
		message.Content,
		message.ScheduledAt.UTC(),
		message.CreatedAt,
		message.Credential.authorization,
		authCookies,
	)
	return message, err
}

func (s *scheduledMessageStore) listPending(ctx context.Context, userID string) ([]scheduledMessage, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
  id,
  user_id,
  channel_id,
  content,
  scheduled_at,
  created_at,
  retry_at,
  COALESCE(last_error, ''),
  failed_attempts,
  COALESCE(auth_authorization, ''),
  auth_cookies
FROM scheduled_messages
WHERE user_id = ? AND sent_at IS NULL AND canceled_at IS NULL
ORDER BY scheduled_at, id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanScheduledMessages(rows)
}

func (s *scheduledMessageStore) cancel(ctx context.Context, userID, id string) (bool, error) {
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, `
UPDATE scheduled_messages
SET
  canceled_at = ?,
  locked_by = NULL,
  locked_until = NULL
WHERE id = ?
  AND user_id = ?
  AND sent_at IS NULL
  AND canceled_at IS NULL
  AND (locked_until IS NULL OR locked_until <= ?)`,
		now,
		id,
		userID,
		now,
	)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (s *scheduledMessageStore) lockDue(ctx context.Context, workerID string, now time.Time, limit int, lockDuration time.Duration) ([]scheduledMessage, error) {
	if limit <= 0 {
		return nil, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
SELECT
  id,
  user_id,
  channel_id,
  content,
  scheduled_at,
  created_at,
  retry_at,
  COALESCE(last_error, ''),
  failed_attempts,
  COALESCE(auth_authorization, ''),
  auth_cookies
FROM scheduled_messages
WHERE sent_at IS NULL
  AND canceled_at IS NULL
  AND scheduled_at <= ?
  AND (retry_at IS NULL OR retry_at <= ?)
  AND (locked_until IS NULL OR locked_until <= ?)
ORDER BY scheduled_at, id
LIMIT ?
FOR UPDATE`, now.UTC(), now.UTC(), now.UTC(), limit)
	if err != nil {
		return nil, err
	}
	messages, err := scanScheduledMessages(rows)
	rows.Close()
	if err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		return messages, tx.Commit()
	}

	args := []any{workerID, now.Add(lockDuration).UTC()}
	placeholders := make([]string, 0, len(messages))
	for _, message := range messages {
		placeholders = append(placeholders, "?")
		args = append(args, message.ID)
	}
	_, err = tx.ExecContext(ctx, `
UPDATE scheduled_messages
SET locked_by = ?, locked_until = ?
WHERE id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return nil, err
	}
	return messages, tx.Commit()
}

func (s *scheduledMessageStore) markSent(ctx context.Context, id, workerID string) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE scheduled_messages
SET
  sent_at = ?,
  last_error = NULL,
  locked_by = NULL,
  locked_until = NULL
WHERE id = ? AND locked_by = ?`,
		time.Now().UTC(),
		id,
		workerID,
	)
	return err
}

func (s *scheduledMessageStore) markFailed(ctx context.Context, id, workerID, message string, retryAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE scheduled_messages
SET
  failed_attempts = failed_attempts + 1,
  retry_at = ?,
  last_error = ?,
  locked_by = NULL,
  locked_until = NULL
WHERE id = ? AND locked_by = ? AND sent_at IS NULL AND canceled_at IS NULL`,
		retryAt.UTC(),
		truncateScheduledMessageError(message),
		id,
		workerID,
	)
	return err
}

func scanScheduledMessages(rows *sql.Rows) ([]scheduledMessage, error) {
	messages := []scheduledMessage{}
	for rows.Next() {
		var (
			message       scheduledMessage
			retryAt       sql.NullTime
			authCookies   string
			authorization string
		)
		if err := rows.Scan(
			&message.ID,
			&message.UserID,
			&message.ChannelID,
			&message.Content,
			&message.ScheduledAt,
			&message.CreatedAt,
			&retryAt,
			&message.LastError,
			&message.FailedAttempts,
			&authorization,
			&authCookies,
		); err != nil {
			return nil, err
		}
		if retryAt.Valid {
			retryAtUTC := retryAt.Time.UTC()
			message.RetryAt = &retryAtUTC
		}
		message.ScheduledAt = message.ScheduledAt.UTC()
		message.CreatedAt = message.CreatedAt.UTC()
		cookies, err := decodeCredentialCookies(authCookies)
		if err != nil {
			return nil, err
		}
		message.Credential = traQCredential{
			authorization: authorization,
			cookies:       cookies,
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func createScheduledMessageID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	hexed := hex.EncodeToString(b[:])
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		hexed[0:8],
		hexed[8:12],
		hexed[12:16],
		hexed[16:20],
		hexed[20:32],
	), nil
}

func encodeCredentialCookies(credential traQCredential) (string, error) {
	data, err := json.Marshal(credential.cookies)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func decodeCredentialCookies(raw string) ([]string, error) {
	var cookies []string
	if raw == "" {
		return cookies, nil
	}
	if err := json.Unmarshal([]byte(raw), &cookies); err != nil {
		return nil, err
	}
	return cookies, nil
}

func truncateScheduledMessageError(message string) string {
	const maxLength = 1000
	if len(message) <= maxLength {
		return message
	}
	return message[:maxLength]
}
