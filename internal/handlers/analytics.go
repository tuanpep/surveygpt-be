package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/surveyflow/be/internal/middleware"
	apperr "github.com/surveyflow/be/internal/pkg/errors"
	"github.com/surveyflow/be/internal/pkg/httputil"
	"github.com/surveyflow/be/internal/services"
)

// AnalyticsHandler handles analytics HTTP requests.
type AnalyticsHandler struct {
	analyticsSvc *services.AnalyticsService
}

// NewAnalyticsHandler creates a new AnalyticsHandler.
func NewAnalyticsHandler(analyticsSvc *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsSvc: analyticsSvc}
}

// GetSummary handles GET /surveys/:id/analytics/summary.
func (h *AnalyticsHandler) GetSummary(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")

	summary, err := h.analyticsSvc.GetSummary(c.Request().Context(), orgID, surveyID)
	if err != nil {
		return err
	}
	return httputil.OK(c, summary)
}

// GetQuestionStats handles GET /surveys/:id/analytics/questions/:questionId.
func (h *AnalyticsHandler) GetQuestionStats(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")
	questionID := c.Param("questionId")

	stats, err := h.analyticsSvc.GetQuestionStats(c.Request().Context(), orgID, surveyID, questionID)
	if err != nil {
		return err
	}
	return httputil.OK(c, stats)
}

// GetCrossTab handles GET /surveys/:id/analytics/cross-tab.
func (h *AnalyticsHandler) GetCrossTab(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")
	rowQuestionID := c.QueryParam("row_question_id")
	colQuestionID := c.QueryParam("col_question_id")

	if rowQuestionID == "" || colQuestionID == "" {
		return apperr.BadRequest("row_question_id and col_question_id are required")
	}

	result, err := h.analyticsSvc.GetCrossTab(c.Request().Context(), orgID, surveyID, rowQuestionID, colQuestionID)
	if err != nil {
		return err
	}

	return httputil.OK(c, result)
}

// GetDropoff handles GET /surveys/:id/analytics/dropoff.
func (h *AnalyticsHandler) GetDropoff(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")

	steps, err := h.analyticsSvc.GetDropoff(c.Request().Context(), orgID, surveyID)
	if err != nil {
		return err
	}
	return httputil.OK(c, echo.Map{"steps": steps})
}

// GetAIInsights handles POST /surveys/:id/analytics/ai-insights.
func (h *AnalyticsHandler) GetAIInsights(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")

	insights, err := h.analyticsSvc.GetAIInsights(c.Request().Context(), orgID, surveyID)
	if err != nil {
		return err
	}

	return httputil.OK(c, insights)
}
