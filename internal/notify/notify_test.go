package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadSMTPConfig(t *testing.T) {
	// Save original env vars
	origHost := os.Getenv("GT_SMTP_HOST")
	origPort := os.Getenv("GT_SMTP_PORT")
	origUser := os.Getenv("GT_SMTP_USER")
	origPass := os.Getenv("GT_SMTP_PASS")
	origFrom := os.Getenv("GT_SMTP_FROM")

	// Cleanup
	defer func() {
		os.Setenv("GT_SMTP_HOST", origHost)
		os.Setenv("GT_SMTP_PORT", origPort)
		os.Setenv("GT_SMTP_USER", origUser)
		os.Setenv("GT_SMTP_PASS", origPass)
		os.Setenv("GT_SMTP_FROM", origFrom)
	}()

	t.Run("defaults", func(t *testing.T) {
		os.Unsetenv("GT_SMTP_HOST")
		os.Unsetenv("GT_SMTP_PORT")
		os.Unsetenv("GT_SMTP_USER")
		os.Unsetenv("GT_SMTP_PASS")
		os.Unsetenv("GT_SMTP_FROM")

		cfg := LoadSMTPConfig()

		if cfg.Host != "localhost" {
			t.Errorf("expected host=localhost, got %s", cfg.Host)
		}
		if cfg.Port != "25" {
			t.Errorf("expected port=25, got %s", cfg.Port)
		}
		if cfg.From != "gongshow@localhost" {
			t.Errorf("expected from=gongshow@localhost, got %s", cfg.From)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		os.Setenv("GT_SMTP_HOST", "mail.example.com")
		os.Setenv("GT_SMTP_PORT", "587")
		os.Setenv("GT_SMTP_USER", "user")
		os.Setenv("GT_SMTP_PASS", "pass")
		os.Setenv("GT_SMTP_FROM", "alerts@example.com")

		cfg := LoadSMTPConfig()

		if cfg.Host != "mail.example.com" {
			t.Errorf("expected host=mail.example.com, got %s", cfg.Host)
		}
		if cfg.Port != "587" {
			t.Errorf("expected port=587, got %s", cfg.Port)
		}
		if cfg.Username != "user" {
			t.Errorf("expected username=user, got %s", cfg.Username)
		}
		if cfg.Password != "pass" {
			t.Errorf("expected password=pass, got %s", cfg.Password)
		}
		if cfg.From != "alerts@example.com" {
			t.Errorf("expected from=alerts@example.com, got %s", cfg.From)
		}
	})
}

func TestLoadTwilioConfig(t *testing.T) {
	// Save original env vars
	origSID := os.Getenv("TWILIO_ACCOUNT_SID")
	origToken := os.Getenv("TWILIO_AUTH_TOKEN")
	origFrom := os.Getenv("TWILIO_FROM_NUMBER")

	// Cleanup
	defer func() {
		os.Setenv("TWILIO_ACCOUNT_SID", origSID)
		os.Setenv("TWILIO_AUTH_TOKEN", origToken)
		os.Setenv("TWILIO_FROM_NUMBER", origFrom)
	}()

	t.Run("empty", func(t *testing.T) {
		os.Unsetenv("TWILIO_ACCOUNT_SID")
		os.Unsetenv("TWILIO_AUTH_TOKEN")
		os.Unsetenv("TWILIO_FROM_NUMBER")

		cfg := LoadTwilioConfig()

		if cfg.AccountSID != "" {
			t.Errorf("expected empty AccountSID, got %s", cfg.AccountSID)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		os.Setenv("TWILIO_ACCOUNT_SID", "AC123")
		os.Setenv("TWILIO_AUTH_TOKEN", "token123")
		os.Setenv("TWILIO_FROM_NUMBER", "+15551234567")

		cfg := LoadTwilioConfig()

		if cfg.AccountSID != "AC123" {
			t.Errorf("expected AccountSID=AC123, got %s", cfg.AccountSID)
		}
		if cfg.AuthToken != "token123" {
			t.Errorf("expected AuthToken=token123, got %s", cfg.AuthToken)
		}
		if cfg.FromNumber != "+15551234567" {
			t.Errorf("expected FromNumber=+15551234567, got %s", cfg.FromNumber)
		}
	})
}

func TestSendEmailNoRecipient(t *testing.T) {
	n := &Notification{
		ID:        "esc-001",
		Severity:  "high",
		Title:     "Test escalation",
		Timestamp: time.Now(),
	}

	result := SendEmail("", n)

	if result.Success {
		t.Error("expected failure when no recipient configured")
	}
	if result.Channel != "email" {
		t.Errorf("expected channel=email, got %s", result.Channel)
	}
	if result.Error == nil {
		t.Error("expected error to be set")
	}
}

func TestSendSMSMissingConfig(t *testing.T) {
	// Ensure Twilio env vars are unset
	origSID := os.Getenv("TWILIO_ACCOUNT_SID")
	origToken := os.Getenv("TWILIO_AUTH_TOKEN")
	defer func() {
		os.Setenv("TWILIO_ACCOUNT_SID", origSID)
		os.Setenv("TWILIO_AUTH_TOKEN", origToken)
	}()
	os.Unsetenv("TWILIO_ACCOUNT_SID")
	os.Unsetenv("TWILIO_AUTH_TOKEN")

	n := &Notification{
		ID:        "esc-001",
		Severity:  "critical",
		Title:     "Test",
		Timestamp: time.Now(),
	}

	result := SendSMS("+15551234567", n)

	if result.Success {
		t.Error("expected failure when Twilio not configured")
	}
	if result.Channel != "sms" {
		t.Errorf("expected channel=sms, got %s", result.Channel)
	}
}

func TestSendSlackNoWebhook(t *testing.T) {
	n := &Notification{
		ID:        "esc-001",
		Severity:  "medium",
		Title:     "Test",
		Timestamp: time.Now(),
	}

	result := SendSlack("", n)

	if result.Success {
		t.Error("expected failure when no webhook configured")
	}
	if result.Channel != "slack" {
		t.Errorf("expected channel=slack, got %s", result.Channel)
	}
}

func TestSendSlackSuccess(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type=application/json")
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("failed to decode payload: %v", err)
		}

		// Verify payload has expected fields
		if _, ok := payload["text"]; !ok {
			t.Error("payload missing 'text' field")
		}
		if _, ok := payload["attachments"]; !ok {
			t.Error("payload missing 'attachments' field")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := &Notification{
		ID:          "esc-001",
		Severity:    "high",
		Title:       "Test escalation",
		Body:        "Something went wrong",
		Source:      "gongshow/crew/lisa",
		RelatedBead: "go-123",
		Timestamp:   time.Now(),
	}

	result := SendSlack(server.URL, n)

	if !result.Success {
		t.Errorf("expected success, got error: %v", result.Error)
	}
	if result.Channel != "slack" {
		t.Errorf("expected channel=slack, got %s", result.Channel)
	}
}

func TestSendSlackServerError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	n := &Notification{
		ID:        "esc-001",
		Severity:  "high",
		Title:     "Test",
		Timestamp: time.Now(),
	}

	result := SendSlack(server.URL, n)

	if result.Success {
		t.Error("expected failure on server error")
	}
}

func TestWriteLog(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	n := &Notification{
		ID:          "esc-001",
		Severity:    "critical",
		Title:       "Test escalation",
		Body:        "Something critical happened",
		Source:      "gongshow/crew/lisa",
		RelatedBead: "go-456",
		Timestamp:   time.Now(),
	}

	result := WriteLog(tmpDir, n)

	if !result.Success {
		t.Errorf("expected success, got error: %v", result.Error)
	}
	if result.Channel != "log" {
		t.Errorf("expected channel=log, got %s", result.Channel)
	}

	// Verify log file was created
	logFile := filepath.Join(tmpDir, "logs", "escalations.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Verify content is valid JSON
	var entry map[string]interface{}
	if err := json.Unmarshal(content, &entry); err != nil {
		t.Errorf("log entry is not valid JSON: %v", err)
	}

	// Verify expected fields
	if entry["id"] != "esc-001" {
		t.Errorf("expected id=esc-001, got %v", entry["id"])
	}
	if entry["severity"] != "critical" {
		t.Errorf("expected severity=critical, got %v", entry["severity"])
	}
	if entry["title"] != "Test escalation" {
		t.Errorf("expected title='Test escalation', got %v", entry["title"])
	}
	if entry["related"] != "go-456" {
		t.Errorf("expected related=go-456, got %v", entry["related"])
	}
}

func TestWriteLogAppends(t *testing.T) {
	tmpDir := t.TempDir()

	n1 := &Notification{ID: "esc-001", Severity: "low", Title: "First", Timestamp: time.Now()}
	n2 := &Notification{ID: "esc-002", Severity: "high", Title: "Second", Timestamp: time.Now()}

	result1 := WriteLog(tmpDir, n1)
	result2 := WriteLog(tmpDir, n2)

	if !result1.Success || !result2.Success {
		t.Error("expected both writes to succeed")
	}

	// Read log file
	logFile := filepath.Join(tmpDir, "logs", "escalations.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Should have 2 lines
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
}

func TestBuildEmailBody(t *testing.T) {
	n := &Notification{
		ID:          "esc-test",
		Severity:    "high",
		Title:       "Test Alert",
		Body:        "Details here",
		Source:      "test-agent",
		RelatedBead: "go-789",
		Timestamp:   time.Now(),
	}

	body := buildEmailBody(n)

	if !strings.Contains(body, "esc-test") {
		t.Error("body should contain escalation ID")
	}
	if !strings.Contains(body, "HIGH") {
		t.Error("body should contain severity")
	}
	if !strings.Contains(body, "Test Alert") {
		t.Error("body should contain title")
	}
	if !strings.Contains(body, "Details here") {
		t.Error("body should contain body text")
	}
	if !strings.Contains(body, "go-789") {
		t.Error("body should contain related bead")
	}
	if !strings.Contains(body, "gt escalate ack") {
		t.Error("body should contain ack command")
	}
}

func TestSeverityEmoji(t *testing.T) {
	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "🚨"},
		{"high", "⚠️"},
		{"medium", "📢"},
		{"low", "ℹ️"},
		{"unknown", "📋"},
	}

	for _, tt := range tests {
		got := severityEmoji(tt.severity)
		if got != tt.expected {
			t.Errorf("severityEmoji(%s) = %s, want %s", tt.severity, got, tt.expected)
		}
	}
}

func TestSeverityColor(t *testing.T) {
	tests := []struct {
		severity string
		expected string
	}{
		{"critical", "#FF0000"},
		{"high", "#FFA500"},
		{"medium", "#FFD700"},
		{"low", "#808080"},
		{"unknown", "#0000FF"},
	}

	for _, tt := range tests {
		got := severityColor(tt.severity)
		if got != tt.expected {
			t.Errorf("severityColor(%s) = %s, want %s", tt.severity, got, tt.expected)
		}
	}
}

func TestURLEncode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello world", "hello+world"},
		{"test@example.com", "test%40example.com"},
		{"+1 555-123-4567", "%2B1+555-123-4567"},
	}

	for _, tt := range tests {
		got := urlEncode(tt.input)
		if got != tt.expected {
			t.Errorf("urlEncode(%s) = %s, want %s", tt.input, got, tt.expected)
		}
	}
}

func TestMaskPhoneNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"+15551234567", "********4567"}, // 12 chars, mask first 8
		{"12345678", "****5678"},         // 8 chars, mask first 4
		{"1234", "****"},                 // 4 chars or less, all masked
		{"123", "****"},
		{"", "****"},
	}

	for _, tt := range tests {
		got := maskPhoneNumber(tt.input)
		if got != tt.expected {
			t.Errorf("maskPhoneNumber(%s) = %s, want %s", tt.input, got, tt.expected)
		}
	}
}

func TestBuildSlackPayload(t *testing.T) {
	n := &Notification{
		ID:          "esc-001",
		Severity:    "critical",
		Title:       "Server Down",
		Body:        "Production server is not responding",
		Source:      "gongshow/witness",
		RelatedBead: "go-999",
		Timestamp:   time.Now(),
	}

	payload := buildSlackPayload(n)

	// Check text contains emoji and title
	text, ok := payload["text"].(string)
	if !ok {
		t.Fatal("payload should have text field")
	}
	if !strings.Contains(text, "🚨") {
		t.Error("critical severity should have fire emoji")
	}
	if !strings.Contains(text, "Server Down") {
		t.Error("text should contain title")
	}

	// Check attachments
	attachments, ok := payload["attachments"].([]map[string]interface{})
	if !ok {
		t.Fatal("payload should have attachments")
	}
	if len(attachments) != 1 {
		t.Errorf("expected 1 attachment, got %d", len(attachments))
	}

	// Check color
	if attachments[0]["color"] != "#FF0000" {
		t.Error("critical severity should have red color")
	}
}
