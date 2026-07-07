package main

import (
	"testing"
	"time"
)

func TestNextScheduledMessageRetryDelay(t *testing.T) {
	tests := []struct {
		failedAttempts int
		want           time.Duration
	}{
		{failedAttempts: 0, want: time.Minute},
		{failedAttempts: 1, want: 2 * time.Minute},
		{failedAttempts: 2, want: 4 * time.Minute},
		{failedAttempts: 3, want: 5 * time.Minute},
	}

	for _, tt := range tests {
		if got := nextScheduledMessageRetryDelay(tt.failedAttempts); got != tt.want {
			t.Fatalf("nextScheduledMessageRetryDelay(%d) = %s, want %s", tt.failedAttempts, got, tt.want)
		}
	}
}

func TestScheduledMessageCredentialCookies(t *testing.T) {
	credential := traQCredential{
		cookies:       []string{"r_session=session", "r_csrf=csrf"},
		authorization: "Bearer token",
	}

	raw, err := encodeCredentialCookies(credential)
	if err != nil {
		t.Fatal(err)
	}
	cookies, err := decodeCredentialCookies(raw)
	if err != nil {
		t.Fatal(err)
	}

	if len(cookies) != len(credential.cookies) {
		t.Fatalf("len(cookies) = %d, want %d", len(cookies), len(credential.cookies))
	}
	for i := range cookies {
		if cookies[i] != credential.cookies[i] {
			t.Fatalf("cookies[%d] = %q, want %q", i, cookies[i], credential.cookies[i])
		}
	}
}

func TestCreateScheduledMessageID(t *testing.T) {
	id, err := createScheduledMessageID()
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 36 {
		t.Fatalf("len(id) = %d, want 36", len(id))
	}
}
