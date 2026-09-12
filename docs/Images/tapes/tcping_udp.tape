# A UDP probe only gets an answer if something is listening on the other end,
# so this tape splits the terminal with tmux: the probe on top, the server
# below. Needs tmux on the PATH.
#
# The size is measured rather than guessed. A terminal cell is 0.667 by 1.318
# times the font size, so 89 columns and 33 rows come out as the width and
# height below, padding included. 89 columns is the summary's widest line, the
# one carrying two timestamps, with one to spare, and the 33 rows are the
# probe's 24, the server's 8 and the divider between them, none of them spare.
# The server's 8 are its command line, its 6 lines of output and the row the
# cursor waits on: give it 7 and the command scrolls off the top.
#
# Only the column count is worth arguing about. GitHub scales the GIF down to
# the width of the README, so the text lands at the README width divided by the
# columns: every column we do not need, and every pixel of padding around them,
# costs font size. Rows and the font size below do not, since a taller window
# or a bigger font scales right back down.

Set Shell "bash"
Set FontSize 20
Set Padding 30
Set Width 1250
Set Height 930

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
