package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/surveyflow/be/internal/pkg/errors"
	appjwt "github.com/surveyflow/be/internal/pkg/jwt"
)

const (
	// ContextKeyClaims is the key used to store JWT claims in the Echo context.
	ContextKeyClaims = "claims"

	// ContextKeyUserID is the key used to store the authenticated user ID.
	ContextKeyUserID = "user_id"

	// ContextKeyOrgID is the key used to store the authenticated user's organization ID.
	ContextKeyOrgID = "org_id"

	// ContextKeyRole is the key used to store the authenticated user's role.
	ContextKeyRole = "role"
)

// Auth creates an Echo middleware that validates JWT tokens from the
// Authorization header and sets user claims in the request context.
func Auth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"code":    "UNAUTHORIZED",
					"message": "missing authorization header",
				})
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"code":    "UNAUTHORIZED",
					"message": "invalid authorization header format",
				})
			}

			tokenString := parts[1]
			claims, err := appjwt.ParseToken(tokenString, secret)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"code":    "UNAUTHORIZED",
					"message": "invalid or expired token",
				})
			}

			c.Set(ContextKeyClaims, claims)
			c.Set(ContextKeyUserID, claims.UserID)
			c.Set(ContextKeyOrgID, claims.OrgID)
			c.Set(ContextKeyRole, claims.Role)

			return next(c)
		}
	}
}

// OptionalAuth is like Auth but does not return an error if the token is missing.
// If a valid token is present, claims are set in context. Otherwise the request
// proceeds as unauthenticated.
func OptionalAuth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return next(c)
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				claims, err := appjwt.ParseToken(parts[1], secret)
				if err == nil {
					c.Set(ContextKeyClaims, claims)
					c.Set(ContextKeyUserID, claims.UserID)
					c.Set(ContextKeyOrgID, claims.OrgID)
					c.Set(ContextKeyRole, claims.Role)
				}
			}

			return next(c)
		}
	}
}

// GetClaimsFromContext extracts JWT claims from the Echo context.
// Returns nil if claims are not present (i.e., the request is unauthenticated).
func GetClaimsFromContext(c echo.Context) *appjwt.Claims {
	val := c.Get(ContextKeyClaims)
	if val == nil {
		return nil
	}
	claims, ok := val.(*appjwt.Claims)
	if !ok {
		return nil
	}
	return claims
}

// GetUserIDFromContext extracts the authenticated user ID from the Echo context.
// Returns an empty string if the user is not authenticated.
func GetUserIDFromContext(c echo.Context) string {
	val := c.Get(ContextKeyUserID)
	if val == nil {
		return ""
	}
	s, ok := val.(string)
	if !ok {
		return ""
	}
	return s
}

// GetOrgIDFromContext extracts the authenticated user's organization ID.
func GetOrgIDFromContext(c echo.Context) string {
	val := c.Get(ContextKeyOrgID)
	if val == nil {
		return ""
	}
	s, ok := val.(string)
	if !ok {
		return ""
	}
	return s
}

// RequireRole returns a middleware that checks whether the authenticated user
// has one of the allowed roles. Must be used after Auth middleware.
func RequireRole(allowedRoles ...string) echo.MiddlewareFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		allowed[r] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := GetClaimsFromContext(c)
			if claims == nil {
				return errors.Unauthorized()
			}

			if _, ok := allowed[claims.Role]; !ok {
				return c.JSON(http.StatusForbidden, echo.Map{
					"code":    "FORBIDDEN",
					"message": "insufficient permissions",
				})
			}

			return next(c)
		}
	}
}
