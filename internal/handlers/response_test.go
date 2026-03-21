package handlers_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/surveyflow/be/internal/testutil"
)

func TestResponse_Submit_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Response Survey")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	body := map[string]any{
		"answers": []map[string]any{
			{"question_id": "q1", "value": "5"},
		},
	}

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/responses", body, "")
	require.Equal(t, http.StatusCreated, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.Equal(t, true, data["success"])
	assert.NotEmpty(t, data["response_id"])
}

func TestResponse_Submit_InvalidSurvey(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)

	body := map[string]any{
		"answers": []map[string]any{
			{"question_id": "q1", "value": "5"},
		},
	}

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/00000000-0000-0000-0000-000000000000/responses", body, "")
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestResponse_Submit_NotPublished(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Draft Survey")
	surveyID := survey["id"].(string)

	body := map[string]any{
		"answers": []map[string]any{
			{"question_id": "q1", "value": "5"},
		},
	}

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/responses", body, "")
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestResponse_Submit_MissingRequired(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Required Survey")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	// Submit with no answers — q1 is required.
	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/responses", map[string]any{
		"answers": []map[string]any{},
	}, "")
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestResponse_List_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "List Survey")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	// Submit a response.
	testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/responses", map[string]any{
		"answers": []map[string]any{{"question_id": "q1", "value": "4"}},
	}, "")

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID+"/responses", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseJSON(t, rec)
	assert.NotNil(t, data["data"])
}

func TestResponse_List_Empty(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Empty Survey")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID+"/responses", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseJSON(t, rec)
	list := data["data"].([]any)
	assert.Empty(t, list)
}

func TestResponse_Get_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Get Response Survey")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	// Submit a response.
	submitRec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/responses", map[string]any{
		"answers": []map[string]any{{"question_id": "q1", "value": "3"}},
	}, "")
	require.Equal(t, http.StatusCreated, submitRec.Code)
	responseID := testutil.ParseData(t, submitRec)["response_id"].(string)

	// Get the response.
	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID+"/responses/"+responseID, nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.Equal(t, surveyID, data["survey_id"])
}

func TestResponse_Get_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "NotFound Survey")
	surveyID := survey["id"].(string)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID+"/responses/00000000-0000-0000-0000-000000000000", nil, token)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestResponse_Delete_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Delete Response Survey")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	// Submit a response.
	submitRec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/responses", map[string]any{
		"answers": []map[string]any{{"question_id": "q1", "value": "2"}},
	}, "")
	responseID := testutil.ParseData(t, submitRec)["response_id"].(string)

	// Delete it.
	rec := testutil.DoRequest(t, e, http.MethodDelete, "/api/v1/surveys/"+surveyID+"/responses/"+responseID, nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	// Verify it's gone.
	rec2 := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID+"/responses/"+responseID, nil, token)
	require.Equal(t, http.StatusNotFound, rec2.Code)
}

func TestResponse_Export_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Export Survey")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	// Submit a response.
	testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/responses", map[string]any{
		"answers": []map[string]any{{"question_id": "q1", "value": "5"}},
	}, "")

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/responses/export", map[string]any{
		"format": "json",
	}, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.NotNil(t, data)
}
