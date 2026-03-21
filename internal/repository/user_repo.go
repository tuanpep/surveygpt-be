package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/surveyflow/be/internal/models"
)

// userRepo provides data access for user records.
type userRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(pool *pgxpool.Pool) UserRepo {
	return &userRepo{pool: pool}
}

// Create inserts a new user and returns it with generated fields populated.
func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (email, name, password_hash, avatar_url, email_verified_at,
			two_factor_secret, two_factor_enabled, last_login_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		user.Email, user.Name, user.PasswordHash, user.AvatarURL,
		user.EmailVerifiedAt, user.TwoFactorSecret, user.TwoFactorEnabled,
		user.LastLoginAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetByID fetches a user by primary key.
func (r *userRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, email, name, avatar_url, password_hash, email_verified_at,
			two_factor_secret, two_factor_enabled, last_login_at, created_at, updated_at
		FROM users WHERE id = $1`

	user := &models.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.PasswordHash,
		&user.EmailVerifiedAt, &user.TwoFactorSecret, &user.TwoFactorEnabled,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByEmail fetches a user by email address.
func (r *userRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, name, avatar_url, password_hash, email_verified_at,
			two_factor_secret, two_factor_enabled, last_login_at, created_at, updated_at
		FROM users WHERE email = $1`

	user := &models.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.PasswordHash,
		&user.EmailVerifiedAt, &user.TwoFactorSecret, &user.TwoFactorEnabled,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update applies partial updates to a user.
func (r *userRepo) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET name = $2, avatar_url = $3, password_hash = $4,
			email_verified_at = $5, two_factor_secret = $6, two_factor_enabled = $7,
			last_login_at = $8, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	return r.pool.QueryRow(ctx, query,
		user.ID, user.Name, user.AvatarURL, user.PasswordHash,
		user.EmailVerifiedAt, user.TwoFactorSecret, user.TwoFactorEnabled,
		user.LastLoginAt,
	).Scan(&user.UpdatedAt)
}

// SetPasswordResetToken is a no-op for MVP. Password reset tokens are JWT-based
// and don't need DB storage.
func (r *userRepo) SetPasswordResetToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	return nil
}

// UpdatePassword updates the password hash for a user.
func (r *userRepo) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`,
		userID, passwordHash,
	)
	return err
}

// SetEmailVerified marks a user's email as verified.
func (r *userRepo) SetEmailVerified(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET email_verified_at = NOW(), updated_at = NOW() WHERE id = $1`,
		userID,
	)
	return err
}

// Set2FASecret stores the TOTP secret for a user.
func (r *userRepo) Set2FASecret(ctx context.Context, userID, secret string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET two_factor_secret = $2, updated_at = NOW() WHERE id = $1`,
		userID, secret,
	)
	return err
}

// Enable2FA enables two-factor authentication for a user.
func (r *userRepo) Enable2FA(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET two_factor_enabled = TRUE, updated_at = NOW() WHERE id = $1`,
		userID,
	)
	return err
}

// Disable2FA disables two-factor authentication for a user.
func (r *userRepo) Disable2FA(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET two_factor_enabled = FALSE, two_factor_secret = NULL, updated_at = NOW() WHERE id = $1`,
		userID,
	)
	return err
}
