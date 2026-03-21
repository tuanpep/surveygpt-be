package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

// Logger returns an Echo middleware that logs each request using slog.
// It logs the request ID, method, path, status code, latency, and client IP.
func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			fields := []any{
				"request_id", GetRequestID(c),
				"method",      req.Method,
				"path",       req.URL.Path,
				"status",      res.Status,
				"latency_ms",  time.Since(start).Milliseconds(),
				"remote_ip",   c.RealIP(),
				"user_agent",  req.UserAgent(),
				"content_length", req.ContentLength,
			}

			if res.Status >= 500 {
				slog.Error("request completed", fields...)
			} else if res.Status >= 400 {
				slog.Warn("request completed", fields...)
			} else {
				slog.Info("request completed", fields...)
			}

			return nil
		}
	}
}
