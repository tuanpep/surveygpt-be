package models

import (
	"time"

	"github.com/google/uuid"
)

// ResponseStatus represents the completion state of a survey response.
type ResponseStatus string

const (
	ResponseStatusInProgress ResponseStatus = "in_progress"
	ResponseStatusCompleted  ResponseStatus = "completed"
	ResponseStatusDisqualified ResponseStatus = "disqualified"
	ResponseStatusPartial    ResponseStatus = "partial"
)

// Response represents a single submission to a survey.
type Response struct {
	ID                uuid.UUID      `json:"id" db:"id"`
	SurveyID          uuid.UUID      `json:"survey_id" db:"survey_id"`
	RespondentEmail   *string        `json:"respondent_email,omitempty" db:"respondent_email"`
	RespondentName    *string        `json:"respondent_name,omitempty" db:"respondent_name"`
	RespondentDevice  *string        `json:"respondent_device,omitempty" db:"respondent_device"`
	RespondentBrowser *string        `json:"respondent_browser,omitempty" db:"respondent_browser"`
	RespondentOS      *string        `json:"respondent_os,omitempty" db:"respondent_os"`
	RespondentCountry *string        `json:"respondent_country,omitempty" db:"respondent_country"`
	RespondentIPHash  *string        `json:"respondent_ip_hash,omitempty" db:"respondent_ip_hash"`
	Status            ResponseStatus `json:"status" db:"status"`
	StartedAt         time.Time      `json:"started_at" db:"started_at"`
	CompletedAt       *time.Time     `json:"completed_at,omitempty" db:"completed_at"`
	DurationMs        *int           `json:"duration_ms,omitempty" db:"duration_ms"`
	Source            string         `json:"source" db:"source"`
	SourceDetail      *string        `json:"source_detail,omitempty" db:"source_detail"`
	Language          string         `json:"language" db:"language"`
	EmbeddedData      map[string]any `json:"embedded_data" db:"embedded_data"`
	QualityFlags      QualityFlags   `json:"quality_flags" db:"quality_flags"`
	AIAnalysis        map[string]any `json:"ai_analysis,omitempty" db:"ai_analysis"`
	CreatedAt         time.Time      `json:"created_at" db:"created_at"`
}

// Answer represents a single answer within a survey response.
type Answer struct {
	ID         uuid.UUID      `json:"id" db:"id"`
	ResponseID uuid.UUID      `json:"response_id" db:"response_id"`
	QuestionID string         `json:"question_id" db:"question_id"`
	Value      map[string]any `json:"value" db:"value"`
	Metadata   map[string]any `json:"metadata" db:"metadata"`
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
}

// QualityFlags tracks quality indicators for a response, used for fraud
// detection and data quality scoring.
type QualityFlags struct {
	IsBot            *bool   `json:"is_bot,omitempty"`
	IsDuplicate      *bool   `json:"is_duplicate,omitempty"`
	IsSpeedrun       *bool   `json:"is_speedrun,omitempty"`
	IsVPN            *bool   `json:"is_vpn,omitempty"`
	IsInconsistent   *bool   `json:"is_inconsistent,omitempty"`
	QualityScore     *int    `json:"quality_score,omitempty"` // 0-100
	Flags            []string `json:"flags,omitempty"`
}
