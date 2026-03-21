package testutil

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	appjwt "github.com/surveyflow/be/internal/pkg/jwt"
	"github.com/stretchr/testify/require"
)

// DoRequest makes an HTTP request against the test server.
// If token is non-empty, it adds an Authorization: Bearer header.
func DoRequest(t *testing.T, e *echo.Echo, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&reqBody).Encode(body))
	}

	req := httptest.NewRequest(method, path, &reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// ParseJSON parses the response body as a generic map.
func ParseJSON(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var result map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	return result
}

// ParseData extracts the "data" field from a response.
func ParseData(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	body := ParseJSON(t, rec)
	data, ok := body["data"]
	if ok && data != nil {
		return data.(map[string]any)
	}
	return nil
}

// ParseDataList extracts the "data" field as a slice.
func ParseDataList(t *testing.T, rec *httptest.ResponseRecorder) []any {
	t.Helper()
	body := ParseJSON(t, rec)
	data, ok := body["data"]
	if ok && data != nil {
		return data.([]any)
	}
	return nil
}

// RegisterAndLogin creates a test user via the API and returns an auth token.
func RegisterAndLogin(t *testing.T, e *echo.Echo) (token, userID, orgID string) {
	t.Helper()

	email := randomEmail()
	rec := DoRequest(t, e, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": "TestPass123",
		"name":     "Test User",
	}, "")

	require.Equal(t, http.StatusCreated, rec.Code)

	data := ParseData(t, rec)
	user := data["user"].(map[string]any)
	tokens := data["tokens"].(map[string]any)

	return tokens["access_token"].(string), user["id"].(string), user["org_id"].(string)
}

// RegisterAndLogin2 creates a second test user (different org).
func RegisterAndLogin2(t *testing.T, e *echo.Echo) (token, userID, orgID string) {
	t.Helper()

	email := randomEmail()
	rec := DoRequest(t, e, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    email,
		"password": "TestPass123",
		"name":     "Second User",
	}, "")

	require.Equal(t, http.StatusCreated, rec.Code)

	data := ParseData(t, rec)
	user := data["user"].(map[string]any)
	tokens := data["tokens"].(map[string]any)

	return tokens["access_token"].(string), user["id"].(string), user["org_id"].(string)
}

// GenerateToken creates a JWT directly for testing without going through auth.
func GenerateToken(t *testing.T, userID, orgID, role string) string {
	t.Helper()

	token, err := appjwt.GenerateToken(userID, orgID, role, TestConfig.JWT.Secret, 15*time.Minute)
	require.NoError(t, err)
	return token
}

// randomEmail generates a unique email for test isolation.
func randomEmail() string {
	b := make([]byte, 2)
	rand.Read(b)
	return "test-" + time.Now().Format("150405") + "-" + hex.EncodeToString(b) + "@test.surveyflow.io"
}
