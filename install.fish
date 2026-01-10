#!/usr/bin/env fish
# Fisher install script

echo "📦 Installing fuzz.fish..."

set -l plugin_dir (dirname (status -f))
set -l bin_path "$plugin_dir/bin/fhv"

# Check if Go is installed
if not type -q go
    echo "⚠️  Go is not installed. You'll need Go to use fuzz.fish." >&2
    echo "   Download from: https://golang.org/dl/" >&2
    echo "   Or install via:" >&2
    echo "     • macOS: brew install go" >&2
    echo "     • Ubuntu/Debian: sudo apt install golang" >&2
    echo "     • Arch: sudo pacman -S go" >&2
    exit 1
end

# Check if fzf is installed
if not type -q fzf
    echo "⚠️  fzf is not installed. You'll need fzf to use fuzz.fish." >&2
    echo "   Install via:" >&2
    echo "     • macOS: brew install fzf" >&2
    echo "     • Ubuntu/Debian: sudo apt install fzf" >&2
    echo "     • Arch: sudo pacman -S fzf" >&2
end

# Build the binary
echo "🔨 Building fhv binary..."
mkdir -p "$plugin_dir/bin"

if go build -o "$bin_path" "$plugin_dir/cmd/fhv"
    echo "✅ fuzz.fish installed successfully!"
    echo ""
    echo "   Usage: Press Ctrl+R or type 'fh'"
    echo ""
else
    echo "❌ Build failed!" >&2
    exit 1
end
