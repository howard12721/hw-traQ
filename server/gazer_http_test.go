package main

import (
	"encoding/json"
	"testing"
)

func TestNewGazerResponseReturnsEmptyEntries(t *testing.T) {
	res := newGazerResponse(newGazerService(nil, nil), "user-id", gazerSetting{})

	body, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}

	want := `{"setting":{"entries":[],"enabled":false},"status":{"running":false,"tokenConfigured":false}}`
	if string(body) != want {
		t.Fatalf("body = %s, want %s", body, want)
	}
}
