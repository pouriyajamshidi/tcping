package printers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
)

const (
	// Probes are usually a second apart, so a write that hangs longer than
	// this would hold up the next probe.
	influxDBTimeout = 2 * time.Second

	influxDBWritePath = "/api/v2/write"

	// Used when the caller did not ask for an interval, so a zero value
	// does not end up meaning "send the summary with every probe".
	defaultInfluxDBStatsInterval = 10 * time.Second
)

// A comma, a space and an equals sign are what separate the parts of a line,
// so a tag value holding one has to escape it. Only the hostname is likely
// to, and it comes straight from the command line. Field values need no
// escaping of their own because every field we write is a number.
var influxDBEscaper = strings.NewReplacer(",", `\,`, " ", `\ `, "=", `\=`)

// InfluxDBPrinter sends probe results to InfluxDB as line protocol instead
// of printing them, so a run shows up as a graph rather than as lines of
// text. Line protocol is plain text and the write API is a plain POST, so
// there is nothing here that needs a client library.
type InfluxDBPrinter struct {
	client   *http.Client
	endpoint string
	token    string
	// When the run summary was last written. Zero, so the first probe
	// carries one.
	lastStats time.Time
	// How often the run summary rides along with a probe. Without it the
	// summary would only be written when you press Enter or when tcping
	// exits, which never happens on a long run that no one is watching.
	statsInterval time.Duration
	warned        bool // Whether we already complained about a write that failed.
}

// NewInfluxDBPrinter creates an InfluxDBPrinter pointed at the given
// InfluxDB address. The address can be given with or without the write path,
// so both "http://localhost:8086" and
// "http://localhost:8086/api/v2/write" work.
func NewInfluxDBPrinter(cfg config.PrinterConfig) (*InfluxDBPrinter, error) {
	if cfg.InfluxDBOrg == "" {
		return nil, fmt.Errorf("InfluxDB needs an organization, give it with -influxdb-org")
	}

	if cfg.InfluxDBBucket == "" {
		return nil, fmt.Errorf("InfluxDB needs a bucket, give it with -influxdb-bucket")
	}

	if cfg.InfluxDBToken == "" {
		return nil, fmt.Errorf("InfluxDB needs an API token, put it in the INFLUXDB_TOKEN environment variable")
	}

	endpoint := cfg.InfluxDBURL

	if !strings.Contains(endpoint, "://") {
		endpoint = "http://" + endpoint
	}

	if !strings.HasSuffix(endpoint, influxDBWritePath) {
		endpoint = strings.TrimSuffix(endpoint, "/") + influxDBWritePath
	}

	// The organization and the bucket are whatever the user typed, so let
	// net/url escape them instead of pasting them into the address.
	query := url.Values{}
	query.Set("org", cfg.InfluxDBOrg)
	query.Set("bucket", cfg.InfluxDBBucket)
	query.Set("precision", "ns")

	statsInterval := cfg.InfluxDBStatsInterval
	if statsInterval <= 0 {
		statsInterval = defaultInfluxDBStatsInterval
	}

	return &InfluxDBPrinter{
		client:        &http.Client{Timeout: influxDBTimeout},
		endpoint:      endpoint + "?" + query.Encode(),
		token:         cfg.InfluxDBToken,
		statsInterval: statsInterval,
	}, nil
}

// tags are put on every line, which is what makes them the columns you can
// group and filter by on the other end. The IP can change mid-run when the
// hostname is resolved again, so they are built fresh every time.
func (p *InfluxDBPrinter) tags(s *stats.Statistics) string {
	return fmt.Sprintf("target=%s,ip=%s,port=%s,protocol=%s",
		influxDBEscaper.Replace(s.Hostname),
		influxDBEscaper.Replace(s.IPStr()),
		s.PortStr(),
		influxDBEscaper.Replace(s.ProtocolStr()),
	)
}

// line builds one line of line protocol: what was measured and on which
// target, the values themselves, and the time they were taken. An integer
// field is written with an "i" suffix and anything without one is a float,
// so a field has to keep the same form on every line or InfluxDB rejects
// the later ones.
func (p *InfluxDBPrinter) line(s *stats.Statistics, measurement, fields string) string {
	return fmt.Sprintf("%s,%s %s %d", measurement, p.tags(s), fields, time.Now().UnixNano())
}

// send POSTs one batch of lines to InfluxDB. A failure does not stop the
// probing, and we say so only once, so an InfluxDB that is down does not
// fill the terminal with the same error every second.
func (p *InfluxDBPrinter) send(lines []string) {
	if len(lines) == 0 {
		return
	}

	req, err := http.NewRequest(http.MethodPost, p.endpoint, strings.NewReader(strings.Join(lines, "\n")))
	if err != nil {
		p.warnOnce("could not build the request: %v", err)
		return
	}

	req.Header.Set("Authorization", "Token "+p.token)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	resp, err := p.client.Do(req)
	if err != nil {
		p.warnOnce("could not reach InfluxDB at %s: %v", p.endpoint, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		// InfluxDB says what it did not like in the body, and that is the
		// only way to tell a wrong token from a wrong bucket.
		reason, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		p.warnOnce("InfluxDB rejected the write with %s: %s", resp.Status, strings.TrimSpace(string(reason)))
	}
}

func (p *InfluxDBPrinter) warnOnce(format string, args ...any) {
	if p.warned {
		return
	}

	p.warned = true
	fmt.Fprintf(os.Stderr, "InfluxDB Error: "+format+"\n", args...)
	fmt.Fprintln(os.Stderr, "Probing continues, but the metrics are being dropped.")
}

// probeMeasurement is the name a probe is written under. There is one per
// probe type, so a point carries exactly the fields of the probe that
// produced it instead of a pile of columns that are empty most of the time.
func probeMeasurement(s *stats.Statistics) string {
	switch {
	case s.IsHTTP():
		return "tcping_http"

	case s.IsUDP():
		return "tcping_udp"

	default:
		return "tcping_tcp"
	}
}

// httpFields are the extra timings an HTTP(S) probe learned. They are the
// reason to graph an HTTP target at all: a slow connect and a slow first
// byte mean different things.
func (p *InfluxDBPrinter) httpFields(s *stats.Statistics) string {
	if !s.IsHTTP() || !s.HasHTTPResponse() {
		return ""
	}

	fields := fmt.Sprintf(",status_code=%di,connect_ms=%g,ttfb_ms=%g",
		s.HTTP.StatusCode,
		msOf(s.HTTP.ConnectDuration),
		msOf(s.HTTP.TimeToFirstByte),
	)

	// Plain HTTP has no handshake and no certificate to report on.
	if s.HTTP.TLSVersion != "" {
		fields += fmt.Sprintf(",tls_handshake_ms=%g,certificate_days_remaining=%di",
			msOf(s.HTTP.TLSDuration),
			s.CertDaysRemaining(),
		)
	}

	return fields
}

// udpFields are what a UDP probe learned. A reply that carried our own
// payload back is the only proof that something is really listening, and a
// refusal the only proof that nothing is, so both are worth keeping. The
// probe number goes along too, so a gap in the graph names the probe that
// was lost.
func (p *InfluxDBPrinter) udpFields(s *stats.Statistics) string {
	if !s.IsUDP() {
		return ""
	}

	return fmt.Sprintf(",probe_number=%di,reply_bytes=%di,echoed=%t,rejected=%t",
		s.UDP.ProbeNumber,
		s.UDP.ReplySize,
		s.UDP.Echoed,
		s.UDP.Rejected,
	)
}

// probeLine is what every probe writes, successful or not. The counts ride
// along with each probe so a graph can show both the last probe and how the
// run is going.
func (p *InfluxDBPrinter) probeLine(s *stats.Statistics, succeeded bool) string {
	var up int
	if succeeded {
		up = 1
	}

	fields := fmt.Sprintf("success=%di,successful_probes=%di,unsuccessful_probes=%di",
		up, s.TotalSuccessfulProbes, s.TotalUnsuccessfulProbes)

	// A failed probe has no round trip time to report.
	if succeeded {
		fields += fmt.Sprintf(",rtt_ms=%g", s.LatestRTT)
	}

	fields += p.httpFields(s) + p.udpFields(s)

	return p.line(s, probeMeasurement(s), fields)
}

// statisticsLines are the summary of the run so far: the numbers you would
// otherwise only see when tcping exits.
func (p *InfluxDBPrinter) statisticsLines(s *stats.Statistics) []string {
	fields := fmt.Sprintf("packet_loss_percent=%g,uptime_seconds=%g,downtime_seconds=%g",
		s.PacketLoss(),
		s.TotalUptime.Seconds(),
		s.TotalDowntime.Seconds(),
	)

	// Without a single successful probe there is no latency to summarize,
	// and writing zeros would look like a very fast target.
	if s.TotalSuccessfulProbes > 0 {
		fields += fmt.Sprintf(",rtt_min_ms=%g,rtt_avg_ms=%g,rtt_max_ms=%g,rtt_mdev_ms=%g",
			s.RTTResults.Min,
			s.RTTResults.Average,
			s.RTTResults.Max,
			s.RTTResults.Mdev,
		)
	}

	return []string{p.line(s, "tcping_statistics", fields)}
}

// dueStatistics returns the run summary when enough time has passed since
// the last one, and nothing otherwise. Writing it along with a probe means
// no goroutine and no second request, and it keeps the summary flowing on a
// run that is never going to be stopped by hand.
func (p *InfluxDBPrinter) dueStatistics(s *stats.Statistics) []string {
	if time.Since(p.lastStats) < p.statsInterval {
		return nil
	}

	p.lastStats = time.Now()

	return p.statisticsLines(s)
}

// PrintStart says where the metrics are going, then leaves the terminal
// alone for the rest of the run.
func (p *InfluxDBPrinter) PrintStart(s *stats.Statistics) {
	fmt.Printf("Probing %s on port %d over %s - sending metrics to: %s\n",
		s.Hostname, s.Port, s.ProtocolStr(), p.endpoint)
}

// PrintNameResolutionDuration writes how long the hostname resolution took.
func (p *InfluxDBPrinter) PrintNameResolutionDuration(s *stats.Statistics) {
	p.send([]string{
		p.line(s, "tcping_name_resolution", fmt.Sprintf("duration_ms=%g", msOf(s.NameResolutionDuration))),
	})
}

// PrintProbeSuccess writes the metrics of a successful probe.
func (p *InfluxDBPrinter) PrintProbeSuccess(s *stats.Statistics) {
	p.send(append([]string{p.probeLine(s, true)}, p.dueStatistics(s)...))
}

// PrintProbeFailure writes the metrics of a failed probe.
func (p *InfluxDBPrinter) PrintProbeFailure(s *stats.Statistics) {
	p.send(append([]string{p.probeLine(s, false)}, p.dueStatistics(s)...))
}

// PrintStatistics writes the summary of the run so far.
func (p *InfluxDBPrinter) PrintStatistics(s *stats.Statistics) {
	p.lastStats = time.Now()
	p.send(p.statisticsLines(s))
}

// PrintRetryingToResolve has no number behind it, so it goes to the terminal.
func (p *InfluxDBPrinter) PrintRetryingToResolve(hostname string) {
	fmt.Fprintf(os.Stderr, "retrying to resolve %s\n", hostname)
}

// PrintDownTimeDuration writes how long the outage that just ended lasted.
func (p *InfluxDBPrinter) PrintDownTimeDuration(s *stats.Statistics) {
	p.send([]string{
		p.line(s, "tcping_downtime", fmt.Sprintf("seconds=%g", s.CurrentDowntime.Seconds())),
	})
}

// PrintUpTimeDuration writes how long the target was up for, right as it
// stops responding.
func (p *InfluxDBPrinter) PrintUpTimeDuration(s *stats.Statistics) {
	p.send([]string{
		p.line(s, "tcping_uptime", fmt.Sprintf("seconds=%g", s.CurrentUptime.Seconds())),
	})
}

func (p *InfluxDBPrinter) PrintError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "InfluxDB Error: "+format+"\n", args...)
}

// Shutdown writes the final statistics and closes the connection to
// InfluxDB. Statistics are already finalized by finalizeStatistics by the
// time this runs. It does not exit the program - that decision belongs to
// the caller, not the printer.
func (p *InfluxDBPrinter) Shutdown(s *stats.Statistics) {
	p.PrintStatistics(s)
	p.client.CloseIdleConnections()
}
