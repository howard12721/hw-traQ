package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type gazerService struct {
	store            *gazerStore
	traq             *traQClient
	mu               sync.Mutex
	active           map[string]*gazerWorker
	botUserID        string
	botUserIDFetched bool
}

type gazerStatus struct {
	Running         bool   `json:"running"`
	TokenConfigured bool   `json:"tokenConfigured"`
	BotUserID       string `json:"botUserId,omitempty"`
}

type gazerWorker struct {
	cancel    context.CancelFunc
	signature string
}

type websocketEvent struct {
	Type string          `json:"type"`
	Body json.RawMessage `json:"body"`
}

type messageCreatedBody struct {
	ID string `json:"id"`
}

type compiledGazerEntry struct {
	entry   gazerEntry
	pattern *regexp.Regexp
}

var errInvalidGazerPattern = errors.New("invalid gazer pattern")

func newGazerService(store *gazerStore, traq *traQClient) *gazerService {
	return &gazerService{
		store:  store,
		traq:   traq,
		active: map[string]*gazerWorker{},
	}
}

func (s *gazerService) status(userID string) gazerStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, running := s.active[userID]
	return gazerStatus{Running: running, BotUserID: s.botUserID}
}

func (s *gazerService) refreshBotUserID(ctx context.Context) {
	if s == nil || s.traq == nil || s.traq.botToken == "" {
		return
	}

	s.mu.Lock()
	if s.botUserIDFetched {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	bot, err := s.traq.getBotMe(ctx)
	if err != nil {
		slog.Warn("gazer failed to fetch bot user", "error", err)
		return
	}

	s.mu.Lock()
	s.botUserID = bot.ID
	s.botUserIDFetched = true
	s.mu.Unlock()
}

func (s *gazerService) save(ctx context.Context, user *traQUser, credential traQCredential, setting gazerSetting) (gazerSetting, error) {
	current, err := s.store.get(ctx, user.ID)
	if err != nil {
		return gazerSetting{}, err
	}

	setting.UserID = user.ID
	setting.Entries = normalizeGazerEntries(setting.Entries)
	setting.Enabled = len(setting.Entries) > 0
	setting.AccessToken = current.AccessToken
	for _, entry := range setting.Entries {
		if _, err := regexp.Compile(entry.Pattern); err != nil {
			return gazerSetting{}, errInvalidGazerPattern
		}
	}
	if err := s.store.upsert(ctx, setting); err != nil {
		return gazerSetting{}, err
	}
	s.refreshWithBestCredential(user, credential, setting)
	return setting, nil
}

func (s *gazerService) saveAccessToken(ctx context.Context, user *traQUser, accessToken string) (gazerSetting, error) {
	credential := credentialFromAccessToken(accessToken)
	tokenUser, err := s.traq.getMe(ctx, credential)
	if err != nil {
		return gazerSetting{}, err
	}
	if tokenUser.ID != user.ID {
		return gazerSetting{}, errUnauthenticated
	}

	if err := s.store.updateAccessToken(ctx, user.ID, accessToken); err != nil {
		return gazerSetting{}, err
	}
	setting, err := s.store.get(ctx, user.ID)
	if err != nil {
		return gazerSetting{}, err
	}
	s.refresh(user, credential, setting)
	return setting, nil
}

func (s *gazerService) saveAuthorizationCode(ctx context.Context, user *traQUser, clientID, code, codeVerifier, redirectURI string) (gazerSetting, error) {
	token, err := s.traq.exchangeOAuthCode(ctx, clientID, code, codeVerifier, redirectURI)
	if err != nil {
		return gazerSetting{}, err
	}
	return s.saveAccessToken(ctx, user, token.AccessToken)
}

func (s *gazerService) restore(ctx context.Context) error {
	settings, err := s.store.listRestorable(ctx)
	if err != nil {
		return err
	}
	for _, setting := range settings {
		user := &traQUser{ID: setting.UserID}
		s.refresh(user, credentialFromAccessToken(setting.AccessToken), setting)
	}
	return nil
}

func (s *gazerService) refreshWithBestCredential(user *traQUser, fallback traQCredential, setting gazerSetting) {
	if setting.AccessToken != "" {
		s.refresh(user, credentialFromAccessToken(setting.AccessToken), setting)
		return
	}
	s.refresh(user, fallback, setting)
}

func (s *gazerService) refresh(user *traQUser, credential traQCredential, setting gazerSetting) {
	if !setting.Enabled || len(setting.Entries) == 0 || !credential.hasAuth() {
		s.stop(user.ID)
		return
	}

	compiled, err := compileGazerEntries(setting.Entries)
	if err != nil {
		s.stop(user.ID)
		return
	}

	signature := gazerWorkerSignature(setting, credential)
	s.mu.Lock()
	if worker := s.active[user.ID]; worker != nil {
		if worker.signature == signature {
			s.mu.Unlock()
			return
		}
		worker.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.active[user.ID] = &gazerWorker{cancel: cancel, signature: signature}
	s.mu.Unlock()

	go s.run(ctx, user, credential, setting, compiled)
}

func (s *gazerService) stop(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if worker := s.active[userID]; worker != nil {
		worker.cancel()
		delete(s.active, userID)
	}
}

func (s *gazerService) run(ctx context.Context, user *traQUser, credential traQCredential, setting gazerSetting, entries []compiledGazerEntry) {
	defer s.stopIfCurrent(user.ID, ctx)

	backoff := time.Second
	for {
		if err := s.runOnce(ctx, user, credential, entries); err != nil {
			if ctx.Err() != nil {
				return
			}
			if errors.Is(err, errUnauthenticated) {
				slog.Warn("gazer worker stopped by authentication error", "userID", user.ID)
				return
			}
			slog.Warn("gazer worker disconnected", "userID", user.ID, "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func (s *gazerService) stopIfCurrent(userID string, ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	worker := s.active[userID]
	if worker == nil {
		return
	}
	select {
	case <-ctx.Done():
		delete(s.active, userID)
	default:
	}
}

func gazerWorkerSignature(setting gazerSetting, credential traQCredential) string {
	parts := []string{credential.authorization}
	for _, entry := range setting.Entries {
		parts = append(
			parts,
			entry.Pattern,
			boolSignature(entry.IncludeSelf),
			boolSignature(entry.IncludeBots),
		)
	}
	parts = append(parts, credential.cookies...)
	return strings.Join(parts, "\x00")
}

func boolSignature(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func compileGazerEntries(entries []gazerEntry) ([]compiledGazerEntry, error) {
	compiled := make([]compiledGazerEntry, 0, len(entries))
	for _, entry := range entries {
		pattern, err := regexp.Compile(entry.Pattern)
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, compiledGazerEntry{
			entry:   entry,
			pattern: pattern,
		})
	}
	return compiled, nil
}

func normalizeGazerEntries(entries []gazerEntry) []gazerEntry {
	normalized := make([]gazerEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Pattern == "" {
			continue
		}
		normalized = append(normalized, gazerEntry{
			Pattern:     entry.Pattern,
			IncludeSelf: entry.IncludeSelf,
			IncludeBots: entry.IncludeBots,
		})
	}
	return normalized
}

func (s *gazerService) runOnce(ctx context.Context, user *traQUser, credential traQCredential, entries []compiledGazerEntry) error {
	conn, err := s.traq.dialUserWebSocket(ctx, credential)
	if err != nil {
		return err
	}
	defer conn.Close()
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-done:
		}
	}()
	defer close(done)

	if err := conn.WriteMessage(websocket.TextMessage, []byte("timeline_streaming:on")); err != nil {
		return err
	}

	botCache := map[string]bool{}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		var event websocketEvent
		if err := json.Unmarshal(data, &event); err != nil {
			continue
		}
		if event.Type != "MESSAGE_CREATED" {
			continue
		}

		var body messageCreatedBody
		if err := json.Unmarshal(event.Body, &body); err != nil || body.ID == "" {
			continue
		}
		s.processMessage(ctx, user, credential, entries, botCache, body.ID)
	}
}

func (s *gazerService) processMessage(ctx context.Context, user *traQUser, credential traQCredential, entries []compiledGazerEntry, botCache map[string]bool, messageID string) {
	message, err := s.traq.getMessage(ctx, credential, messageID)
	if err != nil {
		slog.Warn("gazer failed to fetch message", "messageID", messageID, "error", err)
		return
	}

	var (
		isBot      bool
		isBotKnown bool
		matches    []gazerEntry
	)
	for _, entry := range entries {
		if message.UserID == user.ID && !entry.entry.IncludeSelf {
			continue
		}
		if !entry.entry.IncludeBots {
			if !isBotKnown {
				var ok bool
				isBot, ok = botCache[message.UserID]
				if !ok {
					author, err := s.traq.getUser(ctx, credential, message.UserID)
					if err != nil {
						slog.Warn("gazer failed to fetch message author", "userID", message.UserID, "error", err)
						return
					}
					isBot = author.Bot
					botCache[message.UserID] = isBot
				}
				isBotKnown = true
			}
			if isBot {
				continue
			}
		}
		if entry.pattern.MatchString(message.Content) {
			matches = append(matches, entry.entry)
		}
	}
	if len(matches) == 0 {
		return
	}

	dm, err := s.traq.getBotDMChannel(ctx, user.ID)
	if err != nil {
		slog.Warn("gazer failed to get bot dm channel", "userID", user.ID, "error", err)
		return
	}
	for _, entry := range matches {
		payload := gazerNotificationPayload{
			MessageID: message.ID,
			ChannelID: message.ChannelID,
			AuthorID:  message.UserID,
			Content:   message.Content,
			Pattern:   entry.Pattern,
			CreatedAt: message.CreatedAt,
		}
		if err := s.store.createNotification(ctx, gazerNotification{
			UserID:    user.ID,
			MessageID: payload.MessageID,
			ChannelID: payload.ChannelID,
			AuthorID:  payload.AuthorID,
			Content:   payload.Content,
			Pattern:   payload.Pattern,
			CreatedAt: payload.CreatedAt,
		}); err != nil {
			slog.Warn("gazer failed to save notification", "userID", user.ID, "messageID", message.ID, "error", err)
		}
		content, err := formatGazerNotification(payload)
		if err != nil {
			slog.Warn("gazer failed to format notification", "messageID", message.ID, "error", err)
			return
		}
		if err := s.traq.postBotMessage(ctx, dm.ID, content); err != nil {
			slog.Warn("gazer failed to notify", "userID", user.ID, "error", err)
		}
	}
}
