package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// EditWithEditor opens an editor to get user input
// Similar to vcat command
func EditWithEditor(initialContent string) (string, error) {
	// Create temporary file
	tmpfile, err := os.CreateTemp("", "gw-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write initial content if provided
	if initialContent != "" {
		if _, err := tmpfile.WriteString(initialContent); err != nil {
			return "", fmt.Errorf("failed to write to temp file: %w", err)
		}
	}
	tmpfile.Close()

	// Get editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("GW_EDITOR")
	}
	if editor == "" {
		editor = "vim"
	}

	// Open editor
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run editor: %w", err)
	}

	// Read edited content
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read temp file: %w", err)
	}

	// Get first line and trim whitespace
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		return "", nil
	}

	result := strings.TrimSpace(lines[0])
	return result, nil
}

// SelectWithPeco opens peco for interactive selection
func SelectWithPeco(items []string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to select")
	}

	// Create temporary file with items
	tmpfile, err := os.CreateTemp("", "gw-peco-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	content := strings.Join(items, "\n")
	if _, err := tmpfile.WriteString(content); err != nil {
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}
	tmpfile.Close()

	// Run peco
	cmd := exec.Command("peco")
	cmd.Stdin, err = os.Open(tmpfile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to open temp file: %w", err)
	}
	defer cmd.Stdin.(*os.File).Close()

	cmd.Stderr = os.Stderr

	// Get peco's TTY for proper display
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err == nil {
		cmd.Stdin = tty
		cmd.Stdout = tty
		defer tty.Close()
	}

	// Create a pipe to capture peco output
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run peco: %w", err)
	}

	result := strings.TrimSpace(string(output))
	return result, nil
}

// MultiSelectWithFzf opens fzf for multi-select interactive selection
func MultiSelectWithFzf(items []string) ([]string, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no items to select")
	}

	// Run fzf with multi-select mode
	cmd := exec.Command("fzf", "--multi", "--prompt=Select worktrees to remove (Space to select, Enter to confirm): ")

	// Set up stdin with items
	cmd.Stdin = strings.NewReader(strings.Join(items, "\n"))

	// Connect to TTY for interactive mode
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open TTY: %w", err)
	}
	defer tty.Close()

	cmd.Stdin = tty
	cmd.Stderr = tty

	// Capture output
	output, err := cmd.Output()
	if err != nil {
		// User may have cancelled
		return nil, fmt.Errorf("selection cancelled or failed: %w", err)
	}

	// Parse selected items
	selected := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(selected) == 1 && selected[0] == "" {
		return []string{}, nil
	}

	return selected, nil
}

// GetWorktreeRoot returns the root directory for worktrees
func GetWorktreeRoot() (string, error) {
	// Check environment variable first
	if root := os.Getenv("GW_WORKTREE_ROOT"); root != "" {
		return root, nil
	}

	// Default to ~/.worktrees
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, ".worktrees"), nil
}

// Confirm asks for user confirmation
func Confirm(message string) (bool, error) {
	fmt.Printf("%s [y/N]: ", message)

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false, nil
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}
