package tcping

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/statistics"
)

// Sentinel errors for application control flow.
// These allow main.go to handle exits appropriately while keeping
// library code free of os.Exit calls.
var (
	// ErrUsageRequested indicates that usage help was requested
	ErrUsageRequested = errors.New("usage help requested")

	// ErrUpdateCheckFailed indicates update check encountered an error
	ErrUpdateCheckFailed = errors.New("update check failed")

	// ErrUpdateAvailable indicates an update is available
	ErrUpdateAvailable = errors.New("update available")

	// ErrDatabaseTableCreation indicates database table creation failed
	ErrDatabaseTableCreation = errors.New("database table creation failed")

	// ErrCriticalError indicates a critical error that should terminate the program
	ErrCriticalError = errors.New("critical error occurred")
)

// HandleExit processes application errors and exits with appropriate codes.
// This centralizes all exit logic that was previously scattered throughout
// the codebase using os.Exit calls.
func HandleExit(err error) {
	switch {
	case err == nil, errors.Is(err, ErrUpdateAvailable):
		os.Exit(0)

	case errors.Is(err, ErrUsageRequested),
		errors.Is(err, ErrUpdateCheckFailed),
		errors.Is(err, ErrDatabaseTableCreation),
		errors.Is(err, ErrCriticalError):
		os.Exit(1)

	default:
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// Result contains the main data structure for the TCPing program.
// It holds statistics and state about the ongoing pinging process.
type Result struct {
	Settings                  Settings                    // User-specified settings and configuration.
	StartTime                 time.Time                   // Start time of the TCPing operation.
	EndTime                   time.Time                   // End time of the TCPing operation.
	StartOfUptime             time.Time                   // Timestamp when the current uptime started.
	StartOfDowntime           time.Time                   // Timestamp when the current downtime started.
	LastSuccessfulProbe       time.Time                   // Timestamp of the last successful probe.
	LastUnsuccessfulProbe     time.Time                   // Timestamp of the last unsuccessful probe.
	Ticker                    *time.Ticker                // Ticker used to manage the time between probes.
	LongestUptime             statistics.LongestTime      // Data structure holding information about the longest uptime.
	LongestDowntime           statistics.LongestTime      // Data structure holding information about the longest downtime.
	Rtt                       []float32                   // List of RTT results for successful probes.
	HostnameChanges           []statistics.HostnameChange // List of hostname changes encountered.
	OngoingSuccessfulProbes   uint                        // Count of ongoing successful probes.
	OngoingUnsuccessfulProbes uint                        // Count of ongoing unsuccessful probes.
	TotalDowntime             time.Duration               // Total accumulated downtime.
	TotalUptime               time.Duration               // Total accumulated uptime.
	TotalSuccessfulProbes     uint                        // Total successful probes.
	TotalUnsuccessfulProbes   uint                        // Total unsuccessful probes.
	RetriedHostnameLookups    uint                        // Total number of retries for hostname resolution.
	RttResults                statistics.RttResult        // Struct holding the minimum, average, and maximum RTT values.
	DestWasDown               bool                        // Flag indicating if the destination was unreachable previously.
	DestIsIP                  bool                        // Flag indicating whether the destination is an IP address (not a hostname).
}

// Settings holds the configuration provided by the user for the TCPing operation.
type Settings struct {
	IP                       netip.Addr       // IP address to ping.
	Hostname                 string           // Hostname to resolve and ping.
	NetworkInterface         NetworkInterface // Network interface settings for the operation.
	RetryHostnameLookupAfter uint             // Number of failed requests before retrying to resolve the hostname.
	ProbesBeforeQuit         uint             // Number of probes before the program stops.
	Timeout                  time.Duration    // Timeout for each probe.
	IntervalBetweenProbes    time.Duration    // Time between consecutive probes.
	Port                     uint16           // Port number to connect to.
	UseIPv4                  bool             // Flag indicating whether to use IPv4 addresses.
	UseIPv6                  bool             // Flag indicating whether to use IPv6 addresses.
	NonInteractive           bool             // Flag the program will run in the background.
	ShouldRetryResolve       bool             // Flag indicating whether to retry resolving the hostname on failure.
	ShowFailuresOnly         bool             // Flag indicating whether to only show failed probes.
}

// NetworkInterface represents a network interface used for connecting to the target.
type NetworkInterface struct {
	RemoteAddr *net.TCPAddr // Remote address for the network interface.
	Dialer     net.Dialer   // Dialer used to make network connections.
	Use        bool         // Flag indicating whether to use this network interface.
}
