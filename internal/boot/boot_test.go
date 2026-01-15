package boot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	townRoot := "/tmp/test-town"
	b := New(townRoot)

	if b.townRoot != townRoot {
		t.Errorf("townRoot = %q, want %q", b.townRoot, townRoot)
	}

	expectedBootDir := filepath.Join(townRoot, "deacon", "dogs", "boot")
	if b.bootDir != expectedBootDir {
		t.Errorf("bootDir = %q, want %q", b.bootDir, expectedBootDir)
	}

	expectedDeaconDir := filepath.Join(townRoot, "deacon")
	if b.deaconDir != expectedDeaconDir {
		t.Errorf("deaconDir = %q, want %q", b.deaconDir, expectedDeaconDir)
	}

	if b.tmux == nil {
		t.Error("tmux manager should not be nil")
	}
}

func TestBootPaths(t *testing.T) {
	townRoot := "/tmp/test-town"
	b := New(townRoot)

	t.Run("Dir", func(t *testing.T) {
		expected := filepath.Join(townRoot, "deacon", "dogs", "boot")
		if got := b.Dir(); got != expected {
			t.Errorf("Dir() = %q, want %q", got, expected)
		}
	})

	t.Run("DeaconDir", func(t *testing.T) {
		expected := filepath.Join(townRoot, "deacon")
		if got := b.DeaconDir(); got != expected {
			t.Errorf("DeaconDir() = %q, want %q", got, expected)
		}
	})

	t.Run("markerPath", func(t *testing.T) {
		expected := filepath.Join(townRoot, "deacon", "dogs", "boot", MarkerFileName)
		if got := b.markerPath(); got != expected {
			t.Errorf("markerPath() = %q, want %q", got, expected)
		}
	})

	t.Run("statusPath", func(t *testing.T) {
		expected := filepath.Join(townRoot, "deacon", "dogs", "boot", StatusFileName)
		if got := b.statusPath(); got != expected {
			t.Errorf("statusPath() = %q, want %q", got, expected)
		}
	})
}

func TestEnsureDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "boot-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	b := New(tmpDir)

	// Directory should not exist initially
	if _, err := os.Stat(b.bootDir); !os.IsNotExist(err) {
		t.Error("boot dir should not exist initially")
	}

	// EnsureDir should create it
	if err := b.EnsureDir(); err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}

	// Directory should now exist
	info, err := os.Stat(b.bootDir)
	if err != nil {
		t.Fatalf("boot dir should exist after EnsureDir: %v", err)
	}
	if !info.IsDir() {
		t.Error("boot dir should be a directory")
	}

	// Calling EnsureDir again should be idempotent
	if err := b.EnsureDir(); err != nil {
		t.Fatalf("EnsureDir() second call error = %v", err)
	}
}

func TestSaveAndLoadStatus(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "boot-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	b := New(tmpDir)

	t.Run("save and load status", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		status := &Status{
			Running:     true,
			StartedAt:   now,
			LastAction:  "wake",
			Target:      "deacon",
		}

		if err := b.SaveStatus(status); err != nil {
			t.Fatalf("SaveStatus() error = %v", err)
		}

		loaded, err := b.LoadStatus()
		if err != nil {
			t.Fatalf("LoadStatus() error = %v", err)
		}

		if loaded.Running != status.Running {
			t.Errorf("Running = %v, want %v", loaded.Running, status.Running)
		}
		if loaded.LastAction != status.LastAction {
			t.Errorf("LastAction = %q, want %q", loaded.LastAction, status.LastAction)
		}
		if loaded.Target != status.Target {
			t.Errorf("Target = %q, want %q", loaded.Target, status.Target)
		}
	})

	t.Run("load nonexistent returns empty status", func(t *testing.T) {
		// Use a fresh temp dir with no status file
		freshDir, err := os.MkdirTemp("", "boot-test-fresh-*")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.RemoveAll(freshDir) }()

		freshBoot := New(freshDir)
		status, err := freshBoot.LoadStatus()
		if err != nil {
			t.Fatalf("LoadStatus() on nonexistent file should not error: %v", err)
		}
		if status.Running {
			t.Error("empty status should have Running=false")
		}
		if status.LastAction != "" {
			t.Errorf("empty status LastAction = %q, want empty", status.LastAction)
		}
	})

	t.Run("status JSON format", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		status := &Status{
			Running:     true,
			StartedAt:   now,
			CompletedAt: now.Add(time.Minute),
			LastAction:  "nudge",
			Target:      "witness",
			Error:       "test error",
		}

		data, err := json.Marshal(status)
		if err != nil {
			t.Fatalf("json.Marshal error = %v", err)
		}

		var decoded Status
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("json.Unmarshal error = %v", err)
		}

		if decoded.Error != "test error" {
			t.Errorf("Error = %q, want %q", decoded.Error, "test error")
		}
	})
}

func TestAcquireAndReleaseLock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "boot-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	b := New(tmpDir)

	// Note: AcquireLock checks IsRunning() which queries tmux.
	// Since we don't have tmux in tests, we test the file operations directly.

	t.Run("marker file operations", func(t *testing.T) {
		// Ensure directory exists
		if err := b.EnsureDir(); err != nil {
			t.Fatal(err)
		}

		// Create marker file manually (simulating AcquireLock's file creation)
		markerPath := b.markerPath()
		f, err := os.Create(markerPath)
		if err != nil {
			t.Fatalf("creating marker: %v", err)
		}
		if err := f.Close(); err != nil {
			t.Fatalf("closing marker: %v", err)
		}

		// Verify marker exists
		if _, err := os.Stat(markerPath); err != nil {
			t.Errorf("marker should exist after creation: %v", err)
		}

		// ReleaseLock should remove it
		if err := b.ReleaseLock(); err != nil {
			t.Fatalf("ReleaseLock() error = %v", err)
		}

		// Marker should be gone
		if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
			t.Error("marker should not exist after ReleaseLock")
		}
	})
}

func TestIsDegraded(t *testing.T) {
	t.Run("not degraded by default", func(t *testing.T) {
		// Ensure GT_DEGRADED is not set
		originalValue := os.Getenv("GT_DEGRADED")
		os.Unsetenv("GT_DEGRADED")
		defer func() {
			if originalValue != "" {
				os.Setenv("GT_DEGRADED", originalValue)
			}
		}()

		b := New("/tmp/test")
		if b.IsDegraded() {
			t.Error("IsDegraded() should be false when GT_DEGRADED is not set")
		}
	})

	t.Run("degraded when GT_DEGRADED=true", func(t *testing.T) {
		originalValue := os.Getenv("GT_DEGRADED")
		os.Setenv("GT_DEGRADED", "true")
		defer func() {
			if originalValue != "" {
				os.Setenv("GT_DEGRADED", originalValue)
			} else {
				os.Unsetenv("GT_DEGRADED")
			}
		}()

		b := New("/tmp/test")
		if !b.IsDegraded() {
			t.Error("IsDegraded() should be true when GT_DEGRADED=true")
		}
	})
}

func TestTmuxAccessor(t *testing.T) {
	b := New("/tmp/test")
	if b.Tmux() == nil {
		t.Error("Tmux() should return non-nil tmux manager")
	}
}

func TestStatusConstants(t *testing.T) {
	// Verify constants are defined correctly
	if SessionName != "gt-boot" {
		t.Errorf("SessionName = %q, want %q", SessionName, "gt-boot")
	}
	if MarkerFileName != ".boot-running" {
		t.Errorf("MarkerFileName = %q, want %q", MarkerFileName, ".boot-running")
	}
	if StatusFileName != ".boot-status.json" {
		t.Errorf("StatusFileName = %q, want %q", StatusFileName, ".boot-status.json")
	}
}
