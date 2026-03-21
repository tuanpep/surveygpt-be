package services

import (
	"context"
	"fmt"

	"github.com/surveyflow/be/internal/models"
	apperr "github.com/surveyflow/be/internal/pkg/errors"
	"github.com/surveyflow/be/internal/repository"
)

// DistributionService handles survey distribution channels.
type DistributionService struct {
	surveyRepo    repository.SurveyRepo
	emailListRepo repository.EmailListRepo
}

// NewDistributionService creates a new DistributionService.
func NewDistributionService(surveyRepo repository.SurveyRepo, emailListRepo repository.EmailListRepo) *DistributionService {
	return &DistributionService{
		surveyRepo:    surveyRepo,
		emailListRepo: emailListRepo,
	}
}

// --- Input types ---

type QRCodeInput struct {
	SurveyID  string `json:"survey_id" validate:"required"`
	Size      int    `json:"size"`
	ErrorLevel string `json:"error_level"` // L, M, Q, H
}

type QRCodeOutput struct {
	SVG     string `json:"svg"`
	PNGURL  string `json:"png_url,omitempty"`
}

type EmbedCodeInput struct {
	SurveyID string `json:"survey_id" validate:"required"`
	Mode     string `json:"mode"`     // popup, embedded, fullpage
	Trigger  string `json:"trigger"`  // auto, button, custom
	Width    string `json:"width"`
	Height   string `json:"height"`
}

type EmbedCodeOutput struct {
	HTML string `json:"html"`
	JS  string `json:"js"`
}

type CreateEmailListInput struct {
	OrgID    string             `json:"org_id" validate:"required"`
	Name     string             `json:"name" validate:"required"`
	Contacts []EmailListContact `json:"contacts"`
}

type EmailListContact struct {
	Email       string            `json:"email" validate:"required,email"`
	FirstName   string            `json:"first_name"`
	LastName    string            `json:"last_name"`
	Metadata    map[string]any    `json:"metadata"`
}

type SendSurveyEmailsInput struct {
	SurveyID   string `json:"survey_id" validate:"required"`
	ListID     string `json:"list_id" validate:"required"`
	Subject    string `json:"subject" validate:"required"`
	Body       string `json:"body"`
	ScheduleAt *string `json:"schedule_at"` // ISO 8601
}

type SendSurveyEmailsOutput struct {
	JobID   string `json:"job_id"`
	Count   int    `json:"count"`
	Status  string `json:"status"`
}

// --- Methods ---

// GenerateQRCode generates a QR code for a survey.
func (s *DistributionService) GenerateQRCode(ctx context.Context, orgID, surveyID string, input QRCodeInput) (*QRCodeOutput, error) {
	// Verify org owns survey.
	survey, err := s.surveyRepo.GetByID(ctx, surveyID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if survey == nil || survey.OrgID.String() != orgID {
		return nil, apperr.Forbidden()
	}

	// For MVP, generate a simple SVG placeholder.
	// In production, would use a QR code library.
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
	<rect width="100" height="100" fill="white"/>
	<text x="50" y="50" text-anchor="middle" dominant-baseline="middle" font-size="10">QR: %s</text>
</svg>`, surveyID)

	return &QRCodeOutput{
		SVG:    svg,
		PNGURL: "", // Would return presigned URL in production
	}, nil
}

// GetEmbedCode returns embed code for a survey.
func (s *DistributionService) GetEmbedCode(ctx context.Context, orgID, surveyID string, input EmbedCodeInput) (*EmbedCodeOutput, error) {
	// Verify org owns survey.
	survey, err := s.surveyRepo.GetByID(ctx, surveyID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if survey == nil || survey.OrgID.String() != orgID {
		return nil, apperr.Forbidden()
	}

	// Default to popup mode.
	mode := input.Mode
	if mode == "" {
		mode = "popup"
	}

	// Generate embed code.
	html := fmt.Sprintf(`<div id="surveyflow-embed" data-survey-id="%s" data-mode="%s"></div>
<script src="https://surveyflow.com/embed.js"></script>`, surveyID, mode)

	js := fmt.Sprintf(`
SurveyFlow.init({
  surveyId: "%s",
  mode: "%s",
  trigger: "%s",
  autoOpen: %t
});`, surveyID, mode, input.Trigger, mode == "popup" && input.Trigger == "auto")

	return &EmbedCodeOutput{
		HTML: html,
		JS:   js,
	}, nil
}

// GetEmailLists returns email lists for an organization.
func (s *DistributionService) GetEmailLists(ctx context.Context, orgID string) ([]models.EmailList, error) {
	return s.emailListRepo.List(ctx, orgID)
}

// CreateEmailList creates a new email list.
func (s *DistributionService) CreateEmailList(ctx context.Context, input CreateEmailListInput) (*models.EmailList, error) {
	list := &models.EmailList{
		OrgID:  input.OrgID,
		Name:   input.Name,
		Status: models.EmailListStatusActive,
	}

	if err := s.emailListRepo.Create(ctx, list); err != nil {
		return nil, apperr.Internal(err)
	}

	// Add contacts.
	if len(input.Contacts) > 0 {
		contacts := make([]models.EmailListContact, len(input.Contacts))
		for i, c := range input.Contacts {
			contacts[i] = models.EmailListContact{
				ListID:    list.ID.String(),
				Email:     c.Email,
				FirstName: c.FirstName,
				LastName:  c.LastName,
				Metadata:  c.Metadata,
				Status:    models.ContactStatusActive,
			}
		}
		if err := s.emailListRepo.AddContacts(ctx, contacts); err != nil {
			return nil, apperr.Internal(err)
		}
		list.ContactCount = len(input.Contacts)
	}

	return list, nil
}

// SendSurveyEmails sends survey emails to a list.
func (s *DistributionService) SendSurveyEmails(ctx context.Context, orgID, surveyID string, input SendSurveyEmailsInput) (*SendSurveyEmailsOutput, error) {
	// Verify org owns survey.
	survey, err := s.surveyRepo.GetByID(ctx, surveyID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if survey == nil || survey.OrgID.String() != orgID {
		return nil, apperr.Forbidden()
	}

	// Get email list.
	list, err := s.emailListRepo.GetByID(ctx, input.ListID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if list == nil || list.OrgID != orgID {
		return nil, apperr.NotFound("email list")
	}

	// Get contacts.
	contacts, err := s.emailListRepo.GetContacts(ctx, input.ListID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	if len(contacts) == 0 {
		return nil, apperr.BadRequest("email list has no contacts")
	}

	// For MVP, return placeholder.
	// In production, would enqueue email tasks to worker.
	jobID := "job_" + surveyID[:8] + "_" + input.ListID[:8]

	return &SendSurveyEmailsOutput{
		JobID:  jobID,
		Count:  len(contacts),
		Status: "scheduled",
	}, nil
}
