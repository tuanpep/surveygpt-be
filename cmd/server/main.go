package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/surveyflow/be/internal/config"
	"github.com/surveyflow/be/internal/db"
	"github.com/surveyflow/be/internal/handlers"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Load .env file (ignore error if missing — env vars may be set directly).
	_ = godotenv.Load()

	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Set up structured logging based on environment.
	setupLogger(cfg)

	slog.Info("starting server",
		"env", cfg.App.Env,
		"port", cfg.App.Port,
	)

	// Initialize database connection pool.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		return fmt.Errorf("init database: %w", err)
	}
	defer pool.Close()

	slog.Info("connected to database")

	// Initialize Redis client.
	redisClient, err := db.NewRedisClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("init redis: %w", err)
	}
	defer redisClient.Close()

	slog.Info("connected to redis")

	// Set up routes with dependencies.
	deps := &handlers.Dependencies{
		Config: cfg,
		Pool:   pool,
		Redis:  redisClient,
	}

	e := handlers.SetupRoutes(deps)

	// Start the server in a goroutine.
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      e,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server listening", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	// Graceful shutdown with a 30-second timeout.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	slog.Info("server stopped")
	return nil
}

// setupLogger configures the global slog logger based on the app environment.
func setupLogger(cfg *config.Config) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel(cfg.App.Env),
	})
	slog.SetDefault(slog.New(handler))
}

// slogLevel returns the appropriate log level for the given environment.
func slogLevel(env string) slog.Level {
	switch env {
	case "production":
		return slog.LevelInfo
	case "test":
		return slog.LevelWarn
	default:
		return slog.LevelDebug
	}
}
