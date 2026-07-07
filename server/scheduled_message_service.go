package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

const (
	scheduledMessageMaxContentLength = 10000
	scheduledMessagePollInterval     = 10 * time.Second
	scheduledMessageLockDuration     = 1 * time.Minute
	scheduledMessageDueLimit         = 10
	scheduledMessageRetryBaseDelay   = 1 * time.Minute
	scheduledMessageRetryMaxDelay    = 5 * time.Minute
)

var errInvalidScheduledMessage = errors.New("invalid scheduled message")

type scheduledMessageService struct {
	store        *scheduledMessageStore
	traq         *traQClient
	workerID     string
	pollInterval time.Duration
}

func newScheduledMessageService(store *scheduledMessageStore, traq *traQClient) *scheduledMessageService {
	return &scheduledMessageService{
		store:        store,
		traq:         traq,
		workerID:     newScheduledMessageWorkerID(),
		pollInterval: scheduledMessagePollInterval,
	}
}

func (s *scheduledMessageService) start(ctx context.Context) {
	if s == nil || s.store == nil || s.traq == nil {
		return
	}
	go s.run(ctx)
}

func (s *scheduledMessageService) create(ctx context.Context, user *traQUser, credential traQCredential, channelID, content string, scheduledAt time.Time) (scheduledMessage, error) {
	content = strings.TrimRight(content, "\r\n")
	if user == nil || user.ID == "" || channelID == "" || content == "" || !credential.hasAuth() {
		return scheduledMessage{}, errInvalidScheduledMessage
	}
	if len([]rune(content)) > scheduledMessageMaxContentLength {
		return scheduledMessage{}, errInvalidScheduledMessage
	}
	if !scheduledAt.After(time.Now()) {
		return scheduledMessage{}, errInvalidScheduledMessage
	}

	return s.store.create(ctx, scheduledMessage{
		UserID:      user.ID,
		ChannelID:   channelID,
		Content:     content,
		ScheduledAt: scheduledAt.UTC(),
		Credential:  credential,
	})
}

func (s *scheduledMessageService) list(ctx context.Context, userID string) ([]scheduledMessage, error) {
	return s.store.listPending(ctx, userID)
}

func (s *scheduledMessageService) cancel(ctx context.Context, userID, id string) (bool, error) {
	if id == "" {
		return false, errInvalidScheduledMessage
	}
	return s.store.cancel(ctx, userID, id)
}

func (s *scheduledMessageService) run(ctx context.Context) {
	if err := s.dispatchDueMessages(ctx); err != nil && ctx.Err() == nil {
		slog.Warn("scheduled message dispatch failed", "error", err)
	}

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		if err := s.dispatchDueMessages(ctx); err != nil && ctx.Err() == nil {
			slog.Warn("scheduled message dispatch failed", "error", err)
		}
	}
}

func (s *scheduledMessageService) dispatchDueMessages(ctx context.Context) error {
	for {
		messages, err := s.store.lockDue(
			ctx,
			s.workerID,
			time.Now().UTC(),
			scheduledMessageDueLimit,
			scheduledMessageLockDuration,
		)
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			return nil
		}

		for _, message := range messages {
			if err := s.traq.postUserMessage(ctx, message.Credential, message.ChannelID, message.Content); err != nil {
				slog.Warn(
					"scheduled message post failed",
					"id",
					message.ID,
					"userID",
					message.UserID,
					"channelID",
					message.ChannelID,
					"error",
					err,
				)
				if err := s.store.markFailed(
					ctx,
					message.ID,
					s.workerID,
					err.Error(),
					time.Now().Add(nextScheduledMessageRetryDelay(message.FailedAttempts)).UTC(),
				); err != nil {
					return err
				}
				continue
			}
			if err := s.store.markSent(ctx, message.ID, s.workerID); err != nil {
				return err
			}
		}

		if len(messages) < scheduledMessageDueLimit {
			return nil
		}
	}
}

func nextScheduledMessageRetryDelay(failedAttempts int) time.Duration {
	delay := scheduledMessageRetryBaseDelay
	for i := 0; i < failedAttempts; i++ {
		delay *= 2
		if delay >= scheduledMessageRetryMaxDelay {
			return scheduledMessageRetryMaxDelay
		}
	}
	return delay
}

func newScheduledMessageWorkerID() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "unknown"
	}
	return fmt.Sprintf("%s:%d:%s", host, os.Getpid(), time.Now().UTC().Format("20060102150405.000000000"))
}
