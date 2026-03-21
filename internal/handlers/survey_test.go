package handlers_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/surveyflow/be/internal/testutil"
)

func TestSurvey_Create_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurvey(t, e, token, "My Survey")

	require.NotNil(t, survey)
	assert.Equal(t, "My Survey", survey["title"])
	assert.Equal(t, "draft", survey["status"])
	assert.NotEmpty(t, survey["id"])
}

func TestSurvey_Create_MissingTitle(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys", map[string]string{}, token)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSurvey_List_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	testutil.CreateSurvey(t, e, token, "Survey 1")
	testutil.CreateSurvey(t, e, token, "Survey 2")

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	list := testutil.ParseDataList(t, rec)
	assert.Len(t, list, 2)
}

func TestSurvey_List_WithFilter(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	s1 := testutil.CreateSurveyWithStructure(t, e, token, "Published Survey")
	publishID := s1["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, publishID)

	testutil.CreateSurvey(t, e, token, "Draft Survey")

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys?status=published", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	list := testutil.ParseDataList(t, rec)
	assert.Len(t, list, 1)
	assert.Equal(t, "published", list[0].(map[string]any)["status"])
}

func TestSurvey_List_WithSearch(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	testutil.CreateSurvey(t, e, token, "Customer Feedback")
	testutil.CreateSurvey(t, e, token, "Employee Survey")

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys?search=Feedback", nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	list := testutil.ParseDataList(t, rec)
	assert.Len(t, list, 1)
}

func TestSurvey_Get_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurvey(t, e, token, "Get Test")
	surveyID := survey["id"].(string)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID, nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.Equal(t, "Get Test", data["title"])
}

func TestSurvey_Get_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/00000000-0000-0000-0000-000000000000", nil, token)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestSurvey_Get_WrongOrg(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token1, _, _ := testutil.RegisterAndLogin(t, e)
	token2, _, _ := testutil.RegisterAndLogin2(t, e)

	survey := testutil.CreateSurvey(t, e, token1, "Org1 Survey")
	surveyID := survey["id"].(string)

	// user2 tries to access user1's survey — should be forbidden.
	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID, nil, token2)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestSurvey_Update_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurvey(t, e, token, "Original")
	surveyID := survey["id"].(string)

	rec := testutil.DoRequest(t, e, http.MethodPut, "/api/v1/surveys/"+surveyID, map[string]string{
		"title": "Updated Title",
	}, token)
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.Equal(t, "Updated Title", data["title"])
}

func TestSurvey_Delete_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurvey(t, e, token, "To Delete")
	surveyID := survey["id"].(string)

	rec := testutil.DoRequest(t, e, http.MethodDelete, "/api/v1/surveys/"+surveyID, nil, token)
	require.Equal(t, http.StatusOK, rec.Code)

	// Verify it's gone.
	rec2 := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/surveys/"+surveyID, nil, token)
	require.Equal(t, http.StatusNotFound, rec2.Code)
}

func TestSurvey_Publish_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Publishable")
	surveyID := survey["id"].(string)

	result := testutil.PublishSurvey(t, e, token, surveyID)
	assert.Equal(t, "published", result["status"])
}

func TestSurvey_Publish_NoQuestions(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurvey(t, e, token, "No Questions")
	surveyID := survey["id"].(string)

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/publish", nil, token)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSurvey_Publish_AlreadyPublished(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Double Publish")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/publish", nil, token)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSurvey_Close_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Closable")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	result := testutil.CloseSurveyHelper(t, e, token, surveyID)
	assert.Equal(t, "closed", result["status"])
}

func TestSurvey_Duplicate_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurvey(t, e, token, "Original")
	surveyID := survey["id"].(string)

	rec := testutil.DoRequest(t, e, http.MethodPost, "/api/v1/surveys/"+surveyID+"/duplicate", nil, token)
	require.Equal(t, http.StatusCreated, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.Contains(t, data["title"].(string), "Copy of")
}

func TestSurvey_GetPublic_Success(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurveyWithStructure(t, e, token, "Public Survey")
	surveyID := survey["id"].(string)
	_ = testutil.PublishSurvey(t, e, token, surveyID)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/public/surveys/"+surveyID, nil, "")
	require.Equal(t, http.StatusOK, rec.Code)

	data := testutil.ParseData(t, rec)
	assert.Equal(t, "Public Survey", data["title"])
	assert.NotEmpty(t, data["share_url"])
}

func TestSurvey_GetPublic_NotPublished(t *testing.T) {
	testutil.TruncateTables(t)
	e := newEcho(t)
	token, _, _ := testutil.RegisterAndLogin(t, e)

	survey := testutil.CreateSurvey(t, e, token, "Draft Survey")
	surveyID := survey["id"].(string)

	rec := testutil.DoRequest(t, e, http.MethodGet, "/api/v1/public/surveys/"+surveyID, nil, "")
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
