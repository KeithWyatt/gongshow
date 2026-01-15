package beads

import (
	"strings"
	"testing"
)

func TestAgentFieldsNotificationLevelConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"NotifyVerbose", NotifyVerbose, "verbose"},
		{"NotifyNormal", NotifyNormal, "normal"},
		{"NotifyMuted", NotifyMuted, "muted"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.expected {
				t.Errorf("%s = %q, want %q", tc.name, tc.got, tc.expected)
			}
		})
	}
}

func TestFormatAgentDescription(t *testing.T) {
	t.Run("nil fields returns title only", func(t *testing.T) {
		result := FormatAgentDescription("Test Agent", nil)
		if result != "Test Agent" {
			t.Errorf("FormatAgentDescription(title, nil) = %q, want %q", result, "Test Agent")
		}
	})

	t.Run("empty fields", func(t *testing.T) {
		fields := &AgentFields{}
		result := FormatAgentDescription("Test Agent", fields)

		// Should contain null values for empty fields
		if !strings.Contains(result, "role_type: ") {
			t.Error("should contain role_type")
		}
		if !strings.Contains(result, "rig: null") {
			t.Error("should contain rig: null for empty rig")
		}
		if !strings.Contains(result, "hook_bead: null") {
			t.Error("should contain hook_bead: null for empty hook")
		}
	})

	t.Run("full fields", func(t *testing.T) {
		fields := &AgentFields{
			RoleType:          "polecat",
			Rig:               "gongshow",
			AgentState:        "working",
			HookBead:          "go-abc",
			RoleBead:          "go-role-123",
			CleanupStatus:     "clean",
			ActiveMR:          "mr-456",
			NotificationLevel: "verbose",
		}
		result := FormatAgentDescription("Toast", fields)

		checks := []string{
			"Toast",
			"role_type: polecat",
			"rig: gongshow",
			"agent_state: working",
			"hook_bead: go-abc",
			"role_bead: go-role-123",
			"cleanup_status: clean",
			"active_mr: mr-456",
			"notification_level: verbose",
		}

		for _, check := range checks {
			if !strings.Contains(result, check) {
				t.Errorf("result should contain %q, got:\n%s", check, result)
			}
		}
	})

	t.Run("partial fields with nulls", func(t *testing.T) {
		fields := &AgentFields{
			RoleType:   "witness",
			Rig:        "gongshow",
			AgentState: "running",
			// HookBead, RoleBead, etc. are empty
		}
		result := FormatAgentDescription("Witness", fields)

		if !strings.Contains(result, "role_type: witness") {
			t.Error("should contain role_type: witness")
		}
		if !strings.Contains(result, "rig: gongshow") {
			t.Error("should contain rig: gongshow")
		}
		if !strings.Contains(result, "hook_bead: null") {
			t.Error("should contain hook_bead: null")
		}
		if !strings.Contains(result, "role_bead: null") {
			t.Error("should contain role_bead: null")
		}
	})
}

func TestParseAgentFields(t *testing.T) {
	t.Run("empty description", func(t *testing.T) {
		fields := ParseAgentFields("")
		if fields == nil {
			t.Fatal("ParseAgentFields should never return nil")
		}
		if fields.RoleType != "" {
			t.Errorf("RoleType should be empty, got %q", fields.RoleType)
		}
	})

	t.Run("full fields", func(t *testing.T) {
		description := `Toast

role_type: polecat
rig: gongshow
agent_state: working
hook_bead: go-abc
role_bead: go-role-123
cleanup_status: clean
active_mr: mr-456
notification_level: verbose`

		fields := ParseAgentFields(description)

		if fields.RoleType != "polecat" {
			t.Errorf("RoleType = %q, want %q", fields.RoleType, "polecat")
		}
		if fields.Rig != "gongshow" {
			t.Errorf("Rig = %q, want %q", fields.Rig, "gongshow")
		}
		if fields.AgentState != "working" {
			t.Errorf("AgentState = %q, want %q", fields.AgentState, "working")
		}
		if fields.HookBead != "go-abc" {
			t.Errorf("HookBead = %q, want %q", fields.HookBead, "go-abc")
		}
		if fields.RoleBead != "go-role-123" {
			t.Errorf("RoleBead = %q, want %q", fields.RoleBead, "go-role-123")
		}
		if fields.CleanupStatus != "clean" {
			t.Errorf("CleanupStatus = %q, want %q", fields.CleanupStatus, "clean")
		}
		if fields.ActiveMR != "mr-456" {
			t.Errorf("ActiveMR = %q, want %q", fields.ActiveMR, "mr-456")
		}
		if fields.NotificationLevel != "verbose" {
			t.Errorf("NotificationLevel = %q, want %q", fields.NotificationLevel, "verbose")
		}
	})

	t.Run("null values become empty strings", func(t *testing.T) {
		description := `Test

role_type: polecat
rig: null
hook_bead: null
notification_level: null`

		fields := ParseAgentFields(description)

		if fields.RoleType != "polecat" {
			t.Errorf("RoleType = %q, want %q", fields.RoleType, "polecat")
		}
		if fields.Rig != "" {
			t.Errorf("Rig should be empty for null, got %q", fields.Rig)
		}
		if fields.HookBead != "" {
			t.Errorf("HookBead should be empty for null, got %q", fields.HookBead)
		}
		if fields.NotificationLevel != "" {
			t.Errorf("NotificationLevel should be empty for null, got %q", fields.NotificationLevel)
		}
	})

	t.Run("case insensitive keys", func(t *testing.T) {
		description := `Test

ROLE_TYPE: polecat
Rig: gongshow
Agent_State: working`

		fields := ParseAgentFields(description)

		if fields.RoleType != "polecat" {
			t.Errorf("RoleType = %q, want %q", fields.RoleType, "polecat")
		}
		if fields.Rig != "gongshow" {
			t.Errorf("Rig = %q, want %q", fields.Rig, "gongshow")
		}
		if fields.AgentState != "working" {
			t.Errorf("AgentState = %q, want %q", fields.AgentState, "working")
		}
	})

	t.Run("handles extra whitespace", func(t *testing.T) {
		description := `Test

  role_type:   polecat
  rig:  gongshow  `

		fields := ParseAgentFields(description)

		if fields.RoleType != "polecat" {
			t.Errorf("RoleType = %q, want %q", fields.RoleType, "polecat")
		}
		if fields.Rig != "gongshow" {
			t.Errorf("Rig = %q, want %q", fields.Rig, "gongshow")
		}
	})

	t.Run("ignores lines without colons", func(t *testing.T) {
		description := `Toast
This is a polecat agent
role_type: polecat
Some other text here
rig: gongshow`

		fields := ParseAgentFields(description)

		if fields.RoleType != "polecat" {
			t.Errorf("RoleType = %q, want %q", fields.RoleType, "polecat")
		}
		if fields.Rig != "gongshow" {
			t.Errorf("Rig = %q, want %q", fields.Rig, "gongshow")
		}
	})
}

func TestFormatAndParseAgentFieldsRoundTrip(t *testing.T) {
	original := &AgentFields{
		RoleType:          "polecat",
		Rig:               "gongshow",
		AgentState:        "working",
		HookBead:          "go-abc",
		RoleBead:          "go-role-123",
		CleanupStatus:     "has_uncommitted",
		ActiveMR:          "mr-456",
		NotificationLevel: "muted",
	}

	formatted := FormatAgentDescription("Toast", original)
	parsed := ParseAgentFields(formatted)

	if parsed.RoleType != original.RoleType {
		t.Errorf("RoleType mismatch: got %q, want %q", parsed.RoleType, original.RoleType)
	}
	if parsed.Rig != original.Rig {
		t.Errorf("Rig mismatch: got %q, want %q", parsed.Rig, original.Rig)
	}
	if parsed.AgentState != original.AgentState {
		t.Errorf("AgentState mismatch: got %q, want %q", parsed.AgentState, original.AgentState)
	}
	if parsed.HookBead != original.HookBead {
		t.Errorf("HookBead mismatch: got %q, want %q", parsed.HookBead, original.HookBead)
	}
	if parsed.RoleBead != original.RoleBead {
		t.Errorf("RoleBead mismatch: got %q, want %q", parsed.RoleBead, original.RoleBead)
	}
	if parsed.CleanupStatus != original.CleanupStatus {
		t.Errorf("CleanupStatus mismatch: got %q, want %q", parsed.CleanupStatus, original.CleanupStatus)
	}
	if parsed.ActiveMR != original.ActiveMR {
		t.Errorf("ActiveMR mismatch: got %q, want %q", parsed.ActiveMR, original.ActiveMR)
	}
	if parsed.NotificationLevel != original.NotificationLevel {
		t.Errorf("NotificationLevel mismatch: got %q, want %q", parsed.NotificationLevel, original.NotificationLevel)
	}
}

func TestAgentFieldsEmptyRoundTrip(t *testing.T) {
	// Test that empty fields round-trip correctly through null
	original := &AgentFields{
		RoleType:   "polecat",
		Rig:        "gongshow",
		AgentState: "working",
		// All other fields empty
	}

	formatted := FormatAgentDescription("Toast", original)
	parsed := ParseAgentFields(formatted)

	if parsed.RoleType != original.RoleType {
		t.Errorf("RoleType mismatch: got %q, want %q", parsed.RoleType, original.RoleType)
	}
	if parsed.Rig != original.Rig {
		t.Errorf("Rig mismatch: got %q, want %q", parsed.Rig, original.Rig)
	}
	if parsed.HookBead != "" {
		t.Errorf("HookBead should be empty, got %q", parsed.HookBead)
	}
	if parsed.RoleBead != "" {
		t.Errorf("RoleBead should be empty, got %q", parsed.RoleBead)
	}
	if parsed.CleanupStatus != "" {
		t.Errorf("CleanupStatus should be empty, got %q", parsed.CleanupStatus)
	}
	if parsed.ActiveMR != "" {
		t.Errorf("ActiveMR should be empty, got %q", parsed.ActiveMR)
	}
	if parsed.NotificationLevel != "" {
		t.Errorf("NotificationLevel should be empty, got %q", parsed.NotificationLevel)
	}
}

func TestAgentFieldsRoleTypes(t *testing.T) {
	roleTypes := []string{"polecat", "witness", "refinery", "deacon", "mayor", "crew"}

	for _, roleType := range roleTypes {
		t.Run(roleType, func(t *testing.T) {
			fields := &AgentFields{RoleType: roleType}
			formatted := FormatAgentDescription("Test", fields)
			parsed := ParseAgentFields(formatted)

			if parsed.RoleType != roleType {
				t.Errorf("RoleType = %q, want %q", parsed.RoleType, roleType)
			}
		})
	}
}

func TestAgentFieldsCleanupStatuses(t *testing.T) {
	statuses := []string{"clean", "has_uncommitted", "has_stash", "has_unpushed"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			fields := &AgentFields{
				RoleType:      "polecat",
				CleanupStatus: status,
			}
			formatted := FormatAgentDescription("Test", fields)
			parsed := ParseAgentFields(formatted)

			if parsed.CleanupStatus != status {
				t.Errorf("CleanupStatus = %q, want %q", parsed.CleanupStatus, status)
			}
		})
	}
}
