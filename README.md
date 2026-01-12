
<p align="center">
  <img src="./assets/fuzz.png" />
</p>

# fuzz.fish

fuzz.fish is a Fish Shell plugin that provides fuzzy finding for command history, files, and git branches.


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

#### `ctrl+r` - Command History Search

Search through your command history with context:

- Type to fuzzy search through your command history
- Use arrow keys or ctrl-n, p to navigate
- Press `Enter` to insert the command into your prompt
- Press `ESC` to cancel

#### `alt+b` - Git Branch Search

Search and switch git branches:

- Type to fuzzy search through local and remote branches
- Use arrow keys or ctrl-n, p to navigate
- Press `Enter` to switch to the selected branch
- Press `ESC` to cancel


## License

MIT License - see LICENSE file for details

## Issues

Found a bug or have a feature request? Please open an issue on GitHub.

