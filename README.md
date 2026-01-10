# 🐟 fuzz.fish

Context-aware Fish shell history viewer with fuzzy search powered by fzf.

## ✨ Features

- **Rich Context Display**: View command history with metadata
  - 📁 Execution directory
  - ⏰ Execution time (relative: "2h ago", "3d ago")
  - 📜 Command context (see commands before and after)

- **Fuzzy Search**: Powered by fzf for fast incremental search

- **Interactive Preview**: See command context in the preview window
  - View commands executed before and after
  - Recall entire work sessions easily

- **Keyboard Shortcuts**:
  - `Ctrl+R`: Open history viewer
  - `Enter`: Insert selected command
  - `Ctrl+Y`: Copy command to clipboard
  - `ESC`: Cancel

## 📋 Requirements

- [Fish Shell](https://fishshell.com/) 3.0+
- [fzf](https://github.com/junegunn/fzf) (fuzzy finder)
- [Go](https://golang.org/) 1.21+ (for building)

### Installing Dependencies

**macOS**:
```bash
brew install fish fzf go
```

**Ubuntu/Debian**:
```bash
sudo apt install fish fzf golang
```

**Arch Linux**:
```bash
sudo pacman -S fish fzf go
```

## 🚀 Installation

### Using Fisher (Recommended)

The binary will be automatically built during installation:

```fish
fisher install jedipunkz/fuzz.fish
```

Fisher will:
1. Download the plugin
2. Run `install.fish` to build the Go binary
3. Set up key bindings automatically

### Using Oh My Fish

```fish
omf install https://github.com/jedipunkz/fuzz.fish
```

### Manual Installation

```fish
# Clone the repository
git clone https://github.com/jedipunkz/fuzz.fish.git ~/.config/fish/plugins/fuzz.fish

# Build the binary
cd ~/.config/fish/plugins/fuzz.fish
go build -o bin/fhv ./cmd/fhv

# Source the plugin in your config.fish
echo "source ~/.config/fish/plugins/fuzz.fish/conf.d/fuzz-fish.fish" >> ~/.config/fish/config.fish
echo "set fish_function_path ~/.config/fish/plugins/fuzz.fish/functions \$fish_function_path" >> ~/.config/fish/config.fish
```

**Note**: The Go binary (`bin/fhv`) will be built automatically on first use if it doesn't exist, but using Fisher ensures it's built during installation for a smoother experience.

## 📖 Usage

### Command Line

Run the history viewer directly:

```fish
fh
```

### Keyboard Shortcut

Press `Ctrl+R` in any Fish prompt to open the interactive history viewer.

### In the Viewer

- Type to fuzzy search through your command history
- Use arrow keys to navigate
- Preview window shows:
  - Command details (time, directory)
  - Context: commands before and after
- Press `Enter` to insert the command into your prompt
- Press `Ctrl+Y` to copy without executing
- Press `ESC` to cancel

## 🏗️ Architecture

fuzz.fish combines the power of Go and Fish:

- **Go binary** (`cmd/fhv`): Fast history parsing and formatting
  - Reads Fish history file (`~/.local/share/fish/fish_history`)
  - Formats entries with metadata
  - Generates preview content

- **Fish function** (`functions/fh.fish`): User interface
  - Integrates with fzf
  - Handles command insertion
  - Multi-platform clipboard support (pbcopy/xclip/xsel/wl-copy)

- **Plugin config** (`conf.d/fuzz-fish.fish`): Auto-configuration
  - Sets up `Ctrl+R` key binding
  - Builds binary if missing (on shell startup)
  - Initializes plugin on shell start

- **Install scripts**: Fisher integration
  - `install.fish`: Builds binary during plugin installation
  - `uninstall.fish`: Cleans up binary on uninstall

## 🔧 Configuration

### Custom Key Binding

To use a different key binding, edit `~/.config/fish/conf.d/fuzz-fish.fish`:

```fish
# Use Ctrl+F instead of Ctrl+R
bind \cf fh
```

### Custom fzf Options

You can modify the fzf options in `functions/fh.fish` to customize the appearance and behavior.

## 🐛 Troubleshooting

### fzf not found

Install fzf:
```fish
# macOS
brew install fzf

# Linux
sudo apt install fzf  # or your package manager
```

### Go not found

Install Go:
```fish
# macOS
brew install go

# Linux
sudo apt install golang
```

### Binary not building

Make sure Go is properly installed and in your PATH:
```fish
go version
```

## 📝 License

MIT License - see LICENSE file for details

## 🙏 Acknowledgments

- [fzf](https://github.com/junegunn/fzf) - Amazing fuzzy finder
- [Fish Shell](https://fishshell.com/) - Friendly interactive shell
- Inspired by various history search tools and the need for better context in command history

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📬 Issues

Found a bug or have a feature request? Please open an issue on GitHub.
