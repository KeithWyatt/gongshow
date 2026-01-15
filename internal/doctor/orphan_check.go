package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/KeithWyatt/gongshow/internal/events"
	"github.com/KeithWyatt/gongshow/internal/session"
	"github.com/KeithWyatt/gongshow/internal/tmux"
)

// OrphanSessionCheck detects orphaned tmux sessions that don't match
// the expected GongShow session naming patterns.
type OrphanSessionCheck struct {
	FixableCheck
	sessionLister  SessionLister
	orphanSessions []string // Cached during Run for use in Fix
}

// SessionLister abstracts tmux session listing for testing.
type SessionLister interface {
	ListSessions() ([]string, error)
}

type realSessionLister struct {
	t *tmux.Tmux
}

func (r *realSessionLister) ListSessions() ([]string, error) {
	return r.t.ListSessions()
}

// NewOrphanSessionCheck creates a new orphan session check.
func NewOrphanSessionCheck() *OrphanSessionCheck {
	return &OrphanSessionCheck{
		FixableCheck: FixableCheck{
			BaseCheck: BaseCheck{
				CheckName:        "orphan-sessions",
				CheckDescription: "Detect orphaned tmux sessions",
				CheckCategory:    CategoryCleanup,
			},
		},
	}
}

// NewOrphanSessionCheckWithSessionLister creates a check with a custom session lister (for testing).
func NewOrphanSessionCheckWithSessionLister(lister SessionLister) *OrphanSessionCheck {
	check := NewOrphanSessionCheck()
	check.sessionLister = lister
	return check
}

// Run checks for orphaned GongShow tmux sessions.
func (c *OrphanSessionCheck) Run(ctx *CheckContext) *CheckResult {
	lister := c.sessionLister
	if lister == nil {
		lister = &realSessionLister{t: tmux.NewTmux()}
	}

	sessions, err := lister.ListSessions()
	if err != nil {
		return &CheckResult{
			Name:    c.Name(),
			Status:  StatusWarning,
			Message: "Could not list tmux sessions",
			Details: []string{err.Error()},
		}
	}

	if len(sessions) == 0 {
		return &CheckResult{
			Name:    c.Name(),
			Status:  StatusOK,
			Message: "No tmux sessions found",
		}
	}

	// Get list of valid rigs
	validRigs := c.getValidRigs(ctx.TownRoot)

	// Get session names for mayor/deacon
	mayorSession := session.MayorSessionName()
	deaconSession := session.DeaconSessionName()

	// Check each session
	var orphans []string
	var validCount int

	for _, sess := range sessions {
		if sess == "" {
			continue
		}

		// Only check gt-* sessions (GongShow sessions)
		if !strings.HasPrefix(sess, "gt-") {
			continue
		}

		if c.isValidSession(sess, validRigs, mayorSession, deaconSession) {
			validCount++
		} else {
			orphans = append(orphans, sess)
		}
	}

	// Cache orphans for Fix
	c.orphanSessions = orphans

	if len(orphans) == 0 {
		return &CheckResult{
			Name:    c.Name(),
			Status:  StatusOK,
			Message: fmt.Sprintf("All %d GongShow sessions are valid", validCount),
		}
	}

	details := make([]string, len(orphans))
	for i, session := range orphans {
		details[i] = fmt.Sprintf("Orphan: %s", session)
	}

	return &CheckResult{
		Name:    c.Name(),
		Status:  StatusWarning,
		Message: fmt.Sprintf("Found %d orphaned session(s)", len(orphans)),
		Details: details,
		FixHint: "Run 'gt doctor --fix' to kill orphaned sessions",
	}
}

// Fix kills all orphaned sessions, except crew sessions which are protected.
func (c *OrphanSessionCheck) Fix(ctx *CheckContext) error {
	if len(c.orphanSessions) == 0 {
		return nil
	}

	t := tmux.NewTmux()
	var lastErr error

	for _, sess := range c.orphanSessions {
		// SAFEGUARD: Never auto-kill crew sessions.
		// Crew workers are human-managed and require explicit action.
		if isCrewSession(sess) {
			continue
		}
		// Log pre-death event for crash investigation (before killing)
		_ = events.LogFeed(events.TypeSessionDeath, sess,
			events.SessionDeathPayload(sess, "unknown", "orphan cleanup", "gt doctor"))
		if err := t.KillSession(sess); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// isCrewSession returns true if the session name matches the crew pattern.
// Crew sessions are gt-<rig>-crew-<name> and are protected from auto-cleanup.
func isCrewSession(session string) bool {
	// Pattern: gt-<rig>-crew-<name>
	// Example: gt-gongshow-crew-joe
	parts := strings.Split(session, "-")
	if len(parts) >= 4 && parts[0] == "gt" && parts[2] == "crew" {
		return true
	}
	return false
}

// getValidRigs returns a list of valid rig names from the workspace.
func (c *OrphanSessionCheck) getValidRigs(townRoot string) []string {
	var rigs []string

	// Read rigs.json if it exists
	rigsPath := filepath.Join(townRoot, "mayor", "rigs.json")
	if _, err := os.Stat(rigsPath); err == nil {
		// For simplicity, just scan directories at town root that look like rigs
		entries, err := os.ReadDir(townRoot)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() && entry.Name() != "mayor" && entry.Name() != ".beads" && !strings.HasPrefix(entry.Name(), ".") {
					// Check if it looks like a rig (has polecats/ or crew/ directory)
					polecatsDir := filepath.Join(townRoot, entry.Name(), "polecats")
					crewDir := filepath.Join(townRoot, entry.Name(), "crew")
					if _, err := os.Stat(polecatsDir); err == nil {
						rigs = append(rigs, entry.Name())
					} else if _, err := os.Stat(crewDir); err == nil {
						rigs = append(rigs, entry.Name())
					}
				}
			}
		}
	}

	return rigs
}

// isValidSession checks if a session name matches expected GongShow patterns.
// Valid patterns:
//   - gt-{town}-mayor (dynamic based on town name)
//   - gt-{town}-deacon (dynamic based on town name)
//   - gt-<rig>-witness
//   - gt-<rig>-refinery
//   - gt-<rig>-<polecat> (where polecat is any name)
//
// Note: We can't verify polecat names without reading state, so we're permissive.
func (c *OrphanSessionCheck) isValidSession(sess string, validRigs []string, mayorSession, deaconSession string) bool {
	// Mayor session is always valid (dynamic name based on town)
	if mayorSession != "" && sess == mayorSession {
		return true
	}

	// Deacon session is always valid (dynamic name based on town)
	if deaconSession != "" && sess == deaconSession {
		return true
	}

	// For rig-specific sessions, extract rig name
	// Pattern: gt-<rig>-<role>
	parts := strings.SplitN(sess, "-", 3)
	if len(parts) < 3 {
		// Invalid format - must be gt-<rig>-<something>
		return false
	}

	rigName := parts[1]

	// Check if this rig exists
	rigFound := false
	for _, r := range validRigs {
		if r == rigName {
			rigFound = true
			break
		}
	}

	if !rigFound {
		// Unknown rig - this is an orphan
		return false
	}

	role := parts[2]

	// witness and refinery are valid roles
	if role == "witness" || role == "refinery" {
		return true
	}

	// Any other name is assumed to be a polecat or crew member
	// We can't easily verify without reading state, so accept it
	return true
}

// OrphanProcessCheck detects runtime processes that are not
// running inside a tmux session. These may be user's personal sessions
// or legitimately orphaned processes from crashed GongShow sessions.
// When --fix is used, orphaned processes are killed after verifying they
// have no tmux pane ancestor (using ancestry tracing up to 8 levels).
type OrphanProcessCheck struct {
	FixableCheck
	processLister   ProcessLister
	orphanProcesses []processInfo // Cached during Run for use in Fix
}

// ProcessLister abstracts process listing for testing.
type ProcessLister interface {
	// ListTmuxServerPIDs returns PIDs of tmux server processes.
	ListTmuxServerPIDs() ([]int, error)
	// ListPanePIDs returns PIDs of shells inside tmux panes.
	ListPanePIDs() ([]int, error)
	// ListRuntimeProcesses returns info about running runtime CLI processes.
	ListRuntimeProcesses() ([]processInfo, error)
	// GetParentPID returns the parent PID of a given process.
	GetParentPID(pid int) (int, error)
}

// realProcessLister implements ProcessLister using actual system commands.
type realProcessLister struct{}

func (r *realProcessLister) ListTmuxServerPIDs() ([]int, error) {
	var pids []int
	// Find tmux server processes using ps.
	// Match "tmux", "tmux: server", or paths ending in /tmux.
	// On Linux, long-running tmux servers show as "tmux: server" in comm field.
	out, err := exec.Command("sh", "-c", `ps ax -o pid,comm | awk '$2 == "tmux" || $2 ~ /^tmux:/ || $2 ~ /\/tmux$/ { print $1 }'`).Output()
	if err != nil {
		return pids, nil // No tmux server running
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		var pid int
		if _, err := fmt.Sscanf(line, "%d", &pid); err == nil {
			pids = append(pids, pid)
		}
	}
	return pids, nil
}

func (r *realProcessLister) ListPanePIDs() ([]int, error) {
	var pids []int
	// Use -a flag to get ALL pane PIDs across ALL sessions in one command.
	// This is critical for safety - iterating sessions individually can miss panes
	// if any session query fails, leading to false orphan detection.
	out, err := exec.Command("tmux", "list-panes", "-a", "-F", "#{pane_pid}").Output()
	if err != nil {
		// tmux not running or no sessions - return empty list
		return pids, nil
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		var pid int
		if _, err := fmt.Sscanf(line, "%d", &pid); err == nil {
			pids = append(pids, pid)
		}
	}
	return pids, nil
}

func (r *realProcessLister) ListRuntimeProcesses() ([]processInfo, error) {
	var procs []processInfo
	out, err := exec.Command("ps", "-eo", "pid,ppid,comm").Output()
	if err != nil {
		return nil, err
	}

	// Regex to match runtime CLI processes (not Claude.app)
	runtimePattern := regexp.MustCompile(`(?i)(^claude$|/claude$|^claude-code$|/claude-code$|^codex$|/codex$)`)
	excludePattern := regexp.MustCompile(`(?i)(Claude\.app|claude-native|chrome-native)`)

	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		cmd := strings.Join(fields[2:], " ")
		if excludePattern.MatchString(cmd) {
			continue
		}
		if !runtimePattern.MatchString(cmd) {
			continue
		}
		var pid, ppid int
		if _, err := fmt.Sscanf(fields[0], "%d", &pid); err != nil {
			continue
		}
		if _, err := fmt.Sscanf(fields[1], "%d", &ppid); err != nil {
			continue
		}
		procs = append(procs, processInfo{pid: pid, ppid: ppid, cmd: cmd})
	}
	return procs, nil
}

func (r *realProcessLister) GetParentPID(pid int) (int, error) {
	out, err := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "ppid=").Output() //nolint:gosec // G204: PID is numeric
	if err != nil {
		return 0, err
	}
	var ppid int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &ppid); err != nil {
		return 0, err
	}
	return ppid, nil
}

// NewOrphanProcessCheck creates a new orphan process check.
func NewOrphanProcessCheck() *OrphanProcessCheck {
	return &OrphanProcessCheck{
		FixableCheck: FixableCheck{
			BaseCheck: BaseCheck{
				CheckName:        "orphan-processes",
				CheckDescription: "Detect runtime processes outside tmux",
				CheckCategory:    CategoryCleanup,
			},
		},
		processLister: &realProcessLister{},
	}
}

// NewOrphanProcessCheckWithProcessLister creates a check with a custom process lister (for testing).
func NewOrphanProcessCheckWithProcessLister(lister ProcessLister) *OrphanProcessCheck {
	check := NewOrphanProcessCheck()
	check.processLister = lister
	return check
}

// Run checks for runtime processes running outside tmux.
func (c *OrphanProcessCheck) Run(ctx *CheckContext) *CheckResult {
	// Get list of tmux session PIDs
	tmuxPIDs, err := c.getTmuxSessionPIDs()
	if err != nil {
		return &CheckResult{
			Name:    c.Name(),
			Status:  StatusWarning,
			Message: "Could not get tmux session info",
			Details: []string{err.Error()},
		}
	}

	// Find runtime processes
	runtimeProcs, err := c.processLister.ListRuntimeProcesses()
	if err != nil {
		return &CheckResult{
			Name:    c.Name(),
			Status:  StatusWarning,
			Message: "Could not list runtime processes",
			Details: []string{err.Error()},
		}
	}

	if len(runtimeProcs) == 0 {
		return &CheckResult{
			Name:    c.Name(),
			Status:  StatusOK,
			Message: "No runtime processes found",
		}
	}

	// Check which runtime processes are outside tmux
	var outsideTmux []processInfo
	var insideTmux int

	for _, proc := range runtimeProcs {
		if c.isOrphanProcess(proc, tmuxPIDs) {
			outsideTmux = append(outsideTmux, proc)
		} else {
			insideTmux++
		}
	}

	// Cache orphans for Fix
	c.orphanProcesses = outsideTmux

	if len(outsideTmux) == 0 {
		return &CheckResult{
			Name:    c.Name(),
			Status:  StatusOK,
			Message: fmt.Sprintf("All %d runtime processes are inside tmux", insideTmux),
		}
	}

	details := make([]string, 0, len(outsideTmux)+2)
	details = append(details, "These processes have no tmux pane ancestor (checked 8 levels).")
	details = append(details, "Orphaned processes detected:")
	for _, proc := range outsideTmux {
		details = append(details, fmt.Sprintf("  PID %d: %s (parent: %d)", proc.pid, proc.cmd, proc.ppid))
	}

	return &CheckResult{
		Name:    c.Name(),
		Status:  StatusWarning,
		Message: fmt.Sprintf("Found %d orphaned runtime process(es)", len(outsideTmux)),
		Details: details,
		FixHint: "Run 'gt doctor --fix' to kill orphaned processes",
	}
}

type processInfo struct {
	pid  int
	ppid int
	cmd  string
}

// getTmuxSessionPIDs returns PIDs of all tmux server processes and pane shell PIDs.
func (c *OrphanProcessCheck) getTmuxSessionPIDs() (map[int]bool, error) { //nolint:unparam // error return kept for future use
	pids := make(map[int]bool)

	// Get tmux server PIDs
	serverPIDs, err := c.processLister.ListTmuxServerPIDs()
	if err != nil {
		return pids, nil
	}
	for _, pid := range serverPIDs {
		pids[pid] = true
	}

	// Get shell PIDs inside tmux panes
	panePIDs, err := c.processLister.ListPanePIDs()
	if err != nil {
		return pids, nil
	}
	for _, pid := range panePIDs {
		pids[pid] = true
	}

	return pids, nil
}

// isOrphanProcess checks if a runtime process is orphaned.
// A process is orphaned if its parent (or ancestor) is not a tmux session.
// maxAncestryDepth is the maximum number of parent levels to check when
// determining if a process is orphaned. 8 levels is sufficient for typical
// process trees (tmux -> shell -> shell -> ... -> claude).
const maxAncestryDepth = 8

func (c *OrphanProcessCheck) isOrphanProcess(proc processInfo, tmuxPIDs map[int]bool) bool {
	// Walk up the process tree looking for a tmux pane ancestor.
	// We check up to maxAncestryDepth levels to avoid infinite loops
	// while still catching deep process trees.

	// Start by getting the CURRENT parent PID (not the cached one from proc.ppid)
	// This ensures we catch processes that were reparented between Run() and Fix().
	currentPPID, err := c.processLister.GetParentPID(proc.pid)
	if err != nil {
		// Process may have exited, use cached ppid as fallback
		currentPPID = proc.ppid
	}

	visited := make(map[int]bool)

	for depth := 0; depth < maxAncestryDepth && currentPPID > 1 && !visited[currentPPID]; depth++ {
		visited[currentPPID] = true

		// Check if this is a tmux pane PID
		if tmuxPIDs[currentPPID] {
			return false // Has tmux pane ancestor, not orphaned
		}

		// Get parent's parent
		nextPPID, err := c.processLister.GetParentPID(currentPPID)
		if err != nil {
			break
		}
		currentPPID = nextPPID
	}

	return true // No tmux pane ancestor found within maxAncestryDepth levels
}

// Fix kills all orphaned processes that were detected during Run().
// Safety: Each process is re-verified to have no tmux pane ancestor before killing.
// If ctx.DryRun is true, reports what would be killed without actually killing.
func (c *OrphanProcessCheck) Fix(ctx *CheckContext) error {
	if len(c.orphanProcesses) == 0 {
		return nil
	}

	// Re-fetch current pane PIDs for safety verification.
	// This ensures we don't kill a process that became parented by tmux
	// between Run() and Fix().
	currentPanePIDs, err := c.processLister.ListPanePIDs()
	if err != nil {
		return fmt.Errorf("failed to list current pane PIDs: %w", err)
	}
	panePIDSet := make(map[int]bool)
	for _, pid := range currentPanePIDs {
		panePIDSet[pid] = true
	}

	// Also include tmux server PIDs for completeness
	serverPIDs, _ := c.processLister.ListTmuxServerPIDs()
	for _, pid := range serverPIDs {
		panePIDSet[pid] = true
	}

	var killed int
	var skipped int
	var lastErr error

	for _, proc := range c.orphanProcesses {
		// Re-verify this process is still orphaned before killing.
		// This is a critical safety check.
		if !c.isOrphanProcess(proc, panePIDSet) {
			// Process now has a tmux ancestor - skip it
			skipped++
			continue
		}

		// Verify process still exists before killing
		if err := syscallKill(proc.pid, 0); err != nil {
			// Process no longer exists
			continue
		}

		// Dry-run mode: just count what would be killed
		if ctx.DryRun {
			fmt.Printf("[dry-run] Would kill PID %d: %s\n", proc.pid, proc.cmd)
			killed++
			continue
		}

		// Kill the orphaned process
		if err := syscallKill(proc.pid, syscall.SIGTERM); err != nil {
			lastErr = fmt.Errorf("failed to kill PID %d: %w", proc.pid, err)
			continue
		}
		killed++
	}

	if ctx.DryRun {
		if skipped > 0 {
			fmt.Printf("[dry-run] %d process(es) now have tmux ancestors, would skip\n", skipped)
		}
		return nil
	}

	if lastErr != nil && killed == 0 {
		return lastErr
	}

	return nil
}

// syscallKill wraps syscall.Kill for testability.
var syscallKill = func(pid int, sig syscall.Signal) error {
	return syscall.Kill(pid, sig)
}
