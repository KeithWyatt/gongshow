package witness

import (
	"strings"
	"testing"

	"github.com/KeithWyatt/gongshow/internal/beads"
)

func TestBuildWitnessStartCommand_UsesRoleConfig(t *testing.T) {
	roleConfig := &beads.RoleConfig{
		StartCommand: "exec run --town {town} --rig {rig} --role {role}",
	}

	got, err := buildWitnessStartCommand("/town/rig", "gongshow", "/town", "", roleConfig)
	if err != nil {
		t.Fatalf("buildWitnessStartCommand: %v", err)
	}

	want := "exec run --town /town --rig gongshow --role witness"
	if got != want {
		t.Errorf("buildWitnessStartCommand = %q, want %q", got, want)
	}
}

func TestBuildWitnessStartCommand_DefaultsToRuntime(t *testing.T) {
	got, err := buildWitnessStartCommand("/town/rig", "gongshow", "/town", "", nil)
	if err != nil {
		t.Fatalf("buildWitnessStartCommand: %v", err)
	}

	if !strings.Contains(got, "GT_ROLE=witness") {
		t.Errorf("expected GT_ROLE=witness in command, got %q", got)
	}
	if !strings.Contains(got, "BD_ACTOR=gongshow/witness") {
		t.Errorf("expected BD_ACTOR=gongshow/witness in command, got %q", got)
	}
}

func TestBuildWitnessStartCommand_AgentOverrideWins(t *testing.T) {
	roleConfig := &beads.RoleConfig{
		StartCommand: "exec run --role {role}",
	}

	got, err := buildWitnessStartCommand("/town/rig", "gongshow", "/town", "codex", roleConfig)
	if err != nil {
		t.Fatalf("buildWitnessStartCommand: %v", err)
	}
	if strings.Contains(got, "exec run") {
		t.Fatalf("expected agent override to bypass role start_command, got %q", got)
	}
	if !strings.Contains(got, "GT_ROLE=witness") {
		t.Errorf("expected GT_ROLE=witness in command, got %q", got)
	}
}
