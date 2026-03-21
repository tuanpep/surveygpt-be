package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/surveyflow/be/internal/models"
	"github.com/surveyflow/be/internal/pkg/dbutil"
	"github.com/surveyflow/be/internal/pkg/pagination"
)

// responseRepo provides data access for responses and answers.
type responseRepo struct {
	pool *pgxpool.Pool
}

// NewResponseRepo creates a new ResponseRepo.
func NewResponseRepo(pool *pgxpool.Pool) ResponseRepo {
	return &responseRepo{pool: pool}
}

// Create inserts a new response.
func (r *responseRepo) Create(ctx context.Context, resp *models.Response) error {
	embeddedJSON, err := json.Marshal(resp.EmbeddedData)
	if err != nil {
		return fmt.Errorf("marshal embedded data: %w", err)
	}
	qualityJSON, err := json.Marshal(resp.QualityFlags)
	if err != nil {
		return fmt.Errorf("marshal quality flags: %w", err)
	}

	query := `
		INSERT INTO responses (survey_id, respondent_email, respondent_name,
			respondent_device, respondent_browser, respondent_os, respondent_country,
			respondent_ip_hash, status, source, source_detail, language, embedded_data, quality_flags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, started_at, created_at`

	return r.pool.QueryRow(ctx, query,
		resp.SurveyID, resp.RespondentEmail, resp.RespondentName,
		resp.RespondentDevice, resp.RespondentBrowser, resp.RespondentOS, resp.RespondentCountry,
		resp.RespondentIPHash, resp.Status, resp.Source, resp.SourceDetail, resp.Language,
		embeddedJSON, qualityJSON,
	).Scan(&resp.ID, &resp.StartedAt, &resp.CreatedAt)
}

// GetByID fetches a response by ID.
func (r *responseRepo) GetByID(ctx context.Context, id string) (*models.Response, error) {
	query := `
		SELECT id, survey_id, respondent_email, respondent_name,
			respondent_device, respondent_browser, respondent_os, respondent_country,
			respondent_ip_hash, status, started_at, completed_at, duration_ms,
			source, source_detail, language, embedded_data, quality_flags, ai_analysis, created_at
		FROM responses WHERE id = $1`

	resp := &models.Response{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&resp.ID, &resp.SurveyID, &resp.RespondentEmail, &resp.RespondentName,
		&resp.RespondentDevice, &resp.RespondentBrowser, &resp.RespondentOS, &resp.RespondentCountry,
		&resp.RespondentIPHash, &resp.Status, &resp.StartedAt, &resp.CompletedAt, &resp.DurationMs,
		&resp.Source, &resp.SourceDetail, &resp.Language, &resp.EmbeddedData, &resp.QualityFlags,
		&resp.AIAnalysis, &resp.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// List returns responses for a survey with filtering and pagination.
func (r *responseRepo) List(ctx context.Context, surveyID, status string, dateFrom, dateTo *time.Time, limit int, cursor string) ([]models.Response, string, int, error) {
	q := dbutil.NewQuery("survey_id = $1", surveyID, 2)

	if status != "" {
		q.AndEqual("status", status)
	}
	if dateFrom != nil {
		q.AndGreaterEqual("created_at", dateFrom)
	}
	if dateTo != nil {
		q.AndLessEqual("created_at", dateTo)
	}
	if cursor != "" {
		_, cursorTime, err := pagination.DecodeCursor(cursor)
		if err == nil && cursorTime > 0 {
			q.AndLessThan("created_at", time.Unix(0, cursorTime))
		}
	}

	limitPlaceholder := q.AppendLimit(limit)

	query := `
		SELECT id, survey_id, respondent_email, respondent_name,
			respondent_device, respondent_browser, respondent_os, respondent_country,
			respondent_ip_hash, status, started_at, completed_at, duration_ms,
			source, source_detail, language, embedded_data, quality_flags, ai_analysis, created_at
		FROM responses WHERE ` + q.Where() + `
		ORDER BY created_at DESC
		LIMIT ` + limitPlaceholder

	rows, err := r.pool.Query(ctx, query, q.Args()...)
	if err != nil {
		return nil, "", 0, err
	}
	defer rows.Close()

	var responses []models.Response
	for rows.Next() {
		var resp models.Response
		if err := dbutil.ScanResponseRows(rows, &resp); err != nil {
			return nil, "", 0, err
		}
		responses = append(responses, resp)
	}

	// Get total count.
	countQuery := `SELECT COUNT(*) FROM responses WHERE ` + q.Where()
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, q.ArgsWithoutLast()...).Scan(&total); err != nil {
		return nil, "", 0, err
	}

	// Build next cursor.
	nextCursor := ""
	if len(responses) == limit {
		last := responses[len(responses)-1]
		nextCursor, _ = pagination.EncodeCursor(last.ID.String(), last.CreatedAt.UnixNano())
	}

	if responses == nil {
		responses = []models.Response{}
	}
	return responses, nextCursor, total, nil
}

// Complete marks a response as completed.
func (r *responseRepo) Complete(ctx context.Context, id string, durationMs int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE responses SET status = 'completed', completed_at = NOW(), duration_ms = $2 WHERE id = $1`,
		id, durationMs,
	)
	return err
}

// Delete removes a response.
func (r *responseRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM responses WHERE id = $1`, id)
	return err
}

// CountBySurvey returns the number of responses for a survey.
func (r *responseRepo) CountBySurvey(ctx context.Context, surveyID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM responses WHERE survey_id = $1`, surveyID,
	).Scan(&count)
	return count, err
}

// CountCompletedBySurvey returns completed responses count.
func (r *responseRepo) CountCompletedBySurvey(ctx context.Context, surveyID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM responses WHERE survey_id = $1 AND status = 'completed'`, surveyID,
	).Scan(&count)
	return count, err
}

// GetAvgDuration returns average duration in ms.
func (r *responseRepo) GetAvgDuration(ctx context.Context, surveyID string) (int, error) {
	var avg *float64
	err := r.pool.QueryRow(ctx,
		`SELECT AVG(duration_ms) FROM responses WHERE survey_id = $1 AND status = 'completed'`, surveyID,
	).Scan(&avg)
	if err != nil || avg == nil {
		return 0, err
	}
	return int(*avg), nil
}

// GetResponsesByDay returns daily response counts.
func (r *responseRepo) GetResponsesByDay(ctx context.Context, surveyID string, days int) ([]models.DailyCount, error) {
	query := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM responses
		WHERE survey_id = $1 AND created_at >= NOW() - INTERVAL '1 day' * $2
		GROUP BY DATE(created_at)
		ORDER BY date`

	rows, err := r.pool.Query(ctx, query, surveyID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []models.DailyCount
	for rows.Next() {
		var dc models.DailyCount
		if err := rows.Scan(&dc.Date, &dc.Count); err != nil {
			return nil, err
		}
		counts = append(counts, dc)
	}
	return counts, nil
}

// GetResponsesBySource returns counts by source.
func (r *responseRepo) GetResponsesBySource(ctx context.Context, surveyID string) (map[string]int, error) {
	query := `
		SELECT source, COUNT(*) as count
		FROM responses
		WHERE survey_id = $1
		GROUP BY source`

	rows, err := r.pool.Query(ctx, query, surveyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err != nil {
			return nil, err
		}
		result[source] = count
	}
	return result, nil
}

// GetResponsesByDevice returns counts by device type.
func (r *responseRepo) GetResponsesByDevice(ctx context.Context, surveyID string) (map[string]int, error) {
	query := `
		SELECT COALESCE(respondent_device, 'unknown') as device, COUNT(*) as count
		FROM responses
		WHERE survey_id = $1
		GROUP BY respondent_device`

	rows, err := r.pool.Query(ctx, query, surveyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var device string
		var count int
		if err := rows.Scan(&device, &count); err != nil {
			return nil, err
		}
		result[device] = count
	}
	return result, nil
}

// --- Answers ---

// CreateAnswers batch inserts answers for a response.
func (r *responseRepo) CreateAnswers(ctx context.Context, answers []models.Answer) error {
	if len(answers) == 0 {
		return nil
	}

	query := `
		INSERT INTO answers (response_id, question_id, value, metadata)
		VALUES ($1, $2, $3, $4)`

	batch := &pgx.Batch{}
	for _, a := range answers {
		valueJSON, err := json.Marshal(a.Value)
		if err != nil {
			return fmt.Errorf("marshal answer value: %w", err)
		}
		metadataJSON, err := json.Marshal(a.Metadata)
		if err != nil {
			return fmt.Errorf("marshal answer metadata: %w", err)
		}
		batch.Queue(query, a.ResponseID, a.QuestionID, valueJSON, metadataJSON)
	}

	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()
	for i := 0; i < len(answers); i++ {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}
	return nil
}

// GetAnswersByResponseID fetches all answers for a response.
func (r *responseRepo) GetAnswersByResponseID(ctx context.Context, responseID string) ([]models.Answer, error) {
	query := `
		SELECT id, response_id, question_id, value, metadata, created_at
		FROM answers WHERE response_id = $1
		ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, responseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []models.Answer
	for rows.Next() {
		var a models.Answer
		if err := rows.Scan(&a.ID, &a.ResponseID, &a.QuestionID, &a.Value, &a.Metadata, &a.CreatedAt); err != nil {
			return nil, err
		}
		answers = append(answers, a)
	}
	return answers, nil
}

// GetAnswerCountsByQuestion returns answer counts grouped by question.
func (r *responseRepo) GetAnswerCountsByQuestion(ctx context.Context, surveyID, questionID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM answers a
		 JOIN responses r ON a.response_id = r.id
		 WHERE r.survey_id = $1 AND a.question_id = $2`,
		surveyID, questionID,
	).Scan(&count)
	return count, err
}

// GetChoiceDistribution returns choice distribution for a question.
func (r *responseRepo) GetChoiceDistribution(ctx context.Context, surveyID, questionID string) (map[string]int, error) {
	query := `
		SELECT a.value, COUNT(*) as count
		FROM answers a
		JOIN responses r ON a.response_id = r.id
		WHERE r.survey_id = $1 AND a.question_id = $2
		GROUP BY a.value`

	rows, err := r.pool.Query(ctx, query, surveyID, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var valueJSON []byte
		var count int
		if err := rows.Scan(&valueJSON, &count); err != nil {
			return nil, err
		}

		// Parse the value - could be string or array
		var value any
		if err := json.Unmarshal(valueJSON, &value); err != nil {
			continue
		}

		switch v := value.(type) {
		case string:
			result[v] = count
		case []any:
			for _, item := range v {
				if s, ok := item.(string); ok {
					result[s] += count
				}
			}
		}
	}
	return result, nil
}
