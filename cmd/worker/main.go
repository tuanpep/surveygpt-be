package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/surveyflow/be/internal/config"
	"github.com/surveyflow/be/internal/db"
	"github.com/surveyflow/be/internal/workers"
)

func main() {
	if err := run(); err != nil {
		slog.Error("worker exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Set up structured logging based on environment.
	setupLogger(cfg)

	slog.Info("starting worker",
		"env", cfg.App.Env,
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

	// Initialize Asynq client for enqueueing retry tasks.
	asynqRedisOpt := asynqRedisOptFromURL(cfg.RedisURL)
	asynqClient := asynq.NewClient(asynqRedisOpt)
	defer asynqClient.Close()
	workers.SetAsynqClient(asynqClient)

	// Initialize worker dependencies.
	workers.InitEmailService(cfg)
	workers.InitAnthropicClient(cfg)

	// Create Asynq server with prioritized queues.
	// critical: high priority (6 concurrency), default: normal (3), low: background (1).
	srv := asynq.NewServer(
		asynqRedisOpt,
		asynq.Config{
			Concurrency:     10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				// Exponential backoff: 10s, 20s, 40s, 80s, ... capped at 1 hour.
				d := time.Second * time.Duration(10<<(n-1))
				if d > time.Hour {
					d = time.Hour
				}
				return d
			},
			IsFailure: func(err error) bool {
				// Treat all errors as retryable failures.
				return err != nil
			},
			Logger: newAsynqLogger(),
		},
	)

	// Register task handlers.
	mux := asynq.NewServeMux()

	// Email tasks.
	mux.HandleFunc(workers.TypeSendEmail, workers.HandleSendEmail)
	mux.HandleFunc(workers.TypeSendBulkEmail, workers.HandleSendBulkEmail)

	// AI analysis tasks.
	mux.HandleFunc(workers.TypeAIAnalysis, workers.HandleAIAnalysis)
	mux.HandleFunc(workers.TypeAISentiment, workers.HandleAISentiment)

	// Webhook tasks.
	mux.HandleFunc(workers.TypeWebhookDispatch, workers.HandleWebhookDispatch)
	mux.HandleFunc(workers.TypeWebhookRetry, workers.HandleWebhookRetry)

	// Analytics tasks.
	mux.HandleFunc(workers.TypeAnalyticsCache, workers.HandleAnalyticsCache)

	// Start the worker server in a goroutine.
	go func() {
		slog.Info("worker listening for tasks")
		if err := srv.Run(mux); err != nil {
			slog.Error("worker error", "error", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down worker...")

	// Graceful shutdown.
	srv.Shutdown()

	slog.Info("worker stopped")

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

// asynqLogger adapts slog to asynq's Logger interface.
type asynqLogger struct{}

func newAsynqLogger() *asynqLogger {
	return &asynqLogger{}
}

func (l *asynqLogger) Debug(args ...any) {
	slog.Debug(fmt.Sprint(args...))
}

func (l *asynqLogger) Info(args ...any) {
	slog.Info(fmt.Sprint(args...))
}

func (l *asynqLogger) Warn(args ...any) {
	slog.Warn(fmt.Sprint(args...))
}

func (l *asynqLogger) Error(args ...any) {
	slog.Error(fmt.Sprint(args...))
}

func (l *asynqLogger) Fatal(args ...any) {
	slog.Error(fmt.Sprint(args...))
	os.Exit(1)
}

// asynqRedisOptFromURL parses a Redis URL and returns an asynq.RedisClientOpt.
func asynqRedisOptFromURL(rawURL string) asynq.RedisClientOpt {
	opts, err := redis.ParseURL(rawURL)
	if err != nil {
		return asynq.RedisClientOpt{Addr: rawURL}
	}
	return asynq.RedisClientOpt{
		Addr:     opts.Addr,
		Username: opts.Username,
		Password: opts.Password,
		DB:      opts.DB,
	}
}
