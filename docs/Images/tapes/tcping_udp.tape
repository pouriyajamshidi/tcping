# A UDP probe only gets an answer if something is listening on the other end,
# so this tape splits the terminal with tmux: the probe on top, the server
# below. Needs tmux on the PATH.
#
# The size is measured rather than guessed. A terminal cell is 0.667 by 1.318
# times the font size, so 92 columns and 34 rows come out as the width and
# height below, padding included. 92 columns is the summary's widest line with
# a few to spare, and the 34 rows are the probe's 25, the server's 8 and the
# divider between them. Sizing the window to the output rather than to the
# screen is what keeps the text readable: GitHub scales the GIF down to the
# width of the README, so every column we do not need, and every pixel of
# padding around them, costs font size.

Set Shell "bash"
Set FontSize 20
Set Padding 30
Set Width 1290
Set Height 960

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
Type "tmux split-window -v -l 8 bash --norc"  Enter
Sleep 1s
Type "clear"  Enter
Ctrl+B
Type "o"
Type "clear"  Enter
Ctrl+B
Type "o"
Sleep 1s
Show

# The split left the focus on the new pane, which is the bottom one.
Type "tcping --udp-server 127.0.0.1 9999"
Sleep 500ms  Enter
Sleep 2s

# Ctrl+B o moves to the top pane.
Ctrl+B
Type "o"
Sleep 1s

Type "tcping udp://127.0.0.1 9999 -c 5"
Sleep 500ms  Enter

Sleep 12s

Hide
Type "tmux kill-server"  Enter
