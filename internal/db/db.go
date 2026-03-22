package db

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/surveyflow/be/internal/config"
)

// NewPool creates a new PostgreSQL connection pool from the given configuration.
// It parses the DATABASE_URL and applies sensible pool defaults for connection
// management. Callers must call pool.Close() to release resources.
func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	// Apply production-appropriate pool defaults.
	poolConfig.MaxConns = cfg.DB.MaxConns
	poolConfig.MinConns = cfg.DB.MinConns
	// pgxpool default is 30s; only override if explicitly configured.

	// Enforce SSL in production if not explicitly configured via DATABASE_URL.
	if cfg.IsProduction() && poolConfig.ConnConfig.TLSConfig == nil {
		poolConfig.ConnConfig.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}
