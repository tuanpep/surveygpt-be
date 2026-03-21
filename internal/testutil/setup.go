package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/surveyflow/be/internal/config"
	"github.com/surveyflow/be/internal/db"
)

var (
	TestPool   *pgxpool.Pool
	TestRedis  *redis.Client
	TestConfig *config.Config
)

// TestMain initializes the test database, Redis, and server before all tests.
// This must be called from a TestMain in each test package, or via Init().
func Init() {
	os.Setenv("APP_ENV", "test")
	os.Setenv("DATABASE_URL", "postgresql://surveyflow:devpassword@localhost:5432/surveyflow_test?sslmode=disable")
	os.Setenv("REDIS_URL", "redis://localhost:6379/1")
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")
	os.Setenv("FRONTEND_URL", "http://localhost:5173")

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}
	TestConfig = cfg

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init database: %v\n", err)
		os.Exit(1)
	}
	TestPool = pool

	redisClient, err := db.NewRedisClient(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init redis: %v\n", err)
		os.Exit(1)
	}
	TestRedis = redisClient

	if err := runMigrations(pool); err != nil {
		fmt.Fprintf(os.Stderr, "run migrations: %v\n", err)
		os.Exit(1)
	}
}

// Cleanup closes database and Redis connections.
func Cleanup() {
	if TestPool != nil {
		TestPool.Close()
	}
	if TestRedis != nil {
		TestRedis.Close()
	}
}

// RunMigrations ensures all tables exist. Safe to call multiple times.
// Ignores "already exists" errors.
func RunMigrations() error {
	return applyMigrations(TestPool, false)
}

// runMigrations drops all tables and recreates them from scratch.
func runMigrations(pool *pgxpool.Pool) error {
	return applyMigrations(pool, true)
}

// applyMigrations applies migration files. If dropFirst is true, drops existing tables first.
func applyMigrations(pool *pgxpool.Pool, dropFirst bool) error {
	// Find the module root by locating go.mod from this file's directory.
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("could not determine source file path")
	}
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return fmt.Errorf("could not find module root (go.mod)")
		}
		dir = parent
	}
	migrationsDir := filepath.Join(dir, "internal", "db", "migrations")

	ctx := context.Background()

	if dropFirst {
		pool.Exec(ctx, "DROP TABLE IF EXISTS schema_migrations CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS api_keys, audit_logs, analytics_cache, answers, email_contacts, email_lists, files, integrations, invitations, org_memberships, organizations, responses, surveys, templates, usage_logs, users, webhooks, ai_credits CASCADE")
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint PRIMARY KEY,
			dirty boolean NOT NULL DEFAULT FALSE
		)
	`)

	var upFiles []string
	for _, e := range entries {
		if matched, _ := matchUpFile(e.Name()); matched {
			upFiles = append(upFiles, e.Name())
		}
	}

	for _, fname := range upFiles {
		path := migrationsDir + "/" + fname
		sql, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", fname, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			if !dropFirst {
				continue // already exists from Init()
			}
			return fmt.Errorf("execute migration %s: %w", fname, err)
		}
	}

	return nil
}

// matchUpFile checks if a filename is a golang-migrate up migration file.
func matchUpFile(name string) (bool, error) {
	return strings.HasSuffix(name, ".up.sql"), nil
}

// TruncateTables removes all data from user tables in FK-safe order.
// Uses DELETE FROM instead of TRUNCATE CASCADE to avoid dropping tables.
// Templates are NOT truncated (system seed data).
func TruncateTables(t *testing.T) {
	t.Helper()

	// Flush Redis to clear rate-limit counters and other ephemeral state.
	if err := TestRedis.FlushDB(context.Background()).Err(); err != nil {
		t.Logf("flush redis: %v", err)
	}

	tables := []string{
		"analytics_cache",
		"usage_logs",
		"audit_logs",
		"ai_credits",
		"answers",
		"responses",
		"email_contacts",
		"email_lists",
		"invitations",
		"org_memberships",
		"organizations",
		"users",
		"surveys",
		"webhooks",
		"files",
		"integrations",
		"api_keys",
	}

	ctx := context.Background()
	for _, table := range tables {
		TestPool.Exec(ctx, "DELETE FROM "+table)
	}
}
