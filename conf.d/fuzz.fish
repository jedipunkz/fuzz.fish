# fuzz.fish - Context-aware Fish history viewer
# Initialization and key bindings

# Set the binary path
if set -q __fish_config_dir
    set -Ux FUZZ_FISH_BIN_PATH "$__fish_config_dir/functions/fuzz"
else
    set -Ux FUZZ_FISH_BIN_PATH "$HOME/.config/fish/functions/fuzz"
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
function _fuzz_fish_install --on-event fuzz_install
    echo "📦 fuzz.fish: Running install hook..."
    _fuzz_fish_rebuild_binary
end

# Update hook - rebuild binary when plugin is updated
function _fuzz_fish_update --on-event fuzz_update
    echo "🔄 fuzz.fish: Running update hook..."
    # Force rebuild by removing existing binary
    if test -f "$FUZZ_FISH_BIN_PATH"
        echo "   Removing old binary to force rebuild..."
        rm -f "$FUZZ_FISH_BIN_PATH"
    end
    _fuzz_fish_rebuild_binary
end

# Uninstall hook
function _fuzz_fish_uninstall --on-event fuzz_uninstall
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
    if git clone --depth 1 https://github.com/jedipunkz/fuzz.fish.git "$tmp_dir" >/dev/null 2>&1
        echo "   Clone successful"
    else
        echo "❌ fuzz.fish: Failed to clone repository!" >&2
        rm -rf "$tmp_dir"
        return 1
    end

    echo "   Building binary..."

    # Build from cloned source
    pushd "$tmp_dir" >/dev/null

    # Generate go.sum and download dependencies
    echo "   Downloading dependencies..."
    go mod tidy >/dev/null 2>&1
    go mod download >/dev/null 2>&1

    if go build -o "$bin_path" ./cmd/fuzz
        popd >/dev/null
        echo "✅ fuzz.fish: Build successful!"
        echo "   Binary location: $bin_path"
        ls -lh "$bin_path"
        # Clean up temporary directory
        rm -rf "$tmp_dir"
        return 0
    else
        popd >/dev/null
        echo "❌ fuzz.fish: Build failed!" >&2
        rm -rf "$tmp_dir"
        return 1
    end
end

# Helper function to check binary and rebuild if needed
function _fuzz_ensure_binary_or_error --description 'Internal: Ensure binary exists'
    set -l bin_path "$FUZZ_FISH_BIN_PATH"

    if test -z "$bin_path"; or not test -f "$bin_path"
        if functions -q _fuzz_fish_ensure_binary
            _fuzz_fish_ensure_binary
        else
            echo "❌ fuzz.fish: Binary not found. Please restart your shell." >&2
            return 1
        end
    end

    echo "$bin_path"
end

# Initialize on startup
if status is-interactive
    _fuzz_fish_ensure_binary
end

# History search function
function fh --description 'Fish History viewer with context (TUI)'
    set -l bin_path (_fuzz_ensure_binary_or_error); or return 1

    # Run the TUI binary
    # Redirect stdin/stderr to /dev/tty for TUI interaction,
    # while capturing stdout for the selected command
    set -l cmd ($bin_path </dev/tty 2>/dev/tty)

    if test -n "$cmd"
        # Insert into command line
        commandline -r -- "$cmd"
        commandline -f repaint
    end
end

# File/directory search function
function ff --description 'Search files and directories with preview (TUI)'
    set -l bin_path (_fuzz_ensure_binary_or_error); or return 1

    # Run the TUI binary with 'files' subcommand
    # Redirect stdin/stderr to /dev/tty for TUI interaction,
    # while capturing stdout for the selected file/dir
    set -l result ($bin_path files </dev/tty 2>/dev/tty)

    if test -n "$result"
        # Parse the result: DIR:<path> or FILE:<path>
        if string match -q "DIR:*" -- "$result"
            # It's a directory, cd into it
            set -l dir_path (string replace "DIR:" "" -- "$result")
            cd "$dir_path"
            commandline -f repaint
        else if string match -q "FILE:*" -- "$result"
            # It's a file, insert into command line
            set -l file_path (string replace "FILE:" "" -- "$result")
            commandline -i -- "$file_path"
            commandline -f repaint
        end
    end
end

# Git branch search function
function gb --description 'Git branch search with fuzzy finder (TUI)'
    set -l bin_path (_fuzz_ensure_binary_or_error); or return 1

    # Run the TUI binary with 'git branch' subcommand
    # The binary handles /dev/tty internally for TUI,
    # and outputs the selected branch name to stdout
    set -l branch ($bin_path git branch)

    if test -n "$branch"
        # Checkout the selected branch
        git checkout "$branch"
        commandline -f repaint
    end
end

# Set up Ctrl+R key bindings for history
# Set up Ctrl+Alt+F key bindings for file search
# Set up Alt+B key bindings for git branch search
function __fuzz_fish_key_bindings
    bind \cr fh
    bind \e\cf ff
    bind \eb gb
    if test "$fish_key_bindings" = fish_vi_key_bindings
        bind -M insert \cr fh
        bind -M default \cr fh
        bind -M insert \e\cf ff
        bind -M default \e\cf ff
        bind -M insert \eb gb
        bind -M default \eb gb
    end
end
__fuzz_fish_key_bindings
