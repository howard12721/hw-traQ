package main

import (
	"fmt"
)

type gazerNotificationPayload struct {
	MessageID   string `json:"messageId"`
	ChannelID   string `json:"channelId"`
	AuthorID    string `json:"authorId"`
	Content     string `json:"content"`
	Pattern     string `json:"pattern"`
	DisplayName string `json:"displayName"`
	CreatedAt   string `json:"createdAt"`
}

func formatGazerNotification(payload gazerNotificationPayload) (string, error) {
	displayName := payload.DisplayName
	if displayName == "" {
		displayName = payload.Pattern
	}
	return fmt.Sprintf(
		"%s\nhttps://q.trap.jp/messages/%s",
		displayName,
		payload.MessageID,
	), nil
}
