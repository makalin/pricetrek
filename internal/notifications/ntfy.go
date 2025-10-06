package notifications

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type NtfyNotifier struct {
	topic string
}

func (n *NtfyNotifier) Send(ctx context.Context, message string) error {
	ntfyURL := os.Getenv("PRICETREK_NTFY_URL")
	if ntfyURL == "" {
		ntfyURL = "https://ntfy.sh"
	}

	// Create URL
	apiURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(ntfyURL, "/"), n.topic)

	// Create HTTP client
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(message))
	if err != nil {
		return fmt.Errorf("failed to create ntfy request: %w", err)
	}

	// Set headers
	req.Header.Set("Title", "PriceTrek Alert")
	req.Header.Set("Priority", "default")
	req.Header.Set("Tags", "price,alert")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send ntfy message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ntfy returned status %d", resp.StatusCode)
	}

	return nil
}