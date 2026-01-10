function fh --description 'Fish History viewer with context'
    # Get the plugin directory
    set -l plugin_dir (dirname (dirname (status -f)))
    set -l bin_path "$plugin_dir/bin/fhv"

    # Build the Go binary if it doesn't exist or is outdated
    if not test -f "$bin_path"; or test "$plugin_dir/cmd/fhv/main.go" -nt "$bin_path"
        echo "Building fhv..." >&2
        mkdir -p "$plugin_dir/bin"
        if not go build -o "$bin_path" "$plugin_dir/cmd/fhv"
            echo "Error: Failed to build fhv. Make sure Go is installed." >&2
            return 1
        end
    end

    # Check if fzf is available
    if not type -q fzf
        echo "Error: fzf is required but not installed." >&2
        echo "Install it with: brew install fzf (macOS) or apt install fzf (Linux)" >&2
        return 1
    end

    # Run the history viewer with fzf
    set -l selected ($bin_path | fzf \
        --ansi \
        --height=50% \
        --reverse \
        --preview="$bin_path preview {1}" \
        --preview-window=right:50%:wrap \
        --header='CTRL-Y: Copy | ENTER: Execute | ESC: Cancel' \
        --bind='ctrl-y:execute-silent(echo -n {4..} | pbcopy)+abort' \
        --delimiter='\t' \
        --with-nth=2.. \
        --layout=reverse \
        --border \
        --info=inline \
        --prompt='History > ')

    if test -n "$selected"
        # Extract the command (4th field onward)
        set -l cmd (echo "$selected" | cut -f4-)

        # Insert into command line
        commandline -r -- "$cmd"
        commandline -f repaint
    end
end
