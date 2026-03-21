package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/surveyflow/be/internal/models"
)

// orgRepo provides data access for organizations, memberships, and invitations.
type orgRepo struct {
	pool *pgxpool.Pool
}

// NewOrgRepo creates a new OrgRepo.
func NewOrgRepo(pool *pgxpool.Pool) OrgRepo {
	return &orgRepo{pool: pool}
}

// MemberWithUser joins an OrgMembership with the associated User.
type MemberWithUser struct {
	models.OrgMembership
	User *models.User `json:"user"`
}

// --- Organizations ---

// CreateOrg inserts a new organization.
func (r *orgRepo) CreateOrg(ctx context.Context, org *models.Organization) error {
	query := `
		INSERT INTO organizations (name, slug, plan, billing_email, settings,
			response_limit, survey_limit, member_limit, ai_credits,
			current_period_start, current_period_end)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		org.Name, org.Slug, org.Plan, org.BillingEmail, org.Settings,
		org.ResponseLimit, org.SurveyLimit, org.MemberLimit, org.AICredits,
		org.CurrentPeriodStart, org.CurrentPeriodEnd,
	).Scan(&org.ID, &org.CreatedAt, &org.UpdatedAt)
}

// GetOrgByID fetches an organization by ID.
func (r *orgRepo) GetOrgByID(ctx context.Context, id string) (*models.Organization, error) {
	query := `
		SELECT id, name, slug, plan, stripe_customer_id, stripe_subscription_id,
			billing_email, settings, response_limit, survey_limit, member_limit,
			ai_credits, current_period_start, current_period_end, created_at, updated_at
		FROM organizations WHERE id = $1`

	org := &models.Organization{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Plan,
		&org.StripeCustomerID, &org.StripeSubscriptionID, &org.BillingEmail,
		&org.Settings, &org.ResponseLimit, &org.SurveyLimit, &org.MemberLimit,
		&org.AICredits, &org.CurrentPeriodStart, &org.CurrentPeriodEnd,
		&org.CreatedAt, &org.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return org, nil
}

// GetOrgBySlug fetches an organization by slug.
func (r *orgRepo) GetOrgBySlug(ctx context.Context, slug string) (*models.Organization, error) {
	query := `
		SELECT id, name, slug, plan, stripe_customer_id, stripe_subscription_id,
			billing_email, settings, response_limit, survey_limit, member_limit,
			ai_credits, current_period_start, current_period_end, created_at, updated_at
		FROM organizations WHERE slug = $1`

	org := &models.Organization{}
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&org.ID, &org.Name, &org.Slug, &org.Plan,
		&org.StripeCustomerID, &org.StripeSubscriptionID, &org.BillingEmail,
		&org.Settings, &org.ResponseLimit, &org.SurveyLimit, &org.MemberLimit,
		&org.AICredits, &org.CurrentPeriodStart, &org.CurrentPeriodEnd,
		&org.CreatedAt, &org.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return org, nil
}

// UpdateOrg updates an organization's editable fields.
func (r *orgRepo) UpdateOrg(ctx context.Context, org *models.Organization) error {
	query := `
		UPDATE organizations SET name = $2, slug = $3, billing_email = $4,
			settings = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	return r.pool.QueryRow(ctx, query,
		org.ID, org.Name, org.Slug, org.BillingEmail, org.Settings,
	).Scan(&org.UpdatedAt)
}

// --- Memberships ---

// AddMember creates a new organization membership.
func (r *orgRepo) AddMember(ctx context.Context, orgID, userID, invitedBy string, role models.Role) (*models.OrgMembership, error) {
	query := `
		INSERT INTO org_memberships (org_id, user_id, role, invited_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (org_id, user_id) DO NOTHING
		RETURNING id, joined_at, created_at`

	orgUUID := uuid.MustParse(orgID)
	userUUID := uuid.MustParse(userID)
	var invitedByUUID *uuid.UUID
	if invitedBy != "" {
		u := uuid.MustParse(invitedBy)
		invitedByUUID = &u
	}

	m := &models.OrgMembership{
		OrgID:     orgUUID,
		UserID:    userUUID,
		Role:      role,
		InvitedBy: invitedByUUID,
	}

	err := r.pool.QueryRow(ctx, query, orgUUID, userUUID, role, invitedByUUID).Scan(
		&m.ID, &m.JoinedAt, &m.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		// Conflict — membership already exists
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetMembership fetches a membership by org + user.
func (r *orgRepo) GetMembership(ctx context.Context, orgID, userID string) (*models.OrgMembership, error) {
	query := `
		SELECT id, org_id, user_id, role, invited_by, joined_at, created_at
		FROM org_memberships WHERE org_id = $1 AND user_id = $2`

	m := &models.OrgMembership{}
	err := r.pool.QueryRow(ctx, query, orgID, userID).Scan(
		&m.ID, &m.OrgID, &m.UserID, &m.Role, &m.InvitedBy, &m.JoinedAt, &m.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ListMembers returns all members of an organization with their user info.
func (r *orgRepo) ListMembers(ctx context.Context, orgID string) ([]MemberWithUser, error) {
	query := `
		SELECT m.id, m.org_id, m.user_id, m.role, m.invited_by, m.joined_at, m.created_at,
			u.id, u.email, u.name, u.avatar_url, u.two_factor_enabled, u.created_at, u.updated_at
		FROM org_memberships m
		JOIN users u ON u.id = m.user_id
		WHERE m.org_id = $1
		ORDER BY m.joined_at`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []MemberWithUser
	for rows.Next() {
		var m MemberWithUser
		m.User = &models.User{}
		err := rows.Scan(
			&m.ID, &m.OrgID, &m.UserID, &m.Role, &m.InvitedBy, &m.JoinedAt, &m.CreatedAt,
			&m.User.ID, &m.User.Email, &m.User.Name, &m.User.AvatarURL,
			&m.User.TwoFactorEnabled, &m.User.CreatedAt, &m.User.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

// UpdateMemberRole changes a member's role.
func (r *orgRepo) UpdateMemberRole(ctx context.Context, orgID, userID string, role models.Role) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE org_memberships SET role = $3 WHERE org_id = $1 AND user_id = $2`,
		orgID, userID, role,
	)
	return err
}

// RemoveMember deletes a membership.
func (r *orgRepo) RemoveMember(ctx context.Context, orgID, userID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM org_memberships WHERE org_id = $1 AND user_id = $2`,
		orgID, userID,
	)
	return err
}

// GetMembershipsByUserID returns all org memberships for a given user.
func (r *orgRepo) GetMembershipsByUserID(ctx context.Context, userID string) ([]models.OrgMembership, error) {
	query := `
		SELECT id, org_id, user_id, role, invited_by, joined_at, created_at
		FROM org_memberships WHERE user_id = $1`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []models.OrgMembership
	for rows.Next() {
		var m models.OrgMembership
		if err := rows.Scan(&m.ID, &m.OrgID, &m.UserID, &m.Role, &m.InvitedBy, &m.JoinedAt, &m.CreatedAt); err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}
	return memberships, rows.Err()
}

// GetMemberCount returns the number of members in an organization.
func (r *orgRepo) GetMemberCount(ctx context.Context, orgID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM org_memberships WHERE org_id = $1`, orgID,
	).Scan(&count)
	return count, err
}

// --- Invitations ---

// CreateInvitation inserts a new invitation.
func (r *orgRepo) CreateInvitation(ctx context.Context, inv *models.Invitation) error {
	query := `
		INSERT INTO invitations (org_id, email, role, invited_by, token, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query,
		inv.OrgID, inv.Email, inv.Role, inv.InvitedBy, inv.Token, inv.ExpiresAt,
	).Scan(&inv.ID, &inv.CreatedAt)
}

// GetInvitationByToken fetches an invitation by its token.
func (r *orgRepo) GetInvitationByToken(ctx context.Context, token string) (*models.Invitation, error) {
	query := `
		SELECT id, org_id, email, role, invited_by, token, expires_at, accepted_at, created_at
		FROM invitations WHERE token = $1`

	inv := &models.Invitation{}
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&inv.ID, &inv.OrgID, &inv.Email, &inv.Role, &inv.InvitedBy,
		&inv.Token, &inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// ListInvitations returns pending invitations for an organization.
func (r *orgRepo) ListInvitations(ctx context.Context, orgID string) ([]models.Invitation, error) {
	query := `
		SELECT id, org_id, email, role, invited_by, token, expires_at, accepted_at, created_at
		FROM invitations WHERE org_id = $1 AND accepted_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []models.Invitation
	for rows.Next() {
		var inv models.Invitation
		err := rows.Scan(
			&inv.ID, &inv.OrgID, &inv.Email, &inv.Role, &inv.InvitedBy,
			&inv.Token, &inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		invitations = append(invitations, inv)
	}
	return invitations, nil
}

// AcceptInvitation marks an invitation as accepted.
func (r *orgRepo) AcceptInvitation(ctx context.Context, token string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE invitations SET accepted_at = NOW() WHERE token = $1 AND accepted_at IS NULL AND expires_at > NOW()`,
		token,
	)
	return err
}

// DeleteInvitation removes an invitation.
func (r *orgRepo) DeleteInvitation(ctx context.Context, token string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM invitations WHERE token = $1`, token)
	return err
}

// GetPendingInvitation checks for a pending invitation to an org for an email.
func (r *orgRepo) GetPendingInvitation(ctx context.Context, orgID, email string) (*models.Invitation, error) {
	query := `
		SELECT id, org_id, email, role, invited_by, token, expires_at, accepted_at, created_at
		FROM invitations
		WHERE org_id = $1 AND email = $2 AND accepted_at IS NULL AND expires_at > NOW()`

	inv := &models.Invitation{}
	err := r.pool.QueryRow(ctx, query, orgID, email).Scan(
		&inv.ID, &inv.OrgID, &inv.Email, &inv.Role, &inv.InvitedBy,
		&inv.Token, &inv.ExpiresAt, &inv.AcceptedAt, &inv.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// --- Usage ---

// OrgUsage holds current usage metrics for an organization.
type OrgUsage struct {
	Responses  int `json:"responses"`
	Surveys    int `json:"surveys"`
	Members    int `json:"members"`
	AICredits  struct {
		Total int `json:"total"`
		Used  int `json:"used"`
	} `json:"ai_credits"`
}

// GetUsage fetches current usage for an organization.
func (r *orgRepo) GetUsage(ctx context.Context, orgID string) (*OrgUsage, error) {
	org, err := r.GetOrgByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, pgx.ErrNoRows
	}

	memberCount, err := r.GetMemberCount(ctx, orgID)
	if err != nil {
		return nil, err
	}

	var surveyCount, responseCount, aiUsed int
	err = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM surveys WHERE org_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&surveyCount)
	if err != nil {
		return nil, err
	}

	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(response_count), 0) FROM surveys WHERE org_id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&responseCount)
	if err != nil {
		return nil, err
	}

	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(used_credits, 0) FROM ai_credits WHERE org_id = $1`, orgID,
	).Scan(&aiUsed)
	if err != nil {
		aiUsed = 0 // ai_credits row may not exist yet
	}

	usage := &OrgUsage{
		Responses: responseCount,
		Surveys:   surveyCount,
		Members:   memberCount,
	}
	usage.AICredits.Total = org.AICredits
	usage.AICredits.Used = aiUsed

	return usage, nil
}

