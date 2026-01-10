# fuzz.fish - Context-aware Fish history viewer
# Initialization and key bindings

# Build the Go binary
function __fuzz_fish_build
    set -l plugin_dir (dirname (status -f))
    set -l bin_path "$plugin_dir/../bin/fhv"
    set -l src_path "$plugin_dir/../cmd/fhv/main.go"

    # Check if Go is installed
    if not type -q go
        echo "⚠️  fuzz.fish: Go is not installed. The plugin will not work." >&2
        echo "   Install Go from: https://golang.org/dl/" >&2
        return 1
    end

    # Build if binary doesn't exist or source is newer
    if not test -f "$bin_path"; or test "$src_path" -nt "$bin_path"
        echo "🔨 fuzz.fish: Building binary..." >&2
        mkdir -p "$plugin_dir/../bin"
        if go build -o "$bin_path" "$plugin_dir/../cmd/fhv"
            echo "✅ fuzz.fish: Build successful!" >&2
            return 0
        else
            echo "❌ fuzz.fish: Build failed!" >&2
            return 1
        end
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
