package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

const (
	defaultPort     = "8080"
	shutdownTimeout = 10 * time.Second
)

type healthResponse struct {
	Status string `json:"status"`
}

type pingResponse struct {
	Message string `json:"message"`
}

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config := echo.StartConfig{
		Address:         listenAddr(),
		HideBanner:      true,
		HidePort:        true,
		GracefulTimeout: shutdownTimeout,
	}
	if err := config.Start(ctx, newServer()); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func newServer() *echo.Echo {
	return newServerWithConfig(mustServerConfigFromEnv())
}

type serverConfig struct {
	authenticator *traQAuthenticator
	gazer         *gazerService
	gazerClientID string
}

func mustServerConfigFromEnv() serverConfig {
	db, err := openDBFromEnv()
	if err != nil {
		panic(err)
	}

	store := newGazerStore(db)
	if err := store.migrate(context.Background()); err != nil {
		panic(err)
	}

	traqClient := mustTraQClientFromEnv()
	gazer := newGazerService(store, traqClient)
	if err := gazer.restore(context.Background()); err != nil {
		panic(err)
	}
	return serverConfig{
		authenticator: mustTraQAuthenticatorFromEnv(),
		gazer:         gazer,
		gazerClientID: os.Getenv("GAZER_OAUTH_CLIENT_ID"),
	}
}

func newServerWithConfig(config serverConfig) *echo.Echo {
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	internal := e.Group("/internal")
	internal.GET("/health", getHealth)

	v1 := internal.Group("/v1")
	v1.GET("/ping", getPing)
	v1.GET("/me", getMe, requireTraQUser(config.authenticator))
	if config.gazer != nil {
		v1.GET("/gazer", getGazer(config.gazer), requireTraQUser(config.authenticator))
		v1.PUT("/gazer", putGazer(config.gazer), requireTraQUser(config.authenticator))
		v1.PUT("/gazer/token", putGazerToken(config.gazer), requireTraQUser(config.authenticator))
		v1.GET("/gazer/oauth-client", getGazerOAuthClient(config.gazerClientID), requireTraQUser(config.authenticator))
	}

	return e
}

func listenAddr() string {
	if addr := os.Getenv("ADDR"); addr != "" {
		return addr
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

func getHealth(c *echo.Context) error {
	return c.JSON(http.StatusOK, healthResponse{Status: "ok"})
}

func getPing(c *echo.Context) error {
	return c.JSON(http.StatusOK, pingResponse{Message: "pong"})
}

func getMe(c *echo.Context) error {
	user, ok := currentTraQUser(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	return c.JSON(http.StatusOK, meResponse{User: user})
}
