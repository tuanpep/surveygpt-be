package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/surveyflow/be/internal/models"
)

// analyticsRepo provides data access for analytics queries.
type analyticsRepo struct {
	pool *pgxpool.Pool
}

// NewAnalyticsRepo creates a new AnalyticsRepo.
func NewAnalyticsRepo(pool *pgxpool.Pool) AnalyticsRepo {
	return &analyticsRepo{pool: pool}
}

// GetSummary returns analytics summary for a survey.
func (r *analyticsRepo) GetSummary(ctx context.Context, surveyID string) (*models.AnalyticsSummary, error) {
	summary := &models.AnalyticsSummary{}

	// Total responses.
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM responses WHERE survey_id = $1`, surveyID,
	).Scan(&summary.TotalResponses); err != nil {
		return nil, err
	}

	// Completed responses.
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM responses WHERE survey_id = $1 AND status = 'completed'`, surveyID,
	).Scan(&summary.CompletedResponses); err != nil {
		return nil, err
	}

	// Completion rate.
	if summary.TotalResponses > 0 {
		summary.CompletionRate = float64(summary.CompletedResponses) / float64(summary.TotalResponses)
	}

	// Average duration.
	var avgDuration *float64
	if err := r.pool.QueryRow(ctx,
		`SELECT AVG(duration_ms) FROM responses WHERE survey_id = $1 AND status = 'completed'`, surveyID,
	).Scan(&avgDuration); err != nil {
		return nil, err
	}
	if avgDuration != nil {
		summary.AvgDurationMs = int(*avgDuration)
	}

	// Total views (from survey view_count).
	if err := r.pool.QueryRow(ctx,
		`SELECT view_count FROM surveys WHERE id = $1`, surveyID,
	).Scan(&summary.TotalViews); err != nil {
		return nil, err
	}

	// Conversion rate.
	if summary.TotalViews > 0 {
		summary.ConversionRate = float64(summary.TotalResponses) / float64(summary.TotalViews)
	}

	// First/last response dates.
	if err := r.pool.QueryRow(ctx,
		`SELECT MIN(created_at), MAX(created_at) FROM responses WHERE survey_id = $1`, surveyID,
	).Scan(&summary.FirstResponseAt, &summary.LastResponseAt); err != nil {
		return nil, err
	}

	// Daily counts (last 30 days).
	dailyCounts, _ := r.GetResponsesByDay(ctx, surveyID, 30)
	summary.DailyCounts = dailyCounts

	// Device breakdown.
	deviceBreakdown, _ := r.GetDeviceBreakdown(ctx, surveyID)
	summary.DeviceBreakdown = deviceBreakdown

	// Source breakdown.
	sourceBreakdown, _ := r.GetSourceBreakdown(ctx, surveyID)
	summary.SourceBreakdown = sourceBreakdown

	return summary, nil
}

// GetResponsesByDay returns daily response counts.
func (r *analyticsRepo) GetResponsesByDay(ctx context.Context, surveyID string, days int) ([]models.DailyCount, error) {
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

// GetDeviceBreakdown returns response counts by device type.
func (r *analyticsRepo) GetDeviceBreakdown(ctx context.Context, surveyID string) (map[string]int, error) {
	query := `
		SELECT COALESCE(respondent_device, 'unknown') as device, COUNT(*) as count
		FROM responses WHERE survey_id = $1
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

// GetSourceBreakdown returns response counts by source.
func (r *analyticsRepo) GetSourceBreakdown(ctx context.Context, surveyID string) (map[string]int, error) {
	query := `
		SELECT source, COUNT(*) as count
		FROM responses WHERE survey_id = $1
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

// GetQuestionStats returns stats for a specific question.
func (r *analyticsRepo) GetQuestionStats(ctx context.Context, surveyID, questionID, questionType string) (*models.QuestionStats, error) {
	stats := &models.QuestionStats{
		QuestionID: questionID,
	}

	// Get total answers and skipped count.
	var totalAnswers, totalResponses int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT a.response_id) FROM answers a
		 JOIN responses r ON a.response_id = r.id
		 WHERE r.survey_id = $1 AND a.question_id = $2`,
		surveyID, questionID,
	).Scan(&totalAnswers); err != nil {
		return nil, err
	}

	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM responses WHERE survey_id = $1 AND status = 'completed'`, surveyID,
	).Scan(&totalResponses); err != nil {
		return nil, err
	}

	stats.TotalAnswers = totalAnswers
	stats.SkippedCount = totalResponses - totalAnswers

	// Get choice distribution for choice-type questions.
	if questionType == "multiple_choice" || questionType == "multi_select" || questionType == "dropdown" || questionType == "yes_no" {
		choiceStats := r.getChoiceStats(ctx, surveyID, questionID, totalAnswers)
		stats.ChoiceStats = choiceStats
	}

	// Get numeric stats for rating-type questions.
	if questionType == "rating_likert" || questionType == "rating_star" || questionType == "rating_nps" || questionType == "slider" {
		numericStats := r.getNumericStats(ctx, surveyID, questionID)
		stats.NumericStats = numericStats
	}

	return stats, nil
}

func (r *analyticsRepo) getChoiceStats(ctx context.Context, surveyID, questionID string, totalAnswers int) []models.ChoiceStat {
	query := `
		SELECT a.value, COUNT(*) as count
		FROM answers a
		JOIN responses r ON a.response_id = r.id
		WHERE r.survey_id = $1 AND a.question_id = $2
		GROUP BY a.value
		ORDER BY count DESC`

	rows, err := r.pool.Query(ctx, query, surveyID, questionID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var stats []models.ChoiceStat
	for rows.Next() {
		var valueJSON []byte
		var count int
		if err := rows.Scan(&valueJSON, &count); err != nil {
			continue
		}

		// Parse the value.
		var value any
		if err := json.Unmarshal(valueJSON, &value); err != nil {
			continue
		}

		// Handle both single values and arrays.
		values := []string{}
		switch v := value.(type) {
		case string:
			values = append(values, v)
		case []any:
			for _, item := range v {
				if s, ok := item.(string); ok {
					values = append(values, s)
				}
			}
		case float64:
			values = append(values, string(rune(int(v))))
		}

		for _, v := range values {
			percentage := 0.0
			if totalAnswers > 0 {
				percentage = float64(count) / float64(totalAnswers) * 100
			}
			stats = append(stats, models.ChoiceStat{
				ChoiceID:   v,
				Label:      v,
				Count:      count,
				Percentage: percentage,
			})
		}
	}
	return stats
}

func (r *analyticsRepo) getNumericStats(ctx context.Context, surveyID, questionID string) *models.NumericStats {
	query := `
		SELECT
			AVG(CAST(a.value AS float)) as mean,
			MIN(CAST(a.value AS float)) as min,
			MAX(CAST(a.value AS float)) as max
		FROM answers a
		JOIN responses r ON a.response_id = r.id
		WHERE r.survey_id = $1 AND a.question_id = $2
			AND a.value ? '$number' OR json_typeof(a.value) = 'number'`

	stats := &models.NumericStats{}
	var mean, min, max *float64
	err := r.pool.QueryRow(ctx, query, surveyID, questionID).Scan(&mean, &min, &max)
	if err != nil {
		return stats
	}

	if mean != nil {
		stats.Mean = *mean
	}
	if min != nil {
		stats.Min = *min
	}
	if max != nil {
		stats.Max = *max
	}

	return stats
}

// GetCrossTab returns cross-tabulation between two questions.
func (r *analyticsRepo) GetCrossTab(ctx context.Context, surveyID, rowQuestionID, colQuestionID string) (*models.CrossTabResult, error) {
	// Get all unique values for row question.
	rowQuery := `
		SELECT DISTINCT a.value
		FROM answers a
		JOIN responses r ON a.response_id = r.id
		WHERE r.survey_id = $1 AND a.question_id = $2`

	rowRows, err := r.pool.Query(ctx, rowQuery, surveyID, rowQuestionID)
	if err != nil {
		return nil, err
	}
	defer rowRows.Close()

	var rowLabels []string
	for rowRows.Next() {
		var valueJSON []byte
		if err := rowRows.Scan(&valueJSON); err != nil {
			continue
		}
		var value any
		if json.Unmarshal(valueJSON, &value) == nil {
			if s, ok := value.(string); ok {
				rowLabels = append(rowLabels, s)
			}
		}
	}

	// Get all unique values for column question.
	colQuery := `
		SELECT DISTINCT a.value
		FROM answers a
		JOIN responses r ON a.response_id = r.id
		WHERE r.survey_id = $1 AND a.question_id = $2`

	colRows, err := r.pool.Query(ctx, colQuery, surveyID, colQuestionID)
	if err != nil {
		return nil, err
	}
	defer colRows.Close()

	var colLabels []string
	for colRows.Next() {
		var valueJSON []byte
		if err := colRows.Scan(&valueJSON); err != nil {
			continue
		}
		var value any
		if json.Unmarshal(valueJSON, &value) == nil {
			if s, ok := value.(string); ok {
				colLabels = append(colLabels, s)
			}
		}
	}

	// Get cross-tab counts.
	crossTabQuery := `
		SELECT a1.value as row_val, a2.value as col_val, COUNT(*) as count
		FROM answers a1
		JOIN answers a2 ON a1.response_id = a2.response_id
		JOIN responses r ON a1.response_id = r.id
		WHERE r.survey_id = $1 AND a1.question_id = $2 AND a2.question_id = $3
		GROUP BY a1.value, a2.value`

	crossRows, err := r.pool.Query(ctx, crossTabQuery, surveyID, rowQuestionID, colQuestionID)
	if err != nil {
		return nil, err
	}
	defer crossRows.Close()

	counts := make(map[string]map[string]int)
	for crossRows.Next() {
		var rowValJSON, colValJSON []byte
		var count int
		if err := crossRows.Scan(&rowValJSON, &colValJSON, &count); err != nil {
			continue
		}

		var rowVal, colVal any
		json.Unmarshal(rowValJSON, &rowVal)
		json.Unmarshal(colValJSON, &colVal)

		rowStr, _ := rowVal.(string)
		colStr, _ := colVal.(string)

		if counts[rowStr] == nil {
			counts[rowStr] = make(map[string]int)
		}
		counts[rowStr][colStr] = count
	}

	// Build result.
	result := &models.CrossTabResult{
		RowQuestionID:    rowQuestionID,
		ColumnQuestionID: colQuestionID,
		Headers:          colLabels,
		Rows:             []models.CrossTabRow{},
	}

	for _, rowLabel := range rowLabels {
		row := models.CrossTabRow{
			Label:  rowLabel,
			Values: []int{},
			Total:  0,
		}
		for _, colLabel := range colLabels {
			count := counts[rowLabel][colLabel]
			row.Values = append(row.Values, count)
			row.Total += count
		}
		result.Rows = append(result.Rows, row)
	}

	return result, nil
}

// GetDropoff returns dropoff analysis for a survey.
func (r *analyticsRepo) GetDropoff(ctx context.Context, surveyID string) ([]models.DropoffStep, error) {
	// This is a simplified implementation - in production would need to track per-question views.
	// For now, return empty result.
	return []models.DropoffStep{}, nil
}
