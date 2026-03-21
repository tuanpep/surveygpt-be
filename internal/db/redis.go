package db

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/surveyflow/be/internal/config"
)

// NewRedisClient creates a new Redis client from the given configuration.
// The client is connection-pooled and safe for concurrent use by multiple goroutines.
func NewRedisClient(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	// Apply sensible defaults for a production client.
	opts.PoolSize = 20
	opts.MinIdleConns = 5
	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second

	// Enable TLS in production if not already configured.
	if cfg.IsProduction() && opts.TLSConfig == nil {
		opts.TLSConfig = &tls.Config{}
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}
