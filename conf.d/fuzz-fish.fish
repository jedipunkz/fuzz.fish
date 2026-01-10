# fuzz.fish - Context-aware Fish history viewer
# Initialization and key bindings

# Set the binary path
if set -q __fish_config_dir
    set -Ux FUZZ_FISH_BIN_PATH "$__fish_config_dir/functions/fhv"
else
    set -Ux FUZZ_FISH_BIN_PATH "$HOME/.config/fish/functions/fhv"
end

# Internal function to build/install the binary
function _fuzz_fish_ensure_binary
    set -l bin_path "$FUZZ_FISH_BIN_PATH"

    # If binary exists, nothing to do
    if test -f "$bin_path"
        return 0
    end

    echo "🔨 fuzz.fish: Binary not found. Installing from GitHub..."

    # Ensure functions directory exists
    mkdir -p (dirname "$bin_path")

    # Check dependencies
    if not type -q go
        echo "⚠️  fuzz.fish: Go is not installed." >&2
        echo "   Please install Go to use this plugin." >&2
        return 1
    end

    # Always install from GitHub to ensure latest version
    echo "   Installing from github.com/jedipunkz/fuzz.fish/cmd/fhv@latest..."

    # Use GOBIN to install directly to the target directory
    set -l abs_bin_dir (builtin cd (dirname "$bin_path") && pwd)

    if env GOBIN="$abs_bin_dir" go install github.com/jedipunkz/fuzz.fish/cmd/fhv@latest
        echo "✅ fuzz.fish: Installation successful!"
        return 0
    else
        echo "❌ fuzz.fish: Installation failed!" >&2
        return 1
    end
end

# Install hook - build binary on initial install
function _fuzz_fish_install --on-event fuzz-fish_install
    echo "📦 fuzz.fish: Running install hook..."
    _fuzz_fish_rebuild_binary
end

# Update hook - rebuild binary when plugin is updated
function _fuzz_fish_update --on-event fuzz-fish_update
    echo "🔄 fuzz.fish: Running update hook..."
    _fuzz_fish_rebuild_binary
end

# Uninstall hook
function _fuzz_fish_uninstall --on-event fuzz-fish_uninstall
    if test -f "$FUZZ_FISH_BIN_PATH"
        rm -f "$FUZZ_FISH_BIN_PATH"
        echo "🗑️  fuzz.fish: Removed binary"
    end
end

# Helper function to rebuild binary (used by install and update hooks)
function _fuzz_fish_rebuild_binary
    set -l bin_path "$FUZZ_FISH_BIN_PATH"

    echo "🔨 fuzz.fish: Rebuilding binary from GitHub..."

    # Ensure functions directory exists
    mkdir -p (dirname "$bin_path")

    # Check dependencies
    if not type -q go
        echo "⚠️  fuzz.fish: Go is not installed." >&2
        echo "   Please install Go to use this plugin." >&2
        return 1
    end

    # Remove old binary if exists
    if test -f "$bin_path"
        echo "   Removing old binary..."
        rm -f "$bin_path"
    end

    # Clear module cache for this package to force fresh download
    echo "   Clearing module cache..."
    set -l gopath (go env GOPATH)
    if test -d "$gopath/pkg/mod/github.com/jedipunkz"
        rm -rf "$gopath/pkg/mod/github.com/jedipunkz/fuzz.fish@"*
        echo "   Cache cleared"
    end

    # Always install from GitHub to ensure latest version
    echo "   Installing from github.com/jedipunkz/fuzz.fish/cmd/fhv@latest..."

    # Use GOBIN to install directly to the target directory
    set -l abs_bin_dir (builtin cd (dirname "$bin_path") && pwd)

    if env GOBIN="$abs_bin_dir" go install github.com/jedipunkz/fuzz.fish/cmd/fhv@latest
        echo "✅ fuzz.fish: Installation successful!"
        echo "   Binary location: $bin_path"
        ls -lh "$bin_path"
        return 0
    else
        echo "❌ fuzz.fish: Installation failed!" >&2
        return 1
    end
end

# Initialize on startup
if status is-interactive
    _fuzz_fish_ensure_binary
end

# Set up Ctrl+R key bindings
function __fuzz_fish_key_bindings
    bind \cr fh
    if test "$fish_key_bindings" = fish_vi_key_bindings
        bind -M insert \cr fh
        bind -M default \cr fh
    end
end
__fuzz_fish_key_bindings

function __fuzz_fish_postexec --on-event fish_prompt
end
