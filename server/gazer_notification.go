package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

const gazerNotificationPrefix = "<!-- hw-traq-gazer:"

type gazerNotificationPayload struct {
	MessageID string `json:"messageId"`
	ChannelID string `json:"channelId"`
	AuthorID  string `json:"authorId"`
	Content   string `json:"content"`
	Pattern   string `json:"pattern"`
	CreatedAt string `json:"createdAt"`
}

func formatGazerNotification(payload gazerNotificationPayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	return fmt.Sprintf(
		"%s%s -->\nGazer matched `%s`\n\n> %s\n\n/messages/%s",
		gazerNotificationPrefix,
		encoded,
		payload.Pattern,
		payload.Content,
		payload.MessageID,
	), nil
}
