# fuzz.fish

<img src="./assets/fuzz.png" align="left" width="180" hspace="24" vspace="8" alt="fuzz.fish logo" />

fuzz.fish is a Fish Shell plugin that provides fuzzy finding for command history,
files, git branches, git worktrees, and git commits.

Press `ctrl+r` to open it, type to search, and switch modes with a single key.
No external finder required — a single Go binary ships with the plugin.

[![CI](https://github.com/jedipunkz/fuzz.fish/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/jedipunkz/fuzz.fish/actions/workflows/ci.yml)
[![GitHub Release](https://img.shields.io/github/v/release/jedipunkz/fuzz.fish)](https://github.com/jedipunkz/fuzz.fish/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

<br clear="left" />

<p align="center">
  <img src="./assets/fuzz.gif" width="800" alt="fuzz.fish searching command history, files, and git branches"/>
</p>

## Why fuzz.fish?

- **Nothing else to install.** No `fzf`, `fd`, `ripgrep`, or `bat` alongside it: the plugin is a single Go binary plus Fish keybindings.
- **One keybinding, five modes.** `ctrl+r` opens the finder; `ctrl+s`, `ctrl+w`, `ctrl+g` and `ctrl+x` switch modes without closing it.
- **Every mode has a preview.** History shows when and where the command ran and what surrounded it, files show syntax-highlighted content, and branches, worktrees and commits show their git context.
- **History ranked by frecency.** Match quality ranks first, then `log1p(frequency)` scaled by how recently you last ran the command, so what you actually repeat surfaces first.
- **Git beyond branches.** Worktrees and commits are first-class: `ctrl+x` matches a commit by hash or subject and puts `git show` / `git diff` / `git revert` / `git cherry-pick` on the prompt without running it.

## Requirements

- [Fish Shell](https://fishshell.com/) 3.0+

Installing the plugin downloads a prebuilt binary for macOS and Linux
(`amd64` / `arm64`). On any other platform it falls back to building from
source, which needs [Go](https://golang.org/) 1.25+ and Git.

## Installation

### Using Fisher (Recommended)

```fish
fisher install jedipunkz/fuzz.fish
```

## Usage

Press `ctrl+r` to open fuzz.fish, then type to search. Switch modes at any time with a single key:

| Key | Mode | `enter` does |
|-----|------|--------------|
| `ctrl+r` | Command History Search (default) | Insert the command into your prompt |
| `ctrl+s` | File Search | Insert the file path / `cd` into the directory |
| `ctrl+w` | Git Worktree Search | `cd` into the worktree |
| `ctrl+g` | Git Branch Search | Switch to the selected branch |
| `ctrl+x` | Git Commit Search | Pick a command to run against the commit |

Common keys:

| Key | Action |
|-----|--------|
| `↑`/`↓` or `ctrl+p`/`ctrl+n` | Move the selection |
| `tab` | Complete the query with the selected item |
| `ctrl+y` | Copy the selected item to the clipboard |
| `esc` or `ctrl+c` | Cancel |

Notes:

- Anything already typed on the command line pre-fills the search box, so `vim` then `ctrl+r` starts with history narrowed to `vim`.
- A `*` in the query switches from fuzzy to glob matching in every mode: `nvim *.go` matches `nvim internal/app/filter.go` but not commands that merely contain those letters.
- In Git Branch Search mode, pressing `ctrl+g` again on the current branch runs `git pull origin <branch>`.
- Git Commit Search matches both the short hash and the commit subject. `enter` opens a small action list (`git show`, `git diff`, `git revert`, `git cherry-pick`, `git rebase --onto`, or the bare hash); the chosen command is placed on the prompt without running it. `ctrl+x` outside a git repository shows a warning instead of switching modes.
- File Search skips hidden files and build directories such as `node_modules` and `vendor`.


## License

MIT License - see LICENSE file for details

