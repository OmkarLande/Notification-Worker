package config

// N8NConfig holds settings for integrating with n8n workflow automation
// webhooks. All fields are optional; the n8n channel provider validates them
// when it is initialized.
type N8NConfig struct {
	// WebhookBaseURL is the base URL of the n8n instance.
	// Example: https://n8n.example.com
	WebhookBaseURL string

	// APIKey is the API key used to authenticate requests to the n8n instance.
	APIKey string
}

// loadN8NConfig reads n8n settings from environment variables.
// Both fields are optional; missing values result in an empty config.
func loadN8NConfig() (N8NConfig, error) {
	return N8NConfig{
		WebhookBaseURL: getEnv("N8N_WEBHOOK_BASE_URL", ""),
		APIKey:         getEnv("N8N_API_KEY", ""),
	}, nil
}
