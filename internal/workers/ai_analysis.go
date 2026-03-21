package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/hibiken/asynq"
	"github.com/surveyflow/be/internal/config"
)

// AIWorker holds dependencies for AI analysis tasks.
type AIWorker struct {
	client    anthropic.Client
	msgService anthropic.MessageService
	available bool
}

var aiWorker *AIWorker

// InitAnthropicClient initializes the Anthropic API client.
// Should be called once during worker startup.
func InitAnthropicClient(cfg *config.Config) {
	aiWorker = &AIWorker{}
	if cfg.AnthropicAPIKey == "" {
		slog.Warn("Anthropic API key not configured, AI analysis will be skipped")
		aiWorker.available = false
		return
	}
	aiWorker.client = anthropic.NewClient(option.WithAPIKey(cfg.AnthropicAPIKey))
	aiWorker.msgService = anthropic.NewMessageService(option.WithAPIKey(cfg.AnthropicAPIKey))
	aiWorker.available = true
}

// HandleAIAnalysis processes AI analysis tasks for survey responses.
// It performs sentiment analysis and theme extraction on text answers.
func HandleAIAnalysis(ctx context.Context, t *asynq.Task) error {
	var payload AIAnalysisPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		slog.Error("failed to unmarshal AI analysis payload",
			"error", err,
			"task_id", t.Type(),
		)
		return err
	}

	start := time.Now()
	slog.Info("processing AI analysis",
		"survey_id", payload.SurveyID,
		"response_id", payload.ResponseID,
	)

	if aiWorker == nil || !aiWorker.available {
		slog.Warn("Anthropic client not initialized, skipping AI analysis")
		return nil
	}

	// TODO: Implement AI analysis via Anthropic API.
	// 1. If ResponseID is provided, analyze a single response.
	//    Otherwise, analyze all responses for the survey.
	// 2. Fetch the survey structure and response(s) from the database.
	// 3. For each text-type question (short_text, long_text):
	//    a. Collect all text answers.
	//    b. Build a prompt requesting sentiment analysis and theme extraction.
	//    c. Call the Anthropic API using aiWorker.msgService.New().
	//    d. Parse the response (JSON format: {sentiment, themes, summary}).
	//    e. Store the analysis result in the response's ai_analysis JSONB field.

	slog.Info("AI analysis completed",
		"survey_id", payload.SurveyID,
		"duration", time.Since(start),
	)

	return nil
}

// HandleAISentiment processes sentiment analysis tasks for individual text answers.
func HandleAISentiment(ctx context.Context, t *asynq.Task) error {
	var payload AISentimentPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		slog.Error("failed to unmarshal AI sentiment payload",
			"error", err,
			"task_id", t.Type(),
		)
		return err
	}

	slog.Info("processing AI sentiment analysis",
		"response_id", payload.ResponseID,
		"question_id", payload.QuestionID,
	)

	if aiWorker == nil || !aiWorker.available {
		slog.Warn("Anthropic client not initialized, skipping sentiment analysis")
		return nil
	}

	// TODO: Implement single-text sentiment analysis.
	// Use Haiku for fast, cheap sentiment classification.

	slog.Info("sentiment analysis completed",
		"response_id", payload.ResponseID,
		"question_id", payload.QuestionID,
	)

	return nil
}

// callAnthropic is a helper that calls the Anthropic API with the given prompt.
// It returns the raw text response from the model.
func callAnthropic(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (string, error) {
	if aiWorker == nil || !aiWorker.available {
		return "", fmt.Errorf("Anthropic client not initialized")
	}

	model := anthropic.ModelClaudeSonnet4_5
	if maxTokens <= 256 {
		model = anthropic.ModelClaudeHaiku4_5
	}

	message := anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: int64(maxTokens),
		System:    []anthropic.TextBlockParam{{Type: "text", Text: systemPrompt}},
		Messages: []anthropic.MessageParam{{
			Role: "user",
			Content: []anthropic.ContentBlockParamUnion{
				{OfText: &anthropic.TextBlockParam{Type: "text", Text: userPrompt}},
			},
		}},
	}

	resp, err := aiWorker.msgService.New(ctx, message)
	if err != nil {
		return "", fmt.Errorf("Anthropic API call failed: %w", err)
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("empty response from Anthropic API")
	}

	// Extract text from content blocks
	for _, block := range resp.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}

	return "", fmt.Errorf("no text content in Anthropic response")
}
