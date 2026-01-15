package mayor

import (
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	townRoot := "/tmp/test-town"
	m := NewManager(townRoot)

	if m.townRoot != townRoot {
		t.Errorf("townRoot = %q, want %q", m.townRoot, townRoot)
	}
}

func TestMayorDir(t *testing.T) {
	townRoot := "/tmp/test-town"
	m := NewManager(townRoot)

	expected := filepath.Join(townRoot, "mayor")
	if got := m.mayorDir(); got != expected {
		t.Errorf("mayorDir() = %q, want %q", got, expected)
	}
}

func TestSessionName(t *testing.T) {
	t.Run("package function", func(t *testing.T) {
		name := SessionName()
		if name == "" {
			t.Error("SessionName() should not be empty")
		}
		// Should follow the hq-mayor pattern
		if name != "hq-mayor" {
			t.Errorf("SessionName() = %q, want %q", name, "hq-mayor")
		}
	})

	t.Run("manager method", func(t *testing.T) {
		m := NewManager("/tmp/test")
		name := m.SessionName()
		if name == "" {
			t.Error("Manager.SessionName() should not be empty")
		}
		// Should match package function
		if name != SessionName() {
			t.Errorf("Manager.SessionName() = %q, want %q", name, SessionName())
		}
	})
}

func TestErrors(t *testing.T) {
	t.Run("ErrNotRunning", func(t *testing.T) {
		if ErrNotRunning == nil {
			t.Error("ErrNotRunning should not be nil")
		}
		if ErrNotRunning.Error() != "mayor not running" {
			t.Errorf("ErrNotRunning.Error() = %q, want %q", ErrNotRunning.Error(), "mayor not running")
		}
	})

	t.Run("ErrAlreadyRunning", func(t *testing.T) {
		if ErrAlreadyRunning == nil {
			t.Error("ErrAlreadyRunning should not be nil")
		}
		if ErrAlreadyRunning.Error() != "mayor already running" {
			t.Errorf("ErrAlreadyRunning.Error() = %q, want %q", ErrAlreadyRunning.Error(), "mayor already running")
		}
	})
}

func TestManagerPaths(t *testing.T) {
	tests := []struct {
		name     string
		townRoot string
		expected string
	}{
		{
			name:     "simple path",
			townRoot: "/home/user/gt",
			expected: "/home/user/gt/mayor",
		},
		{
			name:     "nested path",
			townRoot: "/var/lib/gongshow/town",
			expected: "/var/lib/gongshow/town/mayor",
		},
		{
			name:     "relative path",
			townRoot: "./test-town",
			expected: "test-town/mayor",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := NewManager(tc.townRoot)
			got := m.mayorDir()
			if got != tc.expected {
				t.Errorf("mayorDir() = %q, want %q", got, tc.expected)
			}
		})
	}
}
