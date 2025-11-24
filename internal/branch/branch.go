package branch

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// GenerateBranchName generates a branch name with qwtk prefix and current date
func GenerateBranchName(name string) string {
	now := time.Now()
	prefix := os.Getenv("GW_BRANCH_PREFIX")
	if prefix == "" {
		prefix = "qwtk/{date}/"
	}

	// Replace {date} placeholder
	prefix = strings.ReplaceAll(prefix, "{date}", now.Format("2006/01/02"))

	// Sanitize name (replace spaces with hyphens)
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "-")

	return prefix + name
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
