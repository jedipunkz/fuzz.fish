function ff --description 'Search files and directories with preview (TUI)'
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

    # Run the TUI binary with 'files' subcommand
    # It will print the selected file/dir to stdout on exit
    set -l result ($bin_path files)

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
