package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/OmkarLande/notification-worker/internal/channels"
	"github.com/OmkarLande/notification-worker/internal/contracts"
)

type DiscordChannel struct {
	client *http.Client
}

func NewDiscordChannel() *DiscordChannel {
	return &DiscordChannel{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *DiscordChannel) Name() string {
	return "discord"
}

func (c *DiscordChannel) Validate(ctx context.Context, payload contracts.Payload) error {
	p, ok := payload.(*contracts.DiscordPayload)
	if !ok {
		return fmt.Errorf("discord channel requires *contracts.DiscordPayload")
	}

	if p.Content == "" && len(p.Embeds) == 0 {
		return fmt.Errorf("discord payload requires content or embeds")
	}

	return nil
}

func (c *DiscordChannel) Deliver(ctx context.Context, payload contracts.Payload) (*channels.DeliveryResult, error) {
	start := time.Now()

	p, ok := payload.(*contracts.DiscordPayload)
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

	webhookURL := p.WebhookURL
	if webhookURL == "" {
		// Mock webhook for demo if not provided, though it will fail
		webhookURL = "https://discord.com/api/webhooks/mock/mock"
	}

	reqBody := map[string]any{
		"content": p.Content,
	}
	if len(p.Embeds) > 0 {
		reqBody["embeds"] = p.Embeds
	}

	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(b))
	if err != nil {
		return nil, err // fatal
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return &channels.DeliveryResult{
			Success:      false,
			Channel:      c.Name(),
			Duration:     time.Since(start),
			ErrorMessage: err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return &channels.DeliveryResult{
			Success:      false,
			Channel:      c.Name(),
			Duration:     time.Since(start),
			ErrorMessage: fmt.Sprintf("discord api error: %d %s", resp.StatusCode, string(bodyBytes)),
		}, nil
	}

	return &channels.DeliveryResult{
		Success:           true,
		Channel:           c.Name(),
		Duration:          time.Since(start),
		ProviderMessageID: resp.Header.Get("x-ratelimit-reset"), // Just a dummy
		Metadata: map[string]any{
			"payload_size": len(b),
		},
	}, nil
}
