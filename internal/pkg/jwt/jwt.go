package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents the custom JWT claims embedded in access and refresh tokens.
type Claims struct {
	UserID string `json:"uid"`
	OrgID  string `json:"oid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair holds an access token and its associated refresh token.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// GenerateToken creates a signed JWT with the given user claims.
// The token expires after the provided TTL duration.
func GenerateToken(userID, orgID, role, secret string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		OrgID:  orgID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Issuer:    "surveyflow",
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

// GenerateTokenPair creates both an access token and a refresh token.
func GenerateTokenPair(userID, orgID, role, secret string, accessTTL, refreshTTL time.Duration) (*TokenPair, error) {
	accessToken, err := GenerateToken(userID, orgID, role, secret, accessTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := GenerateToken(userID, orgID, role, secret, refreshTTL)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(accessTTL),
	}, nil
}

// ParseToken validates and extracts claims from a JWT string.
// It verifies the signature using the provided secret.
func ParseToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
