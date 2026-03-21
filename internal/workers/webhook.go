package workers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
)

const (
	maxWebhookRetries = 6
	webhookTimeout   = 10 * time.Second
)

// asynqClient is the client used to enqueue retry tasks.
// Must be set via SetAsynqClient before HandleWebhookDispatch is called.
var asynqClient *asynq.Client

// SetAsynqClient sets the asynq client for webhook retry scheduling.
func SetAsynqClient(client *asynq.Client) {
	asynqClient = client
}

// HandleWebhookDispatch processes outgoing webhook delivery tasks.
func HandleWebhookDispatch(ctx context.Context, t *asynq.Task) error {
	var payload WebhookPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		slog.Error("failed to unmarshal webhook payload",
			"error", err,
			"task_id", t.Type(),
		)
		return err
	}

	slog.Info("dispatching webhook",
		"webhook_id", payload.WebhookID,
		"event", payload.Event,
		"org_id", payload.OrgID,
	)

	// Build the webhook body
	body, err := json.Marshal(payload.Payload)
	if err != nil {
		slog.Error("failed to marshal webhook payload",
			"error", err,
			"webhook_id", payload.WebhookID,
		)
		return err
	}

	// Sign the payload with HMAC-SHA256
	signature := computeHMAC(payload.Secret, body)

	// TODO: In production, fetch the actual webhook URL from the database.
	// For now, the URL should be passed in the payload metadata.
	webhookURL, ok := payload.Payload["_webhook_url"].(string)
	if !ok {
		slog.Warn("no webhook URL found in payload, skipping delivery",
			"webhook_id", payload.WebhookID,
		)
		return nil
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		slog.Error("failed to create webhook request",
			"error", err,
			"webhook_id", payload.WebhookID,
		)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Event", payload.Event)
	req.Header.Set("X-Webhook-Delivery", fmt.Sprintf("%d", time.Now().UnixMilli()))
	req.Header.Set("User-Agent", "SurveyFlow-Webhooks/1.0")

	client := &http.Client{Timeout: webhookTimeout}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("webhook delivery failed",
			"error", err,
			"webhook_id", payload.WebhookID,
			"url", webhookURL,
		)
		// Schedule retry
		if payload.RetryCount < maxWebhookRetries {
			return scheduleRetry(ctx, payload)
		}
		return fmt.Errorf("webhook delivery failed after %d retries: %w", payload.RetryCount, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		slog.Warn("webhook endpoint returned error",
			"webhook_id", payload.WebhookID,
			"status_code", resp.StatusCode,
			"response", string(respBody),
		)
		if payload.RetryCount < maxWebhookRetries {
			return scheduleRetry(ctx, payload)
		}
		return fmt.Errorf("webhook endpoint returned status %d after %d retries", resp.StatusCode, payload.RetryCount)
	}

	slog.Info("webhook delivered successfully",
		"webhook_id", payload.WebhookID,
		"status_code", resp.StatusCode,
	)

	return nil
}

// HandleWebhookRetry processes retry attempts for failed webhook deliveries.
func HandleWebhookRetry(ctx context.Context, t *asynq.Task) error {
	var payload WebhookPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		slog.Error("failed to unmarshal webhook retry payload",
			"error", err,
			"task_id", t.Type(),
		)
		return err
	}

	if payload.RetryCount >= maxWebhookRetries {
		slog.Warn("webhook retry limit exceeded, giving up",
			"webhook_id", payload.WebhookID,
			"retry_count", payload.RetryCount,
		)
		return nil
	}

	slog.Info("processing webhook retry",
		"webhook_id", payload.WebhookID,
		"retry_count", payload.RetryCount,
	)

	// Calculate exponential backoff: 30s, 1m, 2m, 4m, 8m, 16m
	backoff := time.Duration(30*(1<<payload.RetryCount)) * time.Second
	if backoff > 30*time.Minute {
		backoff = 30 * time.Minute
	}

	// Schedule the retry
	retryPayload := payload
	retryPayload.RetryCount = payload.RetryCount + 1

	task, err := json.Marshal(retryPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal retry payload: %w", err)
	}

	if asynqClient == nil {
		slog.Error("asynq client not initialized, cannot schedule webhook retry",
			"webhook_id", payload.WebhookID,
		)
		return fmt.Errorf("asynq client not configured")
	}

	_, err = asynqClient.Enqueue(
		asynq.NewTask(TypeWebhookRetry, task),
		asynq.ProcessIn(backoff),
		asynq.Queue("default"),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue webhook retry: %w", err)
	}

	slog.Info("webhook retry scheduled",
		"webhook_id", payload.WebhookID,
		"retry_count", retryPayload.RetryCount,
		"backoff", backoff,
	)

	return nil
}

// computeHMAC generates an HMAC-SHA256 signature for webhook payload verification.
func computeHMAC(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// scheduleRetry creates a retry task with incremented retry count.
func scheduleRetry(ctx context.Context, payload WebhookPayload) error {
	retryPayload := payload
	retryPayload.RetryCount++

	task, err := json.Marshal(retryPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal retry payload: %w", err)
	}

	if asynqClient == nil {
		slog.Error("asynq client not initialized, cannot schedule webhook retry",
			"webhook_id", payload.WebhookID,
		)
		return fmt.Errorf("asynq client not configured")
	}

	// Calculate exponential backoff: 30s, 1m, 2m, 4m, 8m, 16m
	backoff := time.Duration(30*(1<<payload.RetryCount)) * time.Second
	if backoff > 30*time.Minute {
		backoff = 30 * time.Minute
	}

	_, err = asynqClient.Enqueue(
		asynq.NewTask(TypeWebhookRetry, task),
		asynq.ProcessIn(backoff),
		asynq.Queue("default"),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue webhook retry: %w", err)
	}

	slog.Info("webhook retry scheduled",
		"webhook_id", payload.WebhookID,
		"retry_count", retryPayload.RetryCount,
		"backoff", backoff,
	)
	return nil
}
