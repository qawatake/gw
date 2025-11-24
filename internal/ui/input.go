package ui

import (
	"fmt"
	"io"
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

	// Connect to TTY for interactive mode
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("failed to open TTY: %w", err)
	}
	defer tty.Close()

	// Set up pipes before starting
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// peco displays UI on stderr and reads keyboard input from TTY
	cmd.Stderr = tty

	// Start peco
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start peco: %w", err)
	}

	// Write items to stdin
	go func() {
		defer stdin.Close()
		for _, item := range items {
			fmt.Fprintln(stdin, item)
		}
	}()

	// Read stdout
	output, err := io.ReadAll(stdout)
	if err != nil {
		return "", fmt.Errorf("failed to read output: %w", err)
	}

	// Wait for peco to finish
	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("peco failed: %w", err)
	}

	result := strings.TrimSpace(string(output))
	return result, nil
}

// MultiSelect opens an interactive multi-select UI
// Prefers fzf, falls back to peco if fzf is not available
func MultiSelect(items []string) ([]string, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no items to select")
	}

	// Check if fzf is available
	if _, err := exec.LookPath("fzf"); err == nil {
		return multiSelectWithFzf(items)
	}

	// Fallback to peco (single select repeated)
	fmt.Println("Note: fzf not found, using peco (single selection mode)")
	return multiSelectWithPeco(items)
}

// multiSelectWithFzf uses fzf for multi-select
func multiSelectWithFzf(items []string) ([]string, error) {
	cmd := exec.Command("fzf", "--multi", "--prompt=Select worktrees to remove (Space to select, Enter to confirm): ")

	// Connect to TTY for interactive mode
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open TTY: %w", err)
	}
	defer tty.Close()

	cmd.Stdin = tty
	cmd.Stderr = tty

	// Write items to stdin via pipe
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start fzf: %w", err)
	}

	// Write items
	go func() {
		defer stdin.Close()
		for _, item := range items {
			fmt.Fprintln(stdin, item)
		}
	}()

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("selection cancelled or failed: %w", err)
	}

	selected := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(selected) == 1 && selected[0] == "" {
		return []string{}, nil
	}

	return selected, nil
}

// multiSelectWithPeco uses peco for single selection (repeated until done)
func multiSelectWithPeco(items []string) ([]string, error) {
	var selected []string
	remaining := make([]string, len(items))
	copy(remaining, items)

	for {
		if len(remaining) == 0 {
			break
		}

		fmt.Printf("\nSelected %d item(s). Select another or cancel to finish.\n", len(selected))

		choice, err := SelectWithPeco(remaining)
		if err != nil {
			// User cancelled, finish selection
			break
		}

		// Add to selected
		selected = append(selected, choice)

		// Remove from remaining
		newRemaining := []string{}
		for _, item := range remaining {
			if item != choice {
				newRemaining = append(newRemaining, item)
			}
		}
		remaining = newRemaining
	}

	return selected, nil
}

// GetWorktreeRoot returns the root directory for worktrees
// Format: ~/.worktrees/{repo-name}/
func GetWorktreeRoot() (string, error) {
	baseDir := os.Getenv("GW_WORKTREE_ROOT")
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(home, ".worktrees")
	}

	// Get repository directory name
	repoName, err := getRepositoryName()
	if err != nil {
		return "", err
	}

	return filepath.Join(baseDir, repoName), nil
}

// getRepositoryName returns the directory name of the current git repository
func getRepositoryName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get repository root: %w", err)
	}

	repoPath := strings.TrimSpace(string(output))
	repoName := filepath.Base(repoPath)
	return repoName, nil
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
