package statistics

import (
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/types"
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
	LastSuccessfulProbe       time.Time         // Timestamp of the last successful probe.
	LastUnsuccessfulProbe     time.Time         // Timestamp of the last unsuccessful probe.
	TotalDowntime             time.Duration     // Total accumulated downtime.
	TotalUptime               time.Duration     // Total accumulated uptime.
	StartOfUptime             time.Time         // Timestamp when the current uptime started.
	StartOfDowntime           time.Time         // Timestamp when the current downtime started.
	LongestUptime             types.LongestTime // Data structure holding information about the longest uptime.
	LongestDowntime           types.LongestTime // Data structure holding information about the longest downtime.
	HostnameChanges           []types.HostnameChange
	RetriedHostnameLookups    uint
	OngoingSuccessfulProbes   uint // Count of ongoing successful probes.
	OngoingUnsuccessfulProbes uint // Count of ongoing unsuccessful probes.
	LongestUp                 types.LongestTime
	LongestDown               types.LongestTime
	RTT                       []float32
	LatestRTT                 float32
	RTTResults                types.RttResult
	HostChanges               []types.HostnameChange
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
