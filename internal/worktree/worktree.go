package worktree

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Worktree represents a git worktree
type Worktree struct {
	Path   string
	Branch string
	Commit string
	Date   time.Time
	IsMain bool // true if this is the main worktree
}

// List returns all worktrees sorted by date (newest first)
func List() ([]Worktree, error) {
	// Get worktree list
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	worktrees, err := parseWorktreeList(string(output))
	if err != nil {
		return nil, err
	}

	// Get commit dates for sorting
	for i := range worktrees {
		date, err := getCommitDate(worktrees[i].Commit)
		if err != nil {
			// If we can't get the date, use epoch time
			worktrees[i].Date = time.Time{}
		} else {
			worktrees[i].Date = date
		}
	}

	// Sort by date (newest first)
	sort.Slice(worktrees, func(i, j int) bool {
		return worktrees[i].Date.After(worktrees[j].Date)
	})

	return worktrees, nil
}

func parseWorktreeList(output string) ([]Worktree, error) {
	var worktrees []Worktree
	var current Worktree
	isFirst := true

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			if current.Path != "" {
				// First worktree is the main worktree
				if isFirst {
					current.IsMain = true
					isFirst = false
				}
				worktrees = append(worktrees, current)
				current = Worktree{}
			}
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "worktree":
			current.Path = value
		case "HEAD":
			current.Commit = value
		case "branch":
			// branch refs/heads/main -> main
			current.Branch = strings.TrimPrefix(value, "refs/heads/")
		case "detached":
			current.Branch = fmt.Sprintf("(detached at %s)", value[:7])
		}
	}

	// Add last worktree if exists
	if current.Path != "" {
		// First worktree is the main worktree
		if isFirst {
			current.IsMain = true
		}
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}

func getCommitDate(commit string) (time.Time, error) {
	cmd := exec.Command("git", "show", "-s", "--format=%ct", commit)
	output, err := cmd.Output()
	if err != nil {
		return time.Time{}, err
	}

	timestamp := strings.TrimSpace(string(output))
	var unixTime int64
	_, err = fmt.Sscanf(timestamp, "%d", &unixTime)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(unixTime, 0), nil
}

// Format formats a worktree for display
func Format(wt Worktree) string {
	path := wt.Path
	// Try to make path relative to home directory
	if home, err := getHomeDir(); err == nil {
		if rel, err := filepath.Rel(home, path); err == nil && !strings.HasPrefix(rel, "..") {
			path = "~/" + rel
		}
	}

	return fmt.Sprintf("%-40s %s", wt.Branch, path)
}

func getHomeDir() (string, error) {
	cmd := exec.Command("sh", "-c", "echo $HOME")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// Add creates a new worktree
func Add(path, branch string) error {
	cmd := exec.Command("git", "worktree", "add", "-b", branch, path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add worktree: %w\n%s", err, string(output))
	}
	return nil
}

// GenerateWorktreePath generates a filesystem-safe path from branch name
func GenerateWorktreePath(branchName, rootDir string) string {
	// Replace slashes with hyphens
	dirName := strings.ReplaceAll(branchName, "/", "-")
	return filepath.Join(rootDir, dirName)
}

// Remove removes a worktree
func Remove(path string) error {
	cmd := exec.Command("git", "worktree", "remove", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %w\n%s", err, string(output))
	}
	return nil
}

// RemoveBranch removes a git branch
func RemoveBranch(branchName string) error {
	// Use -D to force delete (removes even if not merged)
	cmd := exec.Command("git", "branch", "-D", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove branch: %w\n%s", err, string(output))
	}
	return nil
}
