package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/surveyflow/be/internal/models"
	apperr "github.com/surveyflow/be/internal/pkg/errors"
	"github.com/surveyflow/be/internal/repository"
)

// ResponseService handles response collection and management.
type ResponseService struct {
	responseRepo repository.ResponseRepo
	surveyRepo   repository.SurveyRepo
	pool         *pgxpool.Pool
}

// NewResponseService creates a new ResponseService.
func NewResponseService(responseRepo repository.ResponseRepo, surveyRepo repository.SurveyRepo, pool *pgxpool.Pool) *ResponseService {
	return &ResponseService{
		responseRepo: responseRepo,
		surveyRepo:   surveyRepo,
		pool:         pool,
	}
}

// --- Input types ---

type SubmitResponseInput struct {
	SurveyID string                   `json:"survey_id" validate:"required"`
	Answers  []AnswerInput            `json:"answers" validate:"required"`
	Metadata ResponseMetadataInput    `json:"metadata"`
}

type AnswerInput struct {
	QuestionID string `json:"question_id"`
	Value      any    `json:"value"`
}

type ResponseMetadataInput struct {
	Duration    int               `json:"duration"`
	Source      string            `json:"source"`
	Device      string            `json:"device"`
	Browser     string            `json:"browser"`
	OS          string            `json:"os"`
	Country     string            `json:"country"`
	IPHash      string            `json:"ip_hash"`
	Language    string            `json:"language"`
	EmbeddedData map[string]any   `json:"embedded_data"`
}

type ListResponsesFilter struct {
	Status   string     `query:"status"`
	DateFrom *time.Time `query:"date_from"`
	DateTo   *time.Time `query:"date_to"`
	Cursor   string     `query:"cursor"`
	Limit    int        `query:"limit"`
}

type ResponseWithAnswers struct {
	models.Response
	Answers []models.Answer `json:"answers"`
}

// --- Methods ---

// SubmitResponse validates and saves a survey response.
func (s *ResponseService) SubmitResponse(ctx context.Context, input SubmitResponseInput) (*models.Response, error) {
	// Get survey to validate.
	survey, err := s.surveyRepo.GetByID(ctx, input.SurveyID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if survey == nil {
		return nil, apperr.NotFound("survey")
	}
	if survey.Status != models.SurveyStatusPublished {
		return nil, apperr.BadRequest("survey is not accepting responses")
	}

	// Validate required questions are answered.
	questionMap := make(map[string]models.QuestionDef)
	for _, q := range survey.Structure.Questions {
		questionMap[q.ID] = q
	}

	answerMap := make(map[string]any)
	for _, a := range input.Answers {
		answerMap[a.QuestionID] = a.Value
	}

	for _, q := range survey.Structure.Questions {
		if q.Required {
			val, exists := answerMap[q.ID]
			if !exists || val == nil || val == "" {
				return nil, apperr.BadRequest("required question not answered: " + q.Title)
			}
		}
	}

	// Create response record.
	source := input.Metadata.Source
	if source == "" {
		source = "direct"
	}

	resp := &models.Response{
		SurveyID:     uuid.MustParse(input.SurveyID),
		Status:       models.ResponseStatusCompleted,
		Source:       source,
		Language:     input.Metadata.Language,
		EmbeddedData: input.Metadata.EmbeddedData,
		QualityFlags: models.QualityFlags{},
	}
	if resp.Language == "" {
		resp.Language = "en"
	}
	if input.Metadata.Device != "" {
		resp.RespondentDevice = &input.Metadata.Device
	}
	if input.Metadata.Browser != "" {
		resp.RespondentBrowser = &input.Metadata.Browser
	}
	if input.Metadata.OS != "" {
		resp.RespondentOS = &input.Metadata.OS
	}
	if input.Metadata.Country != "" {
		resp.RespondentCountry = &input.Metadata.Country
	}
	if input.Metadata.IPHash != "" {
		resp.RespondentIPHash = &input.Metadata.IPHash
	}

	// Use a transaction for response + answers + completion + count.
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	defer tx.Rollback(ctx)

	// Create response within transaction.
	createQuery := `
		INSERT INTO responses (survey_id, respondent_email, respondent_name,
			respondent_device, respondent_browser, respondent_os, respondent_country,
			respondent_ip_hash, status, source, source_detail, language, embedded_data, quality_flags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, started_at, created_at`
	embeddedJSON, err := json.Marshal(resp.EmbeddedData)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	qualityJSON, err := json.Marshal(resp.QualityFlags)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	err = tx.QueryRow(ctx, createQuery,
		resp.SurveyID, resp.RespondentEmail, resp.RespondentName,
		resp.RespondentDevice, resp.RespondentBrowser, resp.RespondentOS, resp.RespondentCountry,
		resp.RespondentIPHash, resp.Status, resp.Source, resp.SourceDetail, resp.Language,
		embeddedJSON, qualityJSON,
	).Scan(&resp.ID, &resp.StartedAt, &resp.CreatedAt)
	if err != nil {
		slog.Error("failed to create response", "error", err)
		return nil, apperr.Internal(err)
	}

	// Create answers within transaction.
	answers := make([]models.Answer, len(input.Answers))
	for i, a := range input.Answers {
		var value map[string]any
		switch v := a.Value.(type) {
		case map[string]any:
			value = v
		default:
			value = map[string]any{"value": v}
		}
		answers[i] = models.Answer{
			ResponseID: resp.ID,
			QuestionID: a.QuestionID,
			Value:      value,
			Metadata:   map[string]any{},
		}
	}
	if len(answers) > 0 {
		insertAnswerQuery := `
			INSERT INTO answers (response_id, question_id, value, metadata)
			VALUES ($1, $2, $3, $4)`
		for _, a := range answers {
			valueJSON, err := json.Marshal(a.Value)
			if err != nil {
				return nil, apperr.Internal(err)
			}
			metadataJSON, err := json.Marshal(a.Metadata)
			if err != nil {
				return nil, apperr.Internal(err)
			}
			if _, err := tx.Exec(ctx, insertAnswerQuery, a.ResponseID, a.QuestionID, valueJSON, metadataJSON); err != nil {
				slog.Error("failed to create answer", "error", err)
				return nil, apperr.Internal(err)
			}
		}
	}

	// Complete the response within transaction.
	duration := input.Metadata.Duration
	if _, err := tx.Exec(ctx,
		`UPDATE responses SET status = 'completed', completed_at = NOW(), duration_ms = $2 WHERE id = $1`,
		resp.ID.String(), duration,
	); err != nil {
		slog.Error("failed to complete response", "error", err)
		return nil, apperr.Internal(err)
	}

	// Increment survey response count within transaction.
	if _, err := tx.Exec(ctx,
		`UPDATE surveys SET response_count = response_count + 1, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		input.SurveyID,
	); err != nil {
		slog.Error("failed to increment response count", "error", err)
		return nil, apperr.Internal(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperr.Internal(err)
	}

	return resp, nil
}

// ListResponses returns paginated responses for a survey.
func (s *ResponseService) ListResponses(ctx context.Context, orgID string, surveyID string, filter ListResponsesFilter) ([]models.Response, string, int, error) {
	// Verify org owns survey.
	survey, err := s.surveyRepo.GetByID(ctx, surveyID)
	if err != nil {
		return nil, "", 0, apperr.Internal(err)
	}
	if survey == nil {
		return nil, "", 0, apperr.NotFound("survey")
	}
	if survey.OrgID.String() != orgID {
		return nil, "", 0, apperr.Forbidden()
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return s.responseRepo.List(ctx, surveyID, filter.Status, filter.DateFrom, filter.DateTo, filter.Limit, filter.Cursor)
}

// GetResponse returns a single response with answers.
func (s *ResponseService) GetResponse(ctx context.Context, orgID, surveyID, responseID string) (*ResponseWithAnswers, error) {
	// Verify org owns survey.
	survey, err := s.surveyRepo.GetByID(ctx, surveyID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if survey == nil || survey.OrgID.String() != orgID {
		return nil, apperr.Forbidden()
	}

	resp, err := s.responseRepo.GetByID(ctx, responseID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if resp == nil {
		return nil, apperr.NotFound("response")
	}
	if resp.SurveyID.String() != surveyID {
		return nil, apperr.Forbidden()
	}

	answers, err := s.responseRepo.GetAnswersByResponseID(ctx, responseID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return &ResponseWithAnswers{
		Response: *resp,
		Answers:  answers,
	}, nil
}

// DeleteResponse removes a response.
func (s *ResponseService) DeleteResponse(ctx context.Context, orgID, surveyID, responseID string) error {
	// Verify org owns survey.
	survey, err := s.surveyRepo.GetByID(ctx, surveyID)
	if err != nil {
		return apperr.Internal(err)
	}
	if survey == nil || survey.OrgID.String() != orgID {
		return apperr.Forbidden()
	}

	return s.responseRepo.Delete(ctx, responseID)
}

// ExportResponses generates an export of responses.
func (s *ResponseService) ExportResponses(ctx context.Context, orgID, surveyID, format string, filter ListResponsesFilter) (string, error) {
	// Verify org owns survey.
	survey, err := s.surveyRepo.GetByID(ctx, surveyID)
	if err != nil {
		return "", apperr.Internal(err)
	}
	if survey == nil || survey.OrgID.String() != orgID {
		return "", apperr.Forbidden()
	}

	// Get all responses (without pagination for export).
	responses, _, _, err := s.responseRepo.List(ctx, surveyID, filter.Status, filter.DateFrom, filter.DateTo, 10000, "")
	if err != nil {
		return "", apperr.Internal(err)
	}

	// For MVP, return JSON. In production, would generate CSV/XLSX.
	data, err := json.Marshal(responses)
	if err != nil {
		return "", apperr.Internal(err)
	}
	return string(data), nil
}
