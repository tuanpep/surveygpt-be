package models

import (
	"time"

	"github.com/google/uuid"
)

// SurveyStatus represents the lifecycle state of a survey.
type SurveyStatus string

const (
	SurveyStatusDraft     SurveyStatus = "draft"
	SurveyStatusPublished SurveyStatus = "published"
	SurveyStatusClosed    SurveyStatus = "closed"
	SurveyStatusArchived  SurveyStatus = "archived"
)

// UIMode represents the visual presentation mode for a survey.
type UIMode string

const (
	UIModeClassic       UIMode = "classic"
	UIModeMinimal       UIMode = "minimal"
	UIModeCards         UIMode = "cards"
	UIModeConversational UIMode = "conversational"
)

// Survey represents a survey entity in the database.
type Survey struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	OrgID         uuid.UUID       `json:"org_id" db:"org_id"`
	CreatedBy     uuid.UUID       `json:"created_by" db:"created_by"`
	Title         string          `json:"title" db:"title"`
	Description   string          `json:"description" db:"description"`
	Status        SurveyStatus    `json:"status" db:"status"`
	UIMode        UIMode          `json:"ui_mode" db:"ui_mode"`
	Structure     SurveyStructure `json:"structure" db:"structure"`
	Settings      SurveySettings  `json:"settings" db:"settings"`
	Theme         SurveyTheme     `json:"theme" db:"theme"`
	ResponseCount int             `json:"response_count" db:"response_count"`
	ViewCount     int             `json:"view_count" db:"view_count"`
	PublishedAt   *time.Time      `json:"published_at,omitempty" db:"published_at"`
	ClosedAt      *time.Time      `json:"closed_at,omitempty" db:"closed_at"`
	DeletedAt     *time.Time      `json:"deleted_at,omitempty" db:"deleted_at"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// SurveyStructure defines the full schema of a survey: its questions, pages,
// and conditional flow logic.
type SurveyStructure struct {
	Questions []QuestionDef `json:"questions"`
	Blocks    []BlockDef    `json:"blocks,omitempty"`
	Flow      []FlowStep    `json:"flow,omitempty"`
}

// QuestionDef defines a single survey question.
type QuestionDef struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"` // short_text, long_text, multiple_choice, checkbox, rating, scale, date, file_upload, etc.
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Required    bool              `json:"required"`
	Choices     []Choice          `json:"choices,omitempty"`
	Validation  *Validation       `json:"validation,omitempty"`
	Logic       []QuestionLogic   `json:"logic,omitempty"`
	Scoring     *ScoringConfig    `json:"scoring,omitempty"`
	Properties  map[string]any    `json:"properties,omitempty"`
	Layout      map[string]any    `json:"layout,omitempty"`
}

// Choice represents a selectable option in a multiple choice or checkbox question.
type Choice struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Value    any    `json:"value,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	Weight   int    `json:"weight,omitempty"`
	IsOther  bool   `json:"is_other,omitempty"`
}

// BlockDef groups questions into logical sections or pages.
type BlockDef struct {
	ID          string   `json:"id"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	QuestionIDs []string `json:"question_ids"`
	Conditions  []LogicCondition `json:"conditions,omitempty"`
	Randomize   bool     `json:"randomize,omitempty"`
}

// FlowStep defines a transition in the survey flow, controlling which block
// or question to show next based on conditions.
type FlowStep struct {
	FromID     string           `json:"from_id"`     // question_id or block_id
	ToID       string           `json:"to_id"`       // question_id or block_id
	Conditions []LogicCondition `json:"conditions"`
}

// LogicCondition represents a single condition used in flow logic and block visibility.
type LogicCondition struct {
	QuestionID string `json:"question_id"`
	Operator   string `json:"operator"` // equals, not_equals, contains, greater_than, less_than, is_answered, etc.
	Value      any    `json:"value,omitempty"`
}

// QuestionLogic defines conditional logic for a specific question.
type QuestionLogic struct {
	TargetQuestionID string           `json:"target_question_id"`
	Action           string           `json:"action"` // show, hide, require, skip
	Conditions       []LogicCondition `json:"conditions"`
}

// Validation defines constraints on question answers.
type Validation struct {
	Required       *bool  `json:"required,omitempty"`
	MinLength      *int   `json:"min_length,omitempty"`
	MaxLength      *int   `json:"max_length,omitempty"`
	MinValue       *int   `json:"min_value,omitempty"`
	MaxValue       *int   `json:"max_value,omitempty"`
	MinChoices     *int   `json:"min_choices,omitempty"`
	MaxChoices     *int   `json:"max_choices,omitempty"`
	Pattern        string `json:"pattern,omitempty"`
	AcceptTypes    string `json:"accept_types,omitempty"` // for file_upload
	MaxFileSize    *int   `json:"max_file_size,omitempty"`
	MinDate        string `json:"min_date,omitempty"`
	MaxDate        string `json:"max_date,omitempty"`
	CustomMessage  string `json:"custom_message,omitempty"`
}

// ScoringConfig defines how a question contributes to a total score.
type ScoringConfig struct {
	Enabled    bool              `json:"enabled"`
	Type       string            `json:"type,omitempty"` // points, quiz
	Points     map[string]int    `json:"points,omitempty"` // choice_id -> points
	CorrectIDs []string          `json:"correct_ids,omitempty"` // for quiz mode
	Alias      string            `json:"alias,omitempty"`
}

// SurveySettings holds behavioral and access settings for a survey.
type SurveySettings struct {
	IsPublic            bool            `json:"is_public"`
	Password            string          `json:"password,omitempty"`
	AllowAnonymous      bool            `json:"allow_anonymous"`
	CollectEmail        bool            `json:"collect_email"`
	ProgressBar         bool            `json:"progress_bar"`
	ShowQuestionNumbers bool            `json:"show_question_numbers"`
	ShuffleQuestions    bool            `json:"shuffle_questions"`
	OneQuestionPerPage  bool            `json:"one_question_per_page"`
	AllowBackNavigation bool            `json:"allow_back_navigation"`
	ConfirmationMessage string          `json:"confirmation_message,omitempty"`
	RedirectURL         string          `json:"redirect_url,omitempty"`
	ClosedMessage       string          `json:"closed_message,omitempty"`
	MaxResponses        *int            `json:"max_responses,omitempty"`
	ScheduleStart       *time.Time      `json:"schedule_start,omitempty"`
	ScheduleEnd         *time.Time      `json:"schedule_end,omitempty"`
	Localization        map[string]any  `json:"localization,omitempty"`
	Webhooks            []string        `json:"webhooks,omitempty"`
}

// SurveyTheme defines the visual appearance of a survey.
type SurveyTheme struct {
	Font          string `json:"font,omitempty"`
	PrimaryColor  string `json:"primary_color,omitempty"`
	AccentColor   string `json:"accent_color,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	TextColor     string `json:"text_color,omitempty"`
	LogoURL       string `json:"logo_url,omitempty"`
	FaviconURL    string `json:"favicon_url,omitempty"`
	CoverImageURL string `json:"cover_image_url,omitempty"`
	CSS           string `json:"css,omitempty"`
	CustomCSS     string `json:"custom_css,omitempty"`
	BorderRadius  int    `json:"border_radius,omitempty"`
	QuestionWidth string `json:"question_width,omitempty"`
	Layout        string `json:"layout,omitempty"` // centered, full_width, split
}
