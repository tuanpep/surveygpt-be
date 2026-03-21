package handlers_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/surveyflow/be/internal/testutil"
)

func TestTemplate_List_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/templates", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseJSON(t, rec)
	templates := data["data"]
	assert.NotNil(t, templates)
}

func TestTemplate_List_ByCategory(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/templates?category=feedback", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestTemplate_SaveAsTemplate_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Template Source")
	surveyID := survey["id"].(string)

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/templates", map[string]string{
		"survey_id": surveyID,
		"name":      "My Custom Template",
		"category":  "custom",
	}, token)
	require.Equal(t, http.StatusCreated, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.Equal(t, "My Custom Template", data["title"])
	assert.NotEmpty(t, data["id"])
}

func TestTemplate_Get_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	// Create a template first.
	survey := testutil.CreateSurveyWithStructure(t, e, token, "Get Template Source")
	surveyID := survey["id"].(string)
	createRec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/templates", map[string]string{
		"survey_id": surveyID,
		"name":      "Gettable Template",
	}, token)
	require.Equal(t, http.StatusCreated, createRec.Code)
	templateID := testutil.ParseData(t, createRec)["id"].(string)

	// Get the template.
	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/templates/"+templateID, nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.Equal(t, "Gettable Template", data["title"])
}

func TestTemplate_Get_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/templates/00000000-0000-0000-0000-000000000000", nil, token)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestTemplate_CreateFromTemplate_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	// Create a template first.
	survey := testutil.CreateSurveyWithStructure(t, e, token, "From Template Source")
	surveyID := survey["id"].(string)
	createRec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/templates", map[string]string{
		"survey_id": surveyID,
		"name":      "Source Template",
	}, token)
	require.Equal(t, http.StatusCreated, createRec.Code)
	templateID := testutil.ParseData(t, createRec)["id"].(string)

	// Create a survey from the template.
	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/from-template/"+templateID, nil, token)
	require.Equal(t, http.StatusCreated, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.NotEmpty(t, data["id"])
	assert.Contains(t, data["title"], "Source Template")
}
