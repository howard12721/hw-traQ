package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
)

const (
	defaultTraQAPIBaseURL = "https://q.trap.jp/api/v3"
	traQUserContextKey    = "traqUser"
)

var errUnauthenticated = errors.New("unauthenticated")

type traQUser struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	IconFileID  string   `json:"iconFileId"`
	Bot         bool     `json:"bot"`
	State       int      `json:"state"`
	Permissions []string `json:"permissions,omitempty"`
	Groups      []string `json:"groups,omitempty"`
}

type meResponse struct {
	User *traQUser `json:"user"`
}

type traQAuthenticator struct {
	client     *http.Client
	meEndpoint string
}

type traQCredential struct {
	cookies       []string
	authorization string
}

func credentialFromRequest(r *http.Request) traQCredential {
	return traQCredential{
		cookies:       r.Header.Values("Cookie"),
		authorization: r.Header.Get("Authorization"),
	}
}

func credentialFromAccessToken(accessToken string) traQCredential {
	return traQCredential{authorization: "Bearer " + accessToken}
}

func (c traQCredential) hasAuth() bool {
	return len(c.cookies) > 0 || c.authorization != ""
}

func (c traQCredential) applyToHeader(header http.Header) {
	for _, cookie := range c.cookies {
		header.Add("Cookie", cookie)
	}
	if c.authorization != "" {
		header.Set("Authorization", c.authorization)
	}
}

func newTraQAuthenticator(client *http.Client, apiBaseURL string) (*traQAuthenticator, error) {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	baseURL, err := normalizeTraQAPIBaseURL(apiBaseURL)
	if err != nil {
		return nil, err
	}

	return &traQAuthenticator{
		client:     client,
		meEndpoint: baseURL + "/users/me",
	}, nil
}

func mustTraQAuthenticatorFromEnv() *traQAuthenticator {
	authenticator, err := newTraQAuthenticator(nil, traQAPIBaseURLFromEnv())
	if err != nil {
		panic(err)
	}
	return authenticator
}

func traQAPIBaseURLFromEnv() string {
	if baseURL := os.Getenv("TRAQ_API_BASE_URL"); baseURL != "" {
		return baseURL
	}
	return defaultTraQAPIBaseURL
}

func normalizeTraQAPIBaseURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse traQ api base url: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("traQ api base url must be absolute: %q", rawURL)
	}
	u.RawQuery = ""
	u.Fragment = ""
	u.Path = strings.TrimRight(u.Path, "/")
	return u.String(), nil
}

func requireTraQUser(authenticator *traQAuthenticator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			user, err := authenticator.authenticate(c.Request())
			if err != nil {
				if errors.Is(err, errUnauthenticated) {
					return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
				}
				return echo.NewHTTPError(http.StatusBadGateway, "failed to authenticate with traQ")
			}

			c.Set(traQUserContextKey, user)
			return next(c)
		}
	}
}

func (a *traQAuthenticator) authenticate(r *http.Request) (*traQUser, error) {
	credential := credentialFromRequest(r)
	if !credential.hasAuth() {
		return nil, errUnauthenticated
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, a.meEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	credential.applyToHeader(req.Header)

	res, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, errUnauthenticated
	default:
		return nil, fmt.Errorf("unexpected traQ auth status: %d", res.StatusCode)
	}

	var user traQUser
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func currentTraQUser(c *echo.Context) (*traQUser, bool) {
	user, ok := c.Get(traQUserContextKey).(*traQUser)
	return user, ok
}
