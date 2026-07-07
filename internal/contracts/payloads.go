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

// EmailPayload represents a fully rendered email ready for SMTP delivery.
type EmailPayload struct {
	Subject string
	Html    string
}

// DiscordPayload represents a fully rendered payload ready for Discord Webhooks.
type DiscordPayload struct {
	Content string
	Embeds  []any
}

// SlackPayload represents a fully rendered payload ready for Slack Webhooks.
type SlackPayload struct {
	Blocks []any
}

// WhatsAppPayload represents a fully rendered text payload for WhatsApp.
type WhatsAppPayload struct {
	Text string
}
