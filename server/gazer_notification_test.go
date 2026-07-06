package main

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatGazerNotification(t *testing.T) {
	payload := gazerNotificationPayload{
		MessageID: "message-id",
		ChannelID: "channel-id",
		AuthorID:  "author-id",
		Content:   "障害対応お願いします",
		Pattern:   "障害|deploy",
		CreatedAt: "2026-07-06T12:34:56.000Z",
	}

	content, err := formatGazerNotification(payload)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(content, gazerNotificationPrefix) {
		t.Fatalf("content = %q, want prefix %q", content, gazerNotificationPrefix)
	}
	if !strings.Contains(content, "/messages/message-id") {
		t.Fatalf("content = %q, want message link", content)
	}

	end := strings.Index(content, " -->")
	if end == -1 {
		t.Fatalf("content = %q, want payload terminator", content)
	}
	encoded := strings.TrimPrefix(content[:end], gazerNotificationPrefix)
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}

	var got gazerNotificationPayload
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got != payload {
		t.Fatalf("payload = %#v, want %#v", got, payload)
	}
}

func TestWsURLFromAPIBaseURL(t *testing.T) {
	tests := []struct {
		base string
		want string
	}{
		{base: "https://q.trap.jp/api/v3", want: "wss://q.trap.jp/api/v3/ws"},
		{base: "http://127.0.0.1:3000/api/v3/", want: "ws://127.0.0.1:3000/api/v3/ws"},
	}

	for _, tt := range tests {
		t.Run(tt.base, func(t *testing.T) {
			got, err := wsURLFromAPIBaseURL(tt.base)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("wsURL = %q, want %q", got, tt.want)
			}
		})
	}
}
