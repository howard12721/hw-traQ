package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

type scheduledMessagesResponse struct {
	Messages []scheduledMessage `json:"messages"`
}

type scheduledMessageResponse struct {
	Message scheduledMessage `json:"message"`
}

type createScheduledMessageRequest struct {
	ChannelID   string `json:"channelId"`
	Content     string `json:"content"`
	ScheduledAt string `json:"scheduledAt"`
}

func getScheduledMessages(service *scheduledMessageService) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := currentTraQUser(c)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		messages, err := service.list(c.Request().Context(), user.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get scheduled messages")
		}
		return c.JSON(http.StatusOK, scheduledMessagesResponse{Messages: messages})
	}
}

func postScheduledMessage(service *scheduledMessageService) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := currentTraQUser(c)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		var req createScheduledMessageRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		scheduledAt, err := time.Parse(time.RFC3339Nano, req.ScheduledAt)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid scheduled time")
		}

		message, err := service.create(
			c.Request().Context(),
			user,
			credentialFromRequest(c.Request()),
			req.ChannelID,
			req.Content,
			scheduledAt,
		)
		if err != nil {
			if errors.Is(err, errInvalidScheduledMessage) {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid scheduled message")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create scheduled message")
		}
		return c.JSON(http.StatusCreated, scheduledMessageResponse{Message: message})
	}
}

func deleteScheduledMessage(service *scheduledMessageService) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := currentTraQUser(c)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		deleted, err := service.cancel(c.Request().Context(), user.ID, c.Param("id"))
		if err != nil {
			if errors.Is(err, errInvalidScheduledMessage) {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid scheduled message")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to cancel scheduled message")
		}
		if !deleted {
			return echo.NewHTTPError(http.StatusNotFound, "scheduled message not found")
		}
		return c.NoContent(http.StatusNoContent)
	}
}
