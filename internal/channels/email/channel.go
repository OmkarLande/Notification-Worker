package email

import (
	"bytes"
	"context"
	"fmt"
	"net/smtp"
	"time"

	"github.com/OmkarLande/notification-worker/internal/channels"
	"github.com/OmkarLande/notification-worker/internal/config"
	"github.com/OmkarLande/notification-worker/internal/contracts"
)

type EmailChannel struct {
	cfg config.SMTPConfig
}

func NewEmailChannel(cfg config.SMTPConfig) *EmailChannel {
	return &EmailChannel{
		cfg: cfg,
	}
}

func (c *EmailChannel) Name() string {
	return "email"
}

func (c *EmailChannel) Validate(ctx context.Context, payload contracts.Payload) error {
	p, ok := payload.(*contracts.EmailPayload)
	if !ok {
		return fmt.Errorf("email channel requires *contracts.EmailPayload")
	}

	if p.Html == "" {
		return fmt.Errorf("email html content is empty")
	}

	// For demo, if To is empty, we fallback to a default from env or config.
	// We'll let Deliver handle it if it's completely empty.
	return nil
}

func (c *EmailChannel) Deliver(ctx context.Context, payload contracts.Payload) (*channels.DeliveryResult, error) {
	start := time.Now()

	p, ok := payload.(*contracts.EmailPayload)
	if !ok {
		return nil, fmt.Errorf("invalid payload type")
	}

	if err := c.Validate(ctx, payload); err != nil {
		return &channels.DeliveryResult{
			Success:      false,
			Channel:      c.Name(),
			Duration:     time.Since(start),
			ErrorMessage: err.Error(),
		}, nil
	}

	to := p.To
	if to == "" {
		// Mock recipient for demo if not provided
		to = "test@example.com"
	}

	auth := smtp.PlainAuth("", c.cfg.User, c.cfg.Password, c.cfg.Host)
	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", c.cfg.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", p.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(p.Html)

	err := smtp.SendMail(addr, auth, c.cfg.From, []string{to}, buf.Bytes())
	if err != nil {
		return &channels.DeliveryResult{
			Success:      false,
			Channel:      c.Name(),
			Duration:     time.Since(start),
			ErrorMessage: err.Error(),
		}, nil
	}

	return &channels.DeliveryResult{
		Success:           true,
		Channel:           c.Name(),
		Duration:          time.Since(start),
		ProviderMessageID: "smtp-sent", // Mocked ID
		Metadata: map[string]any{
			"payload_size": len(buf.Bytes()),
			"to":           to,
		},
	}, nil
}
