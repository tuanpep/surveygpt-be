package testutil

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

// CreateSurvey creates a draft survey via the API and returns its data.
func CreateSurvey(t *testing.T, e *echo.Echo, token, title string) map[string]any {
	t.Helper()
	return CreateSurveyWithDesc(t, e, token, title, "")
}

// CreateSurveyWithDesc creates a survey with a description.
func CreateSurveyWithDesc(t *testing.T, e *echo.Echo, token, title, desc string) map[string]any {
	t.Helper()

	body := map[string]string{"title": title}
	if desc != "" {
		body["description"] = desc
	}

	rec := DoRequest(t, e, http.MethodPost, "/api/v1/surveys", body, token)
	require.Equal(t, http.StatusCreated, rec.Code)
	return ParseData(t, rec)
}

// CreateSurveyWithStructure creates a survey with questions (needed for publish).
func CreateSurveyWithStructure(t *testing.T, e *echo.Echo, token, title string) map[string]any {
	t.Helper()

	survey := CreateSurvey(t, e, token, title)
	surveyID := survey["id"].(string)

	structure := map[string]any{
		"questions": []map[string]any{
			{
				"id":       "q1",
				"title":    "How satisfied are you?",
				"type":     "rating",
				"required": true,
			},
		},
	}

	rec := DoRequest(t, e, http.MethodPut, "/api/v1/surveys/"+surveyID, map[string]any{
		"structure": structure,
	}, token)
	require.Equal(t, http.StatusOK, rec.Code)
	return ParseData(t, rec)
}

// PublishSurvey publishes a survey and returns its updated data.
func PublishSurvey(t *testing.T, e *echo.Echo, token, surveyID string) map[string]any {
	t.Helper()

	rec := DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/publish", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)
	return ParseData(t, rec)
}

// CloseSurveyHelper closes a published survey via the API.
func CloseSurveyHelper(t *testing.T, e *echo.Echo, token, surveyID string) map[string]any {
	t.Helper()

	rec := DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/close", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)
	return ParseData(t, rec)
}
