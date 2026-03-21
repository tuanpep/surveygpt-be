package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CORS returns an Echo middleware that handles Cross-Origin Resource Sharing.
// It allows the configured frontend URL and common API request methods and headers.
func CORS(frontendURL string) echo.MiddlewareFunc {
	allowedOrigins := strings.Split(frontendURL, ",")
	// Trim whitespace from each origin.
	for i, origin := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(origin)
	}
	// Build a lookup set for fast origin validation.
	allowedSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowedSet[o] = struct{}{}
	}

	return middleware.CORSWithConfig(middleware.CORSConfig{
		// Use AllowOriginFunc instead of AllowOrigins so we can validate the
		// request Origin dynamically while still setting AllowCredentials: true.
		// When credentials are enabled, browsers reject wildcard/multiple origins.
		AllowOriginFunc: func(origin string) (bool, error) {
			_, ok := allowedSet[origin]
			return ok, nil
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderXRequestID,
			"X-Forwarded-For",
			"X-Real-IP",
		},
		ExposeHeaders: []string{
			echo.HeaderXRequestID,
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
			"X-Total-Count",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	})
}
