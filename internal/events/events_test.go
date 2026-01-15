package events

import (
	"encoding/json"
	"testing"
)

func TestEventConstants(t *testing.T) {
	// Verify visibility constants
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"VisibilityAudit", VisibilityAudit, "audit"},
		{"VisibilityFeed", VisibilityFeed, "feed"},
		{"VisibilityBoth", VisibilityBoth, "both"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.expected {
				t.Errorf("%s = %q, want %q", tc.name, tc.got, tc.expected)
			}
		})
	}
}

func TestEventTypes(t *testing.T) {
	// Verify event type constants are non-empty
	types := []struct {
		name  string
		value string
	}{
		{"TypeSling", TypeSling},
		{"TypeHook", TypeHook},
		{"TypeUnhook", TypeUnhook},
		{"TypeHandoff", TypeHandoff},
		{"TypeDone", TypeDone},
		{"TypeMail", TypeMail},
		{"TypeSpawn", TypeSpawn},
		{"TypeKill", TypeKill},
		{"TypeNudge", TypeNudge},
		{"TypeBoot", TypeBoot},
		{"TypeHalt", TypeHalt},
		{"TypeSessionStart", TypeSessionStart},
		{"TypeSessionEnd", TypeSessionEnd},
		{"TypeSessionDeath", TypeSessionDeath},
		{"TypeMassDeath", TypeMassDeath},
		{"TypePatrolStarted", TypePatrolStarted},
		{"TypePolecatChecked", TypePolecatChecked},
		{"TypePolecatNudged", TypePolecatNudged},
		{"TypeEscalationSent", TypeEscalationSent},
		{"TypeEscalationAcked", TypeEscalationAcked},
		{"TypeEscalationClosed", TypeEscalationClosed},
		{"TypePatrolComplete", TypePatrolComplete},
		{"TypeMergeStarted", TypeMergeStarted},
		{"TypeMerged", TypeMerged},
		{"TypeMergeFailed", TypeMergeFailed},
		{"TypeMergeSkipped", TypeMergeSkipped},
	}

	for _, tc := range types {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value == "" {
				t.Errorf("%s should not be empty", tc.name)
			}
		})
	}
}

func TestEventJSONMarshaling(t *testing.T) {
	event := Event{
		Timestamp:  "2024-01-15T10:00:00Z",
		Source:     "gt",
		Type:       TypeSling,
		Actor:      "gongshow/crew/marge",
		Payload:    map[string]interface{}{"bead": "go-abc", "target": "polecat"},
		Visibility: VisibilityFeed,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}

	if decoded.Timestamp != event.Timestamp {
		t.Errorf("Timestamp = %q, want %q", decoded.Timestamp, event.Timestamp)
	}
	if decoded.Source != event.Source {
		t.Errorf("Source = %q, want %q", decoded.Source, event.Source)
	}
	if decoded.Type != event.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, event.Type)
	}
	if decoded.Actor != event.Actor {
		t.Errorf("Actor = %q, want %q", decoded.Actor, event.Actor)
	}
	if decoded.Visibility != event.Visibility {
		t.Errorf("Visibility = %q, want %q", decoded.Visibility, event.Visibility)
	}
}

func TestSlingPayload(t *testing.T) {
	payload := SlingPayload("go-abc", "polecat")

	if payload["bead"] != "go-abc" {
		t.Errorf("bead = %v, want %q", payload["bead"], "go-abc")
	}
	if payload["target"] != "polecat" {
		t.Errorf("target = %v, want %q", payload["target"], "polecat")
	}
}

func TestHookPayload(t *testing.T) {
	payload := HookPayload("go-xyz")

	if payload["bead"] != "go-xyz" {
		t.Errorf("bead = %v, want %q", payload["bead"], "go-xyz")
	}
}

func TestHandoffPayload(t *testing.T) {
	t.Run("with subject", func(t *testing.T) {
		payload := HandoffPayload("Working on auth", true)

		if payload["to_session"] != true {
			t.Errorf("to_session = %v, want true", payload["to_session"])
		}
		if payload["subject"] != "Working on auth" {
			t.Errorf("subject = %v, want %q", payload["subject"], "Working on auth")
		}
	})

	t.Run("without subject", func(t *testing.T) {
		payload := HandoffPayload("", false)

		if payload["to_session"] != false {
			t.Errorf("to_session = %v, want false", payload["to_session"])
		}
		if _, exists := payload["subject"]; exists {
			t.Error("subject should not be present when empty")
		}
	})
}

func TestDonePayload(t *testing.T) {
	payload := DonePayload("go-abc", "feature/auth")

	if payload["bead"] != "go-abc" {
		t.Errorf("bead = %v, want %q", payload["bead"], "go-abc")
	}
	if payload["branch"] != "feature/auth" {
		t.Errorf("branch = %v, want %q", payload["branch"], "feature/auth")
	}
}

func TestMailPayload(t *testing.T) {
	payload := MailPayload("mayor", "Need help")

	if payload["to"] != "mayor" {
		t.Errorf("to = %v, want %q", payload["to"], "mayor")
	}
	if payload["subject"] != "Need help" {
		t.Errorf("subject = %v, want %q", payload["subject"], "Need help")
	}
}

func TestSpawnPayload(t *testing.T) {
	payload := SpawnPayload("gongshow", "Toast")

	if payload["rig"] != "gongshow" {
		t.Errorf("rig = %v, want %q", payload["rig"], "gongshow")
	}
	if payload["polecat"] != "Toast" {
		t.Errorf("polecat = %v, want %q", payload["polecat"], "Toast")
	}
}

func TestBootPayload(t *testing.T) {
	agents := []string{"witness", "refinery"}
	payload := BootPayload("gongshow", agents)

	if payload["rig"] != "gongshow" {
		t.Errorf("rig = %v, want %q", payload["rig"], "gongshow")
	}
	if agents, ok := payload["agents"].([]string); !ok || len(agents) != 2 {
		t.Errorf("agents = %v, want slice with 2 elements", payload["agents"])
	}
}

func TestMergePayload(t *testing.T) {
	t.Run("with reason", func(t *testing.T) {
		payload := MergePayload("mr-123", "Toast", "feature/auth", "conflicts")

		if payload["mr"] != "mr-123" {
			t.Errorf("mr = %v, want %q", payload["mr"], "mr-123")
		}
		if payload["worker"] != "Toast" {
			t.Errorf("worker = %v, want %q", payload["worker"], "Toast")
		}
		if payload["branch"] != "feature/auth" {
			t.Errorf("branch = %v, want %q", payload["branch"], "feature/auth")
		}
		if payload["reason"] != "conflicts" {
			t.Errorf("reason = %v, want %q", payload["reason"], "conflicts")
		}
	})

	t.Run("without reason", func(t *testing.T) {
		payload := MergePayload("mr-123", "Toast", "feature/auth", "")

		if _, exists := payload["reason"]; exists {
			t.Error("reason should not be present when empty")
		}
	})
}

func TestPatrolPayload(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		payload := PatrolPayload("gongshow", 5, "All healthy")

		if payload["rig"] != "gongshow" {
			t.Errorf("rig = %v, want %q", payload["rig"], "gongshow")
		}
		if payload["polecat_count"] != 5 {
			t.Errorf("polecat_count = %v, want 5", payload["polecat_count"])
		}
		if payload["message"] != "All healthy" {
			t.Errorf("message = %v, want %q", payload["message"], "All healthy")
		}
	})

	t.Run("without message", func(t *testing.T) {
		payload := PatrolPayload("gongshow", 3, "")

		if _, exists := payload["message"]; exists {
			t.Error("message should not be present when empty")
		}
	})
}

func TestPolecatCheckPayload(t *testing.T) {
	t.Run("with issue", func(t *testing.T) {
		payload := PolecatCheckPayload("gongshow", "Toast", "working", "go-abc")

		if payload["rig"] != "gongshow" {
			t.Errorf("rig = %v, want %q", payload["rig"], "gongshow")
		}
		if payload["polecat"] != "Toast" {
			t.Errorf("polecat = %v, want %q", payload["polecat"], "Toast")
		}
		if payload["status"] != "working" {
			t.Errorf("status = %v, want %q", payload["status"], "working")
		}
		if payload["issue"] != "go-abc" {
			t.Errorf("issue = %v, want %q", payload["issue"], "go-abc")
		}
	})

	t.Run("without issue", func(t *testing.T) {
		payload := PolecatCheckPayload("gongshow", "Toast", "idle", "")

		if _, exists := payload["issue"]; exists {
			t.Error("issue should not be present when empty")
		}
	})
}

func TestNudgePayload(t *testing.T) {
	payload := NudgePayload("gongshow", "witness", "patrol needed")

	if payload["rig"] != "gongshow" {
		t.Errorf("rig = %v, want %q", payload["rig"], "gongshow")
	}
	if payload["target"] != "witness" {
		t.Errorf("target = %v, want %q", payload["target"], "witness")
	}
	if payload["reason"] != "patrol needed" {
		t.Errorf("reason = %v, want %q", payload["reason"], "patrol needed")
	}
}

func TestEscalationPayload(t *testing.T) {
	payload := EscalationPayload("gongshow", "Toast", "human", "stuck on auth")

	if payload["rig"] != "gongshow" {
		t.Errorf("rig = %v, want %q", payload["rig"], "gongshow")
	}
	if payload["target"] != "Toast" {
		t.Errorf("target = %v, want %q", payload["target"], "Toast")
	}
	if payload["to"] != "human" {
		t.Errorf("to = %v, want %q", payload["to"], "human")
	}
	if payload["reason"] != "stuck on auth" {
		t.Errorf("reason = %v, want %q", payload["reason"], "stuck on auth")
	}
}

func TestUnhookPayload(t *testing.T) {
	payload := UnhookPayload("go-abc")

	if payload["bead"] != "go-abc" {
		t.Errorf("bead = %v, want %q", payload["bead"], "go-abc")
	}
}

func TestKillPayload(t *testing.T) {
	payload := KillPayload("gongshow", "Toast", "zombie cleanup")

	if payload["rig"] != "gongshow" {
		t.Errorf("rig = %v, want %q", payload["rig"], "gongshow")
	}
	if payload["target"] != "Toast" {
		t.Errorf("target = %v, want %q", payload["target"], "Toast")
	}
	if payload["reason"] != "zombie cleanup" {
		t.Errorf("reason = %v, want %q", payload["reason"], "zombie cleanup")
	}
}

func TestHaltPayload(t *testing.T) {
	services := []string{"daemon", "witness", "refinery"}
	payload := HaltPayload(services)

	if services, ok := payload["services"].([]string); !ok || len(services) != 3 {
		t.Errorf("services = %v, want slice with 3 elements", payload["services"])
	}
}

func TestSessionDeathPayload(t *testing.T) {
	payload := SessionDeathPayload("gt-gongshow-Toast", "gongshow/polecats/Toast", "zombie cleanup", "daemon")

	if payload["session"] != "gt-gongshow-Toast" {
		t.Errorf("session = %v, want %q", payload["session"], "gt-gongshow-Toast")
	}
	if payload["agent"] != "gongshow/polecats/Toast" {
		t.Errorf("agent = %v, want %q", payload["agent"], "gongshow/polecats/Toast")
	}
	if payload["reason"] != "zombie cleanup" {
		t.Errorf("reason = %v, want %q", payload["reason"], "zombie cleanup")
	}
	if payload["caller"] != "daemon" {
		t.Errorf("caller = %v, want %q", payload["caller"], "daemon")
	}
}

func TestMassDeathPayload(t *testing.T) {
	t.Run("with cause", func(t *testing.T) {
		sessions := []string{"sess1", "sess2", "sess3"}
		payload := MassDeathPayload(3, "5s", sessions, "tmux crash")

		if payload["count"] != 3 {
			t.Errorf("count = %v, want 3", payload["count"])
		}
		if payload["window"] != "5s" {
			t.Errorf("window = %v, want %q", payload["window"], "5s")
		}
		if payload["possible_cause"] != "tmux crash" {
			t.Errorf("possible_cause = %v, want %q", payload["possible_cause"], "tmux crash")
		}
	})

	t.Run("without cause", func(t *testing.T) {
		payload := MassDeathPayload(2, "10s", []string{"a", "b"}, "")

		if _, exists := payload["possible_cause"]; exists {
			t.Error("possible_cause should not be present when empty")
		}
	})
}

func TestSessionPayload(t *testing.T) {
	t.Run("with all fields", func(t *testing.T) {
		payload := SessionPayload("uuid-123", "gongshow/crew/marge", "auth bug", "/home/user/project")

		if payload["session_id"] != "uuid-123" {
			t.Errorf("session_id = %v, want %q", payload["session_id"], "uuid-123")
		}
		if payload["role"] != "gongshow/crew/marge" {
			t.Errorf("role = %v, want %q", payload["role"], "gongshow/crew/marge")
		}
		if payload["topic"] != "auth bug" {
			t.Errorf("topic = %v, want %q", payload["topic"], "auth bug")
		}
		if payload["cwd"] != "/home/user/project" {
			t.Errorf("cwd = %v, want %q", payload["cwd"], "/home/user/project")
		}
		// actor_pid should be present
		if _, exists := payload["actor_pid"]; !exists {
			t.Error("actor_pid should be present")
		}
	})

	t.Run("without optional fields", func(t *testing.T) {
		payload := SessionPayload("uuid-456", "mayor", "", "")

		if payload["session_id"] != "uuid-456" {
			t.Errorf("session_id = %v, want %q", payload["session_id"], "uuid-456")
		}
		if _, exists := payload["topic"]; exists {
			t.Error("topic should not be present when empty")
		}
		if _, exists := payload["cwd"]; exists {
			t.Error("cwd should not be present when empty")
		}
	})
}

func TestEventsFile(t *testing.T) {
	if EventsFile != ".events.jsonl" {
		t.Errorf("EventsFile = %q, want %q", EventsFile, ".events.jsonl")
	}
}
