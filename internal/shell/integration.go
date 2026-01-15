// ABOUTME: Shell integration installation and removal for GongShow.
// ABOUTME: Manages the shell hook in RC files with safe block markers.

package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/KeithWyatt/gongshow/internal/state"
)

const (
	markerStart = "# --- GongShow Integration (managed by gt) ---"
	markerEnd   = "# --- End GongShow ---"
)

func hookSourceLine() string {
	return fmt.Sprintf(`[[ -f "%s/shell-hook.sh" ]] && source "%s/shell-hook.sh"`,
		state.ConfigDir(), state.ConfigDir())
}

func Install() error {
	shell := DetectShell()
	rcPath := RCFilePath(shell)

	if err := writeHookScript(); err != nil {
		return fmt.Errorf("writing hook script: %w", err)
	}

	if err := addToRCFile(rcPath); err != nil {
		return fmt.Errorf("updating %s: %w", rcPath, err)
	}

	return state.SetShellIntegration(shell)
}

func Remove() error {
	shell := DetectShell()
	rcPath := RCFilePath(shell)

	if err := removeFromRCFile(rcPath); err != nil {
		return fmt.Errorf("updating %s: %w", rcPath, err)
	}

	hookPath := filepath.Join(state.ConfigDir(), "shell-hook.sh")
	if err := os.Remove(hookPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing hook script: %w", err)
	}

	return nil
}

func DetectShell() string {
	shell := os.Getenv("SHELL")
	if strings.HasSuffix(shell, "zsh") {
		return "zsh"
	}
	if strings.HasSuffix(shell, "bash") {
		return "bash"
	}
	return "zsh"
}

func RCFilePath(shell string) string {
	home, _ := os.UserHomeDir()
	switch shell {
	case "bash":
		return filepath.Join(home, ".bashrc")
	default:
		return filepath.Join(home, ".zshrc")
	}
}

func writeHookScript() error {
	dir := state.ConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	hookPath := filepath.Join(dir, "shell-hook.sh")
	return os.WriteFile(hookPath, []byte(shellHookScript), 0644)
}

func addToRCFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(data)

	if strings.Contains(content, markerStart) {
		return updateRCFile(path, content)
	}

	block := fmt.Sprintf("\n%s\n%s\n%s\n", markerStart, hookSourceLine(), markerEnd)

	if len(data) > 0 {
		backupPath := path + ".gongshow-backup"
		if err := os.WriteFile(backupPath, data, 0644); err != nil {
			return fmt.Errorf("writing backup: %w", err)
		}
	}

	return os.WriteFile(path, []byte(content+block), 0644)
}

func removeFromRCFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := string(data)

	startIdx := strings.Index(content, markerStart)
	if startIdx == -1 {
		return nil
	}

	endIdx := strings.Index(content[startIdx:], markerEnd)
	if endIdx == -1 {
		return nil
	}
	endIdx += startIdx + len(markerEnd)

	if endIdx < len(content) && content[endIdx] == '\n' {
		endIdx++
	}

	if startIdx > 0 && content[startIdx-1] == '\n' {
		startIdx--
	}

	newContent := content[:startIdx] + content[endIdx:]
	return os.WriteFile(path, []byte(newContent), 0644)
}

func updateRCFile(path, content string) error {
	startIdx := strings.Index(content, markerStart)
	endIdx := strings.Index(content[startIdx:], markerEnd)
	if endIdx == -1 {
		return fmt.Errorf("malformed GongShow block in %s", path)
	}
	endIdx += startIdx + len(markerEnd)

	block := fmt.Sprintf("%s\n%s\n%s", markerStart, hookSourceLine(), markerEnd)
	newContent := content[:startIdx] + block + content[endIdx:]

	return os.WriteFile(path, []byte(newContent), 0644)
}

var shellHookScript = `#!/bin/bash
# GongShow Shell Integration
# Installed by: gt install --shell
# Location: ~/.config/gongshow/shell-hook.sh

_gongshow_enabled() {
    [[ -n "$GONGSHOW_DISABLED" ]] && return 1
    [[ -n "$GONGSHOW_ENABLED" ]] && return 0
    local state_file="$HOME/.local/state/gongshow/state.json"
    [[ -f "$state_file" ]] && grep -q '"enabled":\s*true' "$state_file" 2>/dev/null
}

_gongshow_ignored() {
    local dir="$PWD"
    while [[ "$dir" != "/" ]]; do
        [[ -f "$dir/.gongshow-ignore" ]] && return 0
        dir="$(dirname "$dir")"
    done
    return 1
}

_gongshow_already_asked() {
    local repo_root="$1"
    local asked_file="$HOME/.cache/gongshow/asked-repos"
    [[ -f "$asked_file" ]] && grep -qF "$repo_root" "$asked_file" 2>/dev/null
}

_gongshow_mark_asked() {
    local repo_root="$1"
    local asked_file="$HOME/.cache/gongshow/asked-repos"
    mkdir -p "$(dirname "$asked_file")"
    echo "$repo_root" >> "$asked_file"
}

_gongshow_offer_add() {
    local repo_root="$1"
    
    _gongshow_already_asked "$repo_root" && return 0
    
    [[ -t 0 ]] || return 0
    
    local repo_name
    repo_name=$(basename "$repo_root")
    
    echo ""
    echo -n "Add '$repo_name' to GongShow? [y/N/never] "
    read -r response </dev/tty
    
    _gongshow_mark_asked "$repo_root"
    
    case "$response" in
        y|Y|yes)
            echo "Adding to GongShow..."
            local output
            output=$(gt rig quick-add "$repo_root" --yes 2>&1)
            local exit_code=$?
            echo "$output"
            
            if [[ $exit_code -eq 0 ]]; then
                local crew_path
                crew_path=$(echo "$output" | grep "^GT_CREW_PATH=" | cut -d= -f2)
                if [[ -n "$crew_path" && -d "$crew_path" ]]; then
                    echo ""
                    echo "Switching to crew workspace..."
                    cd "$crew_path" || true
                    # Re-run hook to set GT_TOWN_ROOT and GT_RIG
                    _gongshow_hook
                fi
            fi
            ;;
        never)
            touch "$repo_root/.gongshow-ignore"
            echo "Created .gongshow-ignore - won't ask again for this repo."
            ;;
        *)
            echo "Skipped. Run 'gt rig quick-add' later to add manually."
            ;;
    esac
}

_gongshow_hook() {
    local previous_exit_status=$?

    _gongshow_enabled || {
        unset GT_TOWN_ROOT GT_RIG
        return $previous_exit_status
    }

    _gongshow_ignored && {
        unset GT_TOWN_ROOT GT_RIG
        return $previous_exit_status
    }

    if ! git rev-parse --git-dir &>/dev/null; then
        unset GT_TOWN_ROOT GT_RIG
        return $previous_exit_status
    fi

    local repo_root
    repo_root=$(git rev-parse --show-toplevel 2>/dev/null) || {
        unset GT_TOWN_ROOT GT_RIG
        return $previous_exit_status
    }

    local cache_file="$HOME/.cache/gongshow/rigs.cache"
    if [[ -f "$cache_file" ]]; then
        local cached
        cached=$(grep "^${repo_root}:" "$cache_file" 2>/dev/null)
        if [[ -n "$cached" ]]; then
            eval "${cached#*:}"
            return $previous_exit_status
        fi
    fi

    if command -v gt &>/dev/null; then
        local detect_output
        detect_output=$(gt rig detect "$repo_root" 2>/dev/null)
        eval "$detect_output"
        
        if [[ -n "$GT_TOWN_ROOT" ]]; then
            (gt rig detect --cache "$repo_root" &>/dev/null &)
        elif [[ -n "$_GONGSHOW_OFFER_ADD" ]]; then
            _gongshow_offer_add "$repo_root"
            unset _GONGSHOW_OFFER_ADD
        fi
    fi

    return $previous_exit_status
}

_gongshow_chpwd_hook() {
    _GONGSHOW_OFFER_ADD=1
    _gongshow_hook
}

case "${SHELL##*/}" in
    zsh)
        autoload -Uz add-zsh-hook
        add-zsh-hook chpwd _gongshow_chpwd_hook
        add-zsh-hook precmd _gongshow_hook
        ;;
    bash)
        if [[ ";${PROMPT_COMMAND[*]:-};" != *";_gongshow_hook;"* ]]; then
            PROMPT_COMMAND="_gongshow_chpwd_hook${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
        fi
        ;;
esac

_gongshow_hook
`
