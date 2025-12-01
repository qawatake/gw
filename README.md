# gw - Git Worktree Wrapper

A convenient command-line tool to manage git worktrees efficiently.

## Features

- **add**: Create a new branch and corresponding worktree with interactive naming
- **list** (alias: **ls**): Display all worktrees sorted by recent activity
- **cd**: Interactively select and navigate to a worktree
- **rm**: Interactively select and remove multiple worktrees with their branches
- **pr checkout**: Checkout a PR branch and create a worktree for it
- **ln**: Share gitignored files (like `.env`) across worktrees using symlinks

## Requirements

- Go 1.24 or later
- Git
- GitHub CLI (`gh`) (for `gw pr checkout` command)
- peco (for `gw cd` command)
- fzf (for `gw rm` command, optional - falls back to peco)
- vim or $EDITOR (for `gw add` command)

### Installing dependencies

**macOS:**
```bash
brew install gh peco fzf
```

**Linux:**
```bash
# GitHub CLI
# See https://github.com/cli/cli/blob/trunk/docs/install_linux.md

# peco
go install github.com/peco/peco/cmd/peco@latest

# fzf
git clone --depth 1 https://github.com/junegunn/fzf.git ~/.fzf
~/.fzf/install
```

## Installation

```bash
go install github.com/qawatake/gw/cmd/gw@latest
```

### Shell integration

Add the following to your shell configuration file:

**Bash (~/.bashrc):**
```bash
eval "$(gw init)"
```

**Zsh (~/.zshrc):**
```bash
eval "$(gw init)"
```

**Fish (~/.config/fish/config.fish):**
```fish
eval (gw init | string collect)
```

## Usage

### `gw add`

Create a new branch and worktree with interactive naming.

```bash
$ gw add
# Opens your editor (vim by default)
# Enter branch name, e.g., "feature-login"
# Creates branch: sample-user/2025/11/24/feature-login
# Creates worktree at: ~/.worktrees/gw/2025-11-24-feature-login/gw
```

The branch name will automatically be prefixed with `{user-name}/YYYY/MM/DD/` where `{user-name}` is derived from `git config user.name` (lowercased with spaces replaced by hyphens). This can be customized via `GW_BRANCH_PREFIX` environment variable.

Worktrees are organized under `~/.worktrees/{repo-name}/{YYYY-MM-DD-name}/{repo-name}/`. This structure allows you to place additional files (e.g., notes) alongside the worktree.

### `gw list` (alias: `gw ls`)

Display all worktrees sorted by date (newest first).

```bash
$ gw list  # or gw ls
sample-user/2025/11/24/feature-login    ~/.worktrees/gw/2025-11-24-feature-login/gw
sample-user/2025/11/23/bugfix-auth      ~/.worktrees/gw/2025-11-23-bugfix-auth/gw
main                                          ~/src/myproject
```

### `gw cd`

Interactively select and navigate to a worktree using peco.

```bash
$ gw cd
# Opens peco with worktree list
# Select a worktree and press Enter
# Your shell will cd to the selected worktree
```

### `gw rm`

Interactively select and remove worktrees (and their branches) using fzf (or peco as fallback).

```bash
$ gw rm
# Opens fzf with worktree list (or peco if fzf not available)
# Note: Main worktree is not shown (cannot be removed)
# With fzf: Press Space to select/deselect, Enter to confirm
# With peco: Select one at a time, choose "Done" to finish
# Confirms before deletion
# Removes both worktree and associated branch
```

### `gw pr checkout`

Checkout a PR branch and create a new worktree for it. Accepts the same arguments as `gh pr checkout`.

```bash
$ gw pr checkout 123
# Checks out PR #123 and creates a worktree for the branch
# You stay in the current directory

$ gw pr checkout feature-branch
# Checkout by branch name

$ gw pr checkout https://github.com/owner/repo/pull/123
# Checkout by URL
```

### `gw ln`

Share gitignored files (like `.env`, `node_modules/`) across worktrees using symlinks.

```bash
$ gw ln add .env
# Share .env across all worktrees

$ gw ln add node_modules
# Works with directories too

$ gw ln ls
# List shared files/directories

$ gw ln pull
# Pull missing shared files into current worktree
# Creates symlinks for files registered in .gw-links.txt that don't exist yet
# Skips files that already exist (with warning)

$ gw ln rm
# Interactively remove a file from sharing
```

## Configuration

Configure gw using environment variables:

### `GW_WORKTREE_ROOT`

Directory where worktrees are created (default: `~/.worktrees`)

```bash
export GW_WORKTREE_ROOT="$HOME/projects/worktrees"
```

### `GW_EDITOR`

Editor to use for branch name input (default: `$EDITOR` or `vim`)

```bash
export GW_EDITOR="nvim"
```

### `GW_BRANCH_PREFIX`

Branch name prefix format (default: `{git-user-name}/{date}/` where `{git-user-name}` is derived from `git config user.name`)

```bash
export GW_BRANCH_PREFIX="feature/{date}/"
```

The `{date}` placeholder will be replaced with the current date in `YYYY/MM/DD` format.

## Examples

### Creating a new feature worktree

```bash
$ gw add
# Enter "user-authentication" in editor
# Creates: sample-user/2025/11/24/user-authentication
# Path: ~/.worktrees/gw/2025-11-24-user-authentication/gw
```

### Switching between worktrees

```bash
$ gw cd
# Select from list using peco
$ pwd
/Users/username/.worktrees/gw/2025-11-24-user-authentication/gw
```

### Cleaning up old worktrees

```bash
$ gw rm
# With fzf: Select multiple worktrees with Space, press Enter
# With peco: Select one at a time, choose "Done" when finished
# Confirm with 'y'
# Selected worktrees and their branches are removed
```

## How it works

### Shell wrapper for `cd`

The `gw cd` command uses a shell wrapper function to enable directory navigation. When you run `gw init`, it generates a shell function that:

1. Detects when you run `gw cd`
2. Executes the gw binary which outputs a `cd` command
3. Evaluates that command in your current shell

This is the same technique used by tools like [try](https://github.com/tobi/try).

## License

MIT

## Author

qawatake
