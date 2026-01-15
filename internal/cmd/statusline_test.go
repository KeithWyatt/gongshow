package cmd

import "testing"

func TestCategorizeSessionRig(t *testing.T) {
	tests := []struct {
		session string
		wantRig string
	}{
		// Standard polecat sessions
		{"gt-gongshow-slit", "gongshow"},
		{"gt-gongshow-Toast", "gongshow"},
		{"gt-myrig-worker", "myrig"},

		// Crew sessions
		{"gt-gongshow-crew-max", "gongshow"},
		{"gt-myrig-crew-user", "myrig"},

		// Witness sessions (canonical format: gt-<rig>-witness)
		{"gt-gongshow-witness", "gongshow"},
		{"gt-myrig-witness", "myrig"},
		// Legacy format still works as fallback
		{"gt-witness-gongshow", "gongshow"},
		{"gt-witness-myrig", "myrig"},

		// Refinery sessions
		{"gt-gongshow-refinery", "gongshow"},
		{"gt-myrig-refinery", "myrig"},

		// Edge cases
		{"gt-a-b", "a"}, // minimum valid

		// Town-level agents (no rig, use hq- prefix)
		{"hq-mayor", ""},
		{"hq-deacon", ""},
	}

	for _, tt := range tests {
		t.Run(tt.session, func(t *testing.T) {
			agent := categorizeSession(tt.session)
			gotRig := ""
			if agent != nil {
				gotRig = agent.Rig
			}
			if gotRig != tt.wantRig {
				t.Errorf("categorizeSession(%q).Rig = %q, want %q", tt.session, gotRig, tt.wantRig)
			}
		})
	}
}

func TestCategorizeSessionType(t *testing.T) {
	tests := []struct {
		session  string
		wantType AgentType
	}{
		// Polecat sessions
		{"gt-gongshow-slit", AgentPolecat},
		{"gt-gongshow-Toast", AgentPolecat},
		{"gt-myrig-worker", AgentPolecat},
		{"gt-a-b", AgentPolecat},

		// Non-polecat sessions
		{"gt-gongshow-witness", AgentWitness}, // canonical format
		{"gt-witness-gongshow", AgentWitness}, // legacy fallback
		{"gt-gongshow-refinery", AgentRefinery},
		{"gt-gongshow-crew-max", AgentCrew},
		{"gt-myrig-crew-user", AgentCrew},

		// Town-level agents (hq- prefix)
		{"hq-mayor", AgentMayor},
		{"hq-deacon", AgentDeacon},
	}

	for _, tt := range tests {
		t.Run(tt.session, func(t *testing.T) {
			agent := categorizeSession(tt.session)
			if agent == nil {
				t.Fatalf("categorizeSession(%q) returned nil", tt.session)
			}
			if agent.Type != tt.wantType {
				t.Errorf("categorizeSession(%q).Type = %v, want %v", tt.session, agent.Type, tt.wantType)
			}
		})
	}
}
