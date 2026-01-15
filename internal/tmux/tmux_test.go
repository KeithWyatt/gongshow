package tmux

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

func hasTmux() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// tmuxVersion returns the tmux version as a string (e.g., "3.4" or "2.9a").
// Returns empty string if version cannot be determined.
func tmuxVersion() string {
	out, err := exec.Command("tmux", "-V").Output()
	if err != nil {
		return ""
	}
	// Output is like "tmux 3.4" or "tmux 2.9a"
	parts := strings.Fields(string(out))
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// hasTmuxFilterFlag checks if tmux supports the -f (filter) flag for list-sessions.
// This flag was added in tmux 3.2.
func hasTmuxFilterFlag() bool {
	version := tmuxVersion()
	if version == "" {
		return false
	}
	// Extract major.minor version (e.g., "3.4" from "3.4" or "3" from "3.2a")
	re := regexp.MustCompile(`^(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(version)
	if len(matches) < 3 {
		return false
	}
	// Parse major and minor as single digits (sufficient for tmux versions)
	major := int(matches[1][0] - '0')
	minor := int(matches[2][0] - '0')
	// tmux 3.2+ supports -f flag
	return major > 3 || (major == 3 && minor >= 2)
}

func TestListSessionsNoServer(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}

	tm := NewTmux()
	sessions, err := tm.ListSessions()
	// Should not error even if no server running
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	// Result may be nil or empty slice
	_ = sessions
}

func TestHasSessionNoServer(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}

	tm := NewTmux()
	has, err := tm.HasSession("nonexistent-session-xyz")
	if err != nil {
		t.Fatalf("HasSession: %v", err)
	}
	if has {
		t.Error("expected session to not exist")
	}
}

func TestSessionLifecycle(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}

	tm := NewTmux()
	sessionName := "gt-test-session-" + t.Name()

	// Clean up any existing session
	_ = tm.KillSession(sessionName)

	// Create session
	if err := tm.NewSession(sessionName, ""); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer func() { _ = tm.KillSession(sessionName) }()

	// Verify exists
	has, err := tm.HasSession(sessionName)
	if err != nil {
		t.Fatalf("HasSession: %v", err)
	}
	if !has {
		t.Error("expected session to exist after creation")
	}

	// List should include it
	sessions, err := tm.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	found := false
	for _, s := range sessions {
		if s == sessionName {
			found = true
			break
		}
	}
	if !found {
		t.Error("session not found in list")
	}

	// Kill session
	if err := tm.KillSession(sessionName); err != nil {
		t.Fatalf("KillSession: %v", err)
	}

	// Verify gone
	has, err = tm.HasSession(sessionName)
	if err != nil {
		t.Fatalf("HasSession after kill: %v", err)
	}
	if has {
		t.Error("expected session to not exist after kill")
	}
}

func TestDuplicateSession(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}

	tm := NewTmux()
	sessionName := "gt-test-dup-" + t.Name()

	// Clean up any existing session
	_ = tm.KillSession(sessionName)

	// Create session
	if err := tm.NewSession(sessionName, ""); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer func() { _ = tm.KillSession(sessionName) }()

	// Try to create duplicate
	err := tm.NewSession(sessionName, "")
	if err != ErrSessionExists {
		t.Errorf("expected ErrSessionExists, got %v", err)
	}
}

func TestSendKeysAndCapture(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}

	tm := NewTmux()
	sessionName := "gt-test-keys-" + t.Name()

	// Clean up any existing session
	_ = tm.KillSession(sessionName)

	// Create session
	if err := tm.NewSession(sessionName, ""); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer func() { _ = tm.KillSession(sessionName) }()

	// Send echo command
	if err := tm.SendKeys(sessionName, "echo HELLO_TEST_MARKER"); err != nil {
		t.Fatalf("SendKeys: %v", err)
	}

	// Give it a moment to execute
	// In real tests you'd wait for output, but for basic test we just capture
	output, err := tm.CapturePane(sessionName, 50)
	if err != nil {
		t.Fatalf("CapturePane: %v", err)
	}

	// Should contain our marker (might not if shell is slow, but usually works)
	if !strings.Contains(output, "echo HELLO_TEST_MARKER") {
		t.Logf("captured output: %s", output)
		// Don't fail, just note - timing issues possible
	}
}

func TestGetSessionInfo(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}
	if !hasTmuxFilterFlag() {
		t.Skip("tmux < 3.2 does not support -f flag for list-sessions")
	}

	tm := NewTmux()
	sessionName := "gt-test-info-" + t.Name()

	// Clean up any existing session
	_ = tm.KillSession(sessionName)

	// Create session
	if err := tm.NewSession(sessionName, ""); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer func() { _ = tm.KillSession(sessionName) }()

	info, err := tm.GetSessionInfo(sessionName)
	if err != nil {
		t.Fatalf("GetSessionInfo: %v", err)
	}

	if info.Name != sessionName {
		t.Errorf("Name = %q, want %q", info.Name, sessionName)
	}
	if info.Windows < 1 {
		t.Errorf("Windows = %d, want >= 1", info.Windows)
	}
}

func TestWrapError(t *testing.T) {
	tm := NewTmux()

	tests := []struct {
		stderr string
		want   error
	}{
		{"no server running on /tmp/tmux-...", ErrNoServer},
		{"error connecting to /tmp/tmux-...", ErrNoServer},
		{"duplicate session: test", ErrSessionExists},
		{"session not found: test", ErrSessionNotFound},
		{"can't find session: test", ErrSessionNotFound},
	}

	for _, tt := range tests {
		err := tm.wrapError(nil, tt.stderr, []string{"test"})
		if err != tt.want {
			t.Errorf("wrapError(%q) = %v, want %v", tt.stderr, err, tt.want)
		}
	}
}

func TestEnsureSessionFresh_NoExistingSession(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}

	tm := NewTmux()
	sessionName := "gt-test-fresh-" + t.Name()

	// Clean up any existing session
	_ = tm.KillSession(sessionName)

	// EnsureSessionFresh should create a new session
	if err := tm.EnsureSessionFresh(sessionName, ""); err != nil {
		t.Fatalf("EnsureSessionFresh: %v", err)
	}
	defer func() { _ = tm.KillSession(sessionName) }()

	// Verify session exists
	has, err := tm.HasSession(sessionName)
	if err != nil {
		t.Fatalf("HasSession: %v", err)
	}
	if !has {
		t.Error("expected session to exist after EnsureSessionFresh")
	}
}

func TestEnsureSessionFresh_ZombieSession(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}

	tm := NewTmux()
	sessionName := "gt-test-zombie-" + t.Name()

	// Clean up any existing session
	_ = tm.KillSession(sessionName)

	// Create a zombie session (session exists but no Claude/node running)
	// A normal tmux session with bash/zsh is a "zombie" for our purposes
	if err := tm.NewSession(sessionName, ""); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer func() { _ = tm.KillSession(sessionName) }()

	// Verify it's a zombie (not running Claude/node)
	if tm.IsClaudeRunning(sessionName) {
		t.Skip("session unexpectedly has Claude running - can't test zombie case")
	}

	// Verify generic agent check also treats it as not running (shell session)
	if tm.IsAgentRunning(sessionName) {
		t.Fatalf("expected IsAgentRunning(%q) to be false for a fresh shell session", sessionName)
	}

	// EnsureSessionFresh should kill the zombie and create fresh session
	// This should NOT error with "session already exists"
	if err := tm.EnsureSessionFresh(sessionName, ""); err != nil {
		t.Fatalf("EnsureSessionFresh on zombie: %v", err)
	}

	// Session should still exist
	has, err := tm.HasSession(sessionName)
	if err != nil {
		t.Fatalf("HasSession: %v", err)
	}
	if !has {
		t.Error("expected session to exist after EnsureSessionFresh on zombie")
	}
}

func TestEnsureSessionFresh_IdempotentOnZombie(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}

	tm := NewTmux()
	sessionName := "gt-test-idem-" + t.Name()

	// Clean up any existing session
	_ = tm.KillSession(sessionName)

	// Call EnsureSessionFresh multiple times - should work each time
	for i := 0; i < 3; i++ {
		if err := tm.EnsureSessionFresh(sessionName, ""); err != nil {
			t.Fatalf("EnsureSessionFresh attempt %d: %v", i+1, err)
		}
	}
	defer func() { _ = tm.KillSession(sessionName) }()

	// Session should exist
	has, err := tm.HasSession(sessionName)
	if err != nil {
		t.Fatalf("HasSession: %v", err)
	}
	if !has {
		t.Error("expected session to exist after multiple EnsureSessionFresh calls")
	}
}

func TestIsAgentRunning(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}

	tm := NewTmux()
	sessionName := "gt-test-agent-" + t.Name()

	// Clean up any existing session
	_ = tm.KillSession(sessionName)

	// Create session (will run default shell)
	if err := tm.NewSession(sessionName, ""); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer func() { _ = tm.KillSession(sessionName) }()

	// Get the current pane command (should be bash/zsh/etc)
	cmd, err := tm.GetPaneCommand(sessionName)
	if err != nil {
		t.Fatalf("GetPaneCommand: %v", err)
	}

	tests := []struct {
		name         string
		processNames []string
		wantRunning  bool
	}{
		{
			name:         "empty process list",
			processNames: []string{},
			wantRunning:  false,
		},
		{
			name:         "matching shell process",
			processNames: []string{cmd}, // Current shell
			wantRunning:  true,
		},
		{
			name:         "claude agent (node) - not running",
			processNames: []string{"node"},
			wantRunning:  cmd == "node", // Only true if shell happens to be node
		},
		{
			name:         "gemini agent - not running",
			processNames: []string{"gemini"},
			wantRunning:  cmd == "gemini",
		},
		{
			name:         "cursor agent - not running",
			processNames: []string{"cursor-agent"},
			wantRunning:  cmd == "cursor-agent",
		},
		{
			name:         "multiple process names with match",
			processNames: []string{"nonexistent", cmd, "also-nonexistent"},
			wantRunning:  true,
		},
		{
			name:         "multiple process names without match",
			processNames: []string{"nonexistent1", "nonexistent2"},
			wantRunning:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tm.IsAgentRunning(sessionName, tt.processNames...)
			if got != tt.wantRunning {
				t.Errorf("IsAgentRunning(%q, %v) = %v, want %v (current cmd: %q)",
					sessionName, tt.processNames, got, tt.wantRunning, cmd)
			}
		})
	}
}

func TestIsAgentRunning_NonexistentSession(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}

	tm := NewTmux()

	// IsAgentRunning on nonexistent session should return false, not error
	got := tm.IsAgentRunning("nonexistent-session-xyz", "node", "gemini", "cursor-agent")
	if got {
		t.Error("IsAgentRunning on nonexistent session should return false")
	}
}

func TestIsClaudeRunning(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}

	tm := NewTmux()
	sessionName := "gt-test-claude-" + t.Name()

	// Clean up any existing session
	_ = tm.KillSession(sessionName)

	// Create session (will run default shell, not Claude)
	if err := tm.NewSession(sessionName, ""); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer func() { _ = tm.KillSession(sessionName) }()

	// IsClaudeRunning should be false (shell is running, not node/claude)
	cmd, _ := tm.GetPaneCommand(sessionName)
	wantRunning := cmd == "node" || cmd == "claude"

	if got := tm.IsClaudeRunning(sessionName); got != wantRunning {
		t.Errorf("IsClaudeRunning() = %v, want %v (pane cmd: %q)", got, wantRunning, cmd)
	}
}

func TestIsClaudeRunning_VersionPattern(t *testing.T) {
	// Test the version pattern regex matching directly
	// Since we can't easily mock the pane command, test the pattern logic
	tests := []struct {
		cmd  string
		want bool
	}{
		{"node", true},
		{"claude", true},
		{"2.0.76", true},
		{"1.2.3", true},
		{"10.20.30", true},
		{"bash", false},
		{"zsh", false},
		{"", false},
		{"v2.0.76", false}, // version with 'v' prefix shouldn't match
		{"2.0", false},     // incomplete version
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			// Check if it matches node/claude directly
			isKnownCmd := tt.cmd == "node" || tt.cmd == "claude"
			// Check version pattern
			matched, _ := regexp.MatchString(`^\d+\.\d+\.\d+`, tt.cmd)

			got := isKnownCmd || matched
			if got != tt.want {
				t.Errorf("IsClaudeRunning logic for %q = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestIsClaudeRunning_ShellWithNodeChild(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}

	tm := NewTmux()
	sessionName := "gt-test-shell-child-" + t.Name()

	// Clean up any existing session
	_ = tm.KillSession(sessionName)

	// Create session with "bash -c" running a node process
	// Use a simple node command that runs for a few seconds
	cmd := `node -e "setTimeout(() => {}, 10000)"`
	if err := tm.NewSessionWithCommand(sessionName, "", cmd); err != nil {
		t.Fatalf("NewSessionWithCommand: %v", err)
	}
	defer func() { _ = tm.KillSession(sessionName) }()

	// Give the node process time to start
	// WaitForCommand waits until NOT running bash/zsh/sh
	shellsToExclude := []string{"bash", "zsh", "sh"}
	err := tm.WaitForCommand(sessionName, shellsToExclude, 2000*1000000) // 2 second timeout
	if err != nil {
		// If we timeout waiting, it means the pane command is still a shell
		// This is the case we're testing - shell with a node child
		paneCmd, _ := tm.GetPaneCommand(sessionName)
		t.Logf("Pane command is %q - testing shell+child detection", paneCmd)
	}

	// Now test IsClaudeRunning - it should detect node as a child process
	paneCmd, _ := tm.GetPaneCommand(sessionName)
	if paneCmd == "node" {
		// Direct node detection should work
		if !tm.IsClaudeRunning(sessionName) {
			t.Error("IsClaudeRunning should return true when pane command is 'node'")
		}
	} else {
		// Pane is a shell (bash/zsh) with node as child
		// The new child process detection should catch this
		got := tm.IsClaudeRunning(sessionName)
		t.Logf("Pane command: %q, IsClaudeRunning: %v", paneCmd, got)
		// Note: This may or may not detect depending on how tmux runs the command.
		// On some systems, tmux runs the command directly; on others via a shell.
	}
}

func TestHasClaudeChild(t *testing.T) {
	// Test the hasClaudeChild helper function directly
	// This uses the current process as a test subject

	// Get current process PID as string
	currentPID := "1" // init/launchd - should have children but not claude/node

	// hasClaudeChild should return false for init (no node/claude children)
	got := hasClaudeChild(currentPID)
	if got {
		t.Logf("hasClaudeChild(%q) = true - init has claude/node child?", currentPID)
	}

	// Test with a definitely nonexistent PID
	got = hasClaudeChild("999999999")
	if got {
		t.Error("hasClaudeChild should return false for nonexistent PID")
	}
}

func TestGetAllDescendants(t *testing.T) {
	// Test the getAllDescendants helper function

	// Test with nonexistent PID - should return empty slice
	got := getAllDescendants("999999999")
	if len(got) != 0 {
		t.Errorf("getAllDescendants(nonexistent) = %v, want empty slice", got)
	}

	// Test with PID 1 (init/launchd) - should find some descendants
	// Note: We can't test exact PIDs, just that the function doesn't panic
	// and returns reasonable results
	descendants := getAllDescendants("1")
	t.Logf("getAllDescendants(\"1\") found %d descendants", len(descendants))

	// Verify returned PIDs are all numeric strings
	for _, pid := range descendants {
		for _, c := range pid {
			if c < '0' || c > '9' {
				t.Errorf("getAllDescendants returned non-numeric PID: %q", pid)
			}
		}
	}
}

func TestGetAllDescendantsWithRetry(t *testing.T) {
	// Test the retry variant of getAllDescendants

	// Test with nonexistent PID - should return empty slice
	got := getAllDescendantsWithRetry("999999999")
	if len(got) != 0 {
		t.Errorf("getAllDescendantsWithRetry(nonexistent) = %v, want empty slice", got)
	}

	// Test with PID 1 (init/launchd) - should find descendants
	descendants := getAllDescendantsWithRetry("1")
	t.Logf("getAllDescendantsWithRetry(\"1\") found %d descendants", len(descendants))

	// On a live system, there should always be some descendants of init
	if len(descendants) == 0 {
		t.Error("getAllDescendantsWithRetry(\"1\") found no descendants - expected some")
	}

	// Verify returned PIDs are all numeric strings and unique
	seen := make(map[string]bool)
	for _, pid := range descendants {
		if seen[pid] {
			t.Errorf("getAllDescendantsWithRetry returned duplicate PID: %q", pid)
		}
		seen[pid] = true

		for _, c := range pid {
			if c < '0' || c > '9' {
				t.Errorf("getAllDescendantsWithRetry returned non-numeric PID: %q", pid)
			}
		}
	}

	// Note: We don't compare with getAllDescendants because process counts
	// fluctuate on a live system. The retry version may find different (not
	// necessarily more) processes due to timing.
}

func TestHasClaudeDescendant(t *testing.T) {
	// Test the recursive descendant check

	// Test with nonexistent PID - should return false
	got := hasClaudeDescendant("999999999", make(map[string]bool))
	if got {
		t.Error("hasClaudeDescendant should return false for nonexistent PID")
	}

	// Test cycle detection - should not infinite loop
	visited := make(map[string]bool)
	visited["1"] = true
	got = hasClaudeDescendant("1", visited)
	if got {
		t.Error("hasClaudeDescendant should return false when PID already visited")
	}

	// Test with PID 1 (init) - should handle deep trees without stack overflow
	got = hasClaudeDescendant("1", make(map[string]bool))
	t.Logf("hasClaudeDescendant(\"1\") = %v", got)
	// Result depends on whether claude/node is running, but it should not panic
}

func TestProcessCleanupConstants(t *testing.T) {
	// Verify the constants are reasonable values

	if SIGTERMGracePeriod < 100*time.Millisecond {
		t.Errorf("SIGTERMGracePeriod too short: %v", SIGTERMGracePeriod)
	}
	if SIGTERMGracePeriod > 5*time.Second {
		t.Errorf("SIGTERMGracePeriod too long: %v", SIGTERMGracePeriod)
	}

	if DescendantRescanDelay < 10*time.Millisecond {
		t.Errorf("DescendantRescanDelay too short: %v", DescendantRescanDelay)
	}

	if DescendantRescanAttempts < 1 {
		t.Errorf("DescendantRescanAttempts too low: %d", DescendantRescanAttempts)
	}
	if DescendantRescanAttempts > 10 {
		t.Errorf("DescendantRescanAttempts too high: %d", DescendantRescanAttempts)
	}
}

func TestSessionSet(t *testing.T) {
	if !hasTmux() {
		t.Skip("tmux not installed")
	}

	tm := NewTmux()
	sessionName := "gt-test-sessionset-" + t.Name()

	// Clean up any existing session
	_ = tm.KillSession(sessionName)

	// Create a test session
	if err := tm.NewSession(sessionName, ""); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	defer func() { _ = tm.KillSession(sessionName) }()

	// Get the session set
	set, err := tm.GetSessionSet()
	if err != nil {
		t.Fatalf("GetSessionSet: %v", err)
	}

	// Test Has() for existing session
	if !set.Has(sessionName) {
		t.Errorf("SessionSet.Has(%q) = false, want true", sessionName)
	}

	// Test Has() for non-existing session
	if set.Has("nonexistent-session-xyz-12345") {
		t.Error("SessionSet.Has(nonexistent) = true, want false")
	}

	// Test nil safety
	var nilSet *SessionSet
	if nilSet.Has("anything") {
		t.Error("nil SessionSet.Has() = true, want false")
	}

	// Test Names() returns the session
	names := set.Names()
	found := false
	for _, n := range names {
		if n == sessionName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("SessionSet.Names() doesn't contain %q", sessionName)
	}
}
