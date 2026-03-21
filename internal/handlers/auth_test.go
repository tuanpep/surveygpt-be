package handlers_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appjwt "github.com/surveyflow/be/internal/pkg/jwt"
	"github.com/surveyflow/be/internal/testutil"
)

func TestAuth_Register_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    "newuser@test.com",
		"password": "TestPass123",
		"name":     "New User",
	}, "")

	require.Equal(t, http.StatusCreated, rec.Code)

	body := testutil.ParseJSON(t, rec)
	data := body["data"].(map[string]any)

	user := data["user"].(map[string]any)
	assert.NotEmpty(t, user["id"])
	assert.Equal(t, "newuser@test.com", user["email"])
	assert.Equal(t, "New User", user["name"])
	assert.NotEmpty(t, user["org_id"])

	tokens := data["tokens"].(map[string]any)
	assert.NotEmpty(t, tokens["access_token"])
	assert.NotEmpty(t, tokens["refresh_token"])
}

func TestAuth_Register_DuplicateEmail(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	body := map[string]string{"email": "dup@test.com", "password": "TestPass123", "name": "User"}
	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/register", body, "")
	require.Equal(t, http.StatusCreated, rec.Code)

	rec2 := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/register", body, "")
	require.Equal(t, http.StatusConflict, rec2.Code)

	resp := testutil.ParseJSON(t, rec2)
	assert.Equal(t, "CONFLICT", resp["code"])
}

func TestAuth_Register_WeakPassword(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":    "weak@test.com",
		"password": "weak",
		"name":     "User",
	}, "")

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAuth_Register_MissingFields(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	cases := []map[string]string{
		{"password": "TestPass123", "name": "User"},                  // missing email
		{"email": "test@test.com", "name": "User"},                   // missing password
		{"email": "test@test.com", "password": "TestPass123"},        // missing name
	}

	for _, tc := range cases {
		rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/register", tc, "")
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "expected 422 for missing fields: %v", tc)
	}
}

func TestAuth_Login_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	// Register first.
	testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email": "login@test.com", "password": "TestPass123", "name": "User",
	}, "")

	// Login.
	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": "login@test.com", "password": "TestPass123",
	}, "")

	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	user := data["user"].(map[string]any)
	assert.Equal(t, "login@test.com", user["email"])

	tokens := data["tokens"].(map[string]any)
	assert.NotEmpty(t, tokens["access_token"])
}

func TestAuth_Login_InvalidPassword(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email": "badpw@test.com", "password": "TestPass123", "name": "User",
	}, "")

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": "badpw@test.com", "password": "WrongPassword1",
	}, "")

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_Login_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": "nonexistent@test.com", "password": "TestPass123",
	}, "")

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_RefreshToken_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	// Register with a known email.
	testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email": "refresh@test.com", "password": "TestPass123", "name": "User",
	}, "")

	// Login to get a refresh token.
	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": "refresh@test.com", "password": "TestPass123",
	}, "")
	require.Equal(t, http.StatusOK, rec.Code)
	data := testutil.ParseData(t, rec)
	tokens := data["tokens"].(map[string]any)
	refreshToken := tokens["refresh_token"].(string)

	// Use refresh token.
	rec2 := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": refreshToken,
	}, "")

	require.Equal(t, http.StatusOK, rec2.Code)

	data2 := testutil.ParseData(t, rec2)
	assert.NotEmpty(t, data2["access_token"])
	assert.NotEmpty(t, data2["refresh_token"])
}

func TestAuth_RefreshToken_Invalid(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": "invalid-token",
	}, "")

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_ForgotPassword_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	// Register a user.
	testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email": "forgot@test.com", "password": "TestPass123", "name": "User",
	}, "")

	// Forgot password — always returns 200 even if email doesn't exist.
	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
		"email": "forgot@test.com",
	}, "")

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuth_ResetPassword_InvalidToken(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/auth/reset-password", map[string]string{
		"token":        "invalid-token",
		"new_password": "NewPassword123",
	}, "")

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuth_MissingAuth(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/me", nil, "")

	require.Equal(t, http.StatusUnauthorized, rec.Code)

	body := testutil.ParseJSON(t, rec)
	assert.Equal(t, "UNAUTHORIZED", body["code"])
}

func TestAuth_InvalidToken(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/me", nil, "invalid-token")

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_GetMe(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	token, _, _ := testutil.RegisterAndLogin(t, e)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/me", nil, token)

	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.NotEmpty(t, data["id"])
	assert.NotEmpty(t, data["email"])
	assert.NotEmpty(t, data["org_id"])
}

func TestAuth_GenerateTokenDirectly(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	token := testutil.GenerateToken(t, "00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002", "owner")

	// Verify the token can be parsed.
	claims, err := appjwt.ParseToken(token, testutil.TestConfig.JWT.Secret)
	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", claims.UserID)
	assert.Equal(t, "00000000-0000-0000-0000-000000000002", claims.OrgID)
	assert.Equal(t, "owner", claims.Role)

	// Use the token to hit an authenticated endpoint.
	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/me", nil, token)
	// Will return 404 since the user doesn't exist in DB.
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
