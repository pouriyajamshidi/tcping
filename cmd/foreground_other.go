//go:build !unix && !windows

package main

import (
	"os"

	"golang.org/x/term"
)

func isForegroundTerminal() bool {
	// Conservative default: only enable interactive stdin if we have a terminal.
	// These platforms don't have the same SIGTTIN job-control semantics as POSIX.
	return term.IsTerminal(int(os.Stdout.Fd()))
}
