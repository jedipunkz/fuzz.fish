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

    # Check dependencies
    if not type -q go
        echo "⚠️  fuzz.fish: Go is not installed." >&2
        echo "   Please install Go to use this plugin." >&2
        return 1
    end
    if not type -q fzf
        echo "⚠️  fuzz.fish: fzf is not installed." >&2
        echo "   Please install fzf to use this plugin." >&2
    end

    # Build strategy
    if test -d "$local_src"
        echo "   Building from local source: $local_src"
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
        # We use 'cd' to resolve the absolute path safely without relying on realpath
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

# Uninstall hook (clean up on removal)
function _fuzz_fish_uninstall --on-event fuzz.fish_uninstall
    if test -f "$FUZZ_FISH_BIN_PATH"
        rm -f "$FUZZ_FISH_BIN_PATH"
        echo "🗑️  fuzz.fish: Removed binary"
    end
end

# Initialize on startup
if status is-interactive
    _fuzz_fish_ensure_binary
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
