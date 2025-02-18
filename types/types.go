// Package types hosts the types and interfaces used throughout the program
package types

import (
	"net"
	"net/netip"
	"time"
)

// Printer defines a set of methods that any printer implementation must provide.
// Printers are responsible for outputting information, but should not modify data or perform calculations.
type Printer interface {
	// PrintStart prints the first message to indicate the target's address and port.
	// This message is printed only once, at the very beginning.
	PrintStart(hostname string, port uint16)

	// PrintProbeSuccess should print a message after each successful probe.
	// hostname could be empty, meaning it's pinging an address.
	// streak is the number of successful consecuti`ve probes.
	PrintProbeSuccess(startTime time.Time, sourceAddr string, userInput Options, streak uint, rtt string)

	// PrintProbeFail should print a message after each failed probe.
	// hostname could be empty, meaning it's pinging an address.
	// streak is the number of successful consecutive probes.
	PrintProbeFail(startTime time.Time, userInput Options, streak uint)

	// PrintRetryingToResolve should print a message with the hostname
	// it is trying to resolve an IP for.
	//
	// This is only being printed when the -r flag is applied.
	PrintRetryingToResolve(hostname string)

	// PrintTotalDownTime should print a downtime duration.
	//
	// This is being called when host was unavailable for some time
	// but the latest probe was successful (became available).
	PrintTotalDownTime(downtime time.Duration)

	// PrintStatistics should print a message with
	// helpful statistics information.
	//
	// This is being called on exit and when user hits "Enter".
	PrintStatistics(s Tcping)

	// PrintError should print an error message.
	// Printer should also apply \n to the given string, if needed.
	PrintError(format string, args ...any)
}

// Prober represents an object that can probe a target
type Prober interface {
	Printer
	Ping()
}

// Tcping contains the main data structure for the TCPing program.
// It holds statistics and state about the ongoing pinging process.
type Tcping struct {
	Printer                                    // Printer is an embedded interface for outputting information and data.
	StartTime                 time.Time        // Start time of the TCPing operation.
	EndTime                   time.Time        // End time of the TCPing operation.
	StartOfUptime             time.Time        // Timestamp when the current uptime started.
	StartOfDowntime           time.Time        // Timestamp when the current downtime started.
	LastSuccessfulProbe       time.Time        // Timestamp of the last successful probe.
	LastUnsuccessfulProbe     time.Time        // Timestamp of the last unsuccessful probe.
	Ticker                    *time.Ticker     // Ticker used to manage the time between probes.
	LongestUptime             LongestTime      // Data structure holding information about the longest uptime.
	LongestDowntime           LongestTime      // Data structure holding information about the longest downtime.
	Rtt                       []float32        // List of RTT results for successful probes.
	HostnameChanges           []HostnameChange // List of hostname changes encountered.
	Options                   Options          // User-specified settings and configuration.
	OngoingSuccessfulProbes   uint             // Count of ongoing successful probes.
	OngoingUnsuccessfulProbes uint             // Count of ongoing unsuccessful probes.
	TotalDowntime             time.Duration    // Total accumulated downtime.
	TotalUptime               time.Duration    // Total accumulated uptime.
	TotalSuccessfulProbes     uint             // Total successful probes.
	TotalUnsuccessfulProbes   uint             // Total unsuccessful probes.
	RetriedHostnameLookups    uint             // Total number of retries for hostname resolution.
	RttResults                RttResult        // Struct holding the minimum, average, and maximum RTT values.
	DestWasDown               bool             // Flag indicating if the destination was unreachable previously.
	DestIsIP                  bool             // Flag indicating whether the destination is an IP address (not a hostname).
}

// Options holds the configuration provided by the user for the TCPing operation.
type Options struct {
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
	ShouldRetryResolve       bool             // Flag indicating whether to retry resolving the hostname on failure.
}

// RttResult holds statistics for round-trip times (RTT) results.
type RttResult struct {
	Min        float32 // Minimum RTT value.
	Max        float32 // Maximum RTT value.
	Average    float32 // Average RTT value.
	HasResults bool    // Flag indicating whether RTT results are available.
}

// LongestTime holds information about the longest period of uptime or downtime.
type LongestTime struct {
	Start    time.Time     // Start time of the longest period.
	End      time.Time     // End time of the longest period.
	Duration time.Duration // Duration of the longest period.
}

// NewLongestTime creates and returns a LongestTime instance with the provided start time and duration.
func NewLongestTime(startTime time.Time, duration time.Duration) LongestTime {
	return LongestTime{
		Start:    startTime,
		End:      startTime.Add(duration),
		Duration: duration,
	}
}

// NetworkInterface represents a network interface used for connecting to the target.
type NetworkInterface struct {
	RemoteAddr *net.TCPAddr // Remote address for the network interface.
	Dialer     net.Dialer   // Dialer used to make network connections.
	Use        bool         // Flag indicating whether to use this network interface.
}

// HostnameChange represents a change in the IP address associated with a hostname.
type HostnameChange struct {
	Addr netip.Addr `json:"addr,omitempty"` // New IP address associated with the hostname.
	When time.Time  `json:"when,omitempty"` // Timestamp of when the change occurred.
}
