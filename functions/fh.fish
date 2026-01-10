function fh --description 'Fish History viewer with context (TUI)'
    # Get binary path from environment variable
    set -l bin_path "$FUZZ_FISH_BIN_PATH"

    # Check if binary exists
    if test -z "$bin_path"; or not test -f "$bin_path"
        # Try to build if missing
        if functions -q _fuzz_fish_ensure_binary
            _fuzz_fish_ensure_binary
        else
            echo "❌ fuzz.fish: Binary not found. Please restart your shell." >&2
            return 1
        end
    end
    
    # Run the TUI binary
    # It will print the selected command to stdout on exit
    set -l cmd ($bin_path)

    if test -n "$cmd"
        # Insert into command line
        commandline -r -- "$cmd"
        commandline -f repaint
    end
end
