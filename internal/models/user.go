package models

import (
	"time"

	"github.com/google/uuid"
)

// Role defines the permission level within an organization.
type Role string

const (
	RoleOwner   Role = "owner"
	RoleAdmin   Role = "admin"
	RoleMember  Role = "member"
	RoleViewer  Role = "viewer"
)

// User represents a registered user account.
type User struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	Email            string     `json:"email" db:"email"`
	Name             string     `json:"name" db:"name"`
	AvatarURL        *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	PasswordHash     *string    `json:"-" db:"password_hash"` // never serialized to JSON
	EmailVerifiedAt  *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
	TwoFactorSecret  *string    `json:"-" db:"two_factor_secret"`
	TwoFactorEnabled bool       `json:"two_factor_enabled" db:"two_factor_enabled"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// Organization represents a workspace that groups users and surveys together.
type Organization struct {
	ID                   uuid.UUID       `json:"id" db:"id"`
	Name                 string          `json:"name" db:"name"`
	Slug                 string          `json:"slug" db:"slug"`
	Plan                 string          `json:"plan" db:"plan"`
	StripeCustomerID     *string         `json:"stripe_customer_id,omitempty" db:"stripe_customer_id"`
	StripeSubscriptionID *string         `json:"stripe_subscription_id,omitempty" db:"stripe_subscription_id"`
	BillingEmail         *string         `json:"billing_email,omitempty" db:"billing_email"`
	Settings             map[string]any  `json:"settings" db:"settings"`
	ResponseLimit        int             `json:"response_limit" db:"response_limit"`
	SurveyLimit          int             `json:"survey_limit" db:"survey_limit"`
	MemberLimit          int             `json:"member_limit" db:"member_limit"`
	AICredits            int             `json:"ai_credits" db:"ai_credits"`
	CurrentPeriodStart   *time.Time      `json:"current_period_start,omitempty" db:"current_period_start"`
	CurrentPeriodEnd     *time.Time      `json:"current_period_end,omitempty" db:"current_period_end"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`
}

// OrgMembership links a user to an organization with a specific role.
type OrgMembership struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	OrgID     uuid.UUID  `json:"org_id" db:"org_id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	Role      Role       `json:"role" db:"role"`
	InvitedBy *uuid.UUID `json:"invited_by,omitempty" db:"invited_by"`
	JoinedAt  time.Time  `json:"joined_at" db:"joined_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// Invitation represents a pending invitation for a user to join an organization.
type Invitation struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	OrgID       uuid.UUID  `json:"org_id" db:"org_id"`
	Email       string     `json:"email" db:"email"`
	Role        Role       `json:"role" db:"role"`
	InvitedBy   uuid.UUID  `json:"invited_by" db:"invited_by"`
	Token       string     `json:"-" db:"token"`
	ExpiresAt   time.Time  `json:"expires_at" db:"expires_at"`
	AcceptedAt  *time.Time `json:"accepted_at,omitempty" db:"accepted_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// HasPermission checks if a role has at least the specified minimum role level.
func (r Role) HasPermission(minimum Role) bool {
	levels := map[Role]int{
		RoleViewer: 0,
		RoleMember: 1,
		RoleAdmin:  2,
		RoleOwner:  3,
	}
	return levels[r] >= levels[minimum]
}

// IsValid checks whether the role value is a known constant.
func (r Role) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember, RoleViewer:
		return true
	}
	return false
}
