# fuzz.fish - Context-aware Fish history viewer
# Initialization and key bindings

# Drop the universal value older versions stored. It was never read back --
# the path is recomputed from $__fish_config_dir on every start -- and it
# outlived uninstall. Erasing a variable that is already gone is a silent no-op.
set -e -U FUZZ_FISH_BIN_PATH

# Set the binary path. Global, not universal: nothing depends on the value
# surviving a restart. `-u` unexports it -- every reader below is a fish
# function in this shell, so child processes have no use for it. The flag has to
# be explicit: a shell started from one that still exported the variable
# inherits it as an exported global, and a plain `set -g` would keep it that way.
if set -q __fish_config_dir
    set -gu FUZZ_FISH_BIN_PATH "$__fish_config_dir/functions/fuzz"
else
    set -gu FUZZ_FISH_BIN_PATH "$HOME/.config/fish/functions/fuzz"
end

# The release this script expects. Downloads are pinned to it, so a plugin
# revision always installs the binary it was written against instead of
# whatever is newest. The release workflow rewrites this line before tagging.
set -gu __fuzz_fish_version v0.4.1

# Internal function to build/install the binary
function _fuzz_fish_ensure_binary
    set -l bin_path "$FUZZ_FISH_BIN_PATH"

    if test -f "$bin_path"
        # A binary older than --version writes its error to stderr and exits
        # non-zero, leaving an empty capture. That compares unequal like any
        # other mismatch, so no separate case is needed.
        set -l installed ("$bin_path" --version 2>/dev/null | string trim)
        if test "$installed" = "$__fuzz_fish_version"
            return 0
        end
        echo "🔄 fuzz.fish: installed binary ($installed) does not match $__fuzz_fish_version"
    end

    # Missing or stale, build it using the same logic as rebuild
    _fuzz_fish_rebuild_binary
end

# Install hook - install binary on initial install
function _fuzz_fish_install --on-event fuzz_install
    echo "📦 fuzz.fish: Running install hook..."
    # Ensure, not reinstall: sourcing this file already installed the binary in
    # an interactive shell, and fisher fires this event right afterwards.
    _fuzz_fish_ensure_binary
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

# Helper function to install the binary (used by install and update hooks)
function _fuzz_fish_rebuild_binary
    set -l bin_path "$FUZZ_FISH_BIN_PATH"

    echo "📥 fuzz.fish: Installing binary..."

    # Ensure functions directory exists
    mkdir -p (dirname "$bin_path")

    # Remove old binary if exists
    if test -f "$bin_path"
        echo "   Removing old binary..."
        rm -f "$bin_path"
    end

    if _fuzz_fish_download_binary "$bin_path"
        return 0
    end

    echo "   No prebuilt binary available, falling back to building from source..."
    _fuzz_fish_build_binary "$bin_path"
end

# Compare a downloaded asset against its line in the release checksums.txt.
# Returns non-zero when no digest tool is available, when the file has no entry
# for this asset, or when the digests differ, so the caller can fall back to
# building from source rather than installing an unverified binary.
function _fuzz_fish_verify_checksum
    set -l file $argv[1]
    set -l asset $argv[2]
    set -l checksums $argv[3]

    set -l actual
    if type -q shasum
        set actual (shasum -a 256 "$file" | string split -f1 ' ')
    else if type -q sha256sum
        set actual (sha256sum "$file" | string split -f1 ' ')
    else
        echo "fuzz.fish: neither shasum nor sha256sum is available to verify $asset" >&2
        return 1
    end

    # Both tools print "<digest>  <name>"; GNU binary mode prints "<digest> *<name>".
    set -l pattern (string escape --style=regex -- $asset)
    set -l expected (string match -rg '^([0-9a-f]{64}) [ *]'$pattern'$' <"$checksums")

    if test (count $expected) -ne 1
        echo "fuzz.fish: no checksum entry for $asset" >&2
        return 1
    end

    if test "$expected[1]" != "$actual[1]"
        echo "fuzz.fish: checksum mismatch for $asset" >&2
        return 1
    end

    return 0
end

# Download the prebuilt binary for this platform from the pinned release.
# Returns non-zero when the platform has no asset, or the download fails, so
# the caller can fall back to building from source.
function _fuzz_fish_download_binary
    set -l bin_path $argv[1]

    if not type -q curl; or not type -q tar
        return 1
    end

    set -l os
    switch (uname -s)
        case Darwin
            set os darwin
        case Linux
            set os linux
        case '*'
            return 1
    end

    set -l arch
    switch (uname -m)
        case x86_64 amd64
            set arch amd64
        case arm64 aarch64
            set arch arm64
        case '*'
            return 1
    end

    set -l asset "fuzz_"$os"_"$arch".tar.gz"
    # Pinned, not "latest": the asset and the checksums.txt that verifies it
    # must both come from the release this script was written against.
    set -l base "https://github.com/jedipunkz/fuzz.fish/releases/download/$__fuzz_fish_version"
    set -l tmp_dir (mktemp -d)

    echo "   Downloading $asset..."
    if not curl -fsSL "$base/$asset" -o "$tmp_dir/$asset"
        rm -rf "$tmp_dir"
        return 1
    end

    if not curl -fsSL "$base/checksums.txt" -o "$tmp_dir/checksums.txt"
        echo "fuzz.fish: could not download checksums.txt for $asset" >&2
        rm -rf "$tmp_dir"
        return 1
    end

    if not _fuzz_fish_verify_checksum "$tmp_dir/$asset" "$asset" "$tmp_dir/checksums.txt"
        rm -rf "$tmp_dir"
        return 1
    end

    if not tar -xzf "$tmp_dir/$asset" -C "$tmp_dir"; or not test -f "$tmp_dir/fuzz"
        rm -rf "$tmp_dir"
        return 1
    end

    mv "$tmp_dir/fuzz" "$bin_path"
    chmod +x "$bin_path"
    rm -rf "$tmp_dir"

    echo "✅ fuzz.fish: Installed prebuilt binary"
    echo "   Binary location: $bin_path"
    return 0
end

# Fallback used when no release asset matches this platform: build from source.
function _fuzz_fish_build_binary
    set -l bin_path $argv[1]

    # Check dependencies
    if not type -q go
        echo "⚠️  fuzz.fish: no prebuilt binary for this platform and Go is not installed." >&2
        echo "   Please install Go to build the plugin from source." >&2
        return 1
    end

    if not type -q git
        echo "⚠️  fuzz.fish: Git is not installed." >&2
        echo "   Please install Git to use this plugin." >&2
        return 1
    end

    # Create temporary directory
    set -l tmp_dir (mktemp -d)
    echo "   Cloning repository to $tmp_dir..."

    # Clone the pinned release, not the default branch, so the fallback build
    # produces the same version this script expects.
    if git clone --depth 1 --branch "$__fuzz_fish_version" https://github.com/jedipunkz/fuzz.fish.git "$tmp_dir" >/dev/null 2>&1
        echo "   Clone successful"
    else
        echo "❌ fuzz.fish: Failed to clone repository!" >&2
        rm -rf "$tmp_dir"
        return 1
    end

    echo "   Building binary..."

    # Build from cloned source
    pushd "$tmp_dir" >/dev/null

    # Download dependencies pinned by the committed go.mod / go.sum, so this
    # build resolves the same versions as a local `make install`.
    echo "   Downloading dependencies..."
    go mod download >/dev/null 2>&1

    # Stamp the version the release workflow would stamp, otherwise the binary
    # reports "dev" and _fuzz_fish_ensure_binary rebuilds it on every startup.
    if go build -ldflags "-X main.version=$__fuzz_fish_version" -o "$bin_path" ./cmd/fuzz
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
            # Build progress goes to stderr: this function's stdout is captured
            # by the caller as the binary path.
            _fuzz_fish_ensure_binary >&2
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

    # Pre-fill the search box with the current command line buffer so that,
    # e.g. after typing "vim", Ctrl+R opens history already filtered by "vim".
    set -l query (commandline)

    # Run the TUI binary
    # Redirect stdin/stderr to /dev/tty for TUI interaction,
    # while capturing stdout for the selected command/branch/file
    # `string collect` keeps the output as one value: a selected history command
    # may contain newlines, which command substitution would otherwise split.
    set -l result ($bin_path --query "$query" </dev/tty 2>/dev/tty | string collect)

    if test -n "$result"
        if string match -q "CMD:*" -- "$result"
            # It's a history command, replace command line
            set -l cmd (string replace "CMD:" "" -- "$result" | string collect)
            commandline -r -- "$cmd"
            commandline -f repaint
        else if string match -q "BRANCH:*" -- "$result"
            # It's a git branch, switch to it
            set -l branch (string replace "BRANCH:" "" -- "$result")
            # Pass the branch as an argument instead of building a shell string:
            # branch names may contain quotes and semicolons, which a quoted
            # `fish -c` string would execute.
            # Let git report why a switch failed (dirty tree, ambiguous remote
            # branch) instead of silently leaving the shell where it was.
            if not git switch --quiet -- "$branch" >/dev/null
                echo "fuzz.fish: could not switch to '$branch'" >&2
            end
            # Force repaint to update prompt
            commandline -f repaint
        else if string match -q "DIR:*" -- "$result"
            # It's a directory, cd into it
            set -l dir_path (string replace "DIR:" "" -- "$result" | string collect)
            cd "$dir_path"
            commandline -f repaint
        else if string match -q "HASH:*" -- "$result"
            # It's a commit hash, insert into command line
            set -l hash (string replace "HASH:" "" -- "$result" | string collect)
            commandline -i -- "$hash"
            commandline -f repaint
        else if string match -q "FILE:*" -- "$result"
            # It's a file, insert into command line
            set -l file_path (string replace "FILE:" "" -- "$result" | string collect)
            commandline -i -- "$file_path"
            commandline -f repaint
        end
    end
end

# Set up Ctrl+R key bindings for history/git/files unified search
function __fuzz_fish_key_bindings
    bind \cr fh
    if test "$fish_key_bindings" = fish_vi_key_bindings
        bind -M insert \cr fh
        bind -M default \cr fh
    end
end
__fuzz_fish_key_bindings
