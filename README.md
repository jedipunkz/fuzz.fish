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
- [Go](https://golang.org/) 1.21+ (for building)

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
| `Ctrl+G` | Toggle between History and Git Branch mode |
| `Ctrl+S` | Switch to File Search mode |
| `↑/↓` or `Ctrl+P/N` | Navigate through results |
| `Enter` | Select item |
| `Ctrl+Y` | Copy selected item to clipboard |
| `ESC` or `Ctrl+C` | Cancel |

#### History Search Mode (default)

Search through your command history with context.

- Type to fuzzy search
- Press `Enter` to insert the selected command into your prompt

#### Git Branch Mode

Search and switch git branches (available in git repositories).

- Press `Ctrl+G` to toggle from History mode
- Press `Enter` to switch to the selected branch

#### File Search Mode

Search files and directories in the current directory.

- Press `Ctrl+S` to switch to File Search mode
- Type to fuzzy search files and directories
- Press `Enter`:
  - File: insert the file path into your prompt
  - Directory: cd into the selected directory
- Hidden files and common build directories (node_modules, vendor, etc.) are automatically excluded


## License

MIT License - see LICENSE file for details

## Issues

Found a bug or have a feature request? Please open an issue on GitHub.

