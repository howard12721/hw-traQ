package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const defaultRequestBodyLimit = 1 << 20

type traQClient struct {
	apiBaseURL string
	wsURL      string
	httpClient *http.Client
	botToken   string
}

type traQMessage struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	ChannelID string `json:"channelId"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

type traQDMChannel struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
}

type traQOAuthToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

func newTraQClient(apiBaseURL, wsURL, botToken string) (*traQClient, error) {
	apiBaseURL, err := normalizeTraQAPIBaseURL(apiBaseURL)
	if err != nil {
		return nil, err
	}
	if wsURL == "" {
		wsURL, err = wsURLFromAPIBaseURL(apiBaseURL)
		if err != nil {
			return nil, err
		}
	}
	return &traQClient{
		apiBaseURL: apiBaseURL,
		wsURL:      wsURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		botToken:   botToken,
	}, nil
}

func mustTraQClientFromEnv() *traQClient {
	client, err := newTraQClient(
		traQAPIBaseURLFromEnv(),
		os.Getenv("TRAQ_WS_URL"),
		os.Getenv("TRAQ_BOT_TOKEN"),
	)
	if err != nil {
		panic(err)
	}
	return client
}

func wsURLFromAPIBaseURL(apiBaseURL string) (string, error) {
	u, err := url.Parse(apiBaseURL)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	default:
		return "", fmt.Errorf("unsupported traQ api scheme: %s", u.Scheme)
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/ws"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

func (c *traQClient) getMessage(ctx context.Context, credential traQCredential, messageID string) (*traQMessage, error) {
	var message traQMessage
	if err := c.getJSON(ctx, credential, "/messages/"+url.PathEscape(messageID), &message); err != nil {
		return nil, err
	}
	return &message, nil
}

func (c *traQClient) getUser(ctx context.Context, credential traQCredential, userID string) (*traQUser, error) {
	var user traQUser
	if err := c.getJSON(ctx, credential, "/users/"+url.PathEscape(userID), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *traQClient) getMe(ctx context.Context, credential traQCredential) (*traQUser, error) {
	var user traQUser
	if err := c.getJSON(ctx, credential, "/users/me", &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *traQClient) getBotMe(ctx context.Context) (*traQUser, error) {
	var user traQUser
	if err := c.getBotJSON(ctx, "/users/me", &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *traQClient) exchangeOAuthCode(ctx context.Context, clientID, code, codeVerifier, redirectURI string) (*traQOAuthToken, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", clientID)
	form.Set("code", code)
	form.Set("code_verifier", codeVerifier)
	form.Set("redirect_uri", redirectURI)

	var token traQOAuthToken
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.apiBaseURL+"/oauth2/token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := c.doJSON(req, &token); err != nil {
		return nil, err
	}
	if token.AccessToken == "" {
		return nil, fmt.Errorf("traQ OAuth token response does not contain access_token")
	}
	return &token, nil
}

func (c *traQClient) getBotDMChannel(ctx context.Context, userID string) (*traQDMChannel, error) {
	var channel traQDMChannel
	if err := c.getBotJSON(ctx, "/users/"+url.PathEscape(userID)+"/dm-channel", &channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

func (c *traQClient) postBotMessage(ctx context.Context, channelID, content string) error {
	body := map[string]any{
		"content": content,
		"embed":   true,
	}
	return c.postBotJSON(ctx, "/channels/"+url.PathEscape(channelID)+"/messages", body, nil)
}

func (c *traQClient) dialUserWebSocket(ctx context.Context, credential traQCredential) (*websocket.Conn, error) {
	header := http.Header{}
	credential.applyToHeader(header)
	conn, res, err := websocket.DefaultDialer.DialContext(ctx, c.wsURL, header)
	if err != nil && res != nil && (res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden) {
		return nil, errUnauthenticated
	}
	return conn, err
}

func (c *traQClient) getJSON(ctx context.Context, credential traQCredential, path string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	credential.applyToHeader(req.Header)
	return c.doJSON(req, v)
}

func (c *traQClient) getBotJSON(ctx context.Context, path string, v any) error {
	req, err := c.newBotRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return c.doJSON(req, v)
}

func (c *traQClient) postBotJSON(ctx context.Context, path string, body any, v any) error {
	var reader io.Reader
	if body != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return err
		}
		reader = buf
	}
	req, err := c.newBotRequest(ctx, http.MethodPost, path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doJSON(req, v)
}

func (c *traQClient) newBotRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	if c.botToken == "" {
		return nil, fmt.Errorf("TRAQ_BOT_TOKEN is not configured")
	}
	req, err := http.NewRequestWithContext(ctx, method, c.apiBaseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.botToken)
	return req, nil
}

func (c *traQClient) doJSON(req *http.Request, v any) error {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return errUnauthenticated
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("traQ API returned %d", res.StatusCode)
	}
	if v == nil {
		io.Copy(io.Discard, io.LimitReader(res.Body, defaultRequestBodyLimit))
		return nil
	}
	return json.NewDecoder(io.LimitReader(res.Body, defaultRequestBodyLimit)).Decode(v)
}
