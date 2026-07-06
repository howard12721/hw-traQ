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

type saveGazerRequest struct {
	Entries []gazerEntry `json:"entries"`
}

type saveGazerTokenRequest struct {
	AccessToken string `json:"accessToken"`
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
		service.refreshWithBestCredential(user, credentialFromRequest(c.Request()), setting)

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

		return c.JSON(http.StatusOK, newGazerResponse(service, user.ID, setting))
	}
}

func putGazerToken(service *gazerService) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := currentTraQUser(c)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		var req saveGazerTokenRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}
		if req.AccessToken == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "access token is required")
		}

		setting, err := service.saveAccessToken(c.Request().Context(), user, req.AccessToken)
		if err != nil {
			if errors.Is(err, errUnauthenticated) {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid access token")
			}
			return echo.NewHTTPError(http.StatusBadRequest, "failed to save gazer access token")
		}

		return c.JSON(http.StatusOK, newGazerResponse(service, user.ID, setting))
	}
}

func newGazerResponse(service *gazerService, userID string, setting gazerSetting) gazerResponse {
	status := service.status(userID)
	status.TokenConfigured = setting.AccessToken != ""
	return gazerResponse{
		Setting: setting,
		Status:  status,
	}
}
