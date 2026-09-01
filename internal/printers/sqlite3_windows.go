//go:build windows

package printers

import (
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

// DatabasePrinter represents a SQLite database connection for storing TCPing results.
type DatabasePrinter struct{}

// NewDatabasePrinter initializes a new sqlite3 Database instance, creates the data table, and returns a pointer to it.
// If any error occurs during database creation or table initialization, the function exits the program.
func NewDatabasePrinter(_ string, _ uint16, _ string) (*DatabasePrinter, error) {
	panic("sqlite3 support has been removed from Windows")
}

// PrintStart prints a message indicating that TCPing has started for the given hostname and port.
func (p *DatabasePrinter) PrintStart(_ *stats.Statistics) {}

// PrintNameResolutionDuration satisfies the "printer" interface but does nothing in this implementation
func (p *DatabasePrinter) PrintNameResolutionDuration(_ *stats.Statistics) {}

// PrintProbeSuccess satisfies the "printer" interface but does nothing in this implementation
func (p *DatabasePrinter) PrintProbeSuccess(_ *stats.Statistics) {}

// PrintProbeFailure satisfies the "printer" interface but does nothing in this implementation
func (p *DatabasePrinter) PrintProbeFailure(_ *stats.Statistics) {}

// PrintStatistics saves TCPing statistics to the database.
// If an error occurs while saving, it logs the error.
func (p *DatabasePrinter) PrintStatistics(_ *stats.Statistics) {}

// PrintRetryingToResolve prints a message indicating that the program is retrying to resolve the hostname.
func (p *DatabasePrinter) PrintRetryingToResolve(_ string) {}

// PrintDownTimeDuration satisfies the "printer" interface but does nothing in this implementation
func (p *DatabasePrinter) PrintDownTimeDuration(_ *stats.Statistics) {}

// PrintUpTimeDuration satisfies the "printer" interface but does nothing in this implementation
func (p *DatabasePrinter) PrintUpTimeDuration(_ *stats.Statistics) {}

// PrintError prints an error message to stderr and exits the program.
func (p *DatabasePrinter) PrintError(_ string, _ ...any) {}

// Shutdown sets the end time, prints statistics, calls Done() and exits the program.
func (p *DatabasePrinter) Shutdown(_ *stats.Statistics) {}

// Done closes the connection to the database
func (p *DatabasePrinter) Done() {}
