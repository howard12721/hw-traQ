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
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	internal := e.Group("/internal")
	internal.GET("/health", getHealth)

	v1 := internal.Group("/v1")
	v1.GET("/ping", getPing)

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
