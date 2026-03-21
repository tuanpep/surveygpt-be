package repository

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/surveyflow/be/internal/models"
	"github.com/surveyflow/be/internal/pkg/dbutil"
)

// templateRepo provides data access for templates.
type templateRepo struct {
	pool *pgxpool.Pool
}

// NewTemplateRepo creates a new TemplateRepo.
func NewTemplateRepo(pool *pgxpool.Pool) TemplateRepo {
	return &templateRepo{pool: pool}
}

// List returns templates, optionally filtered by category or org.
func (r *templateRepo) List(ctx context.Context, category string, orgID string) ([]models.Template, error) {
	query := `
		SELECT id, org_id, category, title, description, tags, structure, theme,
			cover_image_url, is_featured, use_count, created_at, updated_at
		FROM templates WHERE 1=1`
	args := []any{}
	argN := 1

	if category != "" {
		query += ` AND category = $` + strconv.Itoa(argN)
		args = append(args, category)
		argN++
	}
	if orgID != "" {
		query += ` AND org_id = $` + strconv.Itoa(argN)
		args = append(args, orgID)
		argN++
	} else {
		// System templates (org_id IS NULL) plus any from org.
		query += ` AND (org_id IS NULL OR org_id = $` + strconv.Itoa(argN) + `)`
		args = append(args, orgID)
		argN++
	}

	query += ` ORDER BY is_featured DESC, use_count DESC, created_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.Template
	for rows.Next() {
		var t models.Template
		if err := rows.Scan(
			&t.ID, &t.OrgID, &t.Category, &t.Title, &t.Description, &t.Tags,
			&t.Structure, &t.Theme, &t.CoverImageURL, &t.IsFeatured, &t.UseCount,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}

	if templates == nil {
		templates = []models.Template{}
	}
	return templates, nil
}

// GetByID fetches a template by ID.
func (r *templateRepo) GetByID(ctx context.Context, id string) (*models.Template, error) {
	query := `
		SELECT id, org_id, category, title, description, tags, structure, theme,
			cover_image_url, is_featured, use_count, created_at, updated_at
		FROM templates WHERE id = $1`

	t := &models.Template{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.OrgID, &t.Category, &t.Title, &t.Description, &t.Tags,
		&t.Structure, &t.Theme, &t.CoverImageURL, &t.IsFeatured, &t.UseCount,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

// Create inserts a new template.
func (r *templateRepo) Create(ctx context.Context, t *models.Template) error {
	structureJSON := dbutil.MustMarshal(t.Structure)
	themeJSON := dbutil.MustMarshal(t.Theme)

	query := `
		INSERT INTO templates (org_id, category, title, description, tags, structure, theme, cover_image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, use_count, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		t.OrgID, t.Category, t.Title, t.Description, t.Tags,
		structureJSON, themeJSON, t.CoverImageURL,
	).Scan(&t.ID, &t.UseCount, &t.CreatedAt, &t.UpdatedAt)
}

// IncrementUseCount increases the use count for a template.
func (r *templateRepo) IncrementUseCount(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE templates SET use_count = use_count + 1, updated_at = NOW() WHERE id = $1`,
		id,
	)
	return err
}

// Template is a placeholder for the model — we use the one from models package.
// This alias is needed for the org field type.
var _ = models.Template{}
