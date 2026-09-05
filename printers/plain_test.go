package printers

import (
	"net"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"

	"github.com/pouriyajamshidi/tcping/v3/stats"
)

// plainTestStats is a probe against a hostname that succeeded in 3.5 ms.
// Tests change only the fields they care about.
func plainTestStats() *stats.Statistics {
	return &stats.Statistics{
		Hostname:                  "example.com",
		IP:                        netip.MustParseAddr("93.184.216.34"),
		Port:                      443,
		Protocol:                  config.TCP,
		LatestRTT:                 3.5,
		OngoingSuccessfulProbes:   2,
		OngoingUnsuccessfulProbes: 5,
		NameResolutionDuration:    12 * time.Millisecond,
		StartTime:                 time.Now(),
	}
}

// wantLines checks the pieces a test names are all in the output.
func wantLines(t *testing.T, out string, want ...string) {
	t.Helper()

	for _, line := range want {
		if !strings.Contains(out, line) {
			t.Errorf("output = %q, want it to contain %q", out, line)
		}
	}
}

// wantNoLines checks the pieces a test names are absent.
func wantNoLines(t *testing.T, out string, unwanted ...string) {
	t.Helper()

	for _, line := range unwanted {
		if strings.Contains(out, line) {
			t.Errorf("output = %q, want it to leave out %q", out, line)
		}
	}
}

func TestPlainPrintStart(t *testing.T) {
	t.Run("hostname target reports its resolution time", func(t *testing.T) {
		out := captureStdout(t, func() {
			NewPlainPrinter(Config{}).PrintStart(plainTestStats())
		})

		if out != "Probing example.com on port 443 over TCP (resolved in 12.000 ms)\n" {
			t.Errorf("output = %q", out)
		}
	})

	t.Run("IP target has no resolution time", func(t *testing.T) {
		s := plainTestStats()
		s.Hostname = "93.184.216.34"
		s.DestIsIP = true

		out := captureStdout(t, func() {
			NewPlainPrinter(Config{}).PrintStart(s)
		})

		if out != "Probing 93.184.216.34 on port 443 over TCP\n" {
			t.Errorf("output = %q", out)
		}
	})
}

func TestPlainPrintNameResolutionDuration(t *testing.T) {
	t.Run("prints on its own line", func(t *testing.T) {
		out := captureStdout(t, func() {
			NewPlainPrinter(Config{}).PrintNameResolutionDuration(plainTestStats())
		})

		if out != "Resolved in 12.000 ms\n" {
			t.Errorf("output = %q", out)
		}
	})

	// The probe line carries the resolution time itself when the hostname
	// was resolved for that probe, so a second line would repeat it.
	t.Run("stays quiet when the probe line already says it", func(t *testing.T) {
		s := plainTestStats()
		s.ResolvedThisProbe = true

		out := captureStdout(t, func() {
			NewPlainPrinter(Config{}).PrintNameResolutionDuration(s)
		})

		if out != "" {
			t.Errorf("output = %q, want nothing", out)
		}
	})
}

func TestPlainPrintProbeSuccess(t *testing.T) {
	out := captureStdout(t, func() {
		NewPlainPrinter(Config{}).PrintProbeSuccess(plainTestStats())
	})

	want := "Reply from example.com (93.184.216.34) on port 443 TCP_conn=2 time=3.500 ms\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

func TestPlainPrintProbeFailure(t *testing.T) {
	out := captureStdout(t, func() {
		NewPlainPrinter(Config{}).PrintProbeFailure(plainTestStats())
	})

	want := "No reply from example.com (93.184.216.34) on port 443 TCP_conn=5\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

// An IP target has no hostname to put in front of the address.
func TestPlainProbeShowsIPTargetOnce(t *testing.T) {
	s := plainTestStats()
	s.Hostname = "93.184.216.34"
	s.DestIsIP = true

	out := captureStdout(t, func() {
		NewPlainPrinter(Config{}).PrintProbeSuccess(s)
	})

	wantLines(t, out, "Reply from 93.184.216.34 on port 443")
	wantNoLines(t, out, "(93.184.216.34)")
}

func TestPlainProbeTimestampAndSourceAddress(t *testing.T) {
	s := plainTestStats()
	s.LocalAddr = &net.TCPAddr{IP: net.ParseIP("192.168.1.10")}

	cfg := Config{WithTimestamp: true, WithSourceAddress: true}

	for _, tt := range []struct {
		name  string
		print func(p *PlainPrinter)
	}{
		{"success", func(p *PlainPrinter) { p.PrintProbeSuccess(s) }},
		{"failure", func(p *PlainPrinter) { p.PrintProbeFailure(s) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(t, func() {
				tt.print(NewPlainPrinter(cfg))
			})

			wantLines(t, out, "using 192.168.1.10:0 ", time.Now().Format("2006-01-02"))
		})
	}
}

// Without -I there is no source address to show, so -D has to leave the
// "using" part out rather than print an empty one.
func TestPlainProbeWithoutASourceAddress(t *testing.T) {
	s := plainTestStats()

	for _, tt := range []struct {
		name  string
		print func(p *PlainPrinter)
	}{
		{"success", func(p *PlainPrinter) { p.PrintProbeSuccess(s) }},
		{"failure", func(p *PlainPrinter) { p.PrintProbeFailure(s) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(t, func() {
				tt.print(NewPlainPrinter(Config{WithSourceAddress: true}))
			})

			wantNoLines(t, out, "using ")
		})
	}
}

// When the hostname is resolved for every probe, the probe line says how
// long that took instead of taking up a line of its own.
func TestPlainProbeShowsResolutionInline(t *testing.T) {
	s := plainTestStats()
	s.ResolvedThisProbe = true

	for _, tt := range []struct {
		name  string
		print func(p *PlainPrinter)
	}{
		{"success", func(p *PlainPrinter) { p.PrintProbeSuccess(s) }},
		{"failure", func(p *PlainPrinter) { p.PrintProbeFailure(s) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(t, func() {
				tt.print(NewPlainPrinter(Config{}))
			})

			wantLines(t, out, "(resolved in 12.000 ms)")
		})
	}
}

func TestPlainProbeUDPFailureReason(t *testing.T) {
	tests := []struct {
		name     string
		rejected bool
		want     string
	}{
		{"refused", true, "(port unreachable)"},
		{"silent", false, "(port may still be open)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := plainTestStats()
			s.Protocol = config.UDP
			s.UDP = stats.UDPInfo{ProbeNumber: 7, Rejected: tt.rejected}

			out := captureStdout(t, func() {
				NewPlainPrinter(Config{}).PrintProbeFailure(s)
			})

			wantLines(t, out, "UDP_conn=5 "+tt.want)
		})
	}
}

// Verbose adds an indented block under the probe line, saying which probe a
// reply belongs to.
func TestPlainProbeVerboseUDPDetails(t *testing.T) {
	s := plainTestStats()
	s.Protocol = config.UDP
	s.UDP = stats.UDPInfo{ProbeNumber: 7, Echoed: true, ReplySize: 12}

	out := captureStdout(t, func() {
		NewPlainPrinter(Config{Verbose: true}).PrintProbeSuccess(s)
	})

	wantLines(t, out, "    reply echoed back probe 7\n")
}

func TestPlainProbeHTTPStatus(t *testing.T) {
	s := plainTestStats()
	s.Protocol = config.HTTPS
	s.HTTP = stats.HTTPInfo{
		StatusCode:      200,
		Status:          "200 OK",
		Proto:           "HTTP/2.0",
		TLSVersion:      "TLS 1.3",
		TLSCipherSuite:  "TLS_AES_128_GCM_SHA256",
		CertExpiry:      time.Now().Add(48 * time.Hour),
		ConnectDuration: 10 * time.Millisecond,
		TLSDuration:     5 * time.Millisecond,
		TimeToFirstByte: 20 * time.Millisecond,
	}

	t.Run("the probe line carries the status", func(t *testing.T) {
		out := captureStdout(t, func() {
			NewPlainPrinter(Config{}).PrintProbeSuccess(s)
		})

		wantLines(t, out, "HTTPS_conn=2 status=200 time=3.500 ms")
		wantNoLines(t, out, "TLS 1.3")
	})

	t.Run("verbose adds the details underneath", func(t *testing.T) {
		out := captureStdout(t, func() {
			NewPlainPrinter(Config{Verbose: true}).PrintProbeSuccess(s)
		})

		wantLines(t, out,
			"    HTTP/2.0 200 OK\n",
			"    TLS 1.3 TLS_AES_128_GCM_SHA256\n",
			"connect=10.000 ms",
			"tls=5.000 ms",
			"ttfb=20.000 ms",
		)
	})
}

func TestPlainPrintStatistics(t *testing.T) {
	s := plainTestStats()
	s.TotalSuccessfulProbes = 3
	s.TotalUnsuccessfulProbes = 1
	s.LastSuccessfulProbe = time.Now()
	s.LastUnsuccessfulProbe = time.Now()
	s.TotalUptime = 3 * time.Second
	s.TotalDowntime = time.Second
	s.LongestUptime = stats.LongestTime{Start: time.Now(), End: time.Now(), Duration: 3 * time.Second}
	s.LongestDowntime = stats.LongestTime{Start: time.Now(), End: time.Now(), Duration: time.Second}
	s.RetriedHostnameLookups = 2
	s.RTTResults = stats.RTTResult{Min: 1, Average: 2, Max: 3, Mdev: 0.5}
	s.EndTime = s.StartTime.Add(4 * time.Second)

	out := captureStdout(t, func() {
		NewPlainPrinter(Config{}).PrintStatistics(s)
	})

	wantLines(t, out,
		"--- example.com (93.184.216.34) TCPing statistics ---\n",
		"4 TCP probes transmitted on port 443 | 3 received, 25.00% packet loss\n",
		"successful probes:   3\n",
		"unsuccessful probes: 1\n",
		"total uptime: 3 seconds\n",
		"total downtime: 1 second\n",
		"longest consecutive uptime:   3 seconds from ",
		"longest consecutive downtime: 1 second from ",
		"retried to resolve hostname 2 times\n",
		"rtt min/avg/max/mdev: 1.000/2.000/3.000/0.500 ms\n",
		"TCPing started at: ",
		"TCPing ended at:   ",
		"duration (HH:MM:SS): 00:00:04\n",
	)
}

// A run that never succeeded, against an IP, has nothing to say about
// latency, hostname lookups or an end time.
func TestPlainStatisticsLeavesOutWhatDidNotHappen(t *testing.T) {
	s := plainTestStats()
	s.Hostname = "93.184.216.34"
	s.DestIsIP = true
	s.TotalUnsuccessfulProbes = 2
	s.LastUnsuccessfulProbe = time.Now()

	out := captureStdout(t, func() {
		NewPlainPrinter(Config{}).PrintStatistics(s)
	})

	wantLines(t, out,
		"--- 93.184.216.34 TCPing statistics ---\n",
		"last successful probe:   Never succeeded\n",
	)

	wantNoLines(t, out,
		"rtt min/avg/max/mdev",
		"retried to resolve hostname",
		"longest consecutive",
		"TCPing ended at",
	)
}

// Only one lookup was retried, so the line has to read "1 time".
func TestPlainStatisticsSingleResolveRetry(t *testing.T) {
	s := plainTestStats()
	s.RetriedHostnameLookups = 1

	out := captureStdout(t, func() {
		NewPlainPrinter(Config{}).PrintStatistics(s)
	})

	wantLines(t, out, "retried to resolve hostname 1 time\n")
}

func TestPlainStatisticsHostnameChanges(t *testing.T) {
	s := plainTestStats()
	s.HostnameChanges = []stats.HostnameChange{
		{Addr: netip.MustParseAddr("93.184.216.34"), When: time.Now(), Duration: 12 * time.Millisecond},
		{Addr: netip.MustParseAddr("93.184.216.35"), When: time.Now(), Duration: 9 * time.Millisecond},
	}

	out := captureStdout(t, func() {
		NewPlainPrinter(Config{}).PrintStatistics(s)
	})

	wantLines(t, out,
		"IP address changes:\n",
		"  from 93.184.216.34 to 93.184.216.35 at ",
		"took 9.000 ms\n",
	)
}

func TestPlainSimpleMessages(t *testing.T) {
	s := plainTestStats()
	s.CurrentDowntime = 2 * time.Second
	s.CurrentUptime = 5 * time.Second

	tests := []struct {
		name  string
		print func(p *PlainPrinter)
		want  string
	}{
		{
			name:  "retry",
			print: func(p *PlainPrinter) { p.PrintRetryingToResolve("example.com") },
			want:  "Retrying to resolve example.com\n",
		},
		{
			name:  "downtime",
			print: func(p *PlainPrinter) { p.PrintDownTimeDuration(s) },
			want:  "No response received for 2 seconds after 5 seconds of uptime\n",
		},
		{
			name:  "uptime",
			print: func(p *PlainPrinter) { p.PrintUpTimeDuration(s) },
			want:  "Responses received for 5 seconds after 2 seconds of downtime\n",
		},
		{
			name: "downtime without a preceding uptime",
			print: func(p *PlainPrinter) {
				p.PrintDownTimeDuration(&stats.Statistics{CurrentDowntime: 2 * time.Second})
			},
			want: "No response received for 2 seconds\n",
		},
		{
			name: "uptime without a preceding downtime",
			print: func(p *PlainPrinter) {
				p.PrintUpTimeDuration(&stats.Statistics{CurrentUptime: 5 * time.Second})
			},
			want: "Responses received for 5 seconds\n",
		},
		{
			name:  "error",
			print: func(p *PlainPrinter) { p.PrintError("could not resolve %s", "example.com") },
			want:  "could not resolve example.com\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(t, func() {
				tt.print(NewPlainPrinter(Config{}))
			})

			if out != tt.want {
				t.Errorf("output = %q, want %q", out, tt.want)
			}
		})
	}
}

// Shutdown is what prints the summary on exit, so it has to print the same
// thing PrintStatistics does.
func TestPlainShutdownPrintsStatistics(t *testing.T) {
	s := plainTestStats()

	out := captureStdout(t, func() {
		NewPlainPrinter(Config{}).Shutdown(s)
	})

	wantLines(t, out, "TCPing statistics ---\n")
}

// With --no-stats, exiting prints nothing at all.
func TestPlainShutdownOmitsStatistics(t *testing.T) {
	s := plainTestStats()

	out := captureStdout(t, func() {
		NewPlainPrinter(Config{OmitStatistics: true}).Shutdown(s)
	})

	if out != "" {
		t.Errorf("output = %q, want it to be empty", out)
	}
}
