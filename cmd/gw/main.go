package main

import (
	"fmt"
	"os"

	"github.com/qawatake/gw/internal/branch"
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
	case "list":
		if err := runList(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "cd":
		if err := runCD(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "clean":
		if err := runClean(args); err != nil {
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
	fmt.Println("  gw list               List all worktrees")
	fmt.Println("  gw cd                 Change directory to a worktree")
	fmt.Println("  gw clean              Remove selected worktrees")
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
	branchName := branch.GenerateBranchName(name)
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

func runClean(args []string) error {
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
