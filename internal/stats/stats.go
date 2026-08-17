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
	IP                        netip.Addr
	Port                      uint16
	Protocol                  consts.Protocol
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
	RTTResults                RTTResult
	HostChanges               []HostnameChange
	HasResults                bool
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
		StartTime: time.Now(),
	}
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
