#!/usr/bin/env fish
# Fisher install script

echo "📦 Installing fuzz.fish..."

# Determine the functions directory
if set -q __fish_config_dir
    set -g _fuzz_fish_functions_dir "$__fish_config_dir/functions"
else
    set -g _fuzz_fish_functions_dir "$HOME/.config/fish/functions"
end

set -l bin_path "$_fuzz_fish_functions_dir/fhv"
set -l plugin_dir (dirname (status -f))
set -l local_src "$plugin_dir/cmd/fhv"

# Check if Go is installed
if not type -q go
    echo "⚠️  Go is not installed. You'll need Go to use fuzz.fish." >&2
    echo "   Download from: https://golang.org/dl/" >&2
    exit 1
end

# Check if fzf is installed
if not type -q fzf
    echo "⚠️  fzf is not installed. You'll need fzf to use fuzz.fish." >&2
end

echo "🔨 Building fuzz.fish binary..."

# Hybrid build strategy:
# 1. If local source exists (fisher install .), use go build
# 2. If remote install (fisher install jedipunkz/fuzz.fish), use go install
if test -d "$local_src"
    echo "   Building from local source: $local_src"
    if go build -o "$bin_path" "$local_src"
        echo "✅ Build successful!"
    else
        echo "❌ Local build failed!" >&2
        exit 1
    end
else
    echo "   Local source not found (remote install detected)."
    echo "   Installing via go install github.com/jedipunkz/fuzz.fish/cmd/fhv@latest..."
    
    # use GOBIN to install directly to the target directory
    # Note: GOBIN must be an absolute path
    set -l abs_functions_dir (realpath "$_fuzz_fish_functions_dir")
    
    if env GOBIN="$abs_functions_dir" go install github.com/jedipunkz/fuzz.fish/cmd/fhv@latest
        echo "✅ Installation successful!"
    else
        echo "❌ Remote installation failed!" >&2
        exit 1
    end
end

echo "   Binary location: $bin_path"
echo ""
echo "   Usage: Press Ctrl+R or type 'fh'"
echo ""
