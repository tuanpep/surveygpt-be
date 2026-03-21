package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/surveyflow/be/internal/config"
	"github.com/surveyflow/be/internal/models"
	apperr "github.com/surveyflow/be/internal/pkg/errors"
	appjwt "github.com/surveyflow/be/internal/pkg/jwt"
	"github.com/surveyflow/be/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication operations.
type AuthService struct {
	userRepo repository.UserRepo
	orgRepo  repository.OrgRepo
	pool     *pgxpool.Pool
	config   *config.Config
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo repository.UserRepo, orgRepo repository.OrgRepo, pool *pgxpool.Pool, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		orgRepo:  orgRepo,
		pool:     pool,
		config:   cfg,
	}
}

// --- Request/Response types ---

type SignUpInput struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128"`
	Name     string `json:"name" validate:"required,min=1,max=255"`
}

type SignInInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=128"`
}

type SignUpOutput struct {
	User   *UserWithOrg     `json:"user"`
	Tokens *appjwt.TokenPair `json:"tokens"`
}

type SignInOutput struct {
	User        *UserWithOrg     `json:"user,omitempty"`
	Tokens      *appjwt.TokenPair `json:"tokens,omitempty"`
	Requires2FA bool             `json:"requires_2fa"`
}

// UserWithOrg extends User info with organization context.
type UserWithOrg struct {
	ID               string     `json:"id"`
	Email            string     `json:"email"`
	Name             string     `json:"name"`
	AvatarURL        *string    `json:"avatar_url,omitempty"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at,omitempty"`
	TwoFactorEnabled bool       `json:"two_factor_enabled"`
	OrgID            string     `json:"org_id"`
	OrgName          string     `json:"org_name"`
	Role             string     `json:"role"`
	CreatedAt        time.Time  `json:"created_at"`
}

// --- Methods ---

// SignUp registers a new user, creates a default organization, and returns tokens.
func (s *AuthService) SignUp(ctx context.Context, input SignUpInput) (*SignUpOutput, error) {
	existing, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		slog.Error("failed to check existing user", "error", err)
		return nil, apperr.Internal(err)
	}
	if existing != nil {
		return nil, apperr.Conflict("email already registered")
	}

	if err := validatePasswordStrength(input.Password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	hashStr := string(hash)

	user := &models.User{
		Email:        input.Email,
		Name:         input.Name,
		PasswordHash: &hashStr,
	}

	// Create default organization.
	slug := slugify(input.Name) + "-" + randomSuffix(6)
	org := &models.Organization{
		Name:          input.Name + "'s Workspace",
		Slug:          slug,
		Plan:          "free",
		ResponseLimit: s.config.PlanDefaults.ResponseLimit,
		SurveyLimit:   s.config.PlanDefaults.SurveyLimit,
		MemberLimit:   s.config.PlanDefaults.MemberLimit,
		AICredits:     s.config.PlanDefaults.AICredits,
		Settings:      map[string]any{},
	}

	// Use a transaction to ensure user + org + membership are created atomically.
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	defer tx.Rollback(ctx)

	// Create user within transaction.
	userQuery := `
		INSERT INTO users (email, name, password_hash, avatar_url, email_verified_at,
			two_factor_secret, two_factor_enabled, last_login_at)
		VALUES ($1, $2, $3, NULL, NULL, NULL, FALSE, NULL)
		RETURNING id, created_at, updated_at`
	if err := tx.QueryRow(ctx, userQuery,
		user.Email, user.Name, user.PasswordHash,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt); err != nil {
		slog.Error("failed to create user", "error", err)
		return nil, apperr.Internal(err)
	}

	// Create organization within transaction.
	orgQuery := `
		INSERT INTO organizations (name, slug, plan, billing_email, settings,
			response_limit, survey_limit, member_limit, ai_credits,
			current_period_start, current_period_end)
		VALUES ($1, $2, $3, NULL, $4, $5, $6, $7, $8, NULL, NULL)
		RETURNING id, created_at, updated_at`
	if err := tx.QueryRow(ctx, orgQuery,
		org.Name, org.Slug, org.Plan, org.Settings,
		org.ResponseLimit, org.SurveyLimit, org.MemberLimit, org.AICredits,
	).Scan(&org.ID, &org.CreatedAt, &org.UpdatedAt); err != nil {
		slog.Error("failed to create organization", "error", err)
		return nil, apperr.Internal(err)
	}

	// Add owner membership within transaction.
	memberQuery := `
		INSERT INTO org_memberships (org_id, user_id, role, invited_by)
		VALUES ($1, $2, $3, NULL)
		ON CONFLICT (org_id, user_id) DO NOTHING
		RETURNING id`
	if _, err := tx.Exec(ctx, memberQuery, org.ID, user.ID, models.RoleOwner); err != nil {
		slog.Error("failed to add owner membership", "error", err)
		return nil, apperr.Internal(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperr.Internal(err)
	}

	tokens, err := appjwt.GenerateTokenPair(
		user.ID.String(), org.ID.String(), "owner",
		s.config.JWT.Secret,
		s.config.JWT.AccessTTL,
		s.config.JWT.RefreshTTL,
	)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return &SignUpOutput{
		User:   toUserWithOrg(user, org, "owner"),
		Tokens: tokens,
	}, nil
}

// SignIn authenticates a user and returns tokens.
func (s *AuthService) SignIn(ctx context.Context, input SignInInput) (*SignInOutput, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		slog.Error("failed to get user by email", "error", err)
		return nil, apperr.Internal(err)
	}
	if user == nil || user.PasswordHash == nil {
		return nil, apperr.Unauthorized()
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, apperr.Unauthorized()
	}

	if user.TwoFactorEnabled {
		return &SignInOutput{Requires2FA: true}, nil
	}

	return s.completeSignIn(ctx, user)
}

// SignInWith2FA completes sign-in after 2FA verification.
func (s *AuthService) SignInWith2FA(ctx context.Context, email, code string) (*SignInOutput, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if user == nil {
		return nil, apperr.Unauthorized()
	}
	if !user.TwoFactorEnabled {
		return nil, apperr.BadRequest("2FA is not enabled for this account")
	}

	// TODO: Validate the TOTP code against the user's 2FA secret.
	// This requires storing the TOTP secret in the user record and using
	// a TOTP library (e.g., github.com/pquerna/otp) to validate.
	// For now, reject empty or obviously invalid codes.
	if len(code) < 6 {
		return nil, apperr.Unauthorized()
	}

	return s.completeSignIn(ctx, user)
}

func (s *AuthService) completeSignIn(ctx context.Context, user *models.User) (*SignInOutput, error) {
	memberships, err := s.orgRepo.GetMembershipsByUserID(ctx, user.ID.String())
	if err != nil || len(memberships) == 0 {
		slog.Error("failed to get user memberships", "error", err)
		return nil, apperr.Internal(err)
	}

	m := memberships[0]
	org, err := s.orgRepo.GetOrgByID(ctx, m.OrgID.String())
	if err != nil {
		return nil, apperr.Internal(err)
	}

	tokens, err := appjwt.GenerateTokenPair(
		user.ID.String(), m.OrgID.String(), string(m.Role),
		s.config.JWT.Secret,
		s.config.JWT.AccessTTL,
		s.config.JWT.RefreshTTL,
	)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return &SignInOutput{
		User:   toUserWithOrg(user, org, string(m.Role)),
		Tokens: tokens,
	}, nil
}

// GetCurrentUser returns the authenticated user with org context.
func (s *AuthService) GetCurrentUser(ctx context.Context, userID, orgID string) (*UserWithOrg, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if user == nil {
		return nil, apperr.NotFound("user")
	}

	membership, err := s.orgRepo.GetMembership(ctx, orgID, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if membership == nil {
		return nil, apperr.Forbidden()
	}

	org, err := s.orgRepo.GetOrgByID(ctx, orgID)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return toUserWithOrg(user, org, string(membership.Role)), nil
}

// ForgotPassword initiates the password reset flow.
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return apperr.Internal(err)
	}
	if user == nil {
		return nil // don't reveal email existence
	}

	// TODO: enqueue email task with reset link containing the token.
	// For now, just generate the token (it will be sent via email).
	_, err = appjwt.GenerateToken(user.ID.String(), "", "", s.config.JWT.Secret, 1*time.Hour)
	if err != nil {
		return apperr.Internal(err)
	}

	slog.Info("password reset token generated", "user_id", user.ID)
	return nil
}

// ResetPassword sets a new password using a reset token.
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	claims, err := appjwt.ParseToken(token, s.config.JWT.Secret)
	if err != nil {
		return apperr.BadRequest("invalid or expired reset token")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return apperr.Internal(err)
	}

	return s.userRepo.UpdatePassword(ctx, claims.UserID, string(hash))
}

// ChangePassword verifies current password and sets a new one.
func (s *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return apperr.Internal(err)
	}
	if user == nil || user.PasswordHash == nil {
		return apperr.BadRequest("user has no password set")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(currentPassword)); err != nil {
		return apperr.BadRequest("current password is incorrect")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return apperr.Internal(err)
	}

	return s.userRepo.UpdatePassword(ctx, userID, string(hash))
}

// RefreshToken generates a new token pair from a refresh token.
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*appjwt.TokenPair, error) {
	claims, err := appjwt.ParseToken(refreshToken, s.config.JWT.Secret)
	if err != nil {
		return nil, apperr.Unauthorized()
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if user == nil {
		return nil, apperr.Unauthorized()
	}

	tokens, err := appjwt.GenerateTokenPair(
		claims.UserID, claims.OrgID, claims.Role,
		s.config.JWT.Secret,
		s.config.JWT.AccessTTL,
		s.config.JWT.RefreshTTL,
	)
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return tokens, nil
}

// UpdateProfile updates the user's name and avatar.
func (s *AuthService) UpdateProfile(ctx context.Context, userID, name string, avatarURL *string) (*UserWithOrg, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if user == nil {
		return nil, apperr.NotFound("user")
	}

	user.Name = name
	user.AvatarURL = avatarURL
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, apperr.Internal(err)
	}

	// Return without org context for simplicity.
	return &UserWithOrg{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt,
	}, nil
}

// --- Helpers ---

func toUserWithOrg(user *models.User, org *models.Organization, role string) *UserWithOrg {
	return &UserWithOrg{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		EmailVerifiedAt:  user.EmailVerifiedAt,
		TwoFactorEnabled: user.TwoFactorEnabled,
		OrgID:            org.ID.String(),
		OrgName:          org.Name,
		Role:             role,
		CreatedAt:        user.CreatedAt,
	}
}

func slugify(s string) string {
	var b strings.Builder
	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			b.WriteRune(c)
		case c >= 'A' && c <= 'Z':
			b.WriteRune(c + 32)
		case c == ' ' || c == '-' || c == '_':
			b.WriteByte('-')
		}
	}
	result := b.String()
	if result == "" {
		return "user"
	}
	return result
}

func randomSuffix(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// Fallback: use crypto/rand failure should be extremely rare.
		// Use a slice of the current unix nano time encoded as hex.
		nano := time.Now().UnixNano()
		h := hex.EncodeToString([]byte{
			byte(nano >> 56), byte(nano >> 48), byte(nano >> 40), byte(nano >> 32),
			byte(nano >> 24), byte(nano >> 16), byte(nano >> 8), byte(nano),
		})
		if len(h) >= n {
			return h[:n]
		}
		return h
	}
	return hex.EncodeToString(b)
}

// validatePasswordStrength checks that the password meets minimum security requirements.
func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return apperr.BadRequest("password must be at least 8 characters long")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return apperr.BadRequest("password must contain at least one uppercase letter, one lowercase letter, and one digit")
	}
	return nil
}

