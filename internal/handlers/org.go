package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/surveyflow/be/internal/models"
	"github.com/surveyflow/be/internal/middleware"
	apperr "github.com/surveyflow/be/internal/pkg/errors"
	"github.com/surveyflow/be/internal/pkg/httputil"
	"github.com/surveyflow/be/internal/pkg/validator"
	"github.com/surveyflow/be/internal/services"
)

// OrgHandler handles organization HTTP requests.
type OrgHandler struct {
	orgSvc   *services.OrgService
	validator *validator.Validator
}

// NewOrgHandler creates a new OrgHandler.
func NewOrgHandler(orgSvc *services.OrgService) *OrgHandler {
	return &OrgHandler{
		orgSvc:   orgSvc,
		validator: validator.New(),
	}
}

// Create handles POST /organizations.
func (h *OrgHandler) Create(c echo.Context) error {
	var body struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := c.Bind(&body); err != nil {
		return apperr.BadRequest("invalid request body")
	}
	if body.Name == "" || body.Slug == "" {
		return apperr.BadRequest("name and slug are required")
	}

	userID := middleware.GetUserIDFromContext(c)
	org, err := h.orgSvc.CreateOrganization(c.Request().Context(), userID, body.Name, body.Slug)
	if err != nil {
		return err
	}

	return httputil.Created(c, org)
}

// Get handles GET /organizations/me.
func (h *OrgHandler) Get(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)

	org, err := h.orgSvc.GetOrganization(c.Request().Context(), orgID)
	if err != nil {
		return err
	}

	return httputil.OK(c, org)
}

// Update handles PUT /organizations/me.
func (h *OrgHandler) Update(c echo.Context) error {
	var input services.UpdateOrgInput
	if err := c.Bind(&input); err != nil {
		return apperr.BadRequest("invalid request body")
	}

	orgID := middleware.GetOrgIDFromContext(c)
	org, err := h.orgSvc.UpdateOrganization(c.Request().Context(), orgID, input)
	if err != nil {
		return err
	}

	return httputil.OK(c, org)
}

// ListMembers handles GET /organizations/:orgId/members.
func (h *OrgHandler) ListMembers(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)

	members, err := h.orgSvc.ListMembers(c.Request().Context(), orgID)
	if err != nil {
		return err
	}

	return httputil.OK(c, members)
}

// InviteMember handles POST /organizations/:orgId/members/invite.
func (h *OrgHandler) InviteMember(c echo.Context) error {
	var input services.InviteMemberInput
	if err := c.Bind(&input); err != nil {
		return apperr.BadRequest("invalid request body")
	}

	if errs := h.validator.Validate(input); errs != nil {
		return c.JSON(http.StatusUnprocessableEntity, echo.Map{
			"code":    "VALIDATION_FAILED",
			"message": "validation failed",
			"errors":  errs,
		})
	}

	orgID := middleware.GetOrgIDFromContext(c)
	userID := middleware.GetUserIDFromContext(c)

	inv, err := h.orgSvc.InviteMember(c.Request().Context(), orgID, userID, input)
	if err != nil {
		return err
	}

	return httputil.Created(c, inv)
}

// UpdateMemberRole handles PATCH /organizations/:orgId/members/:memberId/role.
func (h *OrgHandler) UpdateMemberRole(c echo.Context) error {
	var body struct {
		Role string `json:"role"`
	}
	if err := c.Bind(&body); err != nil {
		return apperr.BadRequest("invalid request body")
	}
	if body.Role == "" {
		return apperr.BadRequest("role is required")
	}

	orgID := middleware.GetOrgIDFromContext(c)
	memberID := c.Param("memberId")

	if err := h.orgSvc.UpdateMemberRole(c.Request().Context(), orgID, memberID, models.Role(body.Role)); err != nil {
		return err
	}

	return httputil.Message(c, "role updated successfully")
}

// RemoveMember handles DELETE /organizations/:orgId/members/:memberId.
func (h *OrgHandler) RemoveMember(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)
	memberID := c.Param("memberId")

	if err := h.orgSvc.RemoveMember(c.Request().Context(), orgID, memberID); err != nil {
		return err
	}

	return httputil.Message(c, "member removed successfully")
}

// ListInvitations handles GET /organizations/:orgId/invitations.
func (h *OrgHandler) ListInvitations(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)

	invs, err := h.orgSvc.ListInvitations(c.Request().Context(), orgID)
	if err != nil {
		return err
	}

	return httputil.OK(c, invs)
}

// AcceptInvitation handles POST /invitations/:token/accept.
func (h *OrgHandler) AcceptInvitation(c echo.Context) error {
	token := c.Param("token")
	userID := middleware.GetUserIDFromContext(c)

	m, err := h.orgSvc.AcceptInvitation(c.Request().Context(), token, userID)
	if err != nil {
		return err
	}

	return httputil.OK(c, m)
}

// DeclineInvitation handles POST /invitations/:token/decline.
func (h *OrgHandler) DeclineInvitation(c echo.Context) error {
	token := c.Param("token")

	if err := h.orgSvc.DeclineInvitation(c.Request().Context(), token); err != nil {
		return err
	}

	return httputil.Message(c, "invitation declined")
}

// GetUsage handles GET /usage.
func (h *OrgHandler) GetUsage(c echo.Context) error {
	orgID := middleware.GetOrgIDFromContext(c)

	usage, err := h.orgSvc.GetUsage(c.Request().Context(), orgID)
	if err != nil {
		return err
	}

	return httputil.OK(c, usage)
}
