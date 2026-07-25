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

Press `Ctrl+R` to open fuzz.fish, then type to search. Switch modes at any time with a single key:

| Key | Mode | `Enter` does |
|-----|------|--------------|
| `Ctrl+R` | History (default) | Insert the command into your prompt |
| `Ctrl+G` | Git branches | Switch to the selected branch |
| `Ctrl+S` | Files & directories | Insert the file path / `cd` into the directory |
| `Ctrl+W` | Git worktrees | `cd` into the worktree |

Common keys:

| Key | Action |
|-----|--------|
| `↑`/`↓` or `Ctrl+P`/`Ctrl+N` | Move the selection |
| `Tab` | Complete the query with the selected item |
| `Ctrl+Y` | Copy the selected item to the clipboard |
| `Esc` or `Ctrl+C` | Cancel |

### Tips

- Anything already typed on the command line pre-fills the search box, so `vim` + `Ctrl+R` starts with history narrowed to `vim`.
- A `*` in the query switches from fuzzy to glob matching in every mode: `nvim *.go` matches `nvim internal/app/filter.go` but not unrelated commands.
- In git branch mode, `Ctrl+G` on the current branch runs `git pull origin <branch>`.
- Git branch and worktree modes need a git repository; file mode skips hidden files and build directories such as `node_modules` and `vendor`.


## License

MIT License - see LICENSE file for details

## Issues

Found a bug or have a feature request? Please open an issue on GitHub.

