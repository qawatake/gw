# gw - Git Worktree Wrapper

A convenient command-line tool to manage git worktrees efficiently.

## Features

- **add**: Create a new branch and corresponding worktree with interactive naming
- **list**: Display all worktrees sorted by recent activity
- **cd**: Interactively select and navigate to a worktree
- **clean**: Interactively select and remove multiple worktrees

## Requirements

- Go 1.24 or later
- Git
- peco (for `gw cd` command)
- fzf (for `gw clean` command, optional - falls back to peco)
- vim or $EDITOR (for `gw add` command)

### Installing dependencies

**macOS:**
```bash
brew install peco fzf
```

**Linux:**
```bash
# peco
go install github.com/peco/peco/cmd/peco@latest

# fzf
git clone --depth 1 https://github.com/junegunn/fzf.git ~/.fzf
~/.fzf/install
```

## Installation

### Using Make (recommended)

```bash
git clone https://github.com/qawatake/gw.git
cd gw
make install
```

This will build the binary and install it to `~/bin/gw`.

### Build from source manually

```bash
git clone https://github.com/qawatake/gw.git
cd gw
go build -o ~/bin/gw ./cmd/gw
```

### Available Make targets

```bash
make build        # Build the binary to bin/gw
make install      # Build and install to ~/bin
make uninstall    # Remove from ~/bin
make clean        # Remove build artifacts
make test         # Run tests
make fmt          # Format code
make tidy         # Tidy dependencies
make lint         # Run linter (requires golangci-lint)
make build-all    # Build for all platforms
make help         # Show help message
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
# Creates branch: qwtk/2025/11/23/feature-login
# Creates worktree at: ~/.worktrees/gw/qwtk-2025-11-23-feature-login
```

The branch name will automatically be prefixed with `qwtk/YYYY/MM/DD/` (configurable via `GW_BRANCH_PREFIX`).

Worktrees are organized by repository name under `~/.worktrees/{repo-name}/` to keep multiple projects organized.

### `gw list`

Display all worktrees sorted by date (newest first).

```bash
$ gw list
qwtk/2025/11/23/feature-login    ~/.worktrees/gw/qwtk-2025-11-23-feature-login
qwtk/2025/11/22/bugfix-auth      ~/.worktrees/gw/qwtk-2025-11-22-bugfix-auth
main                              ~/src/myproject
```

### `gw cd`

Interactively select and navigate to a worktree using peco.

```bash
$ gw cd
# Opens peco with worktree list
# Select a worktree and press Enter
# Your shell will cd to the selected worktree
```

### `gw clean`

Interactively select and remove worktrees using fzf (or peco as fallback).

```bash
$ gw clean
# Opens fzf with worktree list (or peco if fzf not available)
# With fzf: Press Space to select/deselect, Enter to confirm
# With peco: Select one at a time, cancel to finish
# Confirms before deletion
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

Branch name prefix format (default: `qwtk/{date}/`)

```bash
export GW_BRANCH_PREFIX="feature/{date}/"
```

The `{date}` placeholder will be replaced with the current date in `YYYY/MM/DD` format.

## Examples

### Creating a new feature worktree

```bash
$ gw add
# Enter "user-authentication" in editor
# Creates: qwtk/2025/11/23/user-authentication
# Path: ~/.worktrees/gw/qwtk-2025-11-23-user-authentication
```

### Switching between worktrees

```bash
$ gw cd
# Select from list using peco
$ pwd
/Users/username/.worktrees/gw/qwtk-2025-11-23-user-authentication
```

### Cleaning up old worktrees

```bash
$ gw clean
# Select multiple worktrees with Space
# Press Enter, confirm with 'y'
# Selected worktrees are removed
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
