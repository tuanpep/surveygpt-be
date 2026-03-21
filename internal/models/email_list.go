package models

import (
	"time"

	"github.com/google/uuid"
)

// EmailListStatus represents the status of an email list.
type EmailListStatus string

const (
	EmailListStatusActive   EmailListStatus = "active"
	EmailListStatusArchived EmailListStatus = "archived"
)

// ContactStatus represents the status of a contact.
type ContactStatus string

const (
	ContactStatusActive    ContactStatus = "active"
	ContactStatusBounced   ContactStatus = "bounced"
	ContactStatusUnsubscribed ContactStatus = "unsubscribed"
)

// EmailList represents an email distribution list.
type EmailList struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	OrgID        string          `json:"org_id" db:"org_id"`
	Name         string          `json:"name" db:"name"`
	Status       EmailListStatus `json:"status" db:"status"`
	ContactCount int             `json:"contact_count" db:"contact_count"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time      `json:"deleted_at,omitempty" db:"deleted_at"`
}

// EmailListContact represents a contact in an email list.
type EmailListContact struct {
	ID        uuid.UUID             `json:"id" db:"id"`
	ListID    string                `json:"list_id" db:"list_id"`
	Email     string                `json:"email" db:"email"`
	FirstName string                `json:"first_name" db:"first_name"`
	LastName  string                `json:"last_name" db:"last_name"`
	Metadata  map[string]any        `json:"metadata" db:"metadata"`
	Status    ContactStatus         `json:"status" db:"status"`
	CreatedAt time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt time.Time             `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time            `json:"deleted_at,omitempty" db:"deleted_at"`
}

// EmailCampaign represents an email campaign.
type EmailCampaign struct {
	ID          uuid.UUID `json:"id" db:"id"`
	OrgID       string    `json:"org_id" db:"org_id"`
	SurveyID    string    `json:"survey_id" db:"survey_id"`
	ListID      string    `json:"list_id" db:"list_id"`
	Subject     string    `json:"subject" db:"subject"`
	Body        string    `json:"body" db:"body"`
	Status      string    `json:"status" db:"status"` // draft, scheduled, sending, sent, failed
	ScheduledAt *time.Time `json:"scheduled_at" db:"scheduled_at"`
	SentAt      *time.Time `json:"sent_at" db:"sent_at"`
	TotalCount  int        `json:"total_count" db:"total_count"`
	SentCount   int        `json:"sent_count" db:"sent_count"`
	OpenedCount int        `json:"opened_count" db:"opened_count"`
	ClickedCount int       `json:"clicked_count" db:"clicked_count"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
