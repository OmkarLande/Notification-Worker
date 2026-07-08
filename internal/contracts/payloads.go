package contracts

// InsightSeverity represents the categorization of a generated insight.
type InsightSeverity string

const (
	SeverityInfo    InsightSeverity = "info"
	SeverityWarning InsightSeverity = "warning"
	SeveritySuccess InsightSeverity = "success"
)

// InsightResult contains rule-based deterministic insights generated from
// the provider's execution output.
type InsightResult struct {
	Title    string
	Summary  string
	Severity InsightSeverity
	Metadata map[string]any
}

// Payload is a marker interface for channel-specific payloads.
// It ensures compile-time safety and prevents passing arbitrary data to channels.
type Payload interface {
	Channel() string
}

// EmailPayload represents a fully rendered email ready for SMTP delivery.
type EmailPayload struct {
	To      string
	Subject string
	Html    string
}

func (EmailPayload) Channel() string { return "email" }

// DiscordPayload represents a fully rendered payload ready for Discord Webhooks.
type DiscordPayload struct {
	WebhookURL string
	Content    string
	Embeds     []any
}

func (DiscordPayload) Channel() string { return "discord" }

// SlackPayload represents a fully rendered payload ready for Slack Webhooks.
type SlackPayload struct {
	Blocks []any
}

func (SlackPayload) Channel() string { return "slack" }

// WhatsAppPayload represents a fully rendered text payload for WhatsApp.
type WhatsAppPayload struct {
	Text string
}

func (WhatsAppPayload) Channel() string { return "whatsapp" }
