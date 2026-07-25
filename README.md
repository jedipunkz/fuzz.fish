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

Press `ctrl+r` to open fuzz.fish, then type to search. Switch modes at any time with a single key:

| Key | Mode | `enter` does |
|-----|------|--------------|
| `ctrl+r` | Command History Search (default) | Insert the command into your prompt |
| `ctrl+s` | File Search | Insert the file path / `cd` into the directory |
| `ctrl+w` | Git Worktree Search | `cd` into the worktree |
| `ctrl+g` | Git Branch Search | Switch to the selected branch |

Common keys:

| Key | Action |
|-----|--------|
| `↑`/`↓` or `ctrl+p`/`ctrl+n` | Move the selection |
| `tab` | Complete the query with the selected item |
| `ctrl+y` | Copy the selected item to the clipboard |
| `esc` or `ctrl+c` | Cancel |

#### Command History Search Mode

`ctrl+r` — searches your Fish command history, ranking frequently and recently used commands first. Anything already typed on the command line pre-fills the search box, so `vim` then `ctrl+r` starts with history narrowed to `vim`.

#### File Search Mode

`ctrl+s` — searches files and directories under the current directory. Hidden files and build directories such as `node_modules` and `vendor` are skipped.

#### Git Worktree Search Mode

`ctrl+w` — lists the worktrees of the current repository with their checked-out branch, marking the current one with `*`.

#### Git Branch Search Mode

`ctrl+g` — lists the branches of the current repository. Pressing `ctrl+g` again on the current branch runs `git pull origin <branch>`.

#### Glob Search

A `*` in the query switches from fuzzy to glob matching in every mode: `nvim *.go` matches `nvim internal/app/filter.go` but not commands that merely contain those letters.


## License

MIT License - see LICENSE file for details

