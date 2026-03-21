package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/surveyflow/be/internal/middleware"
	apperr "github.com/surveyflow/be/internal/pkg/errors"
	"github.com/surveyflow/be/internal/pkg/httputil"
	"github.com/surveyflow/be/internal/services"
)

// BillingHandler handles billing HTTP requests.
type BillingHandler struct {
	billingSvc *services.BillingService
}

// NewBillingHandler creates a new BillingHandler.
func NewBillingHandler(billingSvc *services.BillingService) *BillingHandler {
	return &BillingHandler{billingSvc: billingSvc}
}

// GetPlan handles GET /billing/plan.
func (h *BillingHandler) GetPlan(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)

	plan, err := h.billingSvc.GetPlan(c.Request().Context(), orgID)
	if err != nil {
		return err
	}

	return httputil.OK(c, plan)
}

// GetPlans handles GET /billing/plans.
func (h *BillingHandler) GetPlans(c echo.Context) error {
	plans := h.billingSvc.GetPlans(c.Request().Context())

	return httputil.OK(c, plans)
}

// ChangePlan handles POST /billing/change-plan.
func (h *BillingHandler) ChangePlan(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)

	var input services.ChangePlanInput
	if err := c.Bind(&input); err != nil {
		return apperr.BadRequest("invalid request body")
	}

	result, err := h.billingSvc.ChangePlan(c.Request().Context(), orgID, input)
	if err != nil {
		return err
	}

	return httputil.OK(c, result)
}

// GetPortalURL handles GET /billing/portal.
func (h *BillingHandler) GetPortalURL(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)

	url, err := h.billingSvc.GetPortalURL(c.Request().Context(), orgID)
	if err != nil {
		return err
	}

	return httputil.OK(c, echo.Map{
		"url": url,
	})
}

// GetHistory handles GET /billing/history.
func (h *BillingHandler) GetHistory(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)

	invoices, err := h.billingSvc.GetHistory(c.Request().Context(), orgID)
	if err != nil {
		return err
	}

	return httputil.OK(c, invoices)
}

// HandleStripeWebhook handles POST /webhooks/stripe.
func (h *BillingHandler) HandleStripeWebhook(c echo.Context) error {
	payload, err := readBody(c)
	if err != nil {
		return err
	}
	signature := c.Request().Header.Get("Stripe-Signature")

	if err := h.billingSvc.HandleStripeWebhook(c.Request().Context(), payload, signature); err != nil {
		return err
	}

	return httputil.OK(c, echo.Map{
		"received": true,
	})
}

// readBody is a helper to read request body.
func readBody(c echo.Context) ([]byte, error) {
	body := make([]byte, c.Request().ContentLength)
	_, err := c.Request().Body.Read(body)
	return body, err
}
