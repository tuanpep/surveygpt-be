package services

import (
	"context"
	"fmt"

	resend "github.com/resendlabs/resend-go"
)

// EmailService handles sending emails via the Resend API.
type EmailService struct {
	client *resend.Client
	from   string
}

// NewEmailService creates a new EmailService. If apiKey is empty, emails are logged only.
func NewEmailService(apiKey, from string) *EmailService {
	if apiKey == "" {
		return &EmailService{client: nil, from: from}
	}
	return &EmailService{
		client: resend.NewClient(apiKey),
		from:   from,
	}
}

// SendEmail sends a single email.
func (s *EmailService) SendEmail(ctx context.Context, to, subject, htmlBody, textBody string) error {
	if s.client == nil {
		// Log-only mode when no API key is configured.
		return nil
	}

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email to %s: %w", to, err)
	}

	return nil
}

// SendBulkEmail sends the same email to multiple recipients.
// It respects rate limits by sending in batches.
func (s *EmailService) SendBulkEmail(ctx context.Context, recipients []string, subject, htmlBody, textBody string) error {
	if s.client == nil {
		return nil
	}

	const batchSize = 50 // Resend rate limit: 50 emails per request

	for i := 0; i < len(recipients); i += batchSize {
		end := i + batchSize
		if end > len(recipients) {
			end = len(recipients)
		}
		batch := recipients[i:end]

		params := &resend.SendEmailRequest{
			From:    s.from,
			To:      batch,
			Subject: subject,
			Html:    htmlBody,
			Text:    textBody,
		}

		_, err := s.client.Emails.Send(params)
		if err != nil {
			return fmt.Errorf("failed to send bulk email batch %d: %w", i/batchSize+1, err)
		}
	}

	return nil
}
