package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/surveyflow/be/internal/middleware"
	apperr "github.com/surveyflow/be/internal/pkg/errors"
	"github.com/surveyflow/be/internal/pkg/httputil"
	"github.com/surveyflow/be/internal/services"
)

// DistributionHandler handles distribution HTTP requests.
type DistributionHandler struct {
	distributionSvc *services.DistributionService
}

// NewDistributionHandler creates a new DistributionHandler.
func NewDistributionHandler(distributionSvc *services.DistributionService) *DistributionHandler {
	return &DistributionHandler{distributionSvc: distributionSvc}
}

// GetQRCode handles GET /surveys/:id/qr-code.
func (h *DistributionHandler) GetQRCode(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")

	var input services.QRCodeInput
	input.SurveyID = surveyID
	if size := c.QueryParam("size"); size != "" {
		input.Size = 0 // Would parse in production
	}
	input.ErrorLevel = c.QueryParam("error_level")

	result, err := h.distributionSvc.GenerateQRCode(c.Request().Context(), orgID, surveyID, input)
	if err != nil {
		return err
	}

	return httputil.OK(c, result)
}

// GetEmbedCode handles GET /surveys/:id/embed-code.
func (h *DistributionHandler) GetEmbedCode(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")

	var input services.EmbedCodeInput
	input.SurveyID = surveyID
	input.Mode = c.QueryParam("mode")
	input.Trigger = c.QueryParam("trigger")
	input.Width = c.QueryParam("width")
	input.Height = c.QueryParam("height")

	result, err := h.distributionSvc.GetEmbedCode(c.Request().Context(), orgID, surveyID, input)
	if err != nil {
		return err
	}

	return httputil.OK(c, result)
}

// ListEmailLists handles GET /email-lists.
func (h *DistributionHandler) ListEmailLists(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)

	lists, err := h.distributionSvc.GetEmailLists(c.Request().Context(), orgID)
	if err != nil {
		return err
	}

	return httputil.OK(c, lists)
}

// CreateEmailList handles POST /email-lists.
func (h *DistributionHandler) CreateEmailList(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)

	var input services.CreateEmailListInput
	if err := c.Bind(&input); err != nil {
		return apperr.BadRequest("invalid request body")
	}
	input.OrgID = orgID

	list, err := h.distributionSvc.CreateEmailList(c.Request().Context(), input)
	if err != nil {
		return err
	}

	return httputil.Created(c, list)
}

// GetEmailList handles GET /email-lists/:id.
func (h *DistributionHandler) GetEmailList(c echo.Context) error {
	// Would use emailListRepo to get specific list
	return c.JSON(http.StatusNotImplemented, echo.Map{
		"code":    "NOT_IMPLEMENTED",
		"message": "endpoint not yet implemented",
	})
}

// UpdateEmailList handles PUT /email-lists/:id.
func (h *DistributionHandler) UpdateEmailList(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, echo.Map{
		"code":    "NOT_IMPLEMENTED",
		"message": "endpoint not yet implemented",
	})
}

// DeleteEmailList handles DELETE /email-lists/:id.
func (h *DistributionHandler) DeleteEmailList(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, echo.Map{
		"code":    "NOT_IMPLEMENTED",
		"message": "endpoint not yet implemented",
	})
}

// AddContacts handles POST /email-lists/:id/contacts.
func (h *DistributionHandler) AddContacts(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, echo.Map{
		"code":    "NOT_IMPLEMENTED",
		"message": "endpoint not yet implemented",
	})
}

// RemoveContact handles DELETE /email-lists/:id/contacts/:contactId.
func (h *DistributionHandler) RemoveContact(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, echo.Map{
		"code":    "NOT_IMPLEMENTED",
		"message": "endpoint not yet implemented",
	})
}

// SendSurveyEmails handles POST /surveys/:id/send-emails.
func (h *DistributionHandler) SendSurveyEmails(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")

	var input services.SendSurveyEmailsInput
	if err := c.Bind(&input); err != nil {
		return apperr.BadRequest("invalid request body")
	}
	input.SurveyID = surveyID

	result, err := h.distributionSvc.SendSurveyEmails(c.Request().Context(), orgID, surveyID, input)
	if err != nil {
		return err
	}

	return httputil.OK(c, result)
}

// ListEmailCampaigns handles GET /surveys/:id/email/campaigns.
func (h *DistributionHandler) ListEmailCampaigns(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, echo.Map{
		"code":    "NOT_IMPLEMENTED",
		"message": "endpoint not yet implemented",
	})
}
