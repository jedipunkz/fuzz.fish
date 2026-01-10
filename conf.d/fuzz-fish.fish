# fuzz.fish - Context-aware Fish history viewer
# Initialization and key bindings

# Set the binary path
if set -q __fish_config_dir
    set -Ux FUZZ_FISH_BIN_PATH "$__fish_config_dir/functions/fhv"
else
    set -Ux FUZZ_FISH_BIN_PATH "$HOME/.config/fish/functions/fhv"
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
