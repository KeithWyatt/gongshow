package cmd

import (
	"testing"
)

func TestResolveNudgePattern(t *testing.T) {
	// Create test agent sessions (mayor/deacon use hq- prefix)
	agents := []*AgentSession{
		{Name: "hq-mayor", Type: AgentMayor},
		{Name: "hq-deacon", Type: AgentDeacon},
		{Name: "gt-gongshow-witness", Type: AgentWitness, Rig: "gongshow"},
		{Name: "gt-gongshow-refinery", Type: AgentRefinery, Rig: "gongshow"},
		{Name: "gt-gongshow-crew-max", Type: AgentCrew, Rig: "gongshow", AgentName: "max"},
		{Name: "gt-gongshow-crew-jack", Type: AgentCrew, Rig: "gongshow", AgentName: "jack"},
		{Name: "gt-gongshow-alpha", Type: AgentPolecat, Rig: "gongshow", AgentName: "alpha"},
		{Name: "gt-gongshow-beta", Type: AgentPolecat, Rig: "gongshow", AgentName: "beta"},
		{Name: "gt-beads-witness", Type: AgentWitness, Rig: "beads"},
		{Name: "gt-beads-gamma", Type: AgentPolecat, Rig: "beads", AgentName: "gamma"},
	}

	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:     "mayor special case",
			pattern:  "mayor",
			expected: []string{"hq-mayor"},
		},
		{
			name:     "deacon special case",
			pattern:  "deacon",
			expected: []string{"hq-deacon"},
		},
		{
			name:     "specific witness",
			pattern:  "gongshow/witness",
			expected: []string{"gt-gongshow-witness"},
		},
		{
			name:     "all witnesses",
			pattern:  "*/witness",
			expected: []string{"gt-gongshow-witness", "gt-beads-witness"},
		},
		{
			name:     "specific refinery",
			pattern:  "gongshow/refinery",
			expected: []string{"gt-gongshow-refinery"},
		},
		{
			name:     "all polecats in rig",
			pattern:  "gongshow/polecats/*",
			expected: []string{"gt-gongshow-alpha", "gt-gongshow-beta"},
		},
		{
			name:     "specific polecat",
			pattern:  "gongshow/polecats/alpha",
			expected: []string{"gt-gongshow-alpha"},
		},
		{
			name:     "all crew in rig",
			pattern:  "gongshow/crew/*",
			expected: []string{"gt-gongshow-crew-max", "gt-gongshow-crew-jack"},
		},
		{
			name:     "specific crew member",
			pattern:  "gongshow/crew/max",
			expected: []string{"gt-gongshow-crew-max"},
		},
		{
			name:     "legacy polecat format",
			pattern:  "gongshow/alpha",
			expected: []string{"gt-gongshow-alpha"},
		},
		{
			name:     "no matches",
			pattern:  "nonexistent/polecats/*",
			expected: nil,
		},
		{
			name:     "invalid pattern",
			pattern:  "invalid",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveNudgePattern(tt.pattern, agents)

			if len(got) != len(tt.expected) {
				t.Errorf("resolveNudgePattern(%q) returned %d results, want %d: got %v, want %v",
					tt.pattern, len(got), len(tt.expected), got, tt.expected)
				return
			}

			// Check each expected value is present
			gotMap := make(map[string]bool)
			for _, g := range got {
				gotMap[g] = true
			}
			for _, e := range tt.expected {
				if !gotMap[e] {
					t.Errorf("resolveNudgePattern(%q) missing expected %q, got %v",
						tt.pattern, e, got)
				}
			}
		})
	}
}
