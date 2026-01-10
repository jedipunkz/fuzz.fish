# fuzz.fish - Context-aware Fish history viewer
# Initialization and key bindings

# Set the binary path
if set -q __fish_config_dir
    set -Ux FUZZ_FISH_BIN_PATH "$__fish_config_dir/functions/fhv"
else
    set -Ux FUZZ_FISH_BIN_PATH "$HOME/.config/fish/functions/fhv"
end

# Build/Install logic
function _fuzz_fish_build
    echo "🔨 Building fuzz.fish binary..."

    # Determine paths
    set -l bin_path "$FUZZ_FISH_BIN_PATH"
    set -l plugin_dir (dirname (status -f))/..
    set -l local_src "$plugin_dir/cmd/fhv"
    
    # Ensure functions directory exists
    mkdir -p (dirname "$bin_path")

    # Check dependencies
    if not type -q go
        echo "⚠️  Go is not installed. You'll need Go to use fuzz.fish." >&2
        return 1
    end
    if not type -q fzf
        echo "⚠️  fzf is not installed. You'll need fzf to use fuzz.fish." >&2
    end

    # Build strategy
    if test -d "$local_src"
        echo "   Building from local source: $local_src"
        if go build -o "$bin_path" "$local_src"
            echo "✅ Build successful!"
        else
            echo "❌ Local build failed!" >&2
            return 1
        end
    else
        echo "   Local source not found (remote install detected)."
        echo "   Installing via go install github.com/jedipunkz/fuzz.fish/cmd/fhv@latest..."
        
        # Use GOBIN to install directly to the target directory
        set -l abs_bin_dir (dirname "$bin_path" | xargs realpath)
        
        if env GOBIN="$abs_bin_dir" go install github.com/jedipunkz/fuzz.fish/cmd/fhv@latest
            echo "✅ Installation successful!"
        else
            echo "❌ Remote installation failed!" >&2
            return 1
        end
    end
    
    echo "   Binary location: $bin_path"
    echo "   Usage: Press Ctrl+R or type 'fh'"
end

# Install hook
function _fuzz_fish_install --on-event fuzz_fish_install
    _fuzz_fish_build
end

# Update hook
function _fuzz_fish_update --on-event fuzz_fish_update
    _fuzz_fish_build
end

# Uninstall hook
function _fuzz_fish_uninstall --on-event fuzz_fish_uninstall
    if test -f "$FUZZ_FISH_BIN_PATH"
        rm -f "$FUZZ_FISH_BIN_PATH"
        echo "🗑️  Removed binary: $FUZZ_FISH_BIN_PATH"
    end
end

# Check binary on startup (fallback if install hook missed)
if status is-interactive
    if not test -f "$FUZZ_FISH_BIN_PATH"
        # Only try to build if we are inside the plugin directory (dev mode) 
        # OR if it seems like a fresh install that failed.
        # However, to avoid slow startup, we might just warn or do nothing.
        # For now, let's just rely on the install hooks.
    end
end

# Set up Ctrl+R to open the history viewer
function __fuzz_fish_key_bindings
    # Bind Ctrl+R to fh function
    bind \cr fh

    # For vi mode users, bind in both insert and normal modes
    if test "$fish_key_bindings" = fish_vi_key_bindings
        bind -M insert \cr fh
        bind -M default \cr fh
    end
end

# Initialize key bindings
__fuzz_fish_key_bindings

# Re-initialize bindings when the key binding mode changes
function __fuzz_fish_postexec --on-event fish_prompt
    # This ensures bindings persist across different key binding modes
end
