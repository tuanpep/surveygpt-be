package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"
)

// Recovery returns an Echo middleware that recovers from panics inside handlers,
// logs the stack trace, and returns a 500 Internal Server Error response.
func Recovery() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					stack := debug.Stack()
					slog.Error("panic recovered",
						"request_id", GetRequestID(c),
						"error", r,
						"method", c.Request().Method,
						"path", c.Request().URL.Path,
						"stack", string(stack),
					)

					c.JSON(http.StatusInternalServerError, echo.Map{
						"code":    "INTERNAL_ERROR",
						"message": "an internal error occurred",
					})
				}
			}()

			return next(c)
		}
	}
}
