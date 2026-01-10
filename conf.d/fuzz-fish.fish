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

    # Binary not found, build it using the same logic as rebuild
    _fuzz_fish_rebuild_binary
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

    if not type -q git
        echo "⚠️  fuzz.fish: Git is not installed." >&2
        echo "   Please install Git to use this plugin." >&2
        return 1
    end

    # Remove old binary if exists
    if test -f "$bin_path"
        echo "   Removing old binary..."
        rm -f "$bin_path"
    end

    # Create temporary directory
    set -l tmp_dir (mktemp -d)
    echo "   Cloning repository to $tmp_dir..."

    # Clone repository
    if not git clone --depth 1 https://github.com/jedipunkz/fuzz.fish.git "$tmp_dir" 2>&1 | grep -v "Cloning into"
        echo "❌ fuzz.fish: Failed to clone repository!" >&2
        rm -rf "$tmp_dir"
        return 1
    end

    echo "   Building binary..."

    # Build from cloned source
    pushd "$tmp_dir"
    go mod download
    if go build -o "$bin_path" ./cmd/fhv
        popd
        echo "✅ fuzz.fish: Build successful!"
        echo "   Binary location: $bin_path"
        ls -lh "$bin_path"
        # Clean up temporary directory
        rm -rf "$tmp_dir"
        return 0
    else
        popd
        echo "❌ fuzz.fish: Build failed!" >&2
        rm -rf "$tmp_dir"
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
