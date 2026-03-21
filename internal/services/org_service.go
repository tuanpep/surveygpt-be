package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/surveyflow/be/internal/models"
	apperr "github.com/surveyflow/be/internal/pkg/errors"
	"github.com/surveyflow/be/internal/repository"
)

// OrgService handles organization and membership operations.
type OrgService struct {
	orgRepo  repository.OrgRepo
	userRepo repository.UserRepo
}

// NewOrgService creates a new OrgService.
func NewOrgService(orgRepo repository.OrgRepo, userRepo repository.UserRepo) *OrgService {
	return &OrgService{
		orgRepo:  orgRepo,
		userRepo: userRepo,
	}
}

// --- Input types ---

type UpdateOrgInput struct {
	Name         *string      `json:"name"`
	Slug         *string      `json:"slug"`
	BillingEmail *string      `json:"billing_email"`
	Settings     map[string]any `json:"settings"`
}

type InviteMemberInput struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=owner admin member viewer"`
}

// --- Methods ---

// CreateOrganization creates a new org and adds the creator as owner.
func (s *OrgService) CreateOrganization(ctx context.Context, userID, name, slug string) (*models.Organization, error) {
	existing, err := s.orgRepo.GetOrgBySlug(ctx, slug)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if existing != nil {
		return nil, apperr.Conflict("organization slug already taken")
	}

	org := &models.Organization{
		Name:          name,
		Slug:          slug,
		Plan:          "free",
		ResponseLimit: 100,
		SurveyLimit:   10,
		MemberLimit:   5,
		AICredits:     50,
		Settings:      map[string]any{},
	}

	if err := s.orgRepo.CreateOrg(ctx, org); err != nil {
		slog.Error("failed to create organization", "error", err)
		return nil, apperr.Internal(err)
	}

	if _, err := s.orgRepo.AddMember(ctx, org.ID.String(), userID, "", models.RoleOwner); err != nil {
		slog.Error("failed to add owner membership", "error", err)
		return nil, apperr.Internal(err)
	}

	return org, nil
}

// GetOrganization returns an organization by ID.
func (s *OrgService) GetOrganization(ctx context.Context, orgID string) (*models.Organization, error) {
	org, err := s.orgRepo.GetOrgByID(ctx, orgID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if org == nil {
		return nil, apperr.NotFound("organization")
	}
	return org, nil
}

// UpdateOrganization updates editable fields.
func (s *OrgService) UpdateOrganization(ctx context.Context, orgID string, input UpdateOrgInput) (*models.Organization, error) {
	org, err := s.orgRepo.GetOrgByID(ctx, orgID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if org == nil {
		return nil, apperr.NotFound("organization")
	}

	if input.Name != nil {
		org.Name = *input.Name
	}
	if input.Slug != nil && *input.Slug != org.Slug {
		existing, err := s.orgRepo.GetOrgBySlug(ctx, *input.Slug)
		if err != nil {
			return nil, apperr.Internal(err)
		}
		if existing != nil {
			return nil, apperr.Conflict("slug already taken")
		}
		org.Slug = *input.Slug
	}
	if input.BillingEmail != nil {
		org.BillingEmail = input.BillingEmail
	}
	if input.Settings != nil {
		org.Settings = input.Settings
	}

	if err := s.orgRepo.UpdateOrg(ctx, org); err != nil {
		return nil, apperr.Internal(err)
	}
	return org, nil
}

// ListMembers returns all members of an organization.
func (s *OrgService) ListMembers(ctx context.Context, orgID string) ([]repository.MemberWithUser, error) {
	members, err := s.orgRepo.ListMembers(ctx, orgID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if members == nil {
		return []repository.MemberWithUser{}, nil
	}
	return members, nil
}

// InviteMember creates an invitation for a user to join.
func (s *OrgService) InviteMember(ctx context.Context, orgID, inviterID string, input InviteMemberInput) (*models.Invitation, error) {
	// Check if already a member.
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if user != nil {
		existing, err := s.orgRepo.GetMembership(ctx, orgID, user.ID.String())
		if err != nil {
			return nil, apperr.Internal(err)
		}
		if existing != nil {
			return nil, apperr.Conflict("user is already a member")
		}
	}

	// Check for pending invitation.
	pending, err := s.orgRepo.GetPendingInvitation(ctx, orgID, input.Email)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if pending != nil {
		return nil, apperr.Conflict("invitation already pending")
	}

	// Check member limit.
	count, err := s.orgRepo.GetMemberCount(ctx, orgID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	org, err := s.orgRepo.GetOrgByID(ctx, orgID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if count >= org.MemberLimit {
		return nil, apperr.PaymentRequired("member limit reached for this plan")
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, apperr.Internal(err)
	}

	inviterUUID := uuid.MustParse(inviterID)
	inv := &models.Invitation{
		OrgID:     org.ID,
		Email:     input.Email,
		Role:      models.Role(input.Role),
		InvitedBy: inviterUUID,
		Token:     hex.EncodeToString(tokenBytes),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.orgRepo.CreateInvitation(ctx, inv); err != nil {
		return nil, apperr.Internal(err)
	}

	slog.Info("invitation created", "org_id", orgID, "email", input.Email)
	return inv, nil
}

// AcceptInvitation accepts a pending invitation.
func (s *OrgService) AcceptInvitation(ctx context.Context, token, userID string) (*models.OrgMembership, error) {
	inv, err := s.orgRepo.GetInvitationByToken(ctx, token)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if inv == nil {
		return nil, apperr.NotFound("invitation")
	}
	if inv.AcceptedAt != nil {
		return nil, apperr.BadRequest("invitation already accepted")
	}
	if inv.ExpiresAt.Before(time.Now()) {
		return nil, apperr.BadRequest("invitation has expired")
	}

	if err := s.orgRepo.AcceptInvitation(ctx, token); err != nil {
		return nil, apperr.Internal(err)
	}

	m, err := s.orgRepo.AddMember(ctx, inv.OrgID.String(), userID, inv.InvitedBy.String(), inv.Role)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if m == nil {
		return nil, apperr.Conflict("already a member of this organization")
	}

	return m, nil
}

// DeclineInvitation removes a pending invitation.
func (s *OrgService) DeclineInvitation(ctx context.Context, token string) error {
	inv, err := s.orgRepo.GetInvitationByToken(ctx, token)
	if err != nil {
		return apperr.Internal(err)
	}
	if inv == nil {
		return apperr.NotFound("invitation")
	}

	return s.orgRepo.DeleteInvitation(ctx, token)
}

// RemoveMember removes a member from an organization.
func (s *OrgService) RemoveMember(ctx context.Context, orgID, targetUserID string) error {
	target, err := s.orgRepo.GetMembership(ctx, orgID, targetUserID)
	if err != nil {
		return apperr.Internal(err)
	}
	if target == nil {
		return apperr.NotFound("member")
	}
	if target.Role == models.RoleOwner {
		return apperr.BadRequest("cannot remove the organization owner")
	}

	return s.orgRepo.RemoveMember(ctx, orgID, targetUserID)
}

// UpdateMemberRole changes a member's role.
func (s *OrgService) UpdateMemberRole(ctx context.Context, orgID, targetUserID string, newRole models.Role) error {
	if !newRole.IsValid() {
		return apperr.BadRequest("invalid role")
	}

	target, err := s.orgRepo.GetMembership(ctx, orgID, targetUserID)
	if err != nil {
		return apperr.Internal(err)
	}
	if target == nil {
		return apperr.NotFound("member")
	}
	if target.Role == models.RoleOwner {
		return apperr.BadRequest("cannot change the owner's role")
	}

	return s.orgRepo.UpdateMemberRole(ctx, orgID, targetUserID, newRole)
}

// ListInvitations returns pending invitations for an org.
func (s *OrgService) ListInvitations(ctx context.Context, orgID string) ([]models.Invitation, error) {
	invs, err := s.orgRepo.ListInvitations(ctx, orgID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if invs == nil {
		return []models.Invitation{}, nil
	}
	return invs, nil
}

// GetUsage returns current usage metrics.
func (s *OrgService) GetUsage(ctx context.Context, orgID string) (*repository.OrgUsage, error) {
	usage, err := s.orgRepo.GetUsage(ctx, orgID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return usage, nil
}
