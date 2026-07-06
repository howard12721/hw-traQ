package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExchangeOAuthCode(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/oauth2/token" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/api/v3/oauth2/token")
		}
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Fatalf("content-type = %q, want application/x-www-form-urlencoded", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		wantForm := map[string]string{
			"grant_type":    "authorization_code",
			"client_id":     "client-id",
			"code":          "oauth-code",
			"code_verifier": "code-verifier",
			"redirect_uri":  "https://example.com/settings/gazer",
		}
		for key, want := range wantForm {
			if got := r.Form.Get(key); got != want {
				t.Fatalf("%s = %q, want %q", key, got, want)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(traQOAuthToken{
			AccessToken: "access-token",
			TokenType:   "Bearer",
		}); err != nil {
			t.Fatal(err)
		}
	}))
	defer upstream.Close()

	client, err := newTraQClient(upstream.URL+"/api/v3", "", "")
	if err != nil {
		t.Fatal(err)
	}
	token, err := client.exchangeOAuthCode(
		t.Context(),
		"client-id",
		"oauth-code",
		"code-verifier",
		"https://example.com/settings/gazer",
	)
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "access-token" {
		t.Fatalf("accessToken = %q, want access-token", token.AccessToken)
	}
}
