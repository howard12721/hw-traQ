package main

import (
	"fmt"
	"strings"
)

type gazerNotificationPayload struct {
	MessageID string `json:"messageId"`
	ChannelID string `json:"channelId"`
	AuthorID  string `json:"authorId"`
	Content   string `json:"content"`
	Pattern   string `json:"pattern"`
	CreatedAt string `json:"createdAt"`
}

func formatGazerNotification(payload gazerNotificationPayload) (string, error) {
	return fmt.Sprintf(
		"Gazer通知: 「%s」に一致しました\n%s\n元メッセージ: /messages/%s\nこのDMはプッシュ通知用のゲートウェイです。アプリ内ではサイドバーのGazerタブに表示されます。",
		payload.Pattern,
		gazerNotificationExcerpt(payload.Content, 80),
		payload.MessageID,
	), nil
}

func gazerNotificationExcerpt(content string, maxRunes int) string {
	normalized := strings.Join(strings.Fields(content), " ")
	if normalized == "" {
		return "(本文なし)"
	}
	runes := []rune(normalized)
	if len(runes) <= maxRunes {
		return normalized
	}
	return string(runes[:maxRunes]) + "..."
}
