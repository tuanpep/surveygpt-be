package dbutil

import (
	"github.com/jackc/pgx/v5"
	"github.com/surveyflow/be/internal/models"
)

// ScanSurvey scans a full survey row from pgx into a models.Survey.
func ScanSurvey(row pgx.Row, s *models.Survey) error {
	return row.Scan(
		&s.ID, &s.OrgID, &s.CreatedBy, &s.Title, &s.Description, &s.Status, &s.UIMode,
		&s.Structure, &s.Settings, &s.Theme, &s.ResponseCount, &s.ViewCount,
		&s.PublishedAt, &s.ClosedAt, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt,
	)
}

// ScanSurveyRows scans the current row from a pgx.Rows iterator into a models.Survey.
func ScanSurveyRows(rows pgx.Rows, s *models.Survey) error {
	return rows.Scan(
		&s.ID, &s.OrgID, &s.CreatedBy, &s.Title, &s.Description, &s.Status, &s.UIMode,
		&s.Structure, &s.Settings, &s.Theme, &s.ResponseCount, &s.ViewCount,
		&s.PublishedAt, &s.ClosedAt, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt,
	)
}

// ScanResponse scans a full response row from pgx into a models.Response.
func ScanResponse(row pgx.Row, r *models.Response) error {
	return row.Scan(
		&r.ID, &r.SurveyID, &r.RespondentEmail, &r.RespondentName,
		&r.RespondentDevice, &r.RespondentBrowser, &r.RespondentOS, &r.RespondentCountry,
		&r.RespondentIPHash, &r.Status, &r.StartedAt, &r.CompletedAt, &r.DurationMs,
		&r.Source, &r.SourceDetail, &r.Language, &r.EmbeddedData, &r.QualityFlags,
		&r.AIAnalysis, &r.CreatedAt,
	)
}

// ScanResponseRows scans the current row from a pgx.Rows iterator into a models.Response.
func ScanResponseRows(rows pgx.Rows, r *models.Response) error {
	return rows.Scan(
		&r.ID, &r.SurveyID, &r.RespondentEmail, &r.RespondentName,
		&r.RespondentDevice, &r.RespondentBrowser, &r.RespondentOS, &r.RespondentCountry,
		&r.RespondentIPHash, &r.Status, &r.StartedAt, &r.CompletedAt, &r.DurationMs,
		&r.Source, &r.SourceDetail, &r.Language, &r.EmbeddedData, &r.QualityFlags,
		&r.AIAnalysis, &r.CreatedAt,
	)
}
