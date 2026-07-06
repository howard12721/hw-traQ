package main

import (
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
	if !strings.HasPrefix(content, "Gazer通知:") {
		t.Fatalf("content = %q, want readable notification body", content)
	}
	if !strings.Contains(content, "障害|deploy") {
		t.Fatalf("content = %q, want pattern", content)
	}
	if !strings.Contains(content, "/messages/message-id") {
		t.Fatalf("content = %q, want message link", content)
	}
	if strings.Contains(content, "hw-traq-gazer") || strings.Contains(content, "eyJ") {
		t.Fatalf("content = %q, want no machine-readable payload", content)
	}
}

func TestGazerNotificationExcerpt(t *testing.T) {
	if got := gazerNotificationExcerpt("hello\n  world", 80); got != "hello world" {
		t.Fatalf("excerpt = %q, want normalized text", got)
	}
	if got := gazerNotificationExcerpt("", 80); got != "(本文なし)" {
		t.Fatalf("excerpt = %q, want empty fallback", got)
	}
	if got := gazerNotificationExcerpt("1234567890", 4); got != "1234..." {
		t.Fatalf("excerpt = %q, want truncated text", got)
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
