package branch

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// GetGitUserName returns the git user.name from config
func GetGitUserName() (string, error) {
	cmd := exec.Command("git", "config", "user.name")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git user.name is not configured. Please set it with: git config user.name \"Your Name\"")
	}

	userName := strings.TrimSpace(string(output))
	if userName == "" {
		return "", fmt.Errorf("git user.name is empty. Please set it with: git config user.name \"Your Name\"")
	}

	return userName, nil
}

// GenerateBranchName generates a branch name with user prefix and current date
func GenerateBranchName(name string) (string, error) {
	now := time.Now()

	// Get prefix from environment variable or use git user.name
	prefix := os.Getenv("GW_BRANCH_PREFIX")
	if prefix == "" {
		userName, err := GetGitUserName()
		if err != nil {
			return "", err
		}
		// Convert user name to lowercase and replace spaces with hyphens for branch name
		userPrefix := strings.ToLower(userName)
		userPrefix = strings.ReplaceAll(userPrefix, " ", "-")
		prefix = userPrefix + "/{date}/"
	}

	// Replace {date} placeholder
	prefix = strings.ReplaceAll(prefix, "{date}", now.Format("2006/01/02"))

	// Sanitize name (replace spaces with hyphens)
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "-")

	return prefix + name, nil
}

// Create creates a new branch
func Create(branchName string) error {
	cmd := exec.Command("git", "switch", "-c", branchName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
