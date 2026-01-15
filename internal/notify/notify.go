// Package notify provides external notification channels for escalations.
// Channels include email (SMTP), SMS (Twilio), Slack (webhook), and log files.
package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Notification contains the data to send through notification channels.
type Notification struct {
	ID          string // Escalation ID
	Severity    string // critical, high, medium, low
	Title       string // Short description
	Body        string // Full message body
	Source      string // Who triggered this (agent ID)
	RelatedBead string // Related bead ID if any
	Timestamp   time.Time
}

// Result captures the outcome of a notification attempt.
type Result struct {
	Channel string // email, sms, slack, log
	Success bool
	Error   error
	Message string // Human-readable status
}

// SMTPConfig holds SMTP server configuration.
// Loaded from environment variables:
//   - GT_SMTP_HOST: SMTP server hostname (default: localhost)
//   - GT_SMTP_PORT: SMTP server port (default: 25)
//   - GT_SMTP_USER: SMTP username (optional)
//   - GT_SMTP_PASS: SMTP password (optional)
//   - GT_SMTP_FROM: From address (default: gongshow@localhost)
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// LoadSMTPConfig loads SMTP configuration from environment variables.
func LoadSMTPConfig() *SMTPConfig {
	return &SMTPConfig{
		Host:     getEnvOrDefault("GT_SMTP_HOST", "localhost"),
		Port:     getEnvOrDefault("GT_SMTP_PORT", "25"),
		Username: os.Getenv("GT_SMTP_USER"),
		Password: os.Getenv("GT_SMTP_PASS"),
		From:     getEnvOrDefault("GT_SMTP_FROM", "gongshow@localhost"),
	}
}

// TwilioConfig holds Twilio API configuration.
// Loaded from environment variables:
//   - TWILIO_ACCOUNT_SID: Twilio account SID
//   - TWILIO_AUTH_TOKEN: Twilio auth token
//   - TWILIO_FROM_NUMBER: Phone number to send from
type TwilioConfig struct {
	AccountSID string
	AuthToken  string
	FromNumber string
}

// LoadTwilioConfig loads Twilio configuration from environment variables.
func LoadTwilioConfig() *TwilioConfig {
	return &TwilioConfig{
		AccountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
		AuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
		FromNumber: os.Getenv("TWILIO_FROM_NUMBER"),
	}
}

// SendEmail sends an email notification via SMTP.
func SendEmail(to string, n *Notification) *Result {
	cfg := LoadSMTPConfig()

	if to == "" {
		return &Result{
			Channel: "email",
			Success: false,
			Error:   fmt.Errorf("no recipient email address configured"),
			Message: "Email skipped: no recipient configured",
		}
	}

	// Build email message
	subject := fmt.Sprintf("[%s] Escalation: %s", strings.ToUpper(n.Severity), n.Title)
	body := buildEmailBody(n)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=utf-8\r\n"+
		"X-GongShow-Escalation: %s\r\n"+
		"X-GongShow-Severity: %s\r\n"+
		"\r\n"+
		"%s",
		cfg.From, to, subject, n.ID, n.Severity, body)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	var auth smtp.Auth
	if cfg.Username != "" && cfg.Password != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	err := smtp.SendMail(addr, auth, cfg.From, []string{to}, []byte(msg))
	if err != nil {
		return &Result{
			Channel: "email",
			Success: false,
			Error:   err,
			Message: fmt.Sprintf("Failed to send email to %s: %v", to, err),
		}
	}

	return &Result{
		Channel: "email",
		Success: true,
		Message: fmt.Sprintf("Email sent to %s", to),
	}
}

// buildEmailBody constructs the email body from a notification.
func buildEmailBody(n *Notification) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Escalation ID: %s", n.ID))
	lines = append(lines, fmt.Sprintf("Severity: %s", strings.ToUpper(n.Severity)))
	lines = append(lines, fmt.Sprintf("Time: %s", n.Timestamp.Format(time.RFC1123)))
	lines = append(lines, fmt.Sprintf("Source: %s", n.Source))
	lines = append(lines, "")
	lines = append(lines, n.Title)
	lines = append(lines, "")
	if n.Body != "" {
		lines = append(lines, n.Body)
		lines = append(lines, "")
	}
	if n.RelatedBead != "" {
		lines = append(lines, fmt.Sprintf("Related: %s", n.RelatedBead))
		lines = append(lines, "")
	}
	lines = append(lines, "---")
	lines = append(lines, "To acknowledge: gt escalate ack "+n.ID)
	lines = append(lines, "To close: gt escalate close "+n.ID+" --reason \"resolution\"")

	return strings.Join(lines, "\n")
}

// SendSMS sends an SMS notification via Twilio.
func SendSMS(to string, n *Notification) *Result {
	cfg := LoadTwilioConfig()

	if cfg.AccountSID == "" || cfg.AuthToken == "" {
		return &Result{
			Channel: "sms",
			Success: false,
			Error:   fmt.Errorf("Twilio credentials not configured"),
			Message: "SMS skipped: TWILIO_ACCOUNT_SID and TWILIO_AUTH_TOKEN required",
		}
	}

	if cfg.FromNumber == "" {
		return &Result{
			Channel: "sms",
			Success: false,
			Error:   fmt.Errorf("Twilio from number not configured"),
			Message: "SMS skipped: TWILIO_FROM_NUMBER required",
		}
	}

	if to == "" {
		return &Result{
			Channel: "sms",
			Success: false,
			Error:   fmt.Errorf("no recipient phone number configured"),
			Message: "SMS skipped: no recipient configured",
		}
	}

	// Twilio Messages API endpoint
	url := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", cfg.AccountSID)

	// Build SMS message (keep it short for SMS)
	body := fmt.Sprintf("[%s] %s - %s\nID: %s\nAck: gt escalate ack %s",
		strings.ToUpper(n.Severity), n.Title, n.Source, n.ID, n.ID)

	// Truncate if too long for SMS
	if len(body) > 1600 {
		body = body[:1597] + "..."
	}

	// Build form data
	data := fmt.Sprintf("To=%s&From=%s&Body=%s",
		urlEncode(to), urlEncode(cfg.FromNumber), urlEncode(body))

	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return &Result{
			Channel: "sms",
			Success: false,
			Error:   err,
			Message: fmt.Sprintf("Failed to create SMS request: %v", err),
		}
	}

	req.SetBasicAuth(cfg.AccountSID, cfg.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &Result{
			Channel: "sms",
			Success: false,
			Error:   err,
			Message: fmt.Sprintf("Failed to send SMS to %s: %v", to, err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return &Result{
			Channel: "sms",
			Success: false,
			Error:   fmt.Errorf("Twilio API error: %s - %s", resp.Status, string(respBody)),
			Message: fmt.Sprintf("SMS failed: %s", resp.Status),
		}
	}

	return &Result{
		Channel: "sms",
		Success: true,
		Message: fmt.Sprintf("SMS sent to %s", maskPhoneNumber(to)),
	}
}

// SendSlack posts a notification to a Slack webhook.
func SendSlack(webhookURL string, n *Notification) *Result {
	if webhookURL == "" {
		return &Result{
			Channel: "slack",
			Success: false,
			Error:   fmt.Errorf("no Slack webhook URL configured"),
			Message: "Slack skipped: no webhook URL configured",
		}
	}

	// Build Slack message with blocks for rich formatting
	payload := buildSlackPayload(n)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return &Result{
			Channel: "slack",
			Success: false,
			Error:   err,
			Message: fmt.Sprintf("Failed to build Slack payload: %v", err),
		}
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return &Result{
			Channel: "slack",
			Success: false,
			Error:   err,
			Message: fmt.Sprintf("Failed to create Slack request: %v", err),
		}
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &Result{
			Channel: "slack",
			Success: false,
			Error:   err,
			Message: fmt.Sprintf("Failed to post to Slack: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return &Result{
			Channel: "slack",
			Success: false,
			Error:   fmt.Errorf("Slack webhook error: %s - %s", resp.Status, string(respBody)),
			Message: fmt.Sprintf("Slack post failed: %s", resp.Status),
		}
	}

	return &Result{
		Channel: "slack",
		Success: true,
		Message: "Posted to Slack",
	}
}

// buildSlackPayload creates a rich Slack message payload.
func buildSlackPayload(n *Notification) map[string]interface{} {
	// Emoji based on severity
	emoji := severityEmoji(n.Severity)

	// Color based on severity
	color := severityColor(n.Severity)

	// Build attachment fields
	fields := []map[string]interface{}{
		{"title": "Severity", "value": strings.ToUpper(n.Severity), "short": true},
		{"title": "ID", "value": n.ID, "short": true},
		{"title": "Source", "value": n.Source, "short": true},
	}

	if n.RelatedBead != "" {
		fields = append(fields, map[string]interface{}{
			"title": "Related",
			"value": n.RelatedBead,
			"short": true,
		})
	}

	return map[string]interface{}{
		"text": fmt.Sprintf("%s *Escalation: %s*", emoji, n.Title),
		"attachments": []map[string]interface{}{
			{
				"color":   color,
				"fields":  fields,
				"text":    n.Body,
				"footer":  "GongShow Escalation System",
				"ts":      n.Timestamp.Unix(),
				"actions": []map[string]interface{}{
					{
						"type": "button",
						"text": "Acknowledge",
						"url":  fmt.Sprintf("https://github.com/KeithWyatt/gongshow#escalation-%s", n.ID),
					},
				},
			},
		},
	}
}

// WriteLog writes the notification to an escalation log file.
func WriteLog(townRoot string, n *Notification) *Result {
	logDir := filepath.Join(townRoot, "logs")
	logFile := filepath.Join(logDir, "escalations.log")

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return &Result{
			Channel: "log",
			Success: false,
			Error:   err,
			Message: fmt.Sprintf("Failed to create log directory: %v", err),
		}
	}

	// Build log entry
	entry := buildLogEntry(n)

	// Append to log file
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return &Result{
			Channel: "log",
			Success: false,
			Error:   err,
			Message: fmt.Sprintf("Failed to open log file: %v", err),
		}
	}
	defer f.Close()

	if _, err := f.WriteString(entry + "\n"); err != nil {
		return &Result{
			Channel: "log",
			Success: false,
			Error:   err,
			Message: fmt.Sprintf("Failed to write to log: %v", err),
		}
	}

	return &Result{
		Channel: "log",
		Success: true,
		Message: fmt.Sprintf("Logged to %s", logFile),
	}
}

// buildLogEntry creates a structured log entry.
func buildLogEntry(n *Notification) string {
	entry := map[string]interface{}{
		"timestamp": n.Timestamp.Format(time.RFC3339),
		"id":        n.ID,
		"severity":  n.Severity,
		"title":     n.Title,
		"source":    n.Source,
	}

	if n.RelatedBead != "" {
		entry["related"] = n.RelatedBead
	}

	if n.Body != "" {
		entry["body"] = n.Body
	}

	data, _ := json.Marshal(entry)
	return string(data)
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func severityEmoji(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "🚨"
	case "high":
		return "⚠️"
	case "medium":
		return "📢"
	case "low":
		return "ℹ️"
	default:
		return "📋"
	}
}

func severityColor(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "#FF0000" // Red
	case "high":
		return "#FFA500" // Orange
	case "medium":
		return "#FFD700" // Gold
	case "low":
		return "#808080" // Gray
	default:
		return "#0000FF" // Blue
	}
}

// urlEncode encodes a string for URL form data.
func urlEncode(s string) string {
	// Simple URL encoding for form data
	var result strings.Builder
	for _, c := range s {
		switch {
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'):
			result.WriteRune(c)
		case c == '-' || c == '_' || c == '.' || c == '~':
			result.WriteRune(c)
		case c == ' ':
			result.WriteRune('+')
		default:
			result.WriteString(fmt.Sprintf("%%%02X", c))
		}
	}
	return result.String()
}

// maskPhoneNumber masks most digits of a phone number for privacy in logs.
func maskPhoneNumber(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
}
