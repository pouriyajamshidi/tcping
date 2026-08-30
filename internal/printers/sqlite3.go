//go:build !windows

package printers

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

const (
	probeTableSchema = `CREATE TABLE IF NOT EXISTS %s (
		reachable TEXT,
		timestamp DATETIME,
		hostname TEXT,
		ip_address TEXT,
		port INTEGER,
		source_address TEXT,
		destination_is_ip TEXT,
		latency TEXT,
		ongoing_successful_probes INTEGER,
		ongoing_unsuccessful_probes INTEGER
	);`

	probeTableInsertSchema = `INSERT INTO %s (
		reachable,
		timestamp,
		hostname,
		ip_address,
		port,
		source_address,
		destination_is_ip,
		latency,
		ongoing_successful_probes,
		ongoing_unsuccessful_probes
		)
		VALUES (?,?,?,?,?,?,?,?,?,?);`
)

const (
	statsTableSchema = `CREATE TABLE IF NOT EXISTS %s (
		hostname TEXT,
		ip_address TEXT,
		port INTEGER,
		total_duration TEXT,
		total_uptime TEXT,
		total_downtime TEXT,
		total_packets INTEGER,
		total_successful_packets INTEGER,
		total_unsuccessful_packets INTEGER,
		total_packet_loss_percent TEXT,
		longest_uptime TEXT,
		longest_downtime TEXT,
		hostname_resolve_retries INTEGER,
		hostname_changes BLOB,
		last_successful_probe TEXT,
		last_unsuccessful_probe TEXT,
		longest_consecutive_uptime_start TEXT,
		longest_consecutive_uptime_end TEXT,
		longest_consecutive_downtime_start TEXT,
		longest_consecutive_downtime_end TEXT,
		latency_min TEXT,
		latency_avg TEXT,
		latency_max TEXT,
		start_time TEXT,
		end_time TEXT
	);`

	statsTableInsertSchema = `INSERT INTO %s (
		hostname,
		ip_address,
		port,
		total_duration,
		total_uptime,
		total_downtime,
		total_packets,
		total_successful_packets,
		total_unsuccessful_packets,
		total_packet_loss_percent,
		longest_uptime,
		longest_downtime,
		hostname_resolve_retries,
		hostname_changes,
		last_successful_probe,
		last_unsuccessful_probe,
		longest_consecutive_uptime_start,
		longest_consecutive_uptime_end,
		longest_consecutive_downtime_start,
		longest_consecutive_downtime_end,
		latency_min,
		latency_avg,
		latency_max,
		start_time,
		end_time
		)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);`
)

type probeData struct {
	reachable                 string
	timestamp                 string
	hostname                  string
	ipAddr                    string
	port                      uint16
	sourceAddr                string
	destIsIP                  string
	latency                   string
	ongoingSuccessfulProbes   uint
	ongoingUnsuccessfulProbes uint
}

func (d *probeData) toArgs() []any {
	return []any{
		d.reachable,
		d.timestamp,
		d.hostname,
		d.ipAddr,
		d.port,
		d.sourceAddr,
		d.destIsIP,
		d.latency,
		d.ongoingSuccessfulProbes,
		d.ongoingUnsuccessfulProbes,
	}
}

type probeStats struct {
	hostname                        string
	ipAddr                          string
	port                            uint16
	totalDuration                   string
	totalUptime                     string
	totalDowntime                   string
	totalPackets                    uint
	totalSuccessfulPackets          uint
	totalUnsuccessfulPackets        uint
	totalPacketLossPercent          string
	longestUptime                   string
	longestDowntime                 string
	hostnameResolveRetries          uint
	hostnameChanges                 string
	lastSuccessfulProbe             string
	lastUnsuccessfulProbe           string
	longestConsecutiveUptimeStart   string
	longestConsecutiveUptimeEnd     string
	longestConsecutiveDowntimeStart string
	longestConsecutiveDowntimeEnd   string
	latencyMin                      string
	latencyAvg                      string
	latencyMax                      string
	startTimestamp                  string
	endTimestamp                    string
}

func (d *probeStats) toArgs() []any {
	return []any{
		d.hostname,
		d.ipAddr,
		d.port,
		d.totalDuration,
		d.totalUptime,
		d.totalDowntime,
		d.totalPackets,
		d.totalSuccessfulPackets,
		d.totalUnsuccessfulPackets,
		d.totalPacketLossPercent,
		d.longestUptime,
		d.longestDowntime,
		d.hostnameResolveRetries,
		d.hostnameChanges,
		d.lastSuccessfulProbe,
		d.lastUnsuccessfulProbe,
		d.longestConsecutiveUptimeStart,
		d.longestConsecutiveUptimeEnd,
		d.longestConsecutiveDowntimeStart,
		d.longestConsecutiveDowntimeEnd,
		d.latencyMin,
		d.latencyAvg,
		d.latencyMax,
		d.startTimestamp,
		d.endTimestamp,
	}
}

// DatabasePrinter represents a SQLite database connection for storing probe results.
type DatabasePrinter struct {
	Conn           *sqlite.Conn
	probeTableName string
	statsTableName string
	FilePath       string
}

// NewDatabasePrinter initializes a new sqlite3 Database instance, creates the data and stats table, and returns a pointer to it.
func NewDatabasePrinter(target string, port uint16, filePath string) (*DatabasePrinter, error) {
	portStr := strconv.FormatUint(uint64(port), 10)

	probeTableName := sanitizeTableName(target, portStr)
	statsTableName := probeTableName + "_stats"

	filePath = addDbExtension(filePath)

	conn, err := sqlite.OpenConn(filePath, sqlite.OpenCreate, sqlite.OpenReadWrite)
	if err != nil {
		return nil, fmt.Errorf("error creating the database %q: %w", filePath, err)
	}

	tableSchema := fmt.Sprintf(probeTableSchema, probeTableName)
	if err = sqlitex.Execute(conn, tableSchema, &sqlitex.ExecOptions{}); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error creating the data table: %w", err)
	}

	statsSchema := fmt.Sprintf(statsTableSchema, statsTableName)
	if err = sqlitex.Execute(conn, statsSchema, &sqlitex.ExecOptions{}); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error creating the statistics table: %w", err)
	}

	return &DatabasePrinter{
		Conn:           conn,
		probeTableName: probeTableName,
		statsTableName: statsTableName,
		FilePath:       filePath,
	}, nil
}

func addDbExtension(filename string) string {
	if filename == ":memory:" || strings.HasSuffix(filename, ".db") {
		return filename
	}

	return filename + ".db"
}

// sanitizeTableName returns the sanitized and correctly formatted table name.
// Formatting the table name as "example_com_port__year_month_day_hour_minute_sec".
// Table names cannot contain '.', '-' and cannot start with numbers.
func sanitizeTableName(hostname, port string) string {
	sanitizedHost := strings.ReplaceAll(hostname, ".", "_")
	sanitizedHost = strings.ReplaceAll(sanitizedHost, "-", "_")

	sanitizedTime := strings.ReplaceAll(time.Now().Format(time.DateTime), "-", "_")
	sanitizedTime = strings.ReplaceAll(sanitizedTime, ":", "_")
	sanitizedTime = strings.ReplaceAll(sanitizedTime, " ", "_")

	tableName := fmt.Sprintf("%s_%s__%s",
		sanitizedHost,
		port,
		sanitizedTime,
	)

	// table names cannot start with a number
	if unicode.IsNumber(rune(tableName[0])) {
		tableName = "_" + tableName
	}

	return tableName
}

// PrintStart prints a message indicating that TCPing has started for the given hostname and port.
func (p *DatabasePrinter) PrintStart(s *stats.Statistics) {
	fmt.Printf("TCPinging %s on port %d - saving the results to: %s\n", s.Hostname, s.Port, p.FilePath)
}

// PrintProbeSuccess writes successful probe details to the database.
func (p *DatabasePrinter) PrintProbeSuccess(s *stats.Statistics) {
	data := probeData{
		reachable:               "true",
		hostname:                s.Hostname,
		ipAddr:                  s.IPStr(),
		port:                    s.Port,
		latency:                 s.RTTStr(),
		ongoingSuccessfulProbes: s.OngoingSuccessfulProbes,
	}

	if s.WithTimestamp {
		data.timestamp = s.CurrentTimestamp()
	}

	if s.WithSourceAddress && s.SourceAddr() != "" {
		data.sourceAddr = s.SourceAddr()
	}

	if s.DestIsIP {
		data.destIsIP = "true"
	} else {
		data.destIsIP = "false"
	}

	if err := sqlitex.Execute(
		p.Conn,
		fmt.Sprintf(probeTableInsertSchema, p.probeTableName),
		&sqlitex.ExecOptions{Args: data.toArgs()},
	); err != nil {
		p.PrintError("Failed writing probe success data to database: %v", err)
	}
}

// PrintProbeFailure writes failed probe details to the database.
func (p *DatabasePrinter) PrintProbeFailure(s *stats.Statistics) {
	data := probeData{
		reachable:                 "false",
		hostname:                  s.Hostname,
		ipAddr:                    s.IPStr(),
		port:                      s.Port,
		ongoingUnsuccessfulProbes: s.OngoingUnsuccessfulProbes,
	}

	if s.WithTimestamp {
		data.timestamp = s.CurrentTimestamp()
	}

	if s.WithSourceAddress && s.SourceAddr() != "" {
		data.sourceAddr = s.SourceAddr()
	}

	if s.DestIsIP {
		data.destIsIP = "true"
	} else {
		data.destIsIP = "false"
	}

	if err := sqlitex.Execute(
		p.Conn,
		fmt.Sprintf(probeTableInsertSchema, p.probeTableName),
		&sqlitex.ExecOptions{Args: data.toArgs()},
	); err != nil {
		p.PrintError("Failed writing probe failure data to database: %v", err)
	}
}

// PrintStatistics saves TCPing summary statistics to the database.
func (p *DatabasePrinter) PrintStatistics(s *stats.Statistics) {
	data := probeStats{
		hostname:                 s.Hostname,
		ipAddr:                   s.IPStr(),
		port:                     s.Port,
		totalDuration:            s.RuntimeDuration(),
		totalUptime:              s.TotalUptimeDuration(),
		totalDowntime:            s.TotalDowntimeDuration(),
		totalPackets:             s.TotalProbes(),
		totalSuccessfulPackets:   s.TotalSuccessfulProbes,
		totalUnsuccessfulPackets: s.TotalUnsuccessfulProbes,
		totalPacketLossPercent:   fmt.Sprintf("%.2f", s.PacketLoss()),
		startTimestamp:           s.StartTimeFormatted(),
	}

	if len(s.HostnameChanges) > 1 {
		var changes strings.Builder
		for i := 0; i < len(s.HostnameChanges)-1; i++ {
			if s.HostnameChanges[i].Addr.String() == "" {
				continue
			}

			fmt.Fprintf(&changes, "from %s to %s at %s\n",
				s.HostnameChanges[i].Addr.String(),
				s.HostnameChanges[i+1].Addr.String(),
				s.HostnameChanges[i+1].WhenFormatted(),
			)
		}

		data.hostnameChanges = changes.String()
	}

	if !s.LastSuccessfulProbe.IsZero() {
		data.lastSuccessfulProbe = s.LastSuccessfulProbeFormatted()
	}

	if !s.LastUnsuccessfulProbe.IsZero() {
		data.lastUnsuccessfulProbe = s.LastUnsuccessfulProbeFormatted()
	}

	if s.LongestUp.Duration != 0 {
		data.longestUptime = s.LongestUptimeDuration()
		data.longestConsecutiveUptimeStart = s.LongestUptimeStartTime()
		data.longestConsecutiveUptimeEnd = s.LongestUptimeEndTime()
	}

	if s.LongestDown.Duration != 0 {
		data.longestDowntime = s.LongestDowntimeDuration()
		data.longestConsecutiveDowntimeStart = s.LongestDowntimeStartTime()
		data.longestConsecutiveDowntimeEnd = s.LongestDowntimeEndTime()
	}

	if !s.DestIsIP {
		data.hostnameResolveRetries = s.RetriedHostnameLookups
	}

	if s.RTTResults.HasResults {
		data.latencyMin = fmt.Sprintf("%.3f", s.RTTResults.Min)
		data.latencyAvg = fmt.Sprintf("%.3f", s.RTTResults.Average)
		data.latencyMax = fmt.Sprintf("%.3f", s.RTTResults.Max)
	}

	if !s.EndTime.IsZero() {
		data.endTimestamp = s.EndTimeFormatted()
	}

	if err := sqlitex.Execute(
		p.Conn,
		fmt.Sprintf(statsTableInsertSchema, p.statsTableName),
		&sqlitex.ExecOptions{Args: data.toArgs()},
	); err != nil {
		p.PrintError("Failed writing statistics to database: %v", err)
		return
	}

	fmt.Printf("\nProbe and statistics data for %q have been saved to the table %q and %q, respectively\n",
		s.Hostname,
		p.probeTableName,
		p.statsTableName,
	)
}

// PrintRetryingToResolve prints a message indicating that the program is retrying to resolve the hostname.
func (p *DatabasePrinter) PrintRetryingToResolve(hostname string) {
	fmt.Printf("Retrying to resolve %s\n", hostname)
}

// PrintDownTimeDuration prints the total duration of downtime when no response was received.
func (p *DatabasePrinter) PrintDownTimeDuration(s *stats.Statistics) {
	fmt.Printf("No response received for %s\n", s.DowntimeDuration())
}

// PrintError prints an error message to stderr.
func (p *DatabasePrinter) PrintError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Database Error: "+format+"\n", args...)
}

// Shutdown sets the end time, prints statistics, calls Done() and exits the program.
func (p *DatabasePrinter) Shutdown(s *stats.Statistics) {
	s.EndTime = time.Now()
	PrintStats(p, s)
	p.Done()
	os.Exit(0)
}

// Done closes the connection to the database.
func (p *DatabasePrinter) Done() {
	if p.Conn != nil {
		p.Conn.Close()
	}
}
