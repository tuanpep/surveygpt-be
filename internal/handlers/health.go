package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	pool   *pgxpool.Pool
	redis  *redis.Client
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(pool *pgxpool.Pool, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		pool:  pool,
		redis: redisClient,
	}
}

// Check responds with the health status of the service and its dependencies.
// It verifies connectivity to PostgreSQL and Redis.
func (h *HealthHandler) Check(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
	defer cancel()

	status := "ok"
	deps := make(map[string]string)

	// Check PostgreSQL.
	if err := h.pool.Ping(ctx); err != nil {
		status = "degraded"
		deps["database"] = "unreachable"
	} else {
		deps["database"] = "ok"
	}

	// Check Redis.
	if err := h.redis.Ping(ctx).Err(); err != nil {
		status = "degraded"
		deps["redis"] = "unreachable"
	} else {
		deps["redis"] = "ok"
	}

	code := http.StatusOK
	if status != "ok" {
		code = http.StatusServiceUnavailable
	}

	return c.JSON(code, echo.Map{
		"status": status,
		"deps":   deps,
	})
}

// Ready responds with 200 if the service is ready to accept traffic.
func (h *HealthHandler) Ready(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
	defer cancel()

	if err := h.pool.Ping(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, echo.Map{
			"status": "not ready",
			"error":  "database unreachable",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status": "ready",
	})
}

const timeout = 5 * time.Second
