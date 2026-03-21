package workers

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// Task type constants used to identify the kind of background job.
const (
	TypeSendEmail          = "email:send"
	TypeSendBulkEmail      = "email:send_bulk"
	TypeAIAnalysis         = "ai:analysis"
	TypeAISentiment        = "ai:sentiment"
	TypeWebhookDispatch    = "webhook:dispatch"
	TypeWebhookRetry       = "webhook:retry"
	TypeAnalyticsCache     = "analytics:cache_update"
	TypeExportCSV          = "export:csv"
	TypeExportPDF          = "export:pdf"
	TypeCleanupExpired     = "cleanup:expired"
	TypeSyncIntegration    = "integration:sync"
	TypeProcessUpload      = "files:process_upload"
)

// EmailPayload holds the data needed to send a single email.
type EmailPayload struct {
	To       string            `json:"to"`
	From     string            `json:"from"`
	Subject  string            `json:"subject"`
	HTMLBody string            `json:"html_body"`
	TextBody string            `json:"text_body"`
	ReplyTo  string            `json:"reply_to,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// BulkEmailPayload holds the data needed to send an email to a list of recipients.
type BulkEmailPayload struct {
	Recipients []string          `json:"recipients"`
	From       string            `json:"from"`
	Subject    string            `json:"subject"`
	HTMLBody   string            `json:"html_body"`
	TextBody   string            `json:"text_body"`
	CampaignID string            `json:"campaign_id,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// AIAnalysisPayload holds the data needed to run AI analysis on survey responses.
type AIAnalysisPayload struct {
	SurveyID   string `json:"survey_id"`
	ResponseID string `json:"response_id,omitempty"`
	// Empty ResponseID means analyze all responses for the survey.
}

// AISentimentPayload holds the data needed for sentiment analysis of a text answer.
type AISentimentPayload struct {
	ResponseID string `json:"response_id"`
	QuestionID string `json:"question_id"`
	Text       string `json:"text"`
}

// WebhookPayload holds the data needed to dispatch a webhook event.
type WebhookPayload struct {
	WebhookID  string            `json:"webhook_id"`
	OrgID      string            `json:"org_id"`
	Event      string            `json:"event"`
	Payload    map[string]any    `json:"payload"`
	RetryCount int               `json:"retry_count"`
	Secret     string            `json:"secret"`
}

// AnalyticsCachePayload holds the data needed to refresh analytics cache for a survey.
type AnalyticsCachePayload struct {
	SurveyID string `json:"survey_id"`
}

// ExportPayload holds the data needed to generate an export file.
type ExportPayload struct {
	SurveyID  string `json:"survey_id"`
	Format    string `json:"format"` // csv, pdf, xlsx
	UserID    string `json:"user_id"`
	Filters   map[string]any `json:"filters,omitempty"`
}

// CleanupPayload holds the data for periodic cleanup tasks.
type CleanupPayload struct {
	Task string `json:"task"` // expired_responses, old_audit_logs, etc.
}

// NewEmailTask creates a new asynq Task for sending an email.
func NewEmailTask(payload EmailPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendEmail, data), nil
}

// NewBulkEmailTask creates a new asynq Task for sending a bulk email.
func NewBulkEmailTask(payload BulkEmailPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendBulkEmail, data), nil
}

// NewAIAnalysisTask creates a new asynq Task for running AI analysis.
func NewAIAnalysisTask(payload AIAnalysisPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeAIAnalysis, data), nil
}

// NewWebhookDispatchTask creates a new asynq Task for dispatching a webhook.
func NewWebhookDispatchTask(payload WebhookPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeWebhookDispatch, data), nil
}

// NewAnalyticsCacheTask creates a new asynq Task for refreshing analytics cache.
func NewAnalyticsCacheTask(payload AnalyticsCachePayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeAnalyticsCache, data), nil
}
