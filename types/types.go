// Package types hosts the types and interfaces used throughout the program
package types

import (
	"net"
	"net/netip"
	"time"
)

// Tcping contains the main data structure for the TCPing program.
// It holds statistics and state about the ongoing pinging process.
type Tcping struct {
	Options                   Options          // User-specified settings and configuration.
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
	NonInteractive           bool             // Flag the program will run in the background.
	ShouldRetryResolve       bool             // Flag indicating whether to retry resolving the hostname on failure.
	ShowFailuresOnly         bool             // Flag indicating whether to only show failed probes.
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
