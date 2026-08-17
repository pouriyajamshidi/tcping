//go:build unix

package main

import (
	"os"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// isForegroundTerminal reports whether we are running attached to a terminal
// and are in the foreground process group (i.e., safe to read from stdin).
func isForegroundTerminal() bool {
	// Must be a terminal first
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}

	fgPgrp, err := unix.IoctlGetInt(int(os.Stdout.Fd()), unix.TIOCGPGRP)
	if err != nil {
		return false
	}

	myPgrp := unix.Getpgrp()

	return fgPgrp == myPgrp
}
