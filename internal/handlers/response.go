package handlers

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/surveyflow/be/internal/middleware"
	apperr "github.com/surveyflow/be/internal/pkg/errors"
	"github.com/surveyflow/be/internal/pkg/httputil"
	"github.com/surveyflow/be/internal/services"
)

// ResponseHandler handles response HTTP requests.
type ResponseHandler struct {
	responseSvc *services.ResponseService
}

// NewResponseHandler creates a new ResponseHandler.
func NewResponseHandler(responseSvc *services.ResponseService) *ResponseHandler {
	return &ResponseHandler{responseSvc: responseSvc}
}

// Submit handles POST /surveys/:id/responses (public, no auth).
func (h *ResponseHandler) Submit(c echo.Context) error {
	surveyID := c.Param("id")

	var input struct {
		Answers  []services.AnswerInput         `json:"answers"`
		Metadata services.ResponseMetadataInput `json:"metadata"`
	}
	if err := c.Bind(&input); err != nil {
		return apperr.BadRequest("invalid request body")
	}

	resp, err := h.responseSvc.SubmitResponse(c.Request().Context(), services.SubmitResponseInput{
		SurveyID: surveyID,
		Answers:  input.Answers,
		Metadata: input.Metadata,
	})
	if err != nil {
		return err
	}

	return httputil.Created(c, echo.Map{
		"success":      true,
		"response_id":  resp.ID.String(),
		"redirect_url": "/s/" + surveyID + "/thank-you",
	})
}

// List handles GET /surveys/:id/responses (authenticated).
func (h *ResponseHandler) List(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")

	filter := services.ListResponsesFilter{
		Status: c.QueryParam("status"),
		Cursor: c.QueryParam("cursor"),
	}
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if dateFrom := c.QueryParam("date_from"); dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filter.DateFrom = &t
		}
	}
	if dateTo := c.QueryParam("date_to"); dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			filter.DateTo = &t
		}
	}

	responses, nextCursor, total, err := h.responseSvc.ListResponses(c.Request().Context(), orgID, surveyID, filter)
	if err != nil {
		return err
	}

	return httputil.Paginated(c, responses, nextCursor, total)
}

// Get handles GET /surveys/:id/responses/:responseId (authenticated).
func (h *ResponseHandler) Get(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")
	responseID := c.Param("responseId")

	resp, err := h.responseSvc.GetResponse(c.Request().Context(), orgID, surveyID, responseID)
	if err != nil {
		return err
	}

	return httputil.OK(c, resp)
}

// Delete handles DELETE /surveys/:id/responses/:responseId (authenticated).
func (h *ResponseHandler) Delete(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")
	responseID := c.Param("responseId")

	if err := h.responseSvc.DeleteResponse(c.Request().Context(), orgID, surveyID, responseID); err != nil {
		return err
	}

	return httputil.Message(c, "response deleted successfully")
}

// Export handles POST /surveys/:id/responses/export (authenticated).
func (h *ResponseHandler) Export(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	surveyID := c.Param("id")

	var body struct {
		Format   string `json:"format"`
		Status   string `json:"status"`
		DateFrom string `json:"date_from"`
		DateTo   string `json:"date_to"`
	}
	if err := c.Bind(&body); err != nil {
		return apperr.BadRequest("invalid request body")
	}

	filter := services.ListResponsesFilter{
		Status: body.Status,
	}
	if body.DateFrom != "" {
		if t, err := time.Parse("2006-01-02", body.DateFrom); err == nil {
			filter.DateFrom = &t
		}
	}
	if body.DateTo != "" {
		if t, err := time.Parse("2006-01-02", body.DateTo); err == nil {
			filter.DateTo = &t
		}
	}

	format := body.Format
	if format == "" {
		format = "json"
	}

	data, err := h.responseSvc.ExportResponses(c.Request().Context(), orgID, surveyID, format, filter)
	if err != nil {
		return err
	}

	return httputil.OK(c, echo.Map{
		"data":   data,
		"format": format,
	})
}
