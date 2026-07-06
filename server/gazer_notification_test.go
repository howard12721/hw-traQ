package main

import (
	"testing"
)

func TestFormatGazerNotification(t *testing.T) {
	payload := gazerNotificationPayload{
		MessageID:   "message-id",
		ChannelID:   "channel-id",
		AuthorID:    "author-id",
		Content:     "障害対応お願いします",
		Pattern:     "障害|deploy",
		DisplayName: "障害通知",
		CreatedAt:   "2026-07-06T12:34:56.000Z",
	}

	content, err := formatGazerNotification(payload)
	if err != nil {
		t.Fatal(err)
	}
	want := "障害通知\nhttps://q.trap.jp/messages/message-id"
	if content != want {
		t.Fatalf("content = %q, want %q", content, want)
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
