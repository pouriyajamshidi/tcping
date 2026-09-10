# Showing only the failures needs a target that starts out working, so this
# tape splits the terminal with tmux: the probe on top, the UDP server that
# answers it below. Needs tmux on the PATH.
#
# The server is stopped halfway through the run. Nothing is printed while it
# is up, which is the flag's whole point, and the summary at the end still
# counts the probes that were never shown.
#
# The size is measured rather than guessed. A terminal cell is 0.667 by 1.318
# times the font size, so 100 columns and 42 rows come out as the width and
# height below, padding included. 100 columns is the failure that ends the
# uptime, the widest line here, with a few to spare, and the 42 rows are the
# probe's 30, the server's 11 and the divider between them. Both panes hold a
# couple of rows more than the run needs, so that a probe either way does not
# scroll the command that started it out of the picture.

Set Shell "bash"
Set FontSize 20
Set Padding 30
Set Width 1395
Set Height 1170

# Build the split off camera so the recording starts on two ready panes. The
# panes run "bash --norc" with the recorder's prompt exported into them, so
# they look like every other tape instead of like the recording machine. The
# window is split only after attaching, because a detached session is 80
# columns wide and the pane sizes would be scaled up from that.
Hide
Type "export PS1"  Enter
Type "tmux -f /dev/null new-session -d -s failures bash --norc"  Enter
Type "tmux set-option -t failures status off"  Enter
Type "clear; tmux attach -t failures"  Enter
Sleep 1s
Type "tmux split-window -v -l 11 bash --norc"  Enter
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

Type "tcping udp://127.0.0.1 9999 --failures-only"
Sleep 500ms  Enter

# The probes are being answered, so nothing is printed here at all.
Sleep 6s

# Stop the server, so the probes start failing.
Ctrl+B
Type "o"
Sleep 500ms
Ctrl+C
Sleep 500ms

Ctrl+B
Type "o"
Sleep 7s

Ctrl+C
Sleep 4s

Hide
Type "tmux kill-server"  Enter
