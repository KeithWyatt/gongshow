package mail

import (
	"testing"
)

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern string
		address string
		want    bool
	}{
		// Exact matches
		{"gongshow/witness", "gongshow/witness", true},
		{"mayor/", "mayor/", true},

		// Wildcard matches
		{"*/witness", "gongshow/witness", true},
		{"*/witness", "beads/witness", true},
		{"gongshow/*", "gongshow/witness", true},
		{"gongshow/*", "gongshow/refinery", true},
		{"gongshow/crew/*", "gongshow/crew/max", true},

		// Non-matches
		{"*/witness", "gongshow/refinery", false},
		{"gongshow/*", "beads/witness", false},
		{"gongshow/crew/*", "gongshow/polecats/Toast", false},

		// Different path lengths
		{"gongshow/*", "gongshow/crew/max", false},      // * matches single segment
		{"gongshow/*/*", "gongshow/crew/max", true},     // Multiple wildcards
		{"*/*", "gongshow/witness", true},              // Both wildcards
		{"*/*/*", "gongshow/crew/max", true},           // Three-level wildcard
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.address, func(t *testing.T) {
			got := matchPattern(tt.pattern, tt.address)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.pattern, tt.address, got, tt.want)
			}
		})
	}
}

func TestAgentBeadIDToAddress(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		// Town-level agents
		{"gt-mayor", "mayor/"},
		{"gt-deacon", "deacon/"},

		// Rig singletons
		{"gt-gongshow-witness", "gongshow/witness"},
		{"gt-gongshow-refinery", "gongshow/refinery"},
		{"gt-beads-witness", "beads/witness"},

		// Named agents
		{"gt-gongshow-crew-max", "gongshow/crew/max"},
		{"gt-gongshow-polecat-Toast", "gongshow/polecat/Toast"},
		{"gt-beads-crew-wolf", "beads/crew/wolf"},

		// Agent with hyphen in name
		{"gt-gongshow-crew-max-v2", "gongshow/crew/max-v2"},
		{"gt-gongshow-polecat-my-agent", "gongshow/polecat/my-agent"},

		// Invalid
		{"invalid", ""},
		{"not-gt-prefix", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got := agentBeadIDToAddress(tt.id)
			if got != tt.want {
				t.Errorf("agentBeadIDToAddress(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestResolverResolve_DirectAddresses(t *testing.T) {
	resolver := NewResolver(nil, "")

	tests := []struct {
		name    string
		address string
		want    RecipientType
		wantLen int
	}{
		// Direct agent addresses
		{"direct agent", "gongshow/witness", RecipientAgent, 1},
		{"direct crew", "gongshow/crew/max", RecipientAgent, 1},
		{"mayor", "mayor/", RecipientAgent, 1},

		// Legacy prefixes (pass-through)
		{"list prefix", "list:oncall", RecipientAgent, 1},
		{"announce prefix", "announce:alerts", RecipientAgent, 1},

		// Explicit type prefixes
		{"queue prefix", "queue:work", RecipientQueue, 1},
		{"channel prefix", "channel:alerts", RecipientChannel, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolver.Resolve(tt.address)
			if err != nil {
				t.Fatalf("Resolve(%q) error: %v", tt.address, err)
			}
			if len(got) != tt.wantLen {
				t.Errorf("Resolve(%q) returned %d recipients, want %d", tt.address, len(got), tt.wantLen)
			}
			if len(got) > 0 && got[0].Type != tt.want {
				t.Errorf("Resolve(%q)[0].Type = %v, want %v", tt.address, got[0].Type, tt.want)
			}
		})
	}
}

func TestResolverResolve_AtPatterns(t *testing.T) {
	// Without beads, @patterns are passed through for existing router
	resolver := NewResolver(nil, "")

	tests := []struct {
		address string
	}{
		{"@town"},
		{"@witnesses"},
		{"@rig/gongshow"},
		{"@overseer"},
	}

	for _, tt := range tests {
		t.Run(tt.address, func(t *testing.T) {
			got, err := resolver.Resolve(tt.address)
			if err != nil {
				t.Fatalf("Resolve(%q) error: %v", tt.address, err)
			}
			if len(got) != 1 {
				t.Errorf("Resolve(%q) returned %d recipients, want 1", tt.address, len(got))
			}
			// Without beads, @patterns pass through unchanged
			if got[0].Address != tt.address {
				t.Errorf("Resolve(%q) = %q, want pass-through", tt.address, got[0].Address)
			}
		})
	}
}

func TestResolverResolve_UnknownName(t *testing.T) {
	resolver := NewResolver(nil, "")

	// A bare name without prefix should fail if not found
	_, err := resolver.Resolve("unknown-name")
	if err == nil {
		t.Error("Resolve(\"unknown-name\") should return error for unknown name")
	}
}
