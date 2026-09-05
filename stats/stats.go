// Package stats holds what a run has learned about its target so far, and
// formats those numbers for the printers. It is the one place the probers
// write to and the printers read from.
package stats

import (
	"fmt"
	"math"
	"net"
	"net/netip"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"
)

// HTTPInfo is what the most recent HTTP(S) probe learned about the response.
// Every field stays zero for TCP probes.
type HTTPInfo struct {
	StatusCode      int
	Status          string        // e.g. "200 OK"
	Proto           string        // e.g. "HTTP/1.1", "HTTP/2.0"
	TLSVersion      string        // e.g. "TLS 1.3". Empty for plain HTTP.
	TLSCipherSuite  string        // Empty for plain HTTP.
	CertExpiry      time.Time     // Zero for plain HTTP.
	ConnectDuration time.Duration // TCP connect only.
	TLSDuration     time.Duration // TLS handshake only.
	TimeToFirstByte time.Duration // Request sent until the first response byte.
}

// UDPInfo is what the most recent UDP probe learned. Every field stays zero
// for the other probe types.
type UDPInfo struct {
	Rejected    bool   // The target sent back an ICMP port unreachable, so something refused us.
	Echoed      bool   // The reply was exactly what we sent, so a tcping UDP server or an echo service answered.
	ProbeNumber uint64 // The number this probe sent as its payload. It counts up for the whole run, so it says which probe a reply belongs to and which one was lost.
	ReplySize   int    // Size of the reply in bytes. Zero when nothing answered.
}

type Statistics struct {
	Hostname                  string
	IP                        netip.Addr
	Port                      uint16
	Protocol                  config.Protocol
	LastProbeHadFailed        bool
	DestIsIP                  bool
	LocalAddr                 net.Addr
	StartTime                 time.Time
	EndTime                   time.Time
	CurrentUptime             time.Duration
	CurrentDowntime           time.Duration
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
	LongestUptime             LongestTime   // Longest uptime streak observed during the entire run.
	LongestDowntime           LongestTime   // Longest downtime streak observed during the entire run.
	HostnameChanges           []HostnameChange
	RetriedHostnameLookups    uint
	LatestRTT                 float32       // RTT of the most recent successful probe.
	RTTResults                RTTResult     // Running min/average/max/mdev RTT across the entire run.
	NameResolutionDuration    time.Duration // How long the most recent hostname resolution (initial or a retry) took. Meaningless (and zero) when DestIsIP.
	ResolvedThisProbe         bool          // True when ResolveEveryProbe just resolved successfully for this probe. Lets PrintProbeSuccess/PrintProbeFailure fold NameResolutionDuration into their own line instead of a separate one.
	HTTP                      HTTPInfo      // Details of the most recent HTTP(S) probe. Zero for TCP probes.
	UDP                       UDPInfo       // Details of the most recent UDP probe. Zero for the other probe types.
}

func NewStatistics(cfg config.Config) *Statistics {
	var localAddr net.Addr
	if cfg.NetworkInterface.Use {
		if localIP := cfg.NetworkInterface.LocalIPFor(cfg.IP); localIP != nil {
			localAddr = &net.TCPAddr{IP: localIP}
		}
	}

	return &Statistics{
		Hostname:               cfg.Hostname,
		IP:                     cfg.IP,
		Port:                   cfg.Port,
		DestIsIP:               cfg.TargetIsIP,
		LocalAddr:              localAddr,
		Protocol:               cfg.Protocol,
		NameResolutionDuration: cfg.NameResolutionDuration,
		HostnameChanges: []HostnameChange{{
			Addr:     cfg.IP,
			When:     time.Now(),
			Duration: cfg.NameResolutionDuration,
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
	// Round instead of truncating so this agrees with the uptime and
	// downtime totals, which durationToString also rounds. Truncating here
	// is what made a 6.6 second run report "00:00:06" next to "7 seconds".
	d := s.EndTime.Sub(s.StartTime).Round(time.Second)
	return time.Time{}.Add(d).Format(time.TimeOnly)
}

func (s *Statistics) ProtocolStr() string {
	return string(s.Protocol)
}

func (s *Statistics) RTTStr() string {
	return fmt.Sprintf("%.3f", s.LatestRTT)
}

// DurationToMilliseconds converts d to milliseconds as a float32, preserving
// sub-millisecond precision that Duration.Milliseconds() (which returns an
// int64) would drop. Exported because probers fill in LatestRTT with it.
func DurationToMilliseconds(d time.Duration) float32 {
	return float32(d.Nanoseconds()) / float32(time.Millisecond)
}

// millisecondsStr formats d with millisecond precision, for durations that
// are typically sub-second (hostname resolution, RTT) where durationToString's
// whole-second rounding would be useless.
func millisecondsStr(d time.Duration) string {
	return fmt.Sprintf("%.3f", DurationToMilliseconds(d))
}

// NameResolutionDurationStr formats NameResolutionDuration with
// millisecond precision, matching RTT's, since hostname resolution is
// typically sub-second.
func (s *Statistics) NameResolutionDurationStr() string {
	return millisecondsStr(s.NameResolutionDuration)
}

// IsHTTP reports whether this run probes over HTTP or HTTPS, which is what
// decides whether the HTTP details are worth printing at all.
func (s *Statistics) IsHTTP() bool {
	return s.Protocol == config.HTTP || s.Protocol == config.HTTPS
}

// HasHTTPResponse reports whether the last probe got an actual HTTP response.
// A probe that never reached the server has no status to show.
func (s *Statistics) HasHTTPResponse() bool {
	return s.HTTP.StatusCode != 0
}

// IsUDP reports whether this run probes over UDP, which is what decides
// whether the UDP details are worth printing at all.
func (s *Statistics) IsUDP() bool {
	return s.Protocol == config.UDP
}

// ProbeNumberStr is the number the most recent UDP probe sent as its payload.
func (s *Statistics) ProbeNumberStr() string {
	return fmt.Sprint(s.UDP.ProbeNumber)
}

func (s *Statistics) StatusCodeStr() string {
	return fmt.Sprint(s.HTTP.StatusCode)
}

func (s *Statistics) ConnectDurationStr() string {
	return millisecondsStr(s.HTTP.ConnectDuration)
}

func (s *Statistics) TLSDurationStr() string {
	return millisecondsStr(s.HTTP.TLSDuration)
}

func (s *Statistics) TimeToFirstByteStr() string {
	return millisecondsStr(s.HTTP.TimeToFirstByte)
}

func (s *Statistics) CertExpiryStr() string {
	if s.HTTP.CertExpiry.IsZero() {
		return ""
	}
	return s.HTTP.CertExpiry.Format(time.DateOnly)
}

// CertDaysRemaining returns how many days are left on the server certificate.
// It goes negative once the certificate has expired.
func (s *Statistics) CertDaysRemaining() int {
	if s.HTTP.CertExpiry.IsZero() {
		return 0
	}
	return int(time.Until(s.HTTP.CertExpiry).Hours() / 24)
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
	return durationToString(s.CurrentDowntime)
}

func (s *Statistics) UptimeDuration() string {
	return durationToString(s.CurrentUptime)
}

func (s *Statistics) LastSuccessfulProbeFormatted() string {
	return s.LastSuccessfulProbe.Format(time.DateTime)
}

func (s *Statistics) LastUnsuccessfulProbeFormatted() string {
	return s.LastUnsuccessfulProbe.Format(time.DateTime)
}

func (s *Statistics) TotalUptimeDuration() string {
	return durationToString(s.TotalUptime)
}

func (s *Statistics) TotalDowntimeDuration() string {
	return durationToString(s.TotalDowntime)
}

func (s *Statistics) LongestUptimeDuration() string {
	return durationToString(s.LongestUptime.Duration)
}

func (s *Statistics) LongestUptimeStartTime() string {
	return s.LongestUptime.Start.Format(time.DateTime)
}

func (s *Statistics) LongestUptimeEndTime() string {
	return s.LongestUptime.End.Format(time.DateTime)
}

func (s *Statistics) LongestDowntimeDuration() string {
	return durationToString(s.LongestDowntime.Duration)
}

func (s *Statistics) LongestDowntimeStartTime() string {
	return s.LongestDowntime.Start.Format(time.DateTime)
}

func (s *Statistics) LongestDowntimeEndTime() string {
	return s.LongestDowntime.End.Format(time.DateTime)
}

// durationToString creates a human-readable string for a given duration
func durationToString(d time.Duration) string {
	if d == 0 {
		return "0 seconds"
	}

	// Anything shorter than a second keeps its real value instead of being
	// rounded to a whole second, which would report a gap that did happen
	// as "0 seconds". Tenths keep it in the same unit as the rest of the
	// summary, and milliseconds take over below the point where a tenth
	// would itself round down to zero.
	if d < 50*time.Millisecond {
		return millisecondsStr(d) + " ms"
	}

	if d < time.Second {
		return fmt.Sprintf("%.1f seconds", d.Seconds())
	}

	// Round up front so the hour, minute and second parts are taken from the
	// same value. Rounding only the seconds at the end used to let 1m59.7s
	// print as "1 minute 60 seconds".
	d = d.Round(time.Second)

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

	case seconds == 1:
		return "1 second"

	default:
		return fmt.Sprintf("%.0f seconds", seconds)
	}
}

// RTTResult holds running statistics for round-trip times (RTT) results.
// Its zero value is valid and represents "no samples yet". Callers should
// use Statistics.TotalSuccessfulProbes to tell whether any samples have
// been recorded.
type RTTResult struct {
	Min     float32 // Minimum RTT value.
	Max     float32 // Maximum RTT value.
	Average float32 // Average RTT value.
	// Mdev is how far the samples sit from the average, the same number
	// ping ends its summary with. A small one means every probe took about
	// as long as the last, a large one means the latency is jumping around.
	Mdev float32

	// Running sum of (sample - average) squared, which Mdev is the root of.
	// Accumulated as the samples come in so none of them has to be kept.
	squaredDeltas float32
}

// Update folds a new RTT sample into the running min, max, average and
// mdev. sampleCount must be the total number of successful probes observed
// so far, including this one (e.g. Statistics.TotalSuccessfulProbes).
func (r *RTTResult) Update(rttMs float32, sampleCount uint) {
	if sampleCount <= 1 {
		r.Min = rttMs
		r.Max = rttMs
	} else {
		r.Min = min(r.Min, rttMs)
		r.Max = max(r.Max, rttMs)
	}

	// Running average: avg_n = avg_(n-1) + (x_n - avg_(n-1)) / n
	// The distance from the average is taken before and after that step,
	// which is what keeps the sum from losing precision on a long run.
	delta := rttMs - r.Average
	r.Average += delta / float32(sampleCount)
	r.squaredDeltas += delta * (rttMs - r.Average)

	r.Mdev = float32(math.Sqrt(float64(r.squaredDeltas / float32(sampleCount))))
}

// LongestTime holds information about the longest period of uptime or downtime.
type LongestTime struct {
	Start    time.Time     // Start time of the longest period.
	End      time.Time     // End time of the longest period.
	Duration time.Duration // Duration of the longest period.
}

// newLongestTime creates and returns a LongestTime instance with the provided start time and duration.
func newLongestTime(startTime time.Time, duration time.Duration) LongestTime {
	return LongestTime{
		Start:    startTime,
		End:      startTime.Add(duration),
		Duration: duration,
	}
}

// HostnameChange represents changes in the IP address associated with a hostname.
type HostnameChange struct {
	Addr     netip.Addr    // New IP address associated with the hostname.
	When     time.Time     // Timestamp of when the change occurred.
	Duration time.Duration // How long the resolution that produced Addr took.
}

func (h *HostnameChange) WhenFormatted() string {
	return h.When.Format(time.DateTime)
}

func (h *HostnameChange) DurationStr() string {
	return millisecondsStr(h.Duration)
}

// SetLongestDuration updates the longest uptime or downtime based on the given type.
func SetLongestDuration(start time.Time, duration time.Duration, longest *LongestTime) {
	if start.IsZero() || duration == 0 {
		return
	}

	newLongest := newLongestTime(start, duration)

	if longest.End.IsZero() || newLongest.Duration >= longest.Duration {
		*longest = newLongest
	}
}
