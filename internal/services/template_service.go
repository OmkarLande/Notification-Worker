package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	texttmpl "text/template"

	"github.com/OmkarLande/notification-worker/internal/contracts"
	"github.com/OmkarLande/notification-worker/internal/logger"
)

// TemplateKey encapsulates the resolution hierarchy for finding templates.
type TemplateKey struct {
	App string
	Job string
}

// TemplateService resolves and renders templates, returning strongly typed payloads.
type TemplateService struct {
	templatesDir string
	log          logger.Logger
}

func NewTemplateService(templatesDir string, log logger.Logger) *TemplateService {
	return &TemplateService{
		templatesDir: templatesDir,
		log:          log,
	}
}

// GenerateEmailPayload renders the email.html template and returns an EmailPayload.
func (s *TemplateService) GenerateEmailPayload(ctx context.Context, key TemplateKey, data any) (*contracts.EmailPayload, error) {
	rendered, err := s.renderHTML(key, "email.html", data)
	if err != nil {
		return nil, err
	}

	// In a real app, the subject might be extracted from the template frontmatter or passed separately.
	// For now, we mock the subject based on the Job name.
	subject := fmt.Sprintf("Your %s Update", key.Job)

	return &contracts.EmailPayload{
		Subject: subject,
		Html:    rendered,
	}, nil
}

// GenerateDiscordPayload renders the discord.txt template and returns a DiscordPayload.
func (s *TemplateService) GenerateDiscordPayload(ctx context.Context, key TemplateKey, data any) (*contracts.DiscordPayload, error) {
	rendered, err := s.renderText(key, "discord.txt", data)
	if err != nil {
		return nil, err
	}

	return &contracts.DiscordPayload{
		Content: rendered,
		Embeds:  []any{},
	}, nil
}

// GenerateSlackPayload generates the slack blocks.
func (s *TemplateService) GenerateSlackPayload(ctx context.Context, key TemplateKey, data any) (*contracts.SlackPayload, error) {
	return &contracts.SlackPayload{Blocks: []any{}}, nil
}

// GenerateWhatsAppPayload generates the whatsapp text.
func (s *TemplateService) GenerateWhatsAppPayload(ctx context.Context, key TemplateKey, data any) (*contracts.WhatsAppPayload, error) {
	return &contracts.WhatsAppPayload{Text: "WhatsApp integration pending"}, nil
}

// resolveTemplatePath checks for the app-specific template, falling back to default.
func (s *TemplateService) resolveTemplatePath(key TemplateKey, filename string) (string, error) {
	// 1. Try App -> Job template
	// Ensure Job name is filesystem-friendly (e.g. "Daily Digest" -> "daily-digest")
	// For simplicity in Phase 5, we assume the job name folder is exact or we sanitize it.
	// Assuming folder matches job name exactly.
	jobDir := key.Job
	if jobDir == "Daily Digest" {
		jobDir = "daily-digest"
	}
	
	appPath := filepath.Join(s.templatesDir, key.App, jobDir, filename)
	if _, err := os.Stat(appPath); err == nil {
		return appPath, nil
	}

	// 2. Try Default template
	defaultPath := filepath.Join(s.templatesDir, "default", filename)
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath, nil
	}

	return "", fmt.Errorf("template not found for %s in app or default directories", filename)
}

func (s *TemplateService) renderHTML(key TemplateKey, filename string, data any) (string, error) {
	path, err := s.resolveTemplatePath(key, filename)
	if err != nil {
		return "", err
	}

	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", path, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", path, err)
	}
	return buf.String(), nil
}

func (s *TemplateService) renderText(key TemplateKey, filename string, data any) (string, error) {
	path, err := s.resolveTemplatePath(key, filename)
	if err != nil {
		return "", err
	}

	tmpl, err := texttmpl.ParseFiles(path)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", path, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", path, err)
	}
	return buf.String(), nil
}
