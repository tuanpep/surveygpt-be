package services

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/surveyflow/be/internal/config"
	"github.com/surveyflow/be/internal/models"
	apperr "github.com/surveyflow/be/internal/pkg/errors"
	"github.com/surveyflow/be/internal/repository"
)

// SurveyService handles survey business logic.
type SurveyService struct {
	surveyRepo   repository.SurveyRepo
	templateRepo repository.TemplateRepo
	config       *config.Config
}

// NewSurveyService creates a new SurveyService.
func NewSurveyService(surveyRepo repository.SurveyRepo, templateRepo repository.TemplateRepo, cfg *config.Config) *SurveyService {
	return &SurveyService{
		surveyRepo:   surveyRepo,
		templateRepo: templateRepo,
		config:       cfg,
	}
}

// --- Input types ---

type CreateSurveyInput struct {
	Title       string          `json:"title" validate:"required,min=1,max=500"`
	Description string          `json:"description"`
	UIMode      string          `json:"ui_mode" validate:"omitempty,oneof=classic minimal cards conversational"`
}

type UpdateSurveyInput struct {
	Title       *string         `json:"title"`
	Description *string         `json:"description"`
	UIMode      *string         `json:"ui_mode"`
	Status      *string         `json:"status"`
	Structure   *models.SurveyStructure `json:"structure"`
	Settings    *models.SurveySettings  `json:"settings"`
	Theme       *models.SurveyTheme     `json:"theme"`
}

type ListSurveysFilter struct {
	Status   string `query:"status"`
	Search   string `query:"search"`
	SortBy   string `query:"sort_by"`
	SortDir  string `query:"sort_dir"`
	Cursor   string `query:"cursor"`
	Limit    int    `query:"limit"`
}

type PublicSurvey struct {
	ID          string               `json:"id"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Status      string               `json:"status"`
	UIMode      string               `json:"ui_mode"`
	Structure   models.SurveyStructure `json:"structure"`
	Settings    models.SurveySettings  `json:"settings"`
	Theme       models.SurveyTheme     `json:"theme"`
	ShareURL    string               `json:"share_url"`
}

// --- Methods ---

// CreateSurvey creates a new survey for the given organization.
func (s *SurveyService) CreateSurvey(ctx context.Context, orgID, userID string, input CreateSurveyInput) (*models.Survey, error) {
	uiMode := models.UIModeClassic
	if input.UIMode != "" {
		uiMode = models.UIMode(input.UIMode)
	}

	survey := &models.Survey{
		OrgID:     uuid.MustParse(orgID),
		CreatedBy: uuid.MustParse(userID),
		Title:     input.Title,
		Description: input.Description,
		Status:    models.SurveyStatusDraft,
		UIMode:    uiMode,
		Structure: models.SurveyStructure{},
		Settings:  models.SurveySettings{},
		Theme:     models.SurveyTheme{},
	}

	if err := s.surveyRepo.Create(ctx, survey); err != nil {
		slog.Error("failed to create survey", "error", err)
		return nil, apperr.Internal(err)
	}

	return survey, nil
}

// GetSurvey returns a survey by ID, verifying org ownership.
func (s *SurveyService) GetSurvey(ctx context.Context, orgID, surveyID string) (*models.Survey, error) {
	survey, err := s.surveyRepo.GetByID(ctx, surveyID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if survey == nil {
		return nil, apperr.NotFound("survey")
	}
	if survey.OrgID.String() != orgID {
		return nil, apperr.Forbidden()
	}
	return survey, nil
}

// ListSurveys returns paginated surveys for an organization.
func (s *SurveyService) ListSurveys(ctx context.Context, orgID string, filter ListSurveysFilter) ([]models.Survey, string, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = s.config.Pagination.DefaultLimit
	}
	if filter.Limit > s.config.Pagination.MaxLimit {
		filter.Limit = s.config.Pagination.MaxLimit
	}

	return s.surveyRepo.List(ctx, orgID, filter.Status, filter.Search, filter.SortBy, filter.SortDir, filter.Limit, filter.Cursor)
}

// UpdateSurvey applies partial updates to a survey.
func (s *SurveyService) UpdateSurvey(ctx context.Context, orgID, surveyID string, input UpdateSurveyInput) (*models.Survey, error) {
	survey, err := s.GetSurvey(ctx, orgID, surveyID)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		survey.Title = *input.Title
	}
	if input.Description != nil {
		survey.Description = *input.Description
	}
	if input.UIMode != nil {
		survey.UIMode = models.UIMode(*input.UIMode)
	}
	if input.Structure != nil {
		survey.Structure = *input.Structure
	}
	if input.Settings != nil {
		survey.Settings = *input.Settings
	}
	if input.Theme != nil {
		survey.Theme = *input.Theme
	}

	if err := s.surveyRepo.Update(ctx, survey); err != nil {
		return nil, apperr.Internal(err)
	}
	return survey, nil
}

// PublishSurvey publishes a draft survey.
func (s *SurveyService) PublishSurvey(ctx context.Context, orgID, surveyID string) (*models.Survey, error) {
	survey, err := s.GetSurvey(ctx, orgID, surveyID)
	if err != nil {
		return nil, err
	}
	if survey.Status != models.SurveyStatusDraft {
		return nil, apperr.BadRequest("only draft surveys can be published")
	}

	// Validate that the survey has at least one question.
	if len(survey.Structure.Questions) == 0 {
		return nil, apperr.BadRequest("survey must have at least one question before publishing")
	}

	if err := s.surveyRepo.UpdateStatus(ctx, surveyID, models.SurveyStatusPublished); err != nil {
		return nil, apperr.Internal(err)
	}
	survey.Status = models.SurveyStatusPublished
	return survey, nil
}

// CloseSurvey closes an active survey.
func (s *SurveyService) CloseSurvey(ctx context.Context, orgID, surveyID string) (*models.Survey, error) {
	survey, err := s.GetSurvey(ctx, orgID, surveyID)
	if err != nil {
		return nil, err
	}

	if err := s.surveyRepo.UpdateStatus(ctx, surveyID, models.SurveyStatusClosed); err != nil {
		return nil, apperr.Internal(err)
	}
	survey.Status = models.SurveyStatusClosed
	return survey, nil
}

// DuplicateSurvey creates a copy of a survey.
func (s *SurveyService) DuplicateSurvey(ctx context.Context, orgID, surveyID string) (*models.Survey, error) {
	original, err := s.GetSurvey(ctx, orgID, surveyID)
	if err != nil {
		return nil, err
	}

	dup, err := s.surveyRepo.Duplicate(ctx, surveyID, "Copy of "+original.Title)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return dup, nil
}

// DeleteSurvey soft-deletes a survey.
func (s *SurveyService) DeleteSurvey(ctx context.Context, orgID, surveyID string) error {
	if _, err := s.GetSurvey(ctx, orgID, surveyID); err != nil {
		return err
	}
	return s.surveyRepo.SoftDelete(ctx, surveyID)
}

// GetPublicSurvey returns a survey for public viewing (respondent UI).
func (s *SurveyService) GetPublicSurvey(ctx context.Context, surveyID string, appURL string) (*PublicSurvey, error) {
	survey, err := s.surveyRepo.GetByID(ctx, surveyID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if survey == nil {
		return nil, apperr.NotFound("survey")
	}

	if survey.Status != models.SurveyStatusPublished {
		return nil, apperr.BadRequest("survey is not currently accepting responses")
	}

	// Increment view count (best effort).
	if err := s.surveyRepo.IncrementViewCount(ctx, surveyID); err != nil {
		slog.Warn("failed to increment view count", "survey_id", surveyID, "error", err)
	}

	shareURL := appURL + "/s/" + surveyID
	return &PublicSurvey{
		ID:          survey.ID.String(),
		Title:       survey.Title,
		Description: survey.Description,
		Status:      string(survey.Status),
		UIMode:      string(survey.UIMode),
		Structure:   survey.Structure,
		Settings:    survey.Settings,
		Theme:       survey.Theme,
		ShareURL:    shareURL,
	}, nil
}

// CreateFromTemplate creates a survey from a template.
func (s *SurveyService) CreateFromTemplate(ctx context.Context, orgID, userID, templateID string) (*models.Survey, error) {
	tmpl, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if tmpl == nil {
		return nil, apperr.NotFound("template")
	}

	input := CreateSurveyInput{
		Title:  tmpl.Title,
		UIMode: "classic",
	}
	survey, err := s.CreateSurvey(ctx, orgID, userID, input)
	if err != nil {
		return nil, err
	}

	survey.Structure = tmpl.Structure
	survey.Theme = tmpl.Theme
	if err := s.surveyRepo.Update(ctx, survey); err != nil {
		return nil, apperr.Internal(err)
	}

	// Increment template use count.
	if err := s.templateRepo.IncrementUseCount(ctx, templateID); err != nil {
		slog.Warn("failed to increment template use count", "template_id", templateID, "error", err)
	}

	return survey, nil
}

// SaveAsTemplate saves a survey as a custom template.
func (s *SurveyService) SaveAsTemplate(ctx context.Context, orgID, surveyID, name, category string) (*models.Template, error) {
	survey, err := s.GetSurvey(ctx, orgID, surveyID)
	if err != nil {
		return nil, err
	}

	orgUUID := uuid.MustParse(orgID)
	tmpl := &models.Template{
		OrgID:      &orgUUID,
		Category:  category,
		Title:     name,
		Tags:      []string{},
		Structure: survey.Structure,
		Theme:     survey.Theme,
	}

	if err := s.templateRepo.Create(ctx, tmpl); err != nil {
		return nil, apperr.Internal(err)
	}

	return tmpl, nil
}

// ListTemplates returns templates for browsing (system + org custom).
func (s *SurveyService) ListTemplates(ctx context.Context, category, orgID string) ([]models.Template, error) {
	return s.templateRepo.List(ctx, category, orgID)
}

// GetTemplateByID returns a single template by ID.
func (s *SurveyService) GetTemplateByID(ctx context.Context, id string) (*models.Template, error) {
	return s.templateRepo.GetByID(ctx, id)
}
