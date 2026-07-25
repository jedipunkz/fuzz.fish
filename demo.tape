# fuzz.fish demo
#
# Build the binary first, then record from the repository root:
#   make build
#   vhs demo.tape

Output assets/fuzz.gif

Require fish

Set Shell "fish"
Set FontSize 18
Set Width 1400
Set Height 800
Set Padding 20
Set TypingSpeed 180ms

# Bind Ctrl+R to the locally built ./fuzz instead of the installed binary,
# so the demo always shows the current source. Not recorded.
Hide
Sleep 1s
Type "function fuzz_demo; ./fuzz </dev/tty 2>/dev/tty >/dev/null; commandline -f repaint; end"
Enter
Sleep 1s
Type "bind ctrl-r fuzz_demo"
Enter
Sleep 1s
Type "clear"
Enter
Sleep 2s
Show

Sleep 1s

# History search (Ctrl+R)
Ctrl+r
Sleep 1s
Type "git pull"
Sleep 1s
Up@500ms 3
Sleep 1s
Down@500ms 2
Sleep 1s

# Git branch search (Ctrl+G), then move the cursor through the candidates
Ctrl+g
Sleep 1s
Type "feat/"
Sleep 1s
Up@500ms 3
Sleep 1s
Down@500ms 2
Sleep 1s

# File search (Ctrl+S)
Ctrl+S
Sleep 1s
Type "tui.go"
Sleep 1s
Up@500ms 3
Sleep 1s
Down@500ms 2
Sleep 1s

# Quit
Escape
Sleep 2s
