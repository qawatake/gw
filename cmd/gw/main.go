package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/qawatake/gw/internal/branch"
	"github.com/qawatake/gw/internal/link"
	"github.com/qawatake/gw/internal/shell"
	"github.com/qawatake/gw/internal/ui"
	"github.com/qawatake/gw/internal/worktree"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "init":
		if err := runInit(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "add":
		if err := runAdd(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "list", "ls":
		if err := runList(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "cd":
		if err := runCD(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "rm":
		if err := runRM(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "pr":
		if err := runPR(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "ln":
		if err := runLn(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("gw - git worktree wrapper")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gw init               Initialize shell wrapper")
	fmt.Println("  gw add                Create a new branch and worktree")
	fmt.Println("  gw list (ls)          List all worktrees")
	fmt.Println("  gw cd                 Change directory to a worktree")
	fmt.Println("  gw rm                 Remove selected worktrees")
	fmt.Println("  gw pr checkout        Checkout a PR branch into a new worktree")
	fmt.Println("  gw ln add <path>      Share a file/directory across worktrees")
	fmt.Println("  gw ln ls              List shared files/directories")
	fmt.Println("  gw ln rm              Remove a file/directory from sharing")
}

func runInit(args []string) error {
	script, err := shell.GetInitScript()
	if err != nil {
		return err
	}
	fmt.Print(script)
	return nil
}

func runAdd(args []string) error {
	// Get branch name from user via editor
	name, err := ui.EditWithEditor("")
	if err != nil {
		return fmt.Errorf("failed to get branch name: %w", err)
	}

	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Generate full branch name with prefix
	branchName, err := branch.GenerateBranchName(name)
	if err != nil {
		return err
	}
	fmt.Printf("Creating branch: %s\n", branchName)

	// Get worktree root directory
	rootDir, err := ui.GetWorktreeRoot()
	if err != nil {
		return err
	}

	// Generate worktree path
	wtPath := worktree.GenerateWorktreePath(branchName, rootDir)
	fmt.Printf("Creating worktree at: %s\n", wtPath)

	// Create worktree
	if err := worktree.Add(wtPath, branchName); err != nil {
		return err
	}

	// Create symlinks for shared files
	warnings := link.CreateSymlinks(wtPath, rootDir)
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", w)
	}

	fmt.Printf("✓ Successfully created worktree\n")
	fmt.Printf("  Branch: %s\n", branchName)
	fmt.Printf("  Path: %s\n", wtPath)

	return nil
}

func runList(args []string) error {
	worktrees, err := worktree.List()
	if err != nil {
		return err
	}

	for _, wt := range worktrees {
		fmt.Println(worktree.Format(wt))
	}

	return nil
}

func runCD(args []string) error {
	// Get worktree list
	worktrees, err := worktree.List()
	if err != nil {
		return err
	}

	if len(worktrees) == 0 {
		return fmt.Errorf("no worktrees found")
	}

	// Format worktrees for selection
	items := make([]string, len(worktrees))
	for i, wt := range worktrees {
		items[i] = worktree.Format(wt)
	}

	// Let user select with peco
	selected, err := ui.SelectWithPeco(items)
	if err != nil {
		return fmt.Errorf("failed to select worktree: %w", err)
	}

	// Find the selected worktree
	var selectedWorktree *worktree.Worktree
	for i, item := range items {
		if item == selected {
			selectedWorktree = &worktrees[i]
			break
		}
	}

	if selectedWorktree == nil {
		return fmt.Errorf("selected worktree not found")
	}

	// Output cd command for shell wrapper to evaluate
	fmt.Printf("cd %q", selectedWorktree.Path)

	return nil
}

func runRM(args []string) error {
	// Get worktree list
	allWorktrees, err := worktree.List()
	if err != nil {
		return err
	}

	// Filter out main worktree
	var worktrees []worktree.Worktree
	for _, wt := range allWorktrees {
		if !wt.IsMain {
			worktrees = append(worktrees, wt)
		}
	}

	if len(worktrees) == 0 {
		fmt.Println("No additional worktrees found (main worktree cannot be removed)")
		return nil
	}

	// Format worktrees for selection
	items := make([]string, len(worktrees))
	for i, wt := range worktrees {
		items[i] = worktree.Format(wt)
	}

	// Let user select multiple worktrees (fzf or peco)
	selected, err := ui.MultiSelect(items)
	if err != nil {
		return fmt.Errorf("failed to select worktrees: %w", err)
	}

	if len(selected) == 0 {
		fmt.Println("No worktrees selected")
		return nil
	}

	// Find the selected worktrees
	var selectedWorktrees []worktree.Worktree
	for _, sel := range selected {
		for i, item := range items {
			if item == sel {
				selectedWorktrees = append(selectedWorktrees, worktrees[i])
				break
			}
		}
	}

	// Get default branch name
	defaultBranch, err := branch.GetDefaultBranch()
	if err != nil {
		return err
	}

	// Check if default branch worktree is selected
	for _, wt := range selectedWorktrees {
		if wt.Branch == defaultBranch {
			return fmt.Errorf("cannot remove worktree for default branch %q", defaultBranch)
		}
	}

	// Show what will be deleted
	fmt.Printf("\nThe following worktrees will be removed:\n")
	for _, wt := range selectedWorktrees {
		fmt.Printf("  - %s (%s)\n", wt.Branch, wt.Path)
	}
	fmt.Println()

	// Confirm deletion
	confirmed, err := ui.Confirm("Are you sure you want to remove these worktrees?")
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("Cancelled")
		return nil
	}

	// Remove worktrees and their branches
	for _, wt := range selectedWorktrees {
		fmt.Printf("Removing worktree %s...\n", wt.Branch)
		if err := worktree.Remove(wt.Path); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove worktree %s: %v\n", wt.Branch, err)
			continue
		}
		fmt.Printf("✓ Removed worktree %s\n", wt.Branch)

		// Remove associated branch
		fmt.Printf("Removing branch %s...\n", wt.Branch)
		if err := worktree.RemoveBranch(wt.Branch); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove branch %s: %v\n", wt.Branch, err)
			continue
		}
		fmt.Printf("✓ Removed branch %s\n", wt.Branch)
	}

	return nil
}

func runPR(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("pr subcommand required (e.g., 'gw pr checkout')")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "checkout":
		return runPRCheckout(subArgs)
	default:
		return fmt.Errorf("unknown pr subcommand: %s", subcommand)
	}
}

func runLn(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("ln subcommand required (add, ls, rm)")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "add":
		return runLnAdd(subArgs)
	case "ls":
		return runLnLs(subArgs)
	case "rm":
		return runLnRm(subArgs)
	default:
		return fmt.Errorf("unknown ln subcommand: %s", subcommand)
	}
}

func runLnAdd(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("path required: gw ln add <path>")
	}

	targetPath := args[0]

	// Get worktree root directory
	rootDir, err := ui.GetWorktreeRoot()
	if err != nil {
		return err
	}

	// Add the file/directory to .gw-links
	if err := link.Add(targetPath, rootDir); err != nil {
		return err
	}

	fmt.Printf("✓ Added to shared links: %s\n", targetPath)
	return nil
}

func runLnLs(args []string) error {
	// Get worktree root directory
	rootDir, err := ui.GetWorktreeRoot()
	if err != nil {
		return err
	}

	// List all items in .gw-links
	items, err := link.List(rootDir)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Println("No shared files/directories")
		return nil
	}

	for _, item := range items {
		fmt.Println(item)
	}

	return nil
}

func runLnRm(args []string) error {
	// Get worktree root directory
	rootDir, err := ui.GetWorktreeRoot()
	if err != nil {
		return err
	}

	// List registered items for selection
	items, err := link.List(rootDir)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Println("No shared files/directories to remove")
		return nil
	}

	// Let user select with peco/fzf
	selected, err := ui.SelectWithPeco(items)
	if err != nil {
		return fmt.Errorf("failed to select: %w", err)
	}

	// Get main worktree path
	worktrees, err := worktree.List()
	if err != nil {
		return err
	}

	if len(worktrees) == 0 {
		return fmt.Errorf("no worktrees found")
	}

	mainWorktreePath := worktrees[0].Path

	// Remove from .gw-links and move to main worktree
	if err := link.Remove(selected, rootDir, mainWorktreePath); err != nil {
		return err
	}

	fmt.Printf("✓ Removed from shared links: %s\n", selected)
	fmt.Printf("  Moved to: %s\n", mainWorktreePath)
	return nil
}

func runPRCheckout(args []string) error {
	// Run gh pr checkout and capture the branch name
	ghArgs := append([]string{"pr", "checkout"}, args...)
	cmd := exec.Command("gh", ghArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("gh pr checkout failed: %w", err)
	}

	// Get current branch name (gh pr checkout switches to the PR branch)
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	branchName := strings.TrimSpace(string(branchOutput))

	if branchName == "" {
		return fmt.Errorf("failed to determine checked out branch")
	}

	fmt.Printf("Checked out branch: %s\n", branchName)

	// Get worktree root directory
	rootDir, err := ui.GetWorktreeRoot()
	if err != nil {
		return err
	}

	// Generate worktree path
	wtPath := worktree.GenerateWorktreePath(branchName, rootDir)
	fmt.Printf("Creating worktree at: %s\n", wtPath)

	// Switch back to previous branch before creating worktree
	// (we need to detach the branch from current worktree)
	prevBranchCmd := exec.Command("git", "checkout", "-")
	prevBranchCmd.Stdout = os.Stdout
	prevBranchCmd.Stderr = os.Stderr
	if err := prevBranchCmd.Run(); err != nil {
		return fmt.Errorf("failed to switch back to previous branch: %w", err)
	}

	// Create worktree for existing branch
	if err := worktree.AddExistingBranch(wtPath, branchName); err != nil {
		return err
	}

	// Create symlinks for shared files
	warnings := link.CreateSymlinks(wtPath, rootDir)
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", w)
	}

	fmt.Printf("✓ Successfully created worktree\n")
	fmt.Printf("  Branch: %s\n", branchName)
	fmt.Printf("  Path: %s\n", wtPath)

	// Print any output from gh pr checkout
	if len(output) > 0 {
		fmt.Print(string(output))
	}

	return nil
}
