package models

import (
	"time"

	"github.com/google/uuid"
)

// Template represents a reusable survey template.
type Template struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	OrgID         *uuid.UUID      `json:"org_id,omitempty" db:"org_id"`
	Category      string          `json:"category" db:"category"`
	Title         string          `json:"title" db:"title"`
	Description   string          `json:"description" db:"description"`
	Tags          []string        `json:"tags" db:"tags"`
	Structure     SurveyStructure `json:"structure" db:"structure"`
	Theme         SurveyTheme     `json:"theme" db:"theme"`
	CoverImageURL *string         `json:"cover_image_url,omitempty" db:"cover_image_url"`
	IsFeatured    bool            `json:"is_featured" db:"is_featured"`
	UseCount      int             `json:"use_count" db:"use_count"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}
