# Streaming the JSON output somewhere else needs both ends, so this tape
# splits the terminal with tmux: the probe on top, the collector that receives
# its events below. Needs tmux on the PATH.
#
# The probe's pane is the small one on purpose. Nothing is printed there once
# the events are going to a URL, and that is the point being made: the run
# shows up in the other pane.
#
# The size is measured rather than guessed. A terminal cell is 0.667 by 1.318
# times the font size, so 87 columns and 33 rows come out as the width and
# height below, padding included. 87 columns is the probe's command line with
# a couple to spare, which also leaves room for an event carrying a full
# length IPv6 address, and the 33 rows are the probe's 4, the collector's 28 and
# the divider between them. 28 rows is one whole statistics event and a little
# of the probe before it. Sizing the window to the output rather than to the
# screen is what keeps the text readable: GitHub scales the GIF down to the
# width of the README, so every column we do not need, and every pixel of
# padding around them, costs font size.

Set Shell "bash"
Set FontSize 20
Set Padding 30
Set Width 1220
Set Height 930

# Build the split off camera so the recording starts on two ready panes. The
# panes run "bash --norc" with the recorder's prompt exported into them, so
# they look like every other tape instead of like the recording machine. The
# window is split only after attaching, because a detached session is 80
# columns wide and the pane sizes would be scaled up from that.
Hide
Type "export PS1"  Enter
Type "tmux -f /dev/null new-session -d -s json bash --norc"  Enter
Type "tmux set-option -t json status off"  Enter
Type "clear; tmux attach -t json"  Enter
Sleep 1s
Type "tmux split-window -v -l 28 bash --norc"  Enter
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
Type "tcping --json-server 127.0.0.1 8000"
Sleep 500ms  Enter
Sleep 2s

# Ctrl+B o moves to the top pane.
Ctrl+B
Type "o"
Sleep 1s

Type "tcping example.com 80 --pretty --source-label web1 --json-url http://127.0.0.1:8000"
Sleep 500ms  Enter

# Long enough for a few probes to arrive, then the statistics event that the
# probe sends on its way out.
Sleep 8s
Ctrl+C
Sleep 4s

Hide
Type "tmux kill-server"  Enter
