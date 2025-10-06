package notifications

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type TelegramNotifier struct {
	chatID string
}

func (t *TelegramNotifier) Send(ctx context.Context, message string) error {
	token := os.Getenv("PRICETREK_TELEGRAM_TOKEN")
	if token == "" {
		return fmt.Errorf("PRICETREK_TELEGRAM_TOKEN environment variable not set")
	}

	// Create API URL
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	
	// Prepare form data
	data := url.Values{}
	data.Set("chat_id", t.chatID)
	data.Set("text", message)
	data.Set("parse_mode", "HTML")

	// Create HTTP client
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Send request
	resp, err := client.PostForm(apiURL, data)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	return nil
}