package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/surveyflow/be/internal/middleware"
	apperr "github.com/surveyflow/be/internal/pkg/errors"
	"github.com/surveyflow/be/internal/pkg/httputil"
	"github.com/surveyflow/be/internal/services"
)

// SurveyHandler handles survey HTTP requests.
type SurveyHandler struct {
	surveySvc *services.SurveyService
}

// NewSurveyHandler creates a new SurveyHandler.
func NewSurveyHandler(surveySvc *services.SurveyService) *SurveyHandler {
	return &SurveyHandler{surveySvc: surveySvc}
}

// Create handles POST /surveys.
func (h *SurveyHandler) Create(c echo.Context) error {
	var input services.CreateSurveyInput
	if err := c.Bind(&input); err != nil {
		return apperr.BadRequest("invalid request body")
	}
	if input.Title == "" {
		return apperr.BadRequest("title is required")
	}

	orgID := middleware.GetOrgIDFromContext(c)
	userID := middleware.GetUserIDFromContext(c)

	survey, err := h.surveySvc.CreateSurvey(c.Request().Context(), orgID, userID, input)
	if err != nil {
		return err
	}

	return httputil.Created(c, survey)
}

// List handles GET /surveys.
func (h *SurveyHandler) List(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)

	filter := services.ListSurveysFilter{
		Status:  c.QueryParam("status"),
		Search:  c.QueryParam("search"),
		SortBy:  c.QueryParam("sort_by"),
		SortDir: c.QueryParam("sort_dir"),
		Cursor:  c.QueryParam("cursor"),
		Limit:   20,
	}

	surveys, nextCursor, total, err := h.surveySvc.ListSurveys(c.Request().Context(), orgID, filter)
	if err != nil {
		return err
	}

	return httputil.Paginated(c, surveys, nextCursor, total)
}

// Get handles GET /surveys/:id.
func (h *SurveyHandler) Get(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")

	survey, err := h.surveySvc.GetSurvey(c.Request().Context(), orgID, surveyID)
	if err != nil {
		return err
	}

	return httputil.OK(c, survey)
}

// Update handles PUT /surveys/:id.
func (h *SurveyHandler) Update(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")

	var input services.UpdateSurveyInput
	if err := c.Bind(&input); err != nil {
		return apperr.BadRequest("invalid request body")
	}

	survey, err := h.surveySvc.UpdateSurvey(c.Request().Context(), orgID, surveyID, input)
	if err != nil {
		return err
	}

	return httputil.OK(c, survey)
}

// Delete handles DELETE /surveys/:id.
func (h *SurveyHandler) Delete(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")

	if err := h.surveySvc.DeleteSurvey(c.Request().Context(), orgID, surveyID); err != nil {
		return err
	}

	return httputil.Message(c, "survey deleted successfully")
}

// Publish handles POST /surveys/:id/publish.
func (h *SurveyHandler) Publish(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")

	survey, err := h.surveySvc.PublishSurvey(c.Request().Context(), orgID, surveyID)
	if err != nil {
		return err
	}

	return httputil.OK(c, survey)
}

// Close handles POST /surveys/:id/close.
func (h *SurveyHandler) Close(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")

	survey, err := h.surveySvc.CloseSurvey(c.Request().Context(), orgID, surveyID)
	if err != nil {
		return err
	}

	return httputil.OK(c, survey)
}

// Duplicate handles POST /surveys/:id/duplicate.
func (h *SurveyHandler) Duplicate(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")

	survey, err := h.surveySvc.DuplicateSurvey(c.Request().Context(), orgID, surveyID)
	if err != nil {
		return err
	}

	return httputil.Created(c, survey)
}

// GetPublic handles GET /surveys/:slug (public, no auth).
func (h *SurveyHandler) GetPublic(c echo.Context) error {
	surveyID := c.Param("slug")

	appURL := c.Scheme() + "://" + c.Request().Host
	public, err := h.surveySvc.GetPublicSurvey(c.Request().Context(), surveyID, appURL)
	if err != nil {
		return err
	}

	return httputil.OK(c, public)
}
