package workers

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/surveyflow/be/internal/config"
	"github.com/surveyflow/be/internal/services"
)

// emailWorker holds the email service instance.
var emailWorker *services.EmailService

// InitEmailService initializes the email service with the provided config.
// Should be called once during worker startup.
func InitEmailService(cfg *config.Config) {
	emailWorker = services.NewEmailService(cfg.Resend.APIKey, cfg.Resend.EmailFrom)
}

// HandleSendEmail processes email sending tasks from the queue.
func HandleSendEmail(ctx context.Context, t *asynq.Task) error {
	var payload EmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		slog.Error("failed to unmarshal email payload",
			"error", err,
			"task_id", t.Type(),
		)
		return err
	}

	start := time.Now()
	slog.Info("sending email",
		"to", payload.To,
		"subject", payload.Subject,
	)

	if emailWorker == nil {
		slog.Warn("email service not initialized, skipping email send")
		return nil
	}

	err := emailWorker.SendEmail(ctx, payload.To, payload.Subject, payload.HTMLBody, payload.TextBody)
	if err != nil {
		slog.Error("failed to send email",
			"error", err,
			"to", payload.To,
			"subject", payload.Subject,
			"duration", time.Since(start),
		)
		return err
	}

	slog.Info("email sent successfully",
		"to", payload.To,
		"subject", payload.Subject,
		"duration", time.Since(start),
	)

	return nil
}

// HandleSendBulkEmail processes bulk email sending tasks.
func HandleSendBulkEmail(ctx context.Context, t *asynq.Task) error {
	var payload BulkEmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		slog.Error("failed to unmarshal bulk email payload",
			"error", err,
			"task_id", t.Type(),
		)
		return err
	}

	start := time.Now()
	slog.Info("sending bulk email",
		"recipient_count", len(payload.Recipients),
		"subject", payload.Subject,
		"campaign_id", payload.CampaignID,
	)

	if emailWorker == nil {
		slog.Warn("email service not initialized, skipping bulk email send")
		return nil
	}

	err := emailWorker.SendBulkEmail(ctx, payload.Recipients, payload.Subject, payload.HTMLBody, payload.TextBody)
	if err != nil {
		slog.Error("failed to send bulk email",
			"error", err,
			"recipient_count", len(payload.Recipients),
			"duration", time.Since(start),
		)
		return err
	}

	slog.Info("bulk email sent successfully",
		"recipient_count", len(payload.Recipients),
		"duration", time.Since(start),
	)

	return nil
}
