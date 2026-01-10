{: align="center"}
![fuzz.fish](./assets/fuzz.png)

# fuzz.fish

fuzz.fish is Fish Plugin for Context-Aware Command History Search with Fuzzy Find.


## Requirements

- [Fish Shell](https://fishshell.com/) 3.0+
- [fzf](https://github.com/junegunn/fzf) (fuzzy finder)
- [Go](https://golang.org/) 1.21+ (for building)

### Installing Dependencies

macOS:
```bash
brew install fish fzf go
```

Ubuntu/Debian:
```bash
sudo apt install fish fzf golang
```

Arch Linux:
```bash
sudo pacman -S fish fzf go
```

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

