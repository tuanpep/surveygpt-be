package models

import "time"

// AnalyticsSummary provides a high-level overview of survey response data.
type AnalyticsSummary struct {
	TotalResponses    int              `json:"total_responses"`
	CompletedResponses int             `json:"completed_responses"`
	CompletionRate    float64          `json:"completion_rate"`
	AvgDurationMs     int              `json:"avg_duration_ms"`
	MedianDurationMs  int              `json:"median_duration_ms"`
	TotalViews        int              `json:"total_views"`
	ConversionRate    float64          `json:"conversion_rate"`
	FirstResponseAt   *time.Time       `json:"first_response_at,omitempty"`
	LastResponseAt    *time.Time       `json:"last_response_at,omitempty"`
	DailyCounts       []DailyCount     `json:"daily_counts,omitempty"`
	DeviceBreakdown   map[string]int   `json:"device_breakdown,omitempty"`
	BrowserBreakdown  map[string]int   `json:"browser_breakdown,omitempty"`
	CountryBreakdown  map[string]int   `json:"country_breakdown,omitempty"`
	SourceBreakdown   map[string]int   `json:"source_breakdown,omitempty"`
}

// DailyCount holds the number of responses received on a specific day.
type DailyCount struct {
	Date  time.Time `json:"date"`
	Count int       `json:"count"`
}

// QuestionStats provides aggregated statistics for a single question.
type QuestionStats struct {
	QuestionID    string         `json:"question_id"`
	QuestionType  string         `json:"question_type"`
	QuestionTitle string         `json:"question_title"`
	TotalAnswers  int            `json:"total_answers"`
	SkippedCount  int            `json:"skipped_count"`
	ChoiceStats   []ChoiceStat   `json:"choice_stats,omitempty"`
	NumericStats  *NumericStats  `json:"numeric_stats,omitempty"`
	TextStats     *TextStats     `json:"text_stats,omitempty"`
}

// ChoiceStat holds the count and percentage for a single choice option.
type ChoiceStat struct {
	ChoiceID   string  `json:"choice_id"`
	Label      string  `json:"label"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// NumericStats holds statistical measures for numeric question types.
type NumericStats struct {
	Mean     float64 `json:"mean"`
	Median   float64 `json:"median"`
	Mode     float64 `json:"mode"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	StdDev   float64 `json:"std_dev"`
}

// TextStats holds aggregated data for text question types.
type TextStats struct {
	TotalWords     int            `json:"total_words"`
	AvgWordCount   float64        `json:"avg_word_count"`
	TopKeywords    []KeywordCount `json:"top_keywords,omitempty"`
	SentimentScore *float64       `json:"sentiment_score,omitempty"`
}

// KeywordCount holds a keyword and its occurrence count.
type KeywordCount struct {
	Keyword string `json:"keyword"`
	Count   int    `json:"count"`
}

// CrossTabResult represents a cross-tabulation between two questions.
type CrossTabResult struct {
	RowQuestionID    string         `json:"row_question_id"`
	ColumnQuestionID string         `json:"column_question_id"`
	Headers          []string       `json:"headers"`
	Rows             []CrossTabRow  `json:"rows"`
}

// CrossTabRow represents a single row in a cross-tabulation result.
type CrossTabRow struct {
	Label    string  `json:"label"`
	Values   []int   `json:"values"`
	Total    int     `json:"total"`
}

// DropoffStep tracks where respondents abandon a survey.
type DropoffStep struct {
	StepID      string  `json:"step_id"`
	StepType    string  `json:"step_type"`
	StepLabel   string  `json:"step_label"`
	Views       int     `json:"views"`
	Dropoffs    int     `json:"dropoffs"`
	DropoffRate float64 `json:"dropoff_rate"`
}

// AIInsights holds AI-generated analysis results for a survey.
type AIInsights struct {
	Summary         string            `json:"summary"`
	KeyFindings     []string          `json:"key_findings"`
	Recommendations []string          `json:"recommendations"`
	Sentiment       string            `json:"sentiment"`
	Themes          []AITheme         `json:"themes,omitempty"`
	Anomalies       []AIAnomaly       `json:"anomalies,omitempty"`
	GeneratedAt     time.Time         `json:"generated_at"`
}

// AITheme represents a recurring theme identified by AI analysis.
type AITheme struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Mentions    int     `json:"mentions"`
}

// AIAnomaly represents an unusual pattern detected by AI analysis.
type AIAnomaly struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"` // low, medium, high
	QuestionID  string  `json:"question_id,omitempty"`
}
