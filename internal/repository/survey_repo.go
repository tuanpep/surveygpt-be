package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/surveyflow/be/internal/models"
	"github.com/surveyflow/be/internal/pkg/dbutil"
	"github.com/surveyflow/be/internal/pkg/pagination"
)

// allowedSortColumns maps user-provided sort keys to safe SQL column expressions.
var allowedSortColumns = map[string]string{
	"created_at":     "s.created_at",
	"updated_at":     "s.updated_at",
	"title":          "s.title",
	"status":         "s.status",
	"response_count": "s.response_count",
}

// surveyRepo provides data access for surveys.
type surveyRepo struct {
	pool *pgxpool.Pool
}

// NewSurveyRepo creates a new SurveyRepo.
func NewSurveyRepo(pool *pgxpool.Pool) SurveyRepo {
	return &surveyRepo{pool: pool}
}

// Create inserts a new survey.
func (r *surveyRepo) Create(ctx context.Context, s *models.Survey) error {
	structJSON := dbutil.MustMarshal(s.Structure)
	settingsJSON := dbutil.MustMarshal(s.Settings)
	themeJSON := dbutil.MustMarshal(s.Theme)

	query := `
		INSERT INTO surveys (org_id, created_by, title, description, status, ui_mode,
			structure, settings, theme)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, response_count, view_count, published_at, closed_at, deleted_at, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		s.OrgID, s.CreatedBy, s.Title, s.Description, s.Status, s.UIMode,
		structJSON, settingsJSON, themeJSON,
	).Scan(&s.ID, &s.ResponseCount, &s.ViewCount, &s.PublishedAt, &s.ClosedAt, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt)
}

// GetByID fetches a survey by ID (excludes soft-deleted).
func (r *surveyRepo) GetByID(ctx context.Context, id string) (*models.Survey, error) {
	query := `
		SELECT id, org_id, created_by, title, description, status, ui_mode,
			structure, settings, theme, response_count, view_count,
			published_at, closed_at, deleted_at, created_at, updated_at
		FROM surveys WHERE id = $1 AND deleted_at IS NULL`

	s := &models.Survey{}
	if err := dbutil.ScanSurvey(r.pool.QueryRow(ctx, query, id), s); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

// List returns surveys for an organization with filtering and pagination.
// Filters: status, search (title match), sortBy (created_at, updated_at, title, status, response_count), sortOrder (asc, desc).
func (r *surveyRepo) List(ctx context.Context, orgID, status, search, sortBy, sortOrder string, limit int, cursor string) ([]models.Survey, string, int, error) {
	q := dbutil.NewQuery("org_id = $1 AND deleted_at IS NULL", orgID, 2)

	if status != "" {
		q.AndEqual("status", status)
	}
	if search != "" {
		q.AndLike("title", search)
	}

	// Validate sort column against whitelist to prevent SQL injection.
	orderCol := "s.created_at"
	if sortBy != "" {
		if col, ok := allowedSortColumns[sortBy]; ok {
			orderCol = col
		} else {
			return nil, "", 0, fmt.Errorf("invalid sort column: %s", sortBy)
		}
	}
	orderDir := "DESC"
	if sortOrder == "asc" {
		orderDir = "ASC"
	}

	// Cursor — only supported when sorting by created_at or updated_at.
	if cursor != "" {
		_, cursorTime, err := pagination.DecodeCursor(cursor)
		if err == nil && cursorTime > 0 {
			if sortBy == "" || sortBy == "created_at" || sortBy == "updated_at" {
				q.AndCursor(orderCol, time.Unix(0, cursorTime))
			}
		}
	}

	orderClause := orderCol + " " + orderDir
	if orderCol != "s.created_at" {
		orderClause += ", s.created_at " + orderDir
	}

	limitPlaceholder := q.AppendLimit(limit)

	query := `
		SELECT id, org_id, created_by, title, description, status, ui_mode,
			structure, settings, theme, response_count, view_count,
			published_at, closed_at, deleted_at, created_at, updated_at
		FROM surveys s WHERE ` + q.Where() + `
		ORDER BY ` + orderClause + `
		LIMIT ` + limitPlaceholder

	rows, err := r.pool.Query(ctx, query, q.Args()...)
	if err != nil {
		return nil, "", 0, err
	}
	defer rows.Close()

	var surveys []models.Survey
	for rows.Next() {
		var s models.Survey
		if err := dbutil.ScanSurveyRows(rows, &s); err != nil {
			return nil, "", 0, err
		}
		surveys = append(surveys, s)
	}

	// Get total count for pagination.
	countQuery := `SELECT COUNT(*) FROM surveys WHERE ` + q.Where()
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, q.ArgsWithoutLast()...).Scan(&total); err != nil {
		return nil, "", 0, err
	}

	// Build next cursor.
	nextCursor := ""
	if len(surveys) == limit {
		last := surveys[len(surveys)-1]
		nextCursor, _ = pagination.EncodeCursor(last.ID.String(), last.CreatedAt.UnixNano())
	}

	if surveys == nil {
		surveys = []models.Survey{}
	}
	return surveys, nextCursor, total, nil
}

// Update applies changes to a survey.
func (r *surveyRepo) Update(ctx context.Context, s *models.Survey) error {
	structJSON := dbutil.MustMarshal(s.Structure)
	settingsJSON := dbutil.MustMarshal(s.Settings)
	themeJSON := dbutil.MustMarshal(s.Theme)

	query := `
		UPDATE surveys SET title = $2, description = $3, status = $4, ui_mode = $5,
			structure = $6, settings = $7, theme = $8, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at`

	return r.pool.QueryRow(ctx, query,
		s.ID, s.Title, s.Description, s.Status, s.UIMode,
		structJSON, settingsJSON, themeJSON,
	).Scan(&s.UpdatedAt)
}

// UpdateStatus changes the status of a survey.
func (r *surveyRepo) UpdateStatus(ctx context.Context, id string, status models.SurveyStatus) error {
	switch status {
	case models.SurveyStatusPublished:
		_, err := r.pool.Exec(ctx,
			`UPDATE surveys SET status = $2, published_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
			id, status,
		)
		return err
	default:
		_, err := r.pool.Exec(ctx,
			`UPDATE surveys SET status = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
			id, status,
		)
		return err
	}
}

// SoftDelete marks a survey as deleted.
func (r *surveyRepo) SoftDelete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE surveys SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	return err
}

// Duplicate creates a copy of a survey with a new title.
func (r *surveyRepo) Duplicate(ctx context.Context, id, newTitle string) (*models.Survey, error) {
	query := `
		INSERT INTO surveys (org_id, created_by, title, description, status, ui_mode,
			structure, settings, theme)
		SELECT org_id, created_by, $2, description, 'draft', ui_mode,
			structure, settings, theme
		FROM surveys WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, created_at, updated_at`

	s := &models.Survey{}
	err := r.pool.QueryRow(ctx, query, id, newTitle).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	s.Title = newTitle
	s.Status = models.SurveyStatusDraft
	return s, nil
}

// IncrementViewCount increases the view counter.
func (r *surveyRepo) IncrementViewCount(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE surveys SET view_count = view_count + 1, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	return err
}

// IncrementResponseCount increases the response counter.
func (r *surveyRepo) IncrementResponseCount(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE surveys SET response_count = response_count + 1, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	return err
}

// CountByOrg returns the number of non-deleted surveys for an org.
func (r *surveyRepo) CountByOrg(ctx context.Context, orgID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM surveys WHERE org_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&count)
	return count, err
}
