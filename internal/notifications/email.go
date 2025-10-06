package notifications

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/gomail.v2"
)

type EmailNotifier struct {
	from string
	to   []string
}

func (e *EmailNotifier) Send(ctx context.Context, message string) error {
	// Get SMTP configuration from environment
	smtpHost := os.Getenv("PRICETREK_EMAIL_SMTP")
	smtpPort := os.Getenv("PRICETREK_EMAIL_PORT")
	smtpUser := os.Getenv("PRICETREK_EMAIL_USER")
	smtpPass := os.Getenv("PRICETREK_EMAIL_PASS")

	if smtpHost == "" {
		return fmt.Errorf("PRICETREK_EMAIL_SMTP environment variable not set")
	}
	if smtpUser == "" {
		return fmt.Errorf("PRICETREK_EMAIL_USER environment variable not set")
	}
	if smtpPass == "" {
		return fmt.Errorf("PRICETREK_EMAIL_PASS environment variable not set")
	}

	// Default port
	if smtpPort == "" {
		smtpPort = "587"
	}

	// Create message
	m := gomail.NewMessage()
	m.SetHeader("From", e.from)
	m.SetHeader("To", e.to...)
	m.SetHeader("Subject", "PriceTrek Alert")
	m.SetBody("text/html", fmt.Sprintf(`
		<html>
		<body>
			<h2>PriceTrek Alert</h2>
			<p>%s</p>
			<hr>
			<p><small>This is an automated message from PriceTrek</small></p>
		</body>
		</html>
	`, message))

	// Send email
	d := gomail.NewDialer(smtpHost, 587, smtpUser, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}