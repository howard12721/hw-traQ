package main

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
)

type gazerResponse struct {
	Setting gazerSetting `json:"setting"`
	Status  gazerStatus  `json:"status"`
}

type gazerOAuthClientResponse struct {
	ClientID string `json:"clientId"`
}

type gazerNotificationsResponse struct {
	Notifications []gazerNotification `json:"notifications"`
	BotUserID     string              `json:"botUserId,omitempty"`
}

type saveGazerRequest struct {
	Entries []gazerEntry `json:"entries"`
}

type saveGazerTokenRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"codeVerifier"`
	RedirectURI  string `json:"redirectUri"`
}

func getGazer(service *gazerService) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := currentTraQUser(c)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		setting, err := service.store.get(c.Request().Context(), user.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get gazer setting")
		}
		service.refreshWithAccessToken(user, setting)
		service.refreshBotUserID(c.Request().Context())

		return c.JSON(http.StatusOK, newGazerResponse(service, user.ID, setting))
	}
}

func getGazerOAuthClient(clientID string) echo.HandlerFunc {
	return func(c *echo.Context) error {
		if clientID == "" {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "gazer oauth client is not configured")
		}
		return c.JSON(http.StatusOK, gazerOAuthClientResponse{ClientID: clientID})
	}
}

func putGazer(service *gazerService) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := currentTraQUser(c)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		var req saveGazerRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		setting, err := service.save(c.Request().Context(), user, credentialFromRequest(c.Request()), gazerSetting{
			Entries: req.Entries,
		})
		if err != nil {
			if errors.Is(err, errInvalidGazerPattern) {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid gazer setting")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to save gazer setting")
		}
		service.refreshBotUserID(c.Request().Context())

		return c.JSON(http.StatusOK, newGazerResponse(service, user.ID, setting))
	}
}

func putGazerToken(service *gazerService, clientID string) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := currentTraQUser(c)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		var req saveGazerTokenRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		if clientID == "" {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "gazer oauth client is not configured")
		}
		if req.Code == "" || req.CodeVerifier == "" || req.RedirectURI == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "oauth code, code verifier and redirect uri are required")
		}

		setting, err := service.saveAuthorizationCode(c.Request().Context(), user, clientID, req.Code, req.CodeVerifier, req.RedirectURI)
		if err != nil {
			if errors.Is(err, errUnauthenticated) {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid oauth code")
			}
			return echo.NewHTTPError(http.StatusBadRequest, "failed to save gazer access token")
		}
		service.refreshBotUserID(c.Request().Context())

		return c.JSON(http.StatusOK, newGazerResponse(service, user.ID, setting))
	}
}

func getGazerNotifications(service *gazerService) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := currentTraQUser(c)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}
		service.refreshBotUserID(c.Request().Context())

		notifications, err := service.store.listNotifications(c.Request().Context(), user.ID, 100)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get gazer notifications")
		}
		status := service.status(user.ID)
		return c.JSON(http.StatusOK, gazerNotificationsResponse{
			Notifications: notifications,
			BotUserID:     status.BotUserID,
		})
	}
}

func postGazerNotificationsRead(service *gazerService) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := currentTraQUser(c)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		if err := service.store.markNotificationsRead(c.Request().Context(), user.ID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to mark gazer notifications as read")
		}
		return c.NoContent(http.StatusNoContent)
	}
}

func newGazerResponse(service *gazerService, userID string, setting gazerSetting) gazerResponse {
	if setting.Entries == nil {
		setting.Entries = []gazerEntry{}
	}
	status := service.status(userID)
	status.TokenConfigured = setting.AccessToken != ""
	return gazerResponse{
		Setting: setting,
		Status:  status,
	}
}
