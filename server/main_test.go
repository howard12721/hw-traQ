package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestInternalRoutes(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "health",
			path:       "/internal/health",
			wantStatus: http.StatusOK,
			wantBody:   `{"status":"ok"}` + "\n",
		},
		{
			name:       "ping",
			path:       "/internal/v1/ping",
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"pong"}` + "\n",
		},
		{
			name:       "api is not handled here",
			path:       "/api/v3/users/me",
			wantStatus: http.StatusNotFound,
			wantBody:   "{\"message\":\"Not Found\"}\n",
		},
	}

	e := newTestServer(t, "http://127.0.0.1/api/v3")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if rec.Body.String() != tt.wantBody {
				t.Fatalf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestGetGazerOAuthClient(t *testing.T) {
	t.Run("configured", func(t *testing.T) {
		e := echo.New()
		e.GET("/oauth-client", getGazerOAuthClient("client-id"))

		req := httptest.NewRequest(http.MethodGet, "/oauth-client", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		if rec.Body.String() != `{"clientId":"client-id"}`+"\n" {
			t.Fatalf("body = %q", rec.Body.String())
		}
	})

	t.Run("not configured", func(t *testing.T) {
		e := echo.New()
		e.GET("/oauth-client", getGazerOAuthClient(""))

		req := httptest.NewRequest(http.MethodGet, "/oauth-client", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusServiceUnavailable, rec.Body.String())
		}
	})
}

func TestListenAddr(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		t.Setenv("ADDR", "")
		t.Setenv("PORT", "")

		if got := listenAddr(); got != ":8080" {
			t.Fatalf("listenAddr() = %q, want %q", got, ":8080")
		}
	})

	t.Run("port", func(t *testing.T) {
		t.Setenv("ADDR", "")
		t.Setenv("PORT", "18080")

		if got := listenAddr(); got != ":18080" {
			t.Fatalf("listenAddr() = %q, want %q", got, ":18080")
		}
	})

	t.Run("addr", func(t *testing.T) {
		t.Setenv("ADDR", "127.0.0.1:19091")
		t.Setenv("PORT", "18080")

		if got := listenAddr(); got != "127.0.0.1:19091" {
			t.Fatalf("listenAddr() = %q, want %q", got, "127.0.0.1:19091")
		}
	})
}
