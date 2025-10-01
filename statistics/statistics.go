package statistics

import (
	"fmt"
	"math"
	"net"
	"net/netip"
	"slices"
	"time"
)

type protocol string

const (
	TCP   protocol = "TCP"
	UDP   protocol = "UDP"
	HTTP  protocol = "HTTP"
	HTTPS protocol = "HTTPS"
	ICMP  protocol = "ICMP"
)

type Statistics struct {
	IP                        netip.Addr
	Port                      uint16
	Protocol                  protocol
	Hostname                  string
	DestWasDown               bool
	DestIsIP                  bool
	LocalAddr                 net.Addr
	StartTime                 time.Time
	EndTime                   time.Time
	UpTime                    time.Duration
	DownTime                  time.Duration
	Successful                int
	Failed                    int
	TotalSuccessfulProbes     uint
	TotalUnsuccessfulProbes   uint
	LastSuccessfulProbe       time.Time     // Timestamp of the last successful probe.
	LastUnsuccessfulProbe     time.Time     // Timestamp of the last unsuccessful probe.
	TotalDowntime             time.Duration // Total accumulated downtime.
	TotalUptime               time.Duration // Total accumulated uptime.
	StartOfUptime             time.Time     // Timestamp when the current uptime started.
	StartOfDowntime           time.Time     // Timestamp when the current downtime started.
	LongestUptime             LongestTime   // Data structure holding information about the longest uptime.
	LongestDowntime           LongestTime   // Data structure holding information about the longest downtime.
	HostnameChanges           []HostnameChange
	RetriedHostnameLookups    uint
	OngoingSuccessfulProbes   uint // Count of ongoing successful probes.
	OngoingUnsuccessfulProbes uint // Count of ongoing unsuccessful probes.
	LongestUp                 LongestTime
	LongestDown               LongestTime
	RTT                       []float32
	LatestRTT                 float32
	RTTResults                RttResult
	HostChanges               []HostnameChange
	HasResults                bool
	WithTimestamp             bool
	WithSourceAddress         bool
}

func (s *Statistics) IPStr() string {
	return s.IP.String()
}

func (s *Statistics) PortStr() string {
	return fmt.Sprint(s.Port)
}

func (s *Statistics) SourceAddr() string {
	return s.LocalAddr.String()
}

func (s *Statistics) StartTimeFormatted() string {
	return s.StartTime.Format(time.DateTime)
}

func (s *Statistics) EndTimeFormatted() string {
	return s.EndTime.Format(time.DateTime)
}

func (s *Statistics) ProtocolStr() string {
	return string(s.Protocol)
}

func (s *Statistics) RTTStr() string {
	return fmt.Sprintf("%.3f", s.LatestRTT)
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

// RttResult holds statistics for round-trip times (RTT) results.
type RttResult struct {
	Min        float32 // Minimum RTT value.
	Max        float32 // Maximum RTT value.
	Average    float32 // Average RTT value.
	HasResults bool    // Flag indicating whether RTT results are available.
}

// HostnameChange represents a change in the IP address associated with a hostname.
type HostnameChange struct {
	Addr netip.Addr `json:"addr,omitempty"` // New IP address associated with the hostname.
	When time.Time  `json:"when,omitempty"` // Timestamp of when the change occurred.
}

// calcMinAvgMaxRttTime calculates min, avg and max RTT values
func CalcMinAvgMaxRttTime(timeArr []float32) RttResult {
	var result RttResult

	arrLen := len(timeArr)
	if arrLen == 0 {
		return result
	}

	var sum float32

	for _, t := range timeArr {
		sum += t
	}

	result.Min = slices.Min(timeArr)
	result.Max = slices.Max(timeArr)
	result.Average = sum / float32(arrLen)
	result.HasResults = true

	return result
}

// SetLongestDuration updates the longest uptime or downtime based on the given type.
func SetLongestDuration(start time.Time, duration time.Duration, longest *LongestTime) {
	if start.IsZero() || duration == 0 {
		return
	}

	newLongest := NewLongestTime(start, duration)

	if longest.End.IsZero() || newLongest.Duration >= longest.Duration {
		*longest = newLongest
	}
}

// DurationToString creates a human-readable string for a given duration
func DurationToString(duration time.Duration) string {
	hours := math.Floor(duration.Hours())
	if hours > 0 {
		duration -= time.Duration(hours * float64(time.Hour))
	}

	minutes := math.Floor(duration.Minutes())
	if minutes > 0 {
		duration -= time.Duration(minutes * float64(time.Minute))
	}

	seconds := duration.Seconds()

	switch {
	// Hours
	case hours >= 2:
		return fmt.Sprintf("%.0f hours %.0f minutes %.0f seconds", hours, minutes, seconds)
	case hours == 1 && minutes == 0 && seconds == 0:
		return fmt.Sprintf("%.0f hour", hours)
	case hours == 1:
		return fmt.Sprintf("%.0f hour %.0f minutes %.0f seconds", hours, minutes, seconds)

	// Minutes
	case minutes >= 2:
		return fmt.Sprintf("%.0f minutes %.0f seconds", minutes, seconds)
	case minutes == 1 && seconds == 0:
		return fmt.Sprintf("%.0f minute", minutes)
	case minutes == 1:
		return fmt.Sprintf("%.0f minute %.0f seconds", minutes, seconds)

	// Seconds
	case seconds == 0 || seconds == 1 || seconds >= 1 && seconds < 1.1:
		return fmt.Sprintf("%.0f second", seconds)
	case seconds < 1:
		return fmt.Sprintf("%.1f seconds", seconds)

	default:
		return fmt.Sprintf("%.0f seconds", seconds)
	}
}

// NanoToMillisecond returns an amount of milliseconds from nanoseconds.
// Using duration.Milliseconds() is not an option, because it drops
// decimal points, returning an int.
func NanoToMillisecond(nano int64) float32 {
	return float32(nano) / float32(time.Millisecond)
}

// SecondsToDuration returns the corresponding duration from seconds expressed with a float.
func SecondsToDuration(seconds float64) time.Duration {
	return time.Duration(1000*seconds) * time.Millisecond
}
