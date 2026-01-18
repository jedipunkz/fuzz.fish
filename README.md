<p align="center">
  <img src="./assets/fuzz.png" width="300" height="300" />
</p>

# fuzz.fish

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

fuzz.fish provides three interactive fuzzy finders:

#### `ctrl+r` - Command History & Git Branch Search

Search through your command history with context, or switch git branches.

- Type to fuzzy search
- **Press `Ctrl+R` again** to toggle between **History Search** and **Git Branch Search** (in git repositories)
- Use arrow keys or ctrl-n, p to navigate
- Press `Enter`:
  - History mode: insert the command into your prompt
  - Git Branch mode: switch to the selected branch
- Press `Ctrl+Y` to copy the selected item to clipboard
- Press `ESC` to cancel


## License

MIT License - see LICENSE file for details

## Issues

Found a bug or have a feature request? Please open an issue on GitHub.

