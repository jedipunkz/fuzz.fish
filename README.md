<p align="center">
  <img src="./assets/fuzz.png" width="300" height="300" />
</p>

# fuzz.fish

[![CI](https://github.com/jedipunkz/fuzz.fish/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/jedipunkz/fuzz.fish/actions/workflows/ci.yml)
[![GitHub Release](https://img.shields.io/github/v/release/jedipunkz/fuzz.fish)](https://github.com/jedipunkz/fuzz.fish/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

fuzz.fish is a Fish Shell plugin that provides fuzzy finding for command history, files, and git branches.

# Screenshot

<p align="center">
  <img src="./assets/fuzz.gif" width="800"/>
</p>

## Requirements

- [Fish Shell](https://fishshell.com/) 3.0+
- [Go](https://golang.org/) 1.24+ (for building)

## Installation

### Using Fisher (Recommended)

```fish
fisher install jedipunkz/fuzz.fish
```

## Usage

### Keyboard Shortcuts

fuzz.fish provides three interactive fuzzy finders accessible through a unified interface:

#### `Ctrl+R` - Unified Fuzzy Finder

Press `Ctrl+R` to open the fuzzy finder. You can switch between different modes:

| Key | Action |
|-----|--------|
| `Ctrl+G` | Switch to Git Branch Search Mode |
| `Ctrl+S` | Switch to File Search Mode |
| `Ctrl+W` | Switch to Git Worktree Search Mode |
| `↑/↓` or `Ctrl+P/N` | Navigate through results |
| `Enter` | Select item |
| `Ctrl+Y` | Copy selected item to clipboard |
| `ESC` or `Ctrl+C` | Cancel |

#### History Search Mode (default)

Search through your command history with context.

- Type to fuzzy search
- Press `Enter` to insert the selected command into your prompt

#### Glob Search

Include a `*` in your query to switch from fuzzy matching to glob matching. Each `*` matches any run of characters, while the literal parts must appear contiguously and in order. This works in every mode.

For example, typing `nvim *.go` lists only the commands where you opened a `.go` file with `nvim` (e.g. `nvim internal/app/filter.go`), instead of scattering those characters fuzzily.

#### Git Branch Mode

Search and switch git branches (available in git repositories).

- Press `Ctrl+G` to toggle from History mode
- Press `Enter` to switch to the selected branch
- Press `Ctrl+G` again (while in Git Branch mode) to `git pull origin <branch>` for the current branch

#### File Search Mode

Search files and directories in the current directory.

- Press `Ctrl+S` to switch to File Search mode
- Type to fuzzy search files and directories
- Press `Enter`:
  - File: insert the file path into your prompt
  - Directory: cd into the selected directory
- Hidden files and common build directories (node_modules, vendor, etc.) are automatically excluded

#### Git Worktree Mode

Search and switch between git worktrees (available in git repositories).

- Press `Ctrl+W` to switch to Git Worktree mode
- Type to fuzzy search worktrees by path
- Each entry shows its checked-out branch; the current worktree is marked with `*`
- Press `Enter` to `cd` into the selected worktree's directory


## License

MIT License - see LICENSE file for details

## Issues

Found a bug or have a feature request? Please open an issue on GitHub.

