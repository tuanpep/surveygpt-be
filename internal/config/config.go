package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration sourced from environment variables.
type Config struct {
	// App is the general application configuration.
	App struct {
		Env  string `json:"env"`
		Port int    `json:"port"`
	} `json:"app"`

	// DB holds database pool configuration.
	DB struct {
		MaxConns int32 `json:"max_conns"`
		MinConns int32 `json:"min_conns"`
	} `json:"db"`

	// Pagination holds default pagination settings.
	Pagination struct {
		DefaultLimit int `json:"default_limit"`
		MaxLimit     int `json:"max_limit"`
	} `json:"pagination"`

	// PlanDefaults holds default limits for the free plan.
	PlanDefaults struct {
		ResponseLimit int `json:"response_limit"`
		SurveyLimit   int `json:"survey_limit"`
		MemberLimit   int `json:"member_limit"`
		AICredits     int `json:"ai_credits"`
	} `json:"plan_defaults"`

	// Frontend is the frontend origin for CORS and redirects.
	FrontendURL string `json:"frontend_url"`

	// Database is the PostgreSQL connection URL.
	DatabaseURL string `json:"database_url"`

	// Redis is the Redis connection URL.
	RedisURL string `json:"redis_url"`

	// JWT holds JWT configuration.
	JWT struct {
		Secret     string        `json:"secret"`
		AccessTTL  time.Duration `json:"access_ttl"`
		RefreshTTL time.Duration `json:"refresh_ttl"`
	} `json:"jwt"`

	// Google holds Google OAuth credentials.
	Google struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	} `json:"google"`

	// Microsoft holds Microsoft OAuth credentials.
	Microsoft struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	} `json:"microsoft"`

	// Anthropic holds the Anthropic API key.
	AnthropicAPIKey string `json:"anthropic_api_key"`

	// Resend holds the Resend email service configuration.
	Resend struct {
		APIKey string `json:"api_key"`
		EmailFrom string `json:"email_from"`
	} `json:"resend"`

	// Stripe holds Stripe configuration.
	Stripe struct {
		SecretKey      string `json:"secret_key"`
		WebhookSecret  string `json:"webhook_secret"`
	} `json:"stripe"`

	// R2 holds Cloudflare R2 storage configuration.
	R2 struct {
		AccountID  string `json:"account_id"`
		AccessKey  string `json:"access_key"`
		SecretKey  string `json:"secret_key"`
		Bucket     string `json:"bucket"`
		PublicURL  string `json:"public_url"`
	} `json:"r2"`

	// Twilio holds Twilio SMS configuration.
	Twilio struct {
		AccountSID string `json:"account_sid"`
		AuthToken  string `json:"auth_token"`
	} `json:"twilio"`

	// SentryDSN is the Sentry error tracking DSN.
	SentryDSN string `json:"sentry_dsn"`
}

// Load reads configuration from environment variables with sensible defaults.
// It returns an error if required fields are missing.
func Load() (*Config, error) {
	cfg := &Config{}

	cfg.App.Env = getEnv("APP_ENV", "development")
	cfg.App.Port = getEnvInt("PORT", 8080)

	cfg.DB.MaxConns = int32(getEnvInt("DB_MAX_CONNS", 25))
	cfg.DB.MinConns = int32(getEnvInt("DB_MIN_CONNS", 5))

	cfg.Pagination.DefaultLimit = getEnvInt("PAGINATION_DEFAULT_LIMIT", 20)
	cfg.Pagination.MaxLimit = getEnvInt("PAGINATION_MAX_LIMIT", 100)

	cfg.PlanDefaults.ResponseLimit = getEnvInt("PLAN_RESPONSE_LIMIT", 100)
	cfg.PlanDefaults.SurveyLimit = getEnvInt("PLAN_SURVEY_LIMIT", 10)
	cfg.PlanDefaults.MemberLimit = getEnvInt("PLAN_MEMBER_LIMIT", 5)
	cfg.PlanDefaults.AICredits = getEnvInt("PLAN_AI_CREDITS", 50)

	cfg.FrontendURL = getEnv("FRONTEND_URL", "http://localhost:3000")
	cfg.DatabaseURL = getEnv("DATABASE_URL", "")
	cfg.RedisURL = getEnv("REDIS_URL", "redis://localhost:6379/0")

	cfg.JWT.Secret = getEnv("JWT_SECRET", "")
	cfg.JWT.AccessTTL = getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute)
	cfg.JWT.RefreshTTL = getEnvDuration("JWT_REFRESH_TTL", 7*24*time.Hour)

	cfg.Google.ClientID = getEnv("GOOGLE_CLIENT_ID", "")
	cfg.Google.ClientSecret = getEnv("GOOGLE_CLIENT_SECRET", "")

	cfg.Microsoft.ClientID = getEnv("MICROSOFT_CLIENT_ID", "")
	cfg.Microsoft.ClientSecret = getEnv("MICROSOFT_CLIENT_SECRET", "")

	cfg.AnthropicAPIKey = getEnv("ANTHROPIC_API_KEY", "")

	cfg.Resend.APIKey = getEnv("RESEND_API_KEY", "")
	cfg.Resend.EmailFrom = getEnv("EMAIL_FROM", "noreply@surveyflow.app")

	cfg.Stripe.SecretKey = getEnv("STRIPE_SECRET_KEY", "")
	cfg.Stripe.WebhookSecret = getEnv("STRIPE_WEBHOOK_SECRET", "")

	cfg.R2.AccountID = getEnv("R2_ACCOUNT_ID", "")
	cfg.R2.AccessKey = getEnv("R2_ACCESS_KEY", "")
	cfg.R2.SecretKey = getEnv("R2_SECRET_KEY", "")
	cfg.R2.Bucket = getEnv("R2_BUCKET", "surveyflow")
	cfg.R2.PublicURL = getEnv("R2_PUBLIC_URL", "")

	cfg.Twilio.AccountSID = getEnv("TWILIO_ACCOUNT_SID", "")
	cfg.Twilio.AuthToken = getEnv("TWILIO_AUTH_TOKEN", "")

	cfg.SentryDSN = getEnv("SENTRY_DSN", "")

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks that required configuration values are set.
func (c *Config) validate() error {
	required := map[string]string{
		"DATABASE_URL": c.DatabaseURL,
		"JWT_SECRET":   c.JWT.Secret,
	}

	for name, value := range required {
		if value == "" {
			return fmt.Errorf("required environment variable %s is not set", name)
		}
	}

	if c.App.Env == "production" {
		if c.JWT.Secret == "change-me-in-production" || c.JWT.Secret == "" {
			return fmt.Errorf("JWT_SECRET must be set to a non-default value in production")
		}
		if c.FrontendURL == "" || c.FrontendURL == "http://localhost:3000" {
			return fmt.Errorf("FRONTEND_URL must be set in production")
		}
		if _, err := url.Parse(c.FrontendURL); err != nil {
			return fmt.Errorf("FRONTEND_URL is not a valid URL: %w", err)
		}
	}

	return nil
}

// IsDevelopment returns true if the application is running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction returns true if the application is running in production mode.
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// IsTest returns true if the application is running in test mode.
func (c *Config) IsTest() bool {
	return c.App.Env == "test"
}

// getEnv reads an environment variable with a default value.
func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

// getEnvInt reads an environment variable as an integer with a default value.
func getEnvInt(key string, defaultValue int) int {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}

// getEnvDuration reads an environment variable as a duration string with a default value.
// Accepted format: "15m", "1h", "168h" (Go duration format).
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return d
}

// getEnvBool reads an environment variable as a boolean with a default value.
func getEnvBool(key string, defaultValue bool) bool {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return b
}

// getEnvSlice reads an environment variable as a comma-separated string slice.
func getEnvSlice(key string, defaultValue []string) []string {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	return strings.Split(value, ",")
}
