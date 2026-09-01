package stats

import (
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/consts"
	"github.com/pouriyajamshidi/tcping/v3/internal/nic"
)

type Config interface {
	GetHostname() string
	GetIP() netip.Addr
	GetPort() uint16
	GetProtocol() consts.Protocol
	GetUseIPv4() bool
	GetUseIPv6() bool
	GetTimeout() string
	GetProbesBeforeQuit() uint
	GetTargetIsIP() bool
	GetIntervalBetweenProbes() string
	GetShowFailuresOnly() bool
	GetShouldRetryResolve() bool
	GetRetryResolveAfterNFailures() uint
	GetNetworkInterface() nic.NetworkInterface
	GetWithTimestamp() bool
	GetWithSourceAddress() bool
}

type Statistics struct {
	Hostname                  string
	IP                        netip.Addr
	Port                      uint16
	Protocol                  consts.Protocol
	LastProbeHadFailed        bool
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
	OngoingSuccessfulProbes   uint          // Count of ongoing successful probes.
	OngoingUnsuccessfulProbes uint          // Count of ongoing unsuccessful probes.
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
	LongestUp                 LongestTime
	LongestDown               LongestTime
	RTT                       []float32
	LatestRTT                 float32
	RTTResults                RTTResult
	HostChanges               []HostnameChange
	WithTimestamp             bool
	WithSourceAddress         bool
}

func NewStatistics(cfg Config) *Statistics {
	return &Statistics{
		Hostname:          cfg.GetHostname(),
		IP:                cfg.GetIP(),
		Port:              cfg.GetPort(),
		DestIsIP:          cfg.GetTargetIsIP(),
		LocalAddr:         cfg.GetNetworkInterface().Dialer.LocalAddr,
		WithTimestamp:     cfg.GetWithTimestamp(),
		WithSourceAddress: cfg.GetWithSourceAddress(),
		Protocol:          consts.TCP,
		RTTResults:        RTTResult{HasResults: false},
		LongestUptime:     LongestTime{},
		LongestDowntime:   LongestTime{},
		HostnameChanges: []HostnameChange{{
			Addr: cfg.GetIP(),
			When: time.Now(),
		}},
	}
}

func (s *Statistics) IPStr() string {
	return s.IP.String()
}

func (s *Statistics) PortStr() string {
	return fmt.Sprint(s.Port)
}

func (s *Statistics) SourceAddr() string {
	// in case probe failed and -I flag
	// was not used
	if s.LocalAddr == nil {
		return ""
	}

	return s.LocalAddr.String()
}

func (s *Statistics) CurrentTimestamp() string {
	return time.Now().Format(time.DateTime)
}

func (s *Statistics) StartTimeFormatted() string {
	return s.StartTime.Format(time.DateTime)
}

func (s *Statistics) EndTimeFormatted() string {
	return s.EndTime.Format(time.DateTime)
}

func (s *Statistics) RuntimeDuration() string {
	return time.Time{}.Add(s.TotalDowntime + s.TotalUptime).Format(time.TimeOnly)
}

func (s *Statistics) ProtocolStr() string {
	return string(s.Protocol)
}

func (s *Statistics) RTTStr() string {
	return fmt.Sprintf("%.3f", s.LatestRTT)
}

func (s *Statistics) TotalProbes() uint {
	return s.TotalSuccessfulProbes + s.TotalUnsuccessfulProbes
}

func (s *Statistics) PacketLoss() float32 {
	var packetLoss float32
	if s.TotalProbes() > 0 {
		packetLoss = float32(s.TotalUnsuccessfulProbes) / float32(s.TotalProbes()) * 100
	}

	return packetLoss
}

func (s *Statistics) DowntimeDuration() string {
	return DurationToString(s.DownTime)
}

func (s *Statistics) LastSuccessfulProbeFormatted() string {
	return s.LastSuccessfulProbe.Format(time.DateTime)
}

func (s *Statistics) LastUnsuccessfulProbeFormatted() string {
	return s.LastUnsuccessfulProbe.Format(time.DateTime)
}

func (s *Statistics) TotalUptimeDuration() string {
	return DurationToString(s.TotalUptime)
}

func (s *Statistics) TotalDowntimeDuration() string {
	return DurationToString(s.TotalDowntime)
}

func (s *Statistics) LongestUptimeDuration() string {
	return DurationToString(s.LongestUp.Duration)
}

func (s *Statistics) LongestUptimeStartTime() string {
	return s.LongestUp.Start.Format(time.DateTime)
}

func (s *Statistics) LongestUptimeEndTime() string {
	return s.LongestUp.End.Format(time.DateTime)
}

func (s *Statistics) LongestDowntimeDuration() string {
	return DurationToString(s.LongestDown.Duration)
}

func (s *Statistics) LongestDowntimeStartTime() string {
	return s.LongestDown.Start.Format(time.DateTime)
}

func (s *Statistics) LongestDowntimeEndTime() string {
	return s.LongestDown.End.Format(time.DateTime)
}

// DurationToString creates a human-readable string for a given duration
// TODO: unexport this when all printers are using the helper methods
func DurationToString(d time.Duration) string {
	hours := d / time.Hour
	d %= time.Hour

	minutes := d / time.Minute
	d %= time.Minute

	seconds := d.Seconds()

	switch {
	case hours >= 2:
		return fmt.Sprintf("%d hours %d minutes %.0f seconds", hours, minutes, seconds)

	case hours == 1:
		if minutes == 0 && seconds == 0 {
			return "1 hour"
		}
		return fmt.Sprintf("1 hour %d minutes %.0f seconds", minutes, seconds)

	case minutes >= 2:
		return fmt.Sprintf("%d minutes %.0f seconds", minutes, seconds)

	case minutes == 1:
		if seconds == 0 {
			return "1 minute"
		}
		return fmt.Sprintf("1 minute %.0f seconds", seconds)

	case seconds == 0:
		return "0 seconds"

	case seconds < 1.1:
		return "1 second"

	case seconds < 2:
		return fmt.Sprintf("%.1f seconds", seconds)

	default:
		return fmt.Sprintf("%.0f seconds", seconds)
	}
}

// RTTResult holds statistics for round-trip times (RTT) results.
type RTTResult struct {
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

// HostnameChange represents changes in the IP address associated with a hostname.
type HostnameChange struct {
	Addr netip.Addr // New IP address associated with the hostname.
	When time.Time  // Timestamp of when the change occurred.
}

func (h *HostnameChange) WhenFormatted() string {
	return h.When.Format(time.DateTime)
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
