//go:build windows

package printers

import (
	"os"

	"golang.org/x/sys/windows"
)

// A Windows console only renders escape codes once virtual terminal
// processing is turned on. Windows 10 and later support it. Older ones do
// not, and there we turn color off instead of printing the codes raw.
//
// Package level variables are set before init runs, so colorEnabled already
// holds the answer for everything except this.
func init() {
	handle := windows.Handle(os.Stdout.Fd())

	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		// Not a console, e.g. output is redirected. Nothing to turn on.
		return
	}

	if err := windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING); err != nil {
		colorEnabled = false
	}
}
