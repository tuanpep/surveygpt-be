package handlers_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/surveyflow/be/internal/testutil"
)

func TestBilling_GetPlan_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/billing/plan", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.Equal(t, "Free", data["plan_name"])
	assert.NotNil(t, data["limits"])
}

func TestBilling_GetPlans_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/billing/plans", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseJSON(t, rec)
	plans := data["data"].([]any)
	assert.NotEmpty(t, plans)
}

func TestBilling_GetHistory_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/billing/history", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseJSON(t, rec)
	assert.NotNil(t, data["data"])
}

func TestBilling_GetPortalURL_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/billing/portal", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.NotEmpty(t, data["url"])
}

func TestBilling_ChangePlan_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/billing/change-plan", map[string]string{
		"plan_id": "starter",
	}, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.NotEmpty(t, data["checkout_url"])
}
