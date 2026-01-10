function fh --description 'Fish History viewer with context'
    # Get the plugin directory
    set -l plugin_dir (dirname (dirname (status -f)))
    set -l bin_path "$plugin_dir/bin/fhv"

    # Check if binary exists
    if not test -f "$bin_path"
        echo "❌ fuzz.fish: Binary not found. Please restart your shell or run: source ~/.config/fish/config.fish" >&2
        return 1
    end

    # Check if fzf is available
    if not type -q fzf
        echo "❌ fuzz.fish: fzf is required but not installed." >&2
        echo "   Install: brew install fzf (macOS) or apt install fzf (Linux)" >&2
        return 1
    end

    # Detect clipboard command
    set -l clip_cmd
    if type -q pbcopy
        set clip_cmd "pbcopy"
    else if type -q xclip
        set clip_cmd "xclip -selection clipboard"
    else if type -q xsel
        set clip_cmd "xsel --clipboard --input"
    else if type -q wl-copy
        set clip_cmd "wl-copy"
    else
        set clip_cmd "cat"  # Fallback: just print
    end

    # Run the history viewer with fzf
    set -l selected ($bin_path | fzf \
        --ansi \
        --height=50% \
        --reverse \
        --preview="$bin_path preview {1}" \
        --preview-window=right:50%:wrap \
        --header='CTRL-Y: Copy | ENTER: Execute | ESC: Cancel' \
        --bind="ctrl-y:execute-silent(echo -n {4..} | $clip_cmd)+abort" \
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
