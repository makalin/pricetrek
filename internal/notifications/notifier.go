package notifications

import (
	"context"
	"fmt"
	"os"

	"github.com/makalin/pricetrek/internal/config"
)

type Notifier interface {
	Send(ctx context.Context, message string) error
}

type NotificationManager struct {
	notifiers []Notifier
}

func New(cfg *config.Config) *NotificationManager {
	var notifiers []Notifier

	// Email notifier
	if cfg.Notifications.Email.Enabled {
		notifiers = append(notifiers, &EmailNotifier{
			from: cfg.Notifications.Email.From,
			to:   cfg.Notifications.Email.To,
		})
	}

	// Telegram notifier
	if cfg.Notifications.Telegram.Enabled {
		notifiers = append(notifiers, &TelegramNotifier{
			chatID: cfg.Notifications.Telegram.ChatID,
		})
	}

	// Slack notifier
	if cfg.Notifications.Slack.Enabled {
		notifiers = append(notifiers, &SlackNotifier{
			webhook: cfg.Notifications.Slack.Webhook,
		})
	}

	// Ntfy notifier
	if cfg.Notifications.Ntfy.Enabled {
		notifiers = append(notifiers, &NtfyNotifier{
			topic: cfg.Notifications.Ntfy.Topic,
		})
	}

	return &NotificationManager{
		notifiers: notifiers,
	}
}

func (nm *NotificationManager) Send(ctx context.Context, message string) error {
	for _, notifier := range nm.notifiers {
		if err := notifier.Send(ctx, message); err != nil {
			// Log error but continue with other notifiers
			fmt.Fprintf(os.Stderr, "Failed to send notification: %v\n", err)
		}
	}
	return nil
}

type EmailNotifier struct {
	from string
	to   []string
}

func (e *EmailNotifier) Send(ctx context.Context, message string) error {
	// TODO: Implement email notification
	return fmt.Errorf("email notification not implemented yet")
}

type TelegramNotifier struct {
	chatID string
}

func (t *TelegramNotifier) Send(ctx context.Context, message string) error {
	// TODO: Implement telegram notification
	return fmt.Errorf("telegram notification not implemented yet")
}

type SlackNotifier struct {
	webhook string
}

func (s *SlackNotifier) Send(ctx context.Context, message string) error {
	// TODO: Implement slack notification
	return fmt.Errorf("slack notification not implemented yet")
}

type NtfyNotifier struct {
	topic string
}

func (n *NtfyNotifier) Send(ctx context.Context, message string) error {
	// TODO: Implement ntfy notification
	return fmt.Errorf("ntfy notification not implemented yet")
}