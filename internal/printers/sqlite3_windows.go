//go:build windows

package printers

import "errors"

// NewDatabasePrinter reports that -db is unavailable here. The sqlite3
// support is not built on Windows, so the flag fails with a message instead
// of taking the program down.
func NewDatabasePrinter(_ string, _ uint16, _ string) (Printer, error) {
	return nil, errors.New("sqlite3 output is not supported on Windows")
}
