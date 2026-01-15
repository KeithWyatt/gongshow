package beads

import (
	"strings"
	"testing"
)

func TestEscalationStateConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"EscalationOpen", EscalationOpen, "open"},
		{"EscalationAcked", EscalationAcked, "acked"},
		{"EscalationClosed", EscalationClosed, "closed"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.expected {
				t.Errorf("%s = %q, want %q", tc.name, tc.got, tc.expected)
			}
		})
	}
}

func TestFormatEscalationDescription(t *testing.T) {
	t.Run("nil fields returns title only", func(t *testing.T) {
		result := FormatEscalationDescription("Critical Error", nil)
		if result != "Critical Error" {
			t.Errorf("FormatEscalationDescription(title, nil) = %q, want %q", result, "Critical Error")
		}
	})

	t.Run("full fields", func(t *testing.T) {
		fields := &EscalationFields{
			Severity:           "high",
			Reason:             "Build failing repeatedly",
			Source:             "patrol:witness",
			EscalatedBy:        "gongshow/witness",
			EscalatedAt:        "2024-01-15T10:00:00Z",
			AckedBy:            "gongshow/crew/marge",
			AckedAt:            "2024-01-15T10:05:00Z",
			ClosedBy:           "gongshow/crew/marge",
			ClosedReason:       "Fixed the build",
			RelatedBead:        "go-abc",
			OriginalSeverity:   "medium",
			ReescalationCount:  2,
			LastReescalatedAt:  "2024-01-15T10:30:00Z",
			LastReescalatedBy:  "system",
		}
		result := FormatEscalationDescription("Build Failure", fields)

		checks := []string{
			"Build Failure",
			"severity: high",
			"reason: Build failing repeatedly",
			"source: patrol:witness",
			"escalated_by: gongshow/witness",
			"escalated_at: 2024-01-15T10:00:00Z",
			"acked_by: gongshow/crew/marge",
			"acked_at: 2024-01-15T10:05:00Z",
			"closed_by: gongshow/crew/marge",
			"closed_reason: Fixed the build",
			"related_bead: go-abc",
			"original_severity: medium",
			"reescalation_count: 2",
			"last_reescalated_at: 2024-01-15T10:30:00Z",
			"last_reescalated_by: system",
		}

		for _, check := range checks {
			if !strings.Contains(result, check) {
				t.Errorf("result should contain %q, got:\n%s", check, result)
			}
		}
	})

	t.Run("partial fields with nulls", func(t *testing.T) {
		fields := &EscalationFields{
			Severity:    "critical",
			Reason:      "System down",
			EscalatedBy: "daemon",
			EscalatedAt: "2024-01-15T10:00:00Z",
			// AckedBy, ClosedBy, etc. are empty
		}
		result := FormatEscalationDescription("System Down", fields)

		if !strings.Contains(result, "severity: critical") {
			t.Error("should contain severity: critical")
		}
		if !strings.Contains(result, "acked_by: null") {
			t.Error("should contain acked_by: null")
		}
		if !strings.Contains(result, "closed_by: null") {
			t.Error("should contain closed_by: null")
		}
		if !strings.Contains(result, "related_bead: null") {
			t.Error("should contain related_bead: null")
		}
	})
}

func TestParseEscalationFields(t *testing.T) {
	t.Run("empty description", func(t *testing.T) {
		fields := ParseEscalationFields("")
		if fields == nil {
			t.Fatal("ParseEscalationFields should never return nil")
		}
		if fields.Severity != "" {
			t.Errorf("Severity should be empty, got %q", fields.Severity)
		}
	})

	t.Run("full fields", func(t *testing.T) {
		description := `Build Failure

severity: high
reason: Build failing repeatedly
source: patrol:witness
escalated_by: gongshow/witness
escalated_at: 2024-01-15T10:00:00Z
acked_by: gongshow/crew/marge
acked_at: 2024-01-15T10:05:00Z
closed_by: gongshow/crew/marge
closed_reason: Fixed the build
related_bead: go-abc
original_severity: medium
reescalation_count: 2
last_reescalated_at: 2024-01-15T10:30:00Z
last_reescalated_by: system`

		fields := ParseEscalationFields(description)

		if fields.Severity != "high" {
			t.Errorf("Severity = %q, want %q", fields.Severity, "high")
		}
		if fields.Reason != "Build failing repeatedly" {
			t.Errorf("Reason = %q, want %q", fields.Reason, "Build failing repeatedly")
		}
		if fields.Source != "patrol:witness" {
			t.Errorf("Source = %q, want %q", fields.Source, "patrol:witness")
		}
		if fields.EscalatedBy != "gongshow/witness" {
			t.Errorf("EscalatedBy = %q, want %q", fields.EscalatedBy, "gongshow/witness")
		}
		if fields.EscalatedAt != "2024-01-15T10:00:00Z" {
			t.Errorf("EscalatedAt = %q, want %q", fields.EscalatedAt, "2024-01-15T10:00:00Z")
		}
		if fields.AckedBy != "gongshow/crew/marge" {
			t.Errorf("AckedBy = %q, want %q", fields.AckedBy, "gongshow/crew/marge")
		}
		if fields.AckedAt != "2024-01-15T10:05:00Z" {
			t.Errorf("AckedAt = %q, want %q", fields.AckedAt, "2024-01-15T10:05:00Z")
		}
		if fields.ClosedBy != "gongshow/crew/marge" {
			t.Errorf("ClosedBy = %q, want %q", fields.ClosedBy, "gongshow/crew/marge")
		}
		if fields.ClosedReason != "Fixed the build" {
			t.Errorf("ClosedReason = %q, want %q", fields.ClosedReason, "Fixed the build")
		}
		if fields.RelatedBead != "go-abc" {
			t.Errorf("RelatedBead = %q, want %q", fields.RelatedBead, "go-abc")
		}
		if fields.OriginalSeverity != "medium" {
			t.Errorf("OriginalSeverity = %q, want %q", fields.OriginalSeverity, "medium")
		}
		if fields.ReescalationCount != 2 {
			t.Errorf("ReescalationCount = %d, want %d", fields.ReescalationCount, 2)
		}
		if fields.LastReescalatedAt != "2024-01-15T10:30:00Z" {
			t.Errorf("LastReescalatedAt = %q, want %q", fields.LastReescalatedAt, "2024-01-15T10:30:00Z")
		}
		if fields.LastReescalatedBy != "system" {
			t.Errorf("LastReescalatedBy = %q, want %q", fields.LastReescalatedBy, "system")
		}
	})

	t.Run("null values become empty strings", func(t *testing.T) {
		description := `Test

severity: critical
reason: test
acked_by: null
closed_by: null
related_bead: null`

		fields := ParseEscalationFields(description)

		if fields.Severity != "critical" {
			t.Errorf("Severity = %q, want %q", fields.Severity, "critical")
		}
		if fields.AckedBy != "" {
			t.Errorf("AckedBy should be empty for null, got %q", fields.AckedBy)
		}
		if fields.ClosedBy != "" {
			t.Errorf("ClosedBy should be empty for null, got %q", fields.ClosedBy)
		}
		if fields.RelatedBead != "" {
			t.Errorf("RelatedBead should be empty for null, got %q", fields.RelatedBead)
		}
	})

	t.Run("case insensitive keys", func(t *testing.T) {
		description := `Test

SEVERITY: high
Reason: Test reason
Escalated_By: tester`

		fields := ParseEscalationFields(description)

		if fields.Severity != "high" {
			t.Errorf("Severity = %q, want %q", fields.Severity, "high")
		}
		if fields.Reason != "Test reason" {
			t.Errorf("Reason = %q, want %q", fields.Reason, "Test reason")
		}
		if fields.EscalatedBy != "tester" {
			t.Errorf("EscalatedBy = %q, want %q", fields.EscalatedBy, "tester")
		}
	})

	t.Run("reescalation_count parsing", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int
		}{
			{"reescalation_count: 0", 0},
			{"reescalation_count: 1", 1},
			{"reescalation_count: 5", 5},
			{"reescalation_count: invalid", 0}, // Invalid should default to 0
		}

		for _, tc := range tests {
			fields := ParseEscalationFields(tc.input)
			if fields.ReescalationCount != tc.expected {
				t.Errorf("ReescalationCount for %q = %d, want %d", tc.input, fields.ReescalationCount, tc.expected)
			}
		}
	})
}

func TestFormatAndParseEscalationFieldsRoundTrip(t *testing.T) {
	original := &EscalationFields{
		Severity:           "high",
		Reason:             "Memory leak detected",
		Source:             "plugin:memory-monitor",
		EscalatedBy:        "gongshow/deacon",
		EscalatedAt:        "2024-01-15T14:30:00Z",
		AckedBy:            "human",
		AckedAt:            "2024-01-15T14:35:00Z",
		ClosedBy:           "gongshow/crew/joe",
		ClosedReason:       "Fixed memory leak in cache layer",
		RelatedBead:        "go-memory-123",
		OriginalSeverity:   "low",
		ReescalationCount:  3,
		LastReescalatedAt:  "2024-01-15T14:25:00Z",
		LastReescalatedBy:  "witness-patrol",
	}

	formatted := FormatEscalationDescription("Memory Leak Alert", original)
	parsed := ParseEscalationFields(formatted)

	if parsed.Severity != original.Severity {
		t.Errorf("Severity mismatch: got %q, want %q", parsed.Severity, original.Severity)
	}
	if parsed.Reason != original.Reason {
		t.Errorf("Reason mismatch: got %q, want %q", parsed.Reason, original.Reason)
	}
	if parsed.Source != original.Source {
		t.Errorf("Source mismatch: got %q, want %q", parsed.Source, original.Source)
	}
	if parsed.EscalatedBy != original.EscalatedBy {
		t.Errorf("EscalatedBy mismatch: got %q, want %q", parsed.EscalatedBy, original.EscalatedBy)
	}
	if parsed.EscalatedAt != original.EscalatedAt {
		t.Errorf("EscalatedAt mismatch: got %q, want %q", parsed.EscalatedAt, original.EscalatedAt)
	}
	if parsed.AckedBy != original.AckedBy {
		t.Errorf("AckedBy mismatch: got %q, want %q", parsed.AckedBy, original.AckedBy)
	}
	if parsed.AckedAt != original.AckedAt {
		t.Errorf("AckedAt mismatch: got %q, want %q", parsed.AckedAt, original.AckedAt)
	}
	if parsed.ClosedBy != original.ClosedBy {
		t.Errorf("ClosedBy mismatch: got %q, want %q", parsed.ClosedBy, original.ClosedBy)
	}
	if parsed.ClosedReason != original.ClosedReason {
		t.Errorf("ClosedReason mismatch: got %q, want %q", parsed.ClosedReason, original.ClosedReason)
	}
	if parsed.RelatedBead != original.RelatedBead {
		t.Errorf("RelatedBead mismatch: got %q, want %q", parsed.RelatedBead, original.RelatedBead)
	}
	if parsed.OriginalSeverity != original.OriginalSeverity {
		t.Errorf("OriginalSeverity mismatch: got %q, want %q", parsed.OriginalSeverity, original.OriginalSeverity)
	}
	if parsed.ReescalationCount != original.ReescalationCount {
		t.Errorf("ReescalationCount mismatch: got %d, want %d", parsed.ReescalationCount, original.ReescalationCount)
	}
	if parsed.LastReescalatedAt != original.LastReescalatedAt {
		t.Errorf("LastReescalatedAt mismatch: got %q, want %q", parsed.LastReescalatedAt, original.LastReescalatedAt)
	}
	if parsed.LastReescalatedBy != original.LastReescalatedBy {
		t.Errorf("LastReescalatedBy mismatch: got %q, want %q", parsed.LastReescalatedBy, original.LastReescalatedBy)
	}
}

func TestBumpSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"low", "medium"},
		{"medium", "high"},
		{"high", "critical"},
		{"critical", "critical"}, // Can't bump past critical
		{"unknown", "critical"},  // Unknown defaults to critical
		{"", "critical"},         // Empty defaults to critical
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := bumpSeverity(tc.input)
			if got != tc.expected {
				t.Errorf("bumpSeverity(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestReescalationResultStruct(t *testing.T) {
	result := ReescalationResult{
		ID:              "go-esc-123",
		Title:           "Memory Alert",
		OldSeverity:     "low",
		NewSeverity:     "medium",
		ReescalationNum: 1,
		Skipped:         false,
		SkipReason:      "",
	}

	if result.ID != "go-esc-123" {
		t.Errorf("ID = %q, want %q", result.ID, "go-esc-123")
	}
	if result.OldSeverity != "low" {
		t.Errorf("OldSeverity = %q, want %q", result.OldSeverity, "low")
	}
	if result.NewSeverity != "medium" {
		t.Errorf("NewSeverity = %q, want %q", result.NewSeverity, "medium")
	}
	if result.ReescalationNum != 1 {
		t.Errorf("ReescalationNum = %d, want %d", result.ReescalationNum, 1)
	}
	if result.Skipped {
		t.Error("Skipped should be false")
	}
}

func TestReescalationResultSkipped(t *testing.T) {
	result := ReescalationResult{
		ID:              "go-esc-456",
		Title:           "Critical Alert",
		OldSeverity:     "critical",
		NewSeverity:     "critical",
		ReescalationNum: 0,
		Skipped:         true,
		SkipReason:      "already at critical severity",
	}

	if !result.Skipped {
		t.Error("Skipped should be true")
	}
	if result.SkipReason != "already at critical severity" {
		t.Errorf("SkipReason = %q, want %q", result.SkipReason, "already at critical severity")
	}
}

func TestEscalationSeverityLevels(t *testing.T) {
	severities := []string{"low", "medium", "high", "critical"}

	for _, severity := range severities {
		t.Run(severity, func(t *testing.T) {
			fields := &EscalationFields{
				Severity:    severity,
				Reason:      "test",
				EscalatedBy: "test",
				EscalatedAt: "2024-01-15T10:00:00Z",
			}
			formatted := FormatEscalationDescription("Test", fields)
			parsed := ParseEscalationFields(formatted)

			if parsed.Severity != severity {
				t.Errorf("Severity = %q, want %q", parsed.Severity, severity)
			}
		})
	}
}

func TestEscalationEmptyRoundTrip(t *testing.T) {
	// Test that empty optional fields round-trip correctly through null
	original := &EscalationFields{
		Severity:    "medium",
		Reason:      "test reason",
		EscalatedBy: "tester",
		EscalatedAt: "2024-01-15T10:00:00Z",
		// All optional fields empty
	}

	formatted := FormatEscalationDescription("Test", original)
	parsed := ParseEscalationFields(formatted)

	if parsed.Severity != original.Severity {
		t.Errorf("Severity mismatch: got %q, want %q", parsed.Severity, original.Severity)
	}
	if parsed.AckedBy != "" {
		t.Errorf("AckedBy should be empty, got %q", parsed.AckedBy)
	}
	if parsed.ClosedBy != "" {
		t.Errorf("ClosedBy should be empty, got %q", parsed.ClosedBy)
	}
	if parsed.RelatedBead != "" {
		t.Errorf("RelatedBead should be empty, got %q", parsed.RelatedBead)
	}
	if parsed.OriginalSeverity != "" {
		t.Errorf("OriginalSeverity should be empty, got %q", parsed.OriginalSeverity)
	}
	if parsed.ReescalationCount != 0 {
		t.Errorf("ReescalationCount should be 0, got %d", parsed.ReescalationCount)
	}
}
