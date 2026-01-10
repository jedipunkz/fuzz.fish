# fuzz.fish - Context-aware Fish history viewer
# Initialization and key bindings

# Detect the source directory
function __fuzz_fish_get_source_dir
    # Check if fish_plugins exists and contains fuzz.fish path
    if test -f ~/.config/fish/fish_plugins
        set -l fuzz_path (grep "fuzz.fish" ~/.config/fish/fish_plugins | head -1)
        if test -n "$fuzz_path"; and test -d "$fuzz_path"
            echo "$fuzz_path"
            return 0
        end
    end

    # Fallback: try to find it relative to this file
    set -l conf_dir (dirname (status -f))
    if test -f "$conf_dir/../cmd/fhv/main.go"
        realpath "$conf_dir/.."
        return 0
    end

    return 1
end

# Build the Go binary
function __fuzz_fish_build
    set -l source_dir (__fuzz_fish_get_source_dir)
    if test -z "$source_dir"
        return 0  # Silently skip if source not found
    end

    set -l bin_path "$source_dir/bin/fhv"
    set -l src_path "$source_dir/cmd/fhv/main.go"

    # Check if Go is installed
    if not type -q go
        echo "⚠️  fuzz.fish: Go is not installed. The plugin will not work." >&2
        echo "   Install Go from: https://golang.org/dl/" >&2
        return 1
    end

    # Build if binary doesn't exist or source is newer
    if not test -f "$bin_path"; or test "$src_path" -nt "$bin_path"
        echo "🔨 fuzz.fish: Building binary..." >&2
        mkdir -p "$source_dir/bin"
        if go build -o "$bin_path" "$source_dir/cmd/fhv"
            echo "✅ fuzz.fish: Build successful!" >&2
            # Set global variable for fh function to use
            set -Ux FUZZ_FISH_BIN_PATH "$bin_path"
            return 0
        else
            echo "❌ fuzz.fish: Build failed!" >&2
            return 1
        end
    else
        # Binary exists, set the path
        set -Ux FUZZ_FISH_BIN_PATH "$bin_path"
    end
end

# Fisher install hook - build on install
function __fuzz_fish_install --on-event fuzz_fish_install
    __fuzz_fish_build
end

# Fisher update hook - rebuild on update
function __fuzz_fish_update --on-event fuzz_fish_update
    __fuzz_fish_build
end

# Build on shell startup if binary doesn't exist
if status is-interactive
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
