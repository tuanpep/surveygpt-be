package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/surveyflow/be/internal/middleware"
	apperr "github.com/surveyflow/be/internal/pkg/errors"
	"github.com/surveyflow/be/internal/repository"
)

// SSEHandler handles Server-Sent Events for live updates.
type SSEHandler struct {
	redis     *redis.Client
	surveyRepo repository.SurveyRepo
}

// NewSSEHandler creates a new SSEHandler.
func NewSSEHandler(redis *redis.Client, surveyRepo repository.SurveyRepo) *SSEHandler {
	return &SSEHandler{redis: redis, surveyRepo: surveyRepo}
}

// LiveUpdates handles GET /surveys/:id/live - SSE endpoint for real-time survey updates.
func (h *SSEHandler) LiveUpdates(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")

	// Verify the authenticated user's org owns this survey.
	survey, err := h.surveyRepo.GetByID(c.Request().Context(), surveyID)
	if err != nil {
		return apperr.Internal(err)
	}
	if survey == nil || survey.OrgID.String() != orgID {
		return apperr.NotFound("survey")
	}

	// Set SSE headers.
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")

	// Get response writer and check for Flusher interface.
	rw := c.Response().Writer
	flusher, ok := rw.(http.Flusher)
	if !ok {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"code":    "INTERNAL_ERROR",
			"message": "streaming not supported",
		})
	}

	// Create context for this connection.
	ctx := c.Request().Context()

	// Subscribe to Redis pub/sub for this survey's events.
	channel := fmt.Sprintf("survey:%s:events", surveyID)
	pubsub := h.redis.Subscribe(ctx, channel)
	defer pubsub.Close()

	// Send initial connection message.
	h.sendEvent(rw, "connected", map[string]any{
		"survey_id": surveyID,
		"timestamp": time.Now().Unix(),
	})
	flusher.Flush()

	// Also subscribe to org-wide channel for billing updates.
	orgChannel := fmt.Sprintf("org:%s:events", orgID)
	pubsub.Subscribe(ctx, orgChannel)

	// Handle incoming messages.
	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			// Client disconnected.
			return nil

		case msg, ok := <-ch:
			if !ok {
				// Channel closed.
				return nil
			}

			// Parse the message and forward to client.
			// The message payload should contain event type and data.
			h.sendEvent(rw, "update", map[string]any{
				"channel": msg.Channel,
				"payload": msg.Payload,
			})
			flusher.Flush()
		}
	}
}

// sendEvent writes an SSE event to the response writer.
func (h *SSEHandler) sendEvent(rw http.ResponseWriter, event string, data any) {
	fmt.Fprintf(rw, "event: %s\n", event)
	fmt.Fprintf(rw, "data: %s\n\n", formatSSEData(data))
}

// formatSSEData formats data as JSON for SSE.
func formatSSEData(data any) string {
	b, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// PublishEvent publishes an event to Redis (helper for other handlers).
func (h *SSEHandler) PublishEvent(ctx context.Context, surveyID, eventType string, data map[string]any) error {
	channel := fmt.Sprintf("survey:%s:events", surveyID)
	payload := fmt.Sprintf(`{"type":"%s","data":%v}`, eventType, data)
	return h.redis.Publish(ctx, channel, payload).Err()
}

// PublishOrgEvent publishes an org-wide event.
func (h *SSEHandler) PublishOrgEvent(ctx context.Context, orgID, eventType string, data map[string]any) error {
	channel := fmt.Sprintf("org:%s:events", orgID)
	payload := fmt.Sprintf(`{"type":"%s","data":%v}`, eventType, data)
	return h.redis.Publish(ctx, channel, payload).Err()
}
