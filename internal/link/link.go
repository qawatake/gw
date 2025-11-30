package link

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const linksDirName = ".gw-links"
const linksFileName = ".gw-links.txt"

// GetLinksDir returns the path to .gw-links directory
// Format: ~/.worktrees/{repo}/.gw-links/
func GetLinksDir(worktreeRoot string) string {
	return filepath.Join(worktreeRoot, linksDirName)
}

// getLinksFile returns the path to .gw-links.txt file
func getLinksFile(worktreeRoot string) string {
	return filepath.Join(worktreeRoot, linksFileName)
}

// readLinksFile reads registered paths from .gw-links.txt
func readLinksFile(worktreeRoot string) ([]string, error) {
	filePath := getLinksFile(worktreeRoot)
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read links file: %w", err)
	}

	var paths []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}
	return paths, nil
}

// writeLinksFile writes registered paths to .gw-links.txt
func writeLinksFile(worktreeRoot string, paths []string) error {
	filePath := getLinksFile(worktreeRoot)
	content := strings.Join(paths, "\n")
	if len(paths) > 0 {
		content += "\n"
	}
	return os.WriteFile(filePath, []byte(content), 0644)
}

// addToLinksFile adds a path to .gw-links.txt
func addToLinksFile(worktreeRoot string, path string) error {
	paths, err := readLinksFile(worktreeRoot)
	if err != nil {
		return err
	}

	// Check if already exists
	for _, p := range paths {
		if p == path {
			return nil // Already registered
		}
	}

	paths = append(paths, path)
	return writeLinksFile(worktreeRoot, paths)
}

// removeFromLinksFile removes a path from .gw-links.txt
func removeFromLinksFile(worktreeRoot string, path string) error {
	paths, err := readLinksFile(worktreeRoot)
	if err != nil {
		return err
	}

	var newPaths []string
	for _, p := range paths {
		if p != path {
			newPaths = append(newPaths, p)
		}
	}

	return writeLinksFile(worktreeRoot, newPaths)
}

// Add moves a file/directory to .gw-links and creates a symlink
func Add(targetPath string, worktreeRoot string) error {
	// Get absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if target exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file or directory does not exist: %s", targetPath)
	}

	// Get git repository root
	repoRoot, err := getGitRoot()
	if err != nil {
		return err
	}

	// Calculate relative path from repository root
	relPath, err := filepath.Rel(repoRoot, absPath)
	if err != nil {
		return fmt.Errorf("failed to calculate relative path: %w", err)
	}

	// Ensure path is within repository (doesn't start with ..)
	if strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("path must be within the git repository: %s", targetPath)
	}

	// Prepare .gw-links directory
	linksDir := GetLinksDir(worktreeRoot)
	destPath := filepath.Join(linksDir, relPath)
	destDir := filepath.Dir(destPath)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	// Check if already a symlink pointing to .gw-links
	if linkTarget, err := os.Readlink(absPath); err == nil {
		if strings.Contains(linkTarget, linksDirName) {
			return fmt.Errorf("already linked: %s", targetPath)
		}
	}

	// Check if destination already exists in .gw-links
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("already exists in .gw-links: %s", relPath)
	}

	// Move file/directory to .gw-links
	if err := os.Rename(absPath, destPath); err != nil {
		return fmt.Errorf("failed to move %s to %s: %w", absPath, destPath, err)
	}

	// Create symlink (absolute path)
	if err := os.Symlink(destPath, absPath); err != nil {
		// Try to restore on failure
		os.Rename(destPath, absPath)
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Register in .gw-links.txt
	if err := addToLinksFile(worktreeRoot, relPath); err != nil {
		return fmt.Errorf("failed to register link: %w", err)
	}

	return nil
}

// List returns all registered paths from .gw-links.txt
func List(worktreeRoot string) ([]string, error) {
	return readLinksFile(worktreeRoot)
}

// Remove moves a file/directory from .gw-links back to main worktree
func Remove(relPath string, worktreeRoot string, mainWorktreePath string) error {
	linksDir := GetLinksDir(worktreeRoot)
	srcPath := filepath.Join(linksDir, relPath)

	// Check if source exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("not found in .gw-links: %s", relPath)
	}

	// Destination in main worktree
	destPath := filepath.Join(mainWorktreePath, relPath)

	// Create parent directory if needed
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	// Remove symlink in main worktree if it exists
	if info, err := os.Lstat(destPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(destPath); err != nil {
				return fmt.Errorf("failed to remove symlink: %w", err)
			}
		} else {
			return fmt.Errorf("file exists and is not a symlink: %s", destPath)
		}
	}

	// Move from .gw-links to main worktree
	if err := os.Rename(srcPath, destPath); err != nil {
		return fmt.Errorf("failed to move %s to %s: %w", srcPath, destPath, err)
	}

	// Clean up empty parent directories in .gw-links
	cleanEmptyDirs(filepath.Dir(srcPath), linksDir)

	// Remove from .gw-links.txt
	if err := removeFromLinksFile(worktreeRoot, relPath); err != nil {
		return fmt.Errorf("failed to unregister link: %w", err)
	}

	return nil
}

// cleanEmptyDirs removes empty directories up to the root
func cleanEmptyDirs(dir string, root string) {
	for dir != root && dir != filepath.Dir(dir) {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			break
		}
		os.Remove(dir)
		dir = filepath.Dir(dir)
	}
}

// getGitRoot returns the root directory of the current git repository
func getGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git repository root: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// PullResult represents the result of pulling a single link
type PullResult struct {
	Path    string
	Success bool
	Message string
}

// Pull creates symlinks for registered paths that don't exist in current worktree
func Pull(worktreeRoot string) ([]PullResult, error) {
	// Get registered paths
	paths, err := readLinksFile(worktreeRoot)
	if err != nil {
		return nil, err
	}

	if len(paths) == 0 {
		return nil, nil
	}

	// Get current worktree root
	repoRoot, err := getGitRoot()
	if err != nil {
		return nil, err
	}

	linksDir := GetLinksDir(worktreeRoot)
	var results []PullResult

	for _, relPath := range paths {
		destPath := filepath.Join(repoRoot, relPath)
		srcPath := filepath.Join(linksDir, relPath)

		// Check if source exists in .gw-links
		srcInfo, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			results = append(results, PullResult{
				Path:    relPath,
				Success: false,
				Message: "not found in .gw-links",
			})
			continue
		}

		// Check if destination already exists
		if info, err := os.Lstat(destPath); err == nil {
			// If it's already a symlink pointing to the correct location, skip silently
			if info.Mode()&os.ModeSymlink != 0 {
				if target, err := os.Readlink(destPath); err == nil && target == srcPath {
					// Already correctly linked, skip without warning
					continue
				}
			}
			results = append(results, PullResult{
				Path:    relPath,
				Success: false,
				Message: "already exists",
			})
			continue
		}

		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			results = append(results, PullResult{
				Path:    relPath,
				Success: false,
				Message: fmt.Sprintf("failed to create directory: %v", err),
			})
			continue
		}

		// Create symlink
		if srcInfo.IsDir() {
			// For directories, link directly to the directory
			if err := os.Symlink(srcPath, destPath); err != nil {
				results = append(results, PullResult{
					Path:    relPath,
					Success: false,
					Message: fmt.Sprintf("failed to create symlink: %v", err),
				})
				continue
			}
		} else {
			// For files, link to the file
			if err := os.Symlink(srcPath, destPath); err != nil {
				results = append(results, PullResult{
					Path:    relPath,
					Success: false,
					Message: fmt.Sprintf("failed to create symlink: %v", err),
				})
				continue
			}
		}

		results = append(results, PullResult{
			Path:    relPath,
			Success: true,
		})
	}

	return results, nil
}

// CreateSymlinks creates symlinks in a worktree for all items in .gw-links
func CreateSymlinks(worktreePath string, worktreeRoot string) []string {
	linksDir := GetLinksDir(worktreeRoot)
	var warnings []string

	// Check if .gw-links exists
	if _, err := os.Stat(linksDir); os.IsNotExist(err) {
		return nil
	}

	// Walk through .gw-links and create symlinks
	filepath.WalkDir(linksDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Continue on error
		}

		if path == linksDir {
			return nil
		}

		// Get relative path from .gw-links
		relPath, err := filepath.Rel(linksDir, path)
		if err != nil {
			return nil
		}

		destPath := filepath.Join(worktreePath, relPath)

		if d.IsDir() {
			// Create directory structure
			os.MkdirAll(destPath, 0755)
		} else {
			// Check if file already exists
			if _, err := os.Lstat(destPath); err == nil {
				warnings = append(warnings, fmt.Sprintf("skipped (already exists): %s", relPath))
				return nil
			}

			// Ensure parent directory exists
			os.MkdirAll(filepath.Dir(destPath), 0755)

			// Create symlink
			if err := os.Symlink(path, destPath); err != nil {
				warnings = append(warnings, fmt.Sprintf("failed to create symlink: %s: %v", relPath, err))
			}
		}

		return nil
	})

	return warnings
}
