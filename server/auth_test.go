package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestAuthenticatedMe(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/users/me" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/api/v3/users/me")
		}
		if got := r.Header.Get("Cookie"); got != "r_session=session-value" {
			t.Fatalf("cookie = %q, want %q", got, "r_session=session-value")
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(traQUser{
			ID:          "user-id",
			Name:        "howard127",
			DisplayName: "Howard",
			IconFileID:  "icon-file-id",
			State:       1,
			Permissions: []string{"get_me"},
			Groups:      []string{"group-id"},
		}); err != nil {
			t.Fatal(err)
		}
	}))
	defer upstream.Close()

	e := newTestServer(t, upstream.URL+"/api/v3")
	req := httptest.NewRequest(http.MethodGet, "/internal/v1/me", nil)
	req.Header.Set("Cookie", "r_session=session-value")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	want := `{"user":{"id":"user-id","name":"howard127","displayName":"Howard","iconFileId":"icon-file-id","bot":false,"state":1,"permissions":["get_me"],"groups":["group-id"]}}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Fatalf("body = %q, want %q", got, want)
	}
}

func TestAuthenticatedMeForwardsAuthorization(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token-value" {
			t.Fatalf("authorization = %q, want %q", got, "Bearer token-value")
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(traQUser{
			ID:          "user-id",
			Name:        "howard127",
			DisplayName: "Howard",
			IconFileID:  "icon-file-id",
			State:       1,
		}); err != nil {
			t.Fatal(err)
		}
	}))
	defer upstream.Close()

	e := newTestServer(t, upstream.URL+"/api/v3")
	req := httptest.NewRequest(http.MethodGet, "/internal/v1/me", nil)
	req.Header.Set("Authorization", "Bearer token-value")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestAuthenticatedMeRequiresCredentials(t *testing.T) {
	e := newTestServer(t, "http://127.0.0.1/api/v3")
	req := httptest.NewRequest(http.MethodGet, "/internal/v1/me", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestAuthenticatedMeRejectsInvalidTraQSession(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer upstream.Close()

	e := newTestServer(t, upstream.URL+"/api/v3")
	req := httptest.NewRequest(http.MethodGet, "/internal/v1/me", nil)
	req.Header.Set("Cookie", "r_session=invalid")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestAuthenticatedMeReturnsBadGatewayForUpstreamErrors(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer upstream.Close()

	e := newTestServer(t, upstream.URL+"/api/v3")
	req := httptest.NewRequest(http.MethodGet, "/internal/v1/me", nil)
	req.Header.Set("Cookie", "r_session=session-value")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadGateway, rec.Body.String())
	}
}

func TestNormalizeTraQAPIBaseURL(t *testing.T) {
	got, err := normalizeTraQAPIBaseURL("https://q.trap.jp/api/v3/")
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://q.trap.jp/api/v3" {
		t.Fatalf("baseURL = %q, want %q", got, "https://q.trap.jp/api/v3")
	}

	if _, err := normalizeTraQAPIBaseURL("/api/v3"); err == nil {
		t.Fatal("expected error for relative URL")
	}
}

func newTestServer(t *testing.T, apiBaseURL string) *echo.Echo {
	t.Helper()

	authenticator, err := newTraQAuthenticator(nil, apiBaseURL)
	if err != nil {
		t.Fatal(err)
	}
	return newServerWithConfig(serverConfig{authenticator: authenticator})
}
