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

    echo "🔨 fuzz.fish: Binary not found. Setting up..."

    # Determine paths
    set -l plugin_dir (dirname (status -f))/..
    set -l local_src "$plugin_dir/cmd/fhv"
    
    # Ensure functions directory exists
    mkdir -p (dirname "$bin_path")

    # Check dependencies - ONLY GO is needed now, fzf removed
    if not type -q go
        echo "⚠️  fuzz.fish: Go is not installed." >&2
        echo "   Please install Go to use this plugin." >&2
        return 1
    end

    # Build strategy
    if test -d "$local_src"
        echo "   Building from local source: $local_src"
        
        # We need to run go mod tidy probably, or ensure dependencies are fetched
        # Since we changed go.mod, we should run go mod tidy
        pushd "$plugin_dir"
        go mod tidy
        popd
        
        if go build -o "$bin_path" "$local_src"
            echo "✅ fuzz.fish: Build successful!"
        else
            echo "❌ fuzz.fish: Local build failed!" >&2
            return 1
        end
    else
        echo "   Local source not found. Installing from GitHub..."
        echo "   Target: $bin_path"
        
        # Use GOBIN to install directly to the target directory
        set -l abs_bin_dir (builtin cd (dirname "$bin_path") && pwd)
        
        if env GOBIN="$abs_bin_dir" go install github.com/jedipunkz/fuzz.fish/cmd/fhv@latest
            echo "✅ fuzz.fish: Installation successful!"
        else
            echo "❌ fuzz.fish: Remote installation failed!" >&2
            return 1
        end
    end
    
    return 0
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

    echo "🔨 fuzz.fish: Rebuilding binary..."

    # Determine paths
    set -l plugin_dir (dirname (status -f))/..
    set -l local_src "$plugin_dir/cmd/fhv"

    # Ensure functions directory exists
    mkdir -p (dirname "$bin_path")

    # Check dependencies - ONLY GO is needed now
    if not type -q go
        echo "⚠️  fuzz.fish: Go is not installed." >&2
        echo "   Please install Go to use this plugin." >&2
        return 1
    end

    # Remove old binary if exists
    if test -f "$bin_path"
        rm -f "$bin_path"
    end

    # Build from local source
    if test -d "$local_src"
        echo "   Building from local source: $local_src"

        # Debug: Show main.go first line to verify source version
        echo "   Checking source version..."
        head -n 15 "$local_src/main.go" | tail -n 3

        # Debug: Show go.mod
        echo "   go.mod requires:"
        grep "require" "$plugin_dir/go.mod"

        pushd "$plugin_dir"
        # Ensure dependencies are up to date
        echo "   Running go mod tidy..."
        go mod tidy 2>&1 | grep -v "go: downloading" || true
        echo "   Running go mod download..."
        go mod download
        popd

        echo "   Building binary..."
        if go build -o "$bin_path" "$local_src"
            echo "✅ fuzz.fish: Build successful!"
            # Show binary modification time
            ls -lh "$bin_path"
            return 0
        else
            echo "❌ fuzz.fish: Local build failed!" >&2
            return 1
        end
    else
        echo "   Local source not found. Installing from GitHub..."
        echo "   Target: $bin_path"

        # Use GOBIN to install directly to the target directory
        set -l abs_bin_dir (builtin cd (dirname "$bin_path") && pwd)

        if env GOBIN="$abs_bin_dir" go install github.com/jedipunkz/fuzz.fish/cmd/fhv@latest
            echo "✅ fuzz.fish: Installation successful!"
            return 0
        else
            echo "❌ fuzz.fish: Remote installation failed!" >&2
            return 1
        end
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
