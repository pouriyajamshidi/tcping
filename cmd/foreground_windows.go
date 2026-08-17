//go:build windows

package main

import (
	"os"

	"golang.org/x/term"
)

// isForegroundTerminal reports whether we are running attached to a terminal.
// On Windows there is no POSIX-style job control, so we don't need to distinguish
// foreground vs background for the purpose of avoiding SIGTTIN-like behavior.
func isForegroundTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
