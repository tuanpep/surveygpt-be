package handlers_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/surveyflow/be/internal/testutil"
)

func TestDistribution_GetQRCode_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "QR Survey")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID+"/qr-code", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.NotEmpty(t, data["svg"])
}

func TestDistribution_GetEmbedCode_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Embed Survey")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID+"/embed-code", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.NotEmpty(t, data["html"])
	assert.NotEmpty(t, data["js"])
}

func TestDistribution_CreateEmailList_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/email-lists", map[string]any{
		"name":     "Test List",
		"contacts": []map[string]any{},
	}, token)
	require.Equal(t, http.StatusCreated, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.Equal(t, "Test List", data["name"])
}

func TestDistribution_ListEmailLists_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	// Create a list first.
	testutil.DoRequest(t, e, http.MethodPost, "/api/v1/email-lists", map[string]any{
		"name": "List for Listing",
	}, token)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/email-lists", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseJSON(t, rec)
	assert.NotNil(t, data["data"])
}
