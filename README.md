
<p align="center">
  <img src="./assets/fuzz.png" />
</p>

# fuzz.fish

fuzz.fish is Fish Plugin for Context-Aware Command History Search with Fuzzy Find.


## Requirements

- [Fish Shell](https://fishshell.com/) 3.0+
- [Go](https://golang.org/) 1.21+ (for building)
- [bat](https://github.com/sharkdp/bat) (optional, for syntax-highlighted file preview)

## Installation

### Using Fisher (Recommended)

```fish
fisher install jedipunkz/fuzz.fish
```

## Usage

### Keyboard Shortcuts

#### `Ctrl+R` - History Search

Press `Ctrl+R` in any Fish prompt to open the interactive history viewer.

In the viewer:
- Type to fuzzy search through your command history
- Use arrow keys to navigate
- Preview window shows:
  - Command details (time, directory)
  - Context: commands before and after
- Press `Enter` to insert the command into your prompt
- Press `ESC` to cancel

#### `Ctrl+Alt+F` - File/Directory Search

Press `Ctrl+Alt+F` to search files and directories in the current directory.

Features:
- Incremental search for files and directories
- Real-time preview:
  - Files: Syntax-highlighted preview (with `bat`) or plain text
  - Directories: Directory listing
- Directory selection: `cd` into the selected directory
- File selection: Insert file path into the command line
- Press `Ctrl+/` to toggle preview window
- Press `ESC` to cancel


## License

MIT License - see LICENSE file for details



## Issues

Found a bug or have a feature request? Please open an issue on GitHub.

