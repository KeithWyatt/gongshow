// Package proc provides native Go process management via /proc filesystem.
// This eliminates shell spawning overhead for process tree operations.
package proc

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// GetChildren returns direct child PIDs of a process using /proc/<pid>/task/<tid>/children.
// Returns nil on error or if process has no children.
// This is O(1) filesystem reads vs O(1) shell spawn - much faster.
func GetChildren(pid int) []int {
	// Read from /proc/<pid>/task/<pid>/children (Linux 3.5+)
	// This file contains space-separated child PIDs
	path := filepath.Join("/proc", strconv.Itoa(pid), "task", strconv.Itoa(pid), "children")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return nil
	}

	children := make([]int, 0, len(fields))
	for _, f := range fields {
		if cpid, err := strconv.Atoi(f); err == nil {
			children = append(children, cpid)
		}
	}
	return children
}

// GetAllDescendants returns all descendant PIDs in depth-first order (deepest first).
// This is the native Go equivalent of recursive pgrep -P calls.
// Returns PIDs in kill-safe order: children before parents.
func GetAllDescendants(pid int) []int {
	var result []int
	children := GetChildren(pid)
	for _, child := range children {
		// Recursively get grandchildren first (deepest-first order)
		result = append(result, GetAllDescendants(child)...)
		result = append(result, child)
	}
	return result
}

// ProcessInfo contains basic process information read from /proc.
type ProcessInfo struct {
	PID  int
	Comm string // Process command name (from /proc/<pid>/comm)
}

// GetChildrenWithComm returns direct children with their command names.
// More efficient than separate pgrep calls for process identification.
func GetChildrenWithComm(pid int) []ProcessInfo {
	children := GetChildren(pid)
	if len(children) == 0 {
		return nil
	}

	result := make([]ProcessInfo, 0, len(children))
	for _, cpid := range children {
		comm := GetComm(cpid)
		if comm != "" {
			result = append(result, ProcessInfo{PID: cpid, Comm: comm})
		}
	}
	return result
}

// GetComm returns the command name for a process from /proc/<pid>/comm.
// Returns empty string if process doesn't exist or can't be read.
func GetComm(pid int) string {
	path := filepath.Join("/proc", strconv.Itoa(pid), "comm")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// Signal sends a signal to a process using native syscall.
// Returns nil if signal was sent (process may still ignore it).
// Returns error if process doesn't exist or permission denied.
func Signal(pid int, sig syscall.Signal) error {
	return syscall.Kill(pid, sig)
}

// SignalAll sends a signal to multiple processes.
// Continues on error, returns count of successful signals.
// This replaces multiple `kill` shell invocations with direct syscalls.
func SignalAll(pids []int, sig syscall.Signal) int {
	sent := 0
	for _, pid := range pids {
		if err := syscall.Kill(pid, sig); err == nil {
			sent++
		}
	}
	return sent
}

// Exists checks if a process exists by attempting to signal it with signal 0.
func Exists(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

// HasDescendantMatching checks if any descendant's comm matches one of the names.
// Returns true on first match. This replaces recursive pgrep -P -l calls.
func HasDescendantMatching(pid int, names []string, visited map[int]bool) bool {
	if visited[pid] {
		return false
	}
	visited[pid] = true

	children := GetChildrenWithComm(pid)
	for _, child := range children {
		// Check if this child matches
		for _, name := range names {
			if child.Comm == name {
				return true
			}
		}
		// Recursively check grandchildren
		if HasDescendantMatching(child.PID, names, visited) {
			return true
		}
	}
	return false
}

// CountByPattern counts processes matching a command pattern.
// Scans /proc for processes whose comm or cmdline contains the pattern.
// This replaces `pgrep -f pattern | wc -l` shell pipeline.
func CountByPattern(pattern string) int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue // Not a PID directory
		}

		// Check cmdline for pattern (more accurate than comm for multi-word patterns)
		cmdline := getCmdline(pid)
		if strings.Contains(cmdline, pattern) {
			count++
		}
	}
	return count
}

// getCmdline reads /proc/<pid>/cmdline and returns it as a space-joined string.
func getCmdline(pid int) string {
	path := filepath.Join("/proc", strconv.Itoa(pid), "cmdline")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	// cmdline uses null bytes as separators
	return strings.ReplaceAll(string(data), "\x00", " ")
}

// FindByPattern returns PIDs of processes matching a command pattern.
// Scans /proc for processes whose cmdline contains the pattern.
// This replaces `pgrep -f pattern` shell command.
func FindByPattern(pattern string) []int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	var pids []int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue // Not a PID directory
		}

		cmdline := getCmdline(pid)
		if strings.Contains(cmdline, pattern) {
			pids = append(pids, pid)
		}
	}
	return pids
}
