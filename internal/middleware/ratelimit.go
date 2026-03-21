package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// RateLimit creates an Echo middleware that limits the number of requests
// from a single client within a given time window using Redis as the backing store.
//
// keyPrefix is used to namespace the Redis key (e.g., "ratelimit:api").
// rate is the maximum number of requests allowed within the window.
// window is the duration of the sliding window.
func RateLimit(redisClient *redis.Client, keyPrefix string, rate int, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			// Use client IP as the rate limit key identifier.
			clientIP := c.RealIP()
			key := fmt.Sprintf("%s:%s", keyPrefix, clientIP)

			// Use Redis INCR with an expiry for a fixed-window rate limiter.
			count, err := redisClient.Incr(ctx, key).Result()
			if err != nil {
				// On Redis failure, allow the request through to avoid blocking all traffic.
				return next(c)
			}

			// Set the expiry only on the first increment to establish the window.
			if count == 1 {
				if err := redisClient.Expire(ctx, key, window).Err(); err != nil {
					// Log but don't block.
					return next(c)
				}
			}

			// Set rate limit headers for the client.
			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rate))
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, rate-int(count))))
			c.Response().Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

			if int(count) > rate {
				return c.JSON(http.StatusTooManyRequests, echo.Map{
					"code":    "RATE_LIMITED",
					"message": "too many requests, please try again later",
				})
			}

			return next(c)
		}
	}
}
