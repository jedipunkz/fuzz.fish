
<p align="center">
  <img src="./assets/fuzz.png" />
</p>

# fuzz.fish

fuzz.fish is Fish Plugin for Context-Aware Command History Search with Fuzzy Find.


## Requirements

- [Fish Shell](https://fishshell.com/) 3.0+
- [Go](https://golang.org/) 1.21+ (for building)

## Installation

### Using Fisher (Recommended)

```fish
fisher install jedipunkz/fuzz.fish
```

## Usage

### Keyboard Shortcut

Press `Ctrl+R` in any Fish prompt to open the interactive history viewer.

### In the Viewer

- Type to fuzzy search through your command history
- Use arrow keys to navigate
- Preview window shows:
  - Command details (time, directory)
  - Context: commands before and after
- Press `Enter` to insert the command into your prompt
- Press `ESC` to cancel


## License

MIT License - see LICENSE file for details



## Issues

Found a bug or have a feature request? Please open an issue on GitHub.

