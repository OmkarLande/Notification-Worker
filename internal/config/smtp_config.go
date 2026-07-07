package config

import "fmt"

// SMTPConfig holds settings required to connect to an SMTP server for sending
// transactional emails.
type SMTPConfig struct {
	// Host is the SMTP server hostname.
	Host string

	// Port is the SMTP server port (e.g. 465 for TLS, 587 for STARTTLS).
	Port int

	// User is the SMTP authentication username.
	User string

	// Password is the SMTP authentication password.
	Password string

	// From is the sender email address used in outgoing messages.
	From string
}

// loadSMTPConfig reads SMTP settings from environment variables.
func loadSMTPConfig() (SMTPConfig, error) {
	host := getEnv("SMTP_HOST", "")
	if host == "" {
		return SMTPConfig{}, fmt.Errorf("SMTP_HOST is required but not set")
	}

	user := getEnv("SMTP_USER", "")
	if user == "" {
		return SMTPConfig{}, fmt.Errorf("SMTP_USER is required but not set")
	}

	password := getEnv("SMTP_PASSWORD", "")
	if password == "" {
		return SMTPConfig{}, fmt.Errorf("SMTP_PASSWORD is required but not set")
	}

	from := getEnv("SMTP_FROM", "")
	if from == "" {
		return SMTPConfig{}, fmt.Errorf("SMTP_FROM is required but not set")
	}

	port := getEnvInt("SMTP_PORT", 587)
	if port <= 0 || port > 65535 {
		return SMTPConfig{}, fmt.Errorf("SMTP_PORT must be a valid port number (1–65535), got %d", port)
	}

	return SMTPConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		From:     from,
	}, nil
}
