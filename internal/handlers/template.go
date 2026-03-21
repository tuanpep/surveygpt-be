package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/surveyflow/be/internal/middleware"
	apperr "github.com/surveyflow/be/internal/pkg/errors"
	"github.com/surveyflow/be/internal/pkg/httputil"
	"github.com/surveyflow/be/internal/services"
)

// TemplateHandler handles template HTTP requests.
type TemplateHandler struct {
	templateSvc *services.SurveyService
}

// NewTemplateHandler creates a new TemplateHandler.
func NewTemplateHandler(templateSvc *services.SurveyService) *TemplateHandler {
	return &TemplateHandler{templateSvc: templateSvc}
}

// List handles GET /templates.
func (h *TemplateHandler) List(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	category := c.QueryParam("category")

	templates, err := h.templateSvc.ListTemplates(c.Request().Context(), category, orgID)
	if err != nil {
		return err
	}

	return httputil.OK(c, templates)
}

// Get handles GET /templates/:id.
func (h *TemplateHandler) Get(c echo.Context) error {
	templateID := c.Param("id")

	tmpl, err := h.templateSvc.GetTemplateByID(c.Request().Context(), templateID)
	if err != nil {
		return err
	}
	if tmpl == nil {
		return apperr.NotFound("template")
	}

	return httputil.OK(c, tmpl)
}

// Create handles POST /templates (save survey as template).
func (h *TemplateHandler) Create(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)

	var body struct {
		SurveyID string `json:"survey_id" validate:"required"`
		Name     string `json:"name" validate:"required"`
		Category string `json:"category"`
	}
	if err := c.Bind(&body); err != nil {
		return apperr.BadRequest("invalid request body")
	}
	if body.Name == "" || body.SurveyID == "" {
		return apperr.BadRequest("survey_id and name are required")
	}
	if body.Category == "" {
		body.Category = "custom"
	}

	tmpl, err := h.templateSvc.SaveAsTemplate(c.Request().Context(), orgID, body.SurveyID, body.Name, body.Category)
	if err != nil {
		return err
	}

	return httputil.Created(c, tmpl)
}

// CreateFromTemplate handles POST /surveys/:id/from-template.
func (h *TemplateHandler) CreateFromTemplate(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	userID := middleware.GetUserIDFromContext(c)
	templateID := c.Param("id")

	survey, err := h.templateSvc.CreateFromTemplate(c.Request().Context(), orgID, userID, templateID)
	if err != nil {
		return err
	}

	return httputil.Created(c, survey)
}

// Duplicate handles POST /templates/:id/duplicate.
func (h *TemplateHandler) Duplicate(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	userID := middleware.GetUserIDFromContext(c)
	templateID := c.Param("id")

	survey, err := h.templateSvc.CreateFromTemplate(c.Request().Context(), orgID, userID, templateID)
	if err != nil {
		return err
	}

	return httputil.Created(c, survey)
}
