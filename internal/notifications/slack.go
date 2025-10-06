package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type SlackNotifier struct {
	webhook string
}

type SlackMessage struct {
	Text string `json:"text"`
}

func (s *SlackNotifier) Send(ctx context.Context, message string) error {
	webhookURL := s.webhook
	if webhookURL == "" {
		webhookURL = os.Getenv("PRICETREK_SLACK_WEBHOOK")
	}
	if webhookURL == "" {
		return fmt.Errorf("slack webhook URL not configured")
	}

	// Create message
	slackMsg := SlackMessage{
		Text: fmt.Sprintf("🔔 *PriceTrek Alert*\n%s", message),
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(slackMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Send request
	resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}