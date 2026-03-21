package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/surveyflow/be/internal/handlers"
	"github.com/surveyflow/be/internal/testutil"
)

func newEcho(t *testing.T) *echo.Echo {
	t.Helper()
	deps := &handlers.Dependencies{
		Config: testutil.TestConfig,
		Pool:   testutil.TestPool,
		Redis:  testutil.TestRedis,
	}
	return handlers.SetupRoutes(deps)
}

func TestHealthCheck(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	body := testutil.ParseJSON(t, rec)
	assert.Equal(t, "ok", body["status"])

	dbDeps := body["deps"].(map[string]any)
	assert.Equal(t, "ok", dbDeps["database"])
	assert.Equal(t, "ok", dbDeps["redis"])
}

func TestHealthz(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	body := testutil.ParseJSON(t, rec)
	assert.Equal(t, "ok", body["status"])
}

func TestReadyCheck(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	body := testutil.ParseJSON(t, rec)
	assert.Equal(t, "ready", body["status"])
}
