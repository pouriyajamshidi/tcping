# A UDP probe only gets an answer if something is listening on the other end,
# so this tape splits the terminal with tmux: the server on the right, the
# probe on the left. Needs tmux on the PATH.

Set Shell "bash"
Set FontSize 16
Set Width 1800
Set Height 660

# Build the split off camera so the recording starts on two ready panes. The
# panes run "bash --norc" with the recorder's prompt exported into them, so
# they look like every other tape instead of like the recording machine. The
# window is split only after attaching, because a detached session is 80
# columns wide and the pane sizes would be scaled up from that.
Hide
Type "export PS1"  Enter
Type "tmux -f /dev/null new-session -d -s udp bash --norc"  Enter
Type "tmux set-option -t udp status off"  Enter
Type "clear; tmux attach -t udp"  Enter
Sleep 1s
Type "tmux split-window -h -l 62 bash --norc"  Enter
Sleep 1s
Type "clear"  Enter
Ctrl+B
Type "o"
Type "clear"  Enter
Ctrl+B
Type "o"
Sleep 1s
Show

# The split left the focus on the new pane, which is the right one.
Type "tcping --udp-server 127.0.0.1 9999"
Sleep 500ms  Enter
Sleep 2s

# Ctrl+B o moves to the left pane.
Ctrl+B
Type "o"
Sleep 1s

Type "tcping udp://127.0.0.1 9999 -c 5"
Sleep 500ms  Enter

Sleep 12s

Hide
Type "tmux kill-server"  Enter
