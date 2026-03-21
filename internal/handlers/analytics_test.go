package handlers_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/surveyflow/be/internal/testutil"
)

func TestAnalytics_GetSummary_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Analytics Survey")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	// Submit a response.
	testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/responses", map[string]any{
		"answers": []map[string]any{{"question_id": "q1", "value": "5"}},
	}, "")

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID+"/analytics/summary", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.NotNil(t, data["total_responses"])
}

func TestAnalytics_GetSummary_Empty(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Empty Analytics")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID+"/analytics/summary", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.Equal(t, float64(0), data["total_responses"])
}

func TestAnalytics_GetQuestionStats_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Question Stats")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	// Submit responses.
	testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/responses", map[string]any{
		"answers": []map[string]any{{"question_id": "q1", "value": "4"}},
	}, "")
	testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/responses", map[string]any{
		"answers": []map[string]any{{"question_id": "q1", "value": "5"}},
	}, "")

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID+"/analytics/questions/q1", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.Equal(t, "q1", data["question_id"])
}

func TestAnalytics_GetDropoff_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Dropoff Survey")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID+"/analytics/dropoff", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.NotNil(t, data["steps"])
}

func TestAnalytics_GetCrossTab_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "CrossTab Survey")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID+"/analytics/cross-tab?row_question_id=q1&col_question_id=q1", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAnalytics_GetCrossTab_MissingParams(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "CrossTab Missing")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID+"/analytics/cross-tab", nil, token)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
