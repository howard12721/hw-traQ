package main

import (
	"context"
	"testing"
)

func TestRefreshWithAccessTokenRequiresStoredToken(t *testing.T) {
	service := newGazerService(nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	service.active["user-id"] = &gazerWorker{cancel: cancel, signature: "old"}

	service.refreshWithAccessToken(&traQUser{ID: "user-id"}, gazerSetting{
		Enabled: true,
		Entries: []gazerEntry{
			{Pattern: "deploy"},
		},
	})

	if service.status("user-id").Running {
		t.Fatal("worker is running without stored access token")
	}
	select {
	case <-ctx.Done():
	default:
		t.Fatal("worker was not cancelled")
	}
}

func TestCompileGazerEntriesSupportsLookbehind(t *testing.T) {
	compiled, err := compileGazerEntries([]gazerEntry{
		{
			Pattern:     `(?<!!{"type":"user","raw":"@)howard`,
			DisplayName: "howard",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	matched, err := compiled[0].pattern.MatchString("hello howard")
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Fatal("pattern did not match plain text")
	}

	matched, err = compiled[0].pattern.MatchString(`!{"type":"user","raw":"@howard`)
	if err != nil {
		t.Fatal(err)
	}
	if matched {
		t.Fatal("pattern matched mention payload")
	}
}
