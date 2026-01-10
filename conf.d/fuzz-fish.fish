# fuzz.fish - Context-aware Fish history viewer
# Initialization and key bindings

# Set the binary path
set -l install_dir "$HOME/.local/share/fuzz.fish/bin"
set -Ux FUZZ_FISH_BIN_PATH "$install_dir/fhv"

# Helper to build binary (mostly for development or manual updates)
function __fuzz_fish_build
    set -l conf_dir (dirname (status -f))
    
    # Try to find source relative to this file (for development)
    if test -f "$conf_dir/../cmd/fhv/main.go"
        set -l source_dir (realpath "$conf_dir/..")
        set -l bin_path "$source_dir/bin/fhv"
        
        # In development mode, we might want to use the local bin
        if go build -o "$bin_path" "$source_dir/cmd/fhv"
            set -Ux FUZZ_FISH_BIN_PATH "$bin_path"
            return 0
        end
    end
    
    # If we are here, we assume the binary is already installed via install.fish
    if test -f "$FUZZ_FISH_BIN_PATH"
        return 0
    end

    # If binary is missing and we can't build it, warn only when trying to use it
    # (The error is handled in fh.fish)
    return 1
end

# Fisher update hook
function __fuzz_fish_update --on-event fuzz_fish_update
    # Re-run install script logic if needed, but usually install.fish handles this.
    # We can try to trigger a rebuild if source is available.
    __fuzz_fish_build
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
