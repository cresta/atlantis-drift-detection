package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type SlackWebhook struct {
	WebhookURL string
	HTTPClient *http.Client
}

func (s *SlackWebhook) TemporaryError(ctx context.Context, dir string, workspace string, err error) error {
	return s.sendSlackMessage(ctx, fmt.Sprintf("Unknown error in remote\nDirectory: %s\nWorkspace: %s\nError: %s", dir, workspace, err.Error()))
}

func NewSlackWebhook(webhookURL string, HTTPClient *http.Client) *SlackWebhook {
	if webhookURL == "" {
		return nil
	}
	return &SlackWebhook{
		WebhookURL: webhookURL,
		HTTPClient: HTTPClient,
	}
}

type SlackWebhookMessage struct {
	Text string `json:"text"`
}

func (s *SlackWebhook) sendSlackMessage(ctx context.Context, msg string) error {
	body := SlackWebhookMessage{
		Text: msg,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal slack webhook message: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.WebhookURL, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create slack webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack webhook request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send slack webhook request: %w", err)
	}
	return nil
}

func (s *SlackWebhook) ExtraWorkspaceInRemote(ctx context.Context, dir string, workspace string) error {
	return s.sendSlackMessage(ctx, fmt.Sprintf("Extra workspace in remote\nDirectory: %s\nWorkspace: %s", dir, workspace))
}

func (s *SlackWebhook) MissingWorkspaceInRemote(ctx context.Context, dir string, workspace string) error {
	return s.sendSlackMessage(ctx, fmt.Sprintf("Missing workspace in remote\nDirectory: %s\nWorkspace: %s", dir, workspace))
}

func (s *SlackWebhook) PlanDrift(ctx context.Context, dir string, workspace string) error {
	return s.sendSlackMessage(ctx, fmt.Sprintf("Plan Drift workspace in remote\nDirectory: %s\nWorkspace: %s", dir, workspace))
}

var _ Notification = &SlackWebhook{}
