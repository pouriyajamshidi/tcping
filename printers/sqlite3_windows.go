//go:build windows

package printers

import (
	"errors"

	"github.com/pouriyajamshidi/tcping/v3/probe"
)

// NewSQLitePrinter reports that -sqlite is unavailable here. The sqlite3
// support is not built on Windows, so the flag fails with a message instead
// of taking the program down.
func NewSQLitePrinter(_ Config) (probe.Printer, error) {
	return nil, errors.New("sqlite3 output is not supported on Windows")
}
