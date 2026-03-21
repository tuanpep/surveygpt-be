package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// RequestID returns an Echo middleware that generates a unique request ID
// for each incoming request and stores it in the context and response header.
// If a request ID is already present in the X-Request-ID header, it is reused.
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Request().Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = uuid.New().String()
			}

			c.Set("request_id", requestID)
			c.Response().Header().Set(echo.HeaderXRequestID, requestID)

			return next(c)
		}
	}
}

// GetRequestID extracts the request ID from the Echo context.
func GetRequestID(c echo.Context) string {
	val := c.Get("request_id")
	if val == nil {
		return ""
	}
	s, ok := val.(string)
	if !ok {
		return ""
	}
	return s
}
