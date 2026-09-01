//go:build !windows

package printers

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

const (
	probeTableSchema = `CREATE TABLE IF NOT EXISTS %s (
		reachable INTEGER NOT NULL CHECK (reachable IN (0, 1)),
		timestamp TEXT,
		hostname TEXT NOT NULL,
		ip_address TEXT NOT NULL,
		port INTEGER NOT NULL,
		source_address TEXT,
		destination_is_ip INTEGER NOT NULL CHECK (destination_is_ip IN (0, 1)),
		latency REAL,
		ongoing_successful_probes INTEGER NOT NULL DEFAULT 0,
		ongoing_unsuccessful_probes INTEGER NOT NULL DEFAULT 0
	);`

	statsTableSchema = `CREATE TABLE IF NOT EXISTS %s (
		hostname TEXT NOT NULL,
		ip_address TEXT NOT NULL,
		port INTEGER NOT NULL,
		total_duration TEXT,
		total_uptime TEXT,
		total_downtime TEXT,
		total_packets INTEGER NOT NULL,
		total_successful_packets INTEGER NOT NULL,
		total_unsuccessful_packets INTEGER NOT NULL,
		total_packet_loss_percent REAL,
		longest_uptime TEXT,
		longest_downtime TEXT,
		hostname_resolve_retries INTEGER NOT NULL DEFAULT 0,
		hostname_changes TEXT,
		last_successful_probe TEXT,
		last_unsuccessful_probe TEXT,
		longest_consecutive_uptime_start TEXT,
		longest_consecutive_uptime_end TEXT,
		longest_consecutive_downtime_start TEXT,
		longest_consecutive_downtime_end TEXT,
		latency_min REAL,
		latency_avg REAL,
		latency_max REAL,
		start_time TEXT,
		end_time TEXT
	);`
)

const (
	probeInsert = `INSERT INTO %s (
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
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	statsInsert = `INSERT INTO %s (
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
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
)

// DatabasePrinter stores one probe stream and its final statistics in SQLite.
type DatabasePrinter struct {
	Conn           *sqlite.Conn
	probeTableName string
	statsTableName string
	FilePath       string
}

// NewDatabasePrinter opens the database and creates the probe and statistics tables.
func NewDatabasePrinter(target string, port uint16, filePath string) (*DatabasePrinter, error) {
	portStr := strconv.FormatUint(uint64(port), 10)
	probeTableName := sanitizeTableName(target, portStr)
	statsTableName := probeTableName + "_stats"
	filePath = addDbExtension(filePath)

	flags := sqlite.OpenCreate | sqlite.OpenReadWrite
	if filePath != ":memory:" {
		flags |= sqlite.OpenWAL
	}

	conn, err := sqlite.OpenConn(filePath, flags)
	if err != nil {
		return nil, fmt.Errorf("error creating the database %q: %w", filePath, err)
	}

	conn.SetBusyTimeout(5 * time.Second)

	if filePath != ":memory:" {
		if err := sqlitex.Execute(conn, "PRAGMA synchronous=NORMAL;", &sqlitex.ExecOptions{}); err != nil {
			conn.Close()
			return nil, fmt.Errorf("error configuring the database: %w", err)
		}
	}

	if err := sqlitex.Execute(conn, fmt.Sprintf(probeTableSchema, probeTableName), &sqlitex.ExecOptions{}); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error creating the data table: %w", err)
	}

	if err := sqlitex.Execute(conn, fmt.Sprintf(statsTableSchema, statsTableName), &sqlitex.ExecOptions{}); err != nil {
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

// sanitizeTableName formats a target/port pair as a valid SQLite identifier.
func sanitizeTableName(hostname, port string) string {
	sanitizedHost := sanitizeIdentifierPart(hostname)
	sanitizedTime := strings.NewReplacer("-", "_", ":", "_", " ", "_").Replace(time.Now().Format(time.DateTime))

	tableName := fmt.Sprintf("%s_%s__%s", sanitizedHost, port, sanitizedTime)
	if tableName != "" {
		if r, _ := utf8.DecodeRuneInString(tableName); unicode.IsNumber(r) {
			tableName = "_" + tableName
		}
	}
	return tableName
}

// sanitizeIdentifierPart replaces every character that isn't a letter, digit,
// or underscore with an underscore. Table names are built with fmt.Sprintf
// directly into DDL/DML text (SQL identifiers can't be bound as query
// parameters), so this allow-list is what keeps an arbitrary hostname string
// from being able to break out of the intended identifier.
func sanitizeIdentifierPart(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

func (p *DatabasePrinter) insertProbe(
	reachable bool,
	s *stats.Statistics,
	latency string,
	successfulProbes uint,
	unsuccessfulProbes uint,
) {
	timestamp := ""
	if s.WithTimestamp {
		timestamp = s.CurrentTimestamp()
	}

	sourceAddr := ""
	if s.WithSourceAddress {
		sourceAddr = s.SourceAddr()
	}

	latencyValue := any(nil)
	if latency != "" {
		value, err := strconv.ParseFloat(latency, 64)
		if err == nil {
			latencyValue = math.Round(value*1000) / 1000
		} else {
			p.PrintError("Failed parsing latency value %q: %v", latency, err)
		}
	}

	args := []any{
		boolToInt(reachable),
		timestamp,
		s.Hostname,
		s.IPStr(),
		s.Port,
		sourceAddr,
		boolToInt(s.DestIsIP),
		latencyValue,
		successfulProbes,
		unsuccessfulProbes,
	}

	if err := sqlitex.Execute(
		p.Conn,
		fmt.Sprintf(probeInsert, p.probeTableName),
		&sqlitex.ExecOptions{Args: args},
	); err != nil {
		p.PrintError("Failed writing probe data to database: %v", err)
	}
}

// PrintStart prints a message indicating that TCPing has started.
func (p *DatabasePrinter) PrintStart(s *stats.Statistics) {
	fmt.Printf("TCPinging %s on port %d - saving the results to: %s\n", s.Hostname, s.Port, p.FilePath)
}

// PrintProbeSuccess writes successful probe details to the database.
func (p *DatabasePrinter) PrintProbeSuccess(s *stats.Statistics) {
	p.insertProbe(true, s, s.RTTStr(), s.OngoingSuccessfulProbes, 0)
}

// PrintProbeFailure writes failed probe details to the database.
func (p *DatabasePrinter) PrintProbeFailure(s *stats.Statistics) {
	p.insertProbe(false, s, "", 0, s.OngoingUnsuccessfulProbes)
}

// PrintStatistics stores the same summary data that PlainPrinter presents.
func (p *DatabasePrinter) PrintStatistics(s *stats.Statistics) {
	lastSuccessful := ""
	if !s.LastSuccessfulProbe.IsZero() {
		lastSuccessful = s.LastSuccessfulProbeFormatted()
	}

	lastUnsuccessful := ""
	if !s.LastUnsuccessfulProbe.IsZero() {
		lastUnsuccessful = s.LastUnsuccessfulProbeFormatted()
	}

	longestUptime := ""
	longestUptimeStart := ""
	longestUptimeEnd := ""
	if s.LongestUptime.Duration != 0 {
		longestUptime = s.LongestUptimeDuration()
		longestUptimeStart = s.LongestUptimeStartTime()
		longestUptimeEnd = s.LongestUptimeEndTime()
	}

	longestDowntime := ""
	longestDowntimeStart := ""
	longestDowntimeEnd := ""
	if s.LongestDowntime.Duration != 0 {
		longestDowntime = s.LongestDowntimeDuration()
		longestDowntimeStart = s.LongestDowntimeStartTime()
		longestDowntimeEnd = s.LongestDowntimeEndTime()
	}

	latencyMin := any(nil)
	latencyAvg := any(nil)
	latencyMax := any(nil)
	if s.TotalSuccessfulProbes > 0 {
		latencyMin = math.Round(float64(s.RTTResults.Min)*1000) / 1000
		latencyAvg = math.Round(float64(s.RTTResults.Average)*1000) / 1000
		latencyMax = math.Round(float64(s.RTTResults.Max)*1000) / 1000
	}

	endTime := ""
	if !s.EndTime.IsZero() {
		endTime = s.EndTimeFormatted()
	}

	retries := uint(0)
	if !s.DestIsIP {
		retries = s.RetriedHostnameLookups
	}

	args := []any{
		s.Hostname,
		s.IPStr(),
		s.Port,
		s.RuntimeDuration(),
		s.TotalUptimeDuration(),
		s.TotalDowntimeDuration(),
		s.TotalProbes(),
		s.TotalSuccessfulProbes,
		s.TotalUnsuccessfulProbes,
		s.PacketLoss(),
		longestUptime,
		longestDowntime,
		retries,
		hostnameChanges(s),
		lastSuccessful,
		lastUnsuccessful,
		longestUptimeStart,
		longestUptimeEnd,
		longestDowntimeStart,
		longestDowntimeEnd,
		latencyMin,
		latencyAvg,
		latencyMax,
		s.StartTimeFormatted(),
		endTime,
	}

	if err := sqlitex.Execute(
		p.Conn,
		fmt.Sprintf(statsInsert, p.statsTableName),
		&sqlitex.ExecOptions{Args: args},
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

// PrintRetryingToResolve prints a message indicating that the program is retrying to resolve a hostname.
func (p *DatabasePrinter) PrintRetryingToResolve(hostname string) {
	fmt.Printf("Retrying to resolve %s\n", hostname)
}

// PrintDownTimeDuration prints the total duration of downtime when no response was received.
func (p *DatabasePrinter) PrintDownTimeDuration(s *stats.Statistics) {
	fmt.Printf("No response received for %s\n", s.DowntimeDuration())
}

// PrintUpTimeDuration prints how long the target was up for, right as it stops responding.
func (p *DatabasePrinter) PrintUpTimeDuration(s *stats.Statistics) {
	fmt.Printf("No response received after %s of uptime\n", s.UptimeDuration())
}

func (p *DatabasePrinter) PrintError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Database Error: "+format+"\n", args...)
}

// Shutdown prints statistics, calls Done() and exits the program.
// Statistics are already finalized by finalizeStatistics by the time this runs.
func (p *DatabasePrinter) Shutdown(s *stats.Statistics) {
	p.PrintStatistics(s)
	p.Done()
	os.Exit(0)
}

func (p *DatabasePrinter) Done() {
	if p.Conn != nil {
		_ = p.Conn.Close()
	}
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func hostnameChanges(s *stats.Statistics) string {
	if len(s.HostnameChanges) < 2 {
		return ""
	}

	var changes strings.Builder
	for i := 0; i < len(s.HostnameChanges)-1; i++ {
		from := s.HostnameChanges[i].Addr.String()
		if from == "" {
			continue
		}

		fmt.Fprintf(&changes, "from %s to %s at %s\n",
			from,
			s.HostnameChanges[i+1].Addr.String(),
			s.HostnameChanges[i+1].WhenFormatted(),
		)
	}
	return changes.String()
}
