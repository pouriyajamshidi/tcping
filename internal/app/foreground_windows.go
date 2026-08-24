//go:build windows

package app

import (
	"os"

	"golang.org/x/term"
)

// IsForegroundTerminal reports whether we are running attached to a terminal.
// On Windows there is no POSIX-style job control, so we don't need to distinguish
// foreground vs background for the purpose of avoiding SIGTTIN-like behavior.
func IsForegroundTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
