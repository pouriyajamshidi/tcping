// Package printers contains the logic for printing information
package printers

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/pouriyajamshidi/tcping/v3/option"
	"github.com/pouriyajamshidi/tcping/v3/statistics"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

// EventType is a special type for each method
// in the printer interface so that automatic tools
// can understand what kind of an event they've received.
// For instance, probe vs statistics...
type EventType string

const (
	ProbeEvent          EventType = "probe"
	StatisticsEvent     EventType = "statistics"
	HostnameChangeEvent EventType = "hostname change"
)

const (
	dataTableSchema = `CREATE TABLE IF NOT EXISTS %s (
		type TEXT NOT NULL,
		success TEXT,
		timestamp DATETIME,
		ip_address TEXT,
		hostname TEXT,
		port INTEGER,
		source_address TEXT,
		destination_is_ip TEXT,
		time TEXT,
		ongoing_successful_probes INTEGER,
		ongoing_unsuccessful_probes INTEGER
	);`

	dataTableInsertSchema = `INSERT INTO %s (
		type,
		success,
		timestamp,
		ip_address,
		hostname,
		port,
		source_address,
		destination_is_ip,
		time,
		ongoing_successful_probes,
		ongoing_unsuccessful_probes
		)
		VALUES (?,?,?,?,?,?,?,?,?,?,?);`
)

const (
	statsTableSchema = `CREATE TABLE IF NOT EXISTS %s (
		type TEXT NOT NULL,
		timestamp DATETIME,
		ip_address TEXT,
		hostname TEXT,
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
		type,
		timestamp,
		ip_address,
		hostname,
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
		end_time)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);`
)

type dbData struct {
	eventType                 EventType
	success                   string
	timestamp                 string
	ipAddr                    string
	hostname                  string
	port                      uint16
	sourceAddr                string
	destIsIP                  string
	time                      string
	ongoingSuccessfulProbes   uint
	ongoingUnsuccessfulProbes uint
}

func (d *dbData) toArgs() []any {
	return []any{
		d.eventType,
		d.success,
		d.timestamp,
		d.ipAddr,
		d.hostname,
		d.port,
		d.sourceAddr,
		d.destIsIP,
		d.time,
		d.ongoingSuccessfulProbes,
		d.ongoingUnsuccessfulProbes,
	}
}

type dbStats struct {
	eventType                       EventType
	timestamp                       string
	ipAddr                          string
	hostname                        string
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

func (d *dbStats) toArgs() []any {
	return []any{
		d.eventType,
		d.timestamp,
		d.ipAddr,
		d.hostname,
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

// DatabasePrinter represents a SQLite database connection for storing TCPing results.
type DatabasePrinter struct {
	Conn           *sqlite.Conn
	probeTableName string
	statsTableName string
	FilePath       string
	opt            options
}

type DatabasePrinterOption = option.Option[DatabasePrinter]

func (p *DatabasePrinter) options() *options {
	return &p.opt
}

// NewDatabasePrinter initializes a new sqlite3 Database instance, creates the data table, and returns a pointer to it.
// If any error occurs during database creation or table initialization, the function exits the program.
func NewDatabasePrinter(target, port, filePath string, opts ...DatabasePrinterOption) (*DatabasePrinter, error) {
	probeTableName := sanitizeTableName(target, port)
	statsTableName := probeTableName + "_stats"

	filePath = addDbExtension(filePath)

	conn, err := sqlite.OpenConn(filePath, sqlite.OpenCreate, sqlite.OpenReadWrite)
	if err != nil {
		return nil, fmt.Errorf("\ncreate database %q: %w", filePath, err)
	}

	tableSchema := fmt.Sprintf(dataTableSchema, probeTableName)
	if err = sqlitex.Execute(conn, tableSchema, &sqlitex.ExecOptions{}); err != nil {
		return nil, fmt.Errorf("\ncreate data table: %w", err)
	}

	statsTableSchema := fmt.Sprintf(statsTableSchema, statsTableName)
	if err = sqlitex.Execute(conn, statsTableSchema, &sqlitex.ExecOptions{}); err != nil {
		return nil, fmt.Errorf("\ncreate statistics table: %w", err)
	}

	p := &DatabasePrinter{
		Conn:           conn,
		probeTableName: probeTableName,
		statsTableName: statsTableName,
		FilePath:       filePath,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p, nil
}

func addDbExtension(filename string) string {
	if filename == ":memory:" || strings.HasSuffix(filename, ".db") {
		return filename
	}

	return filename + ".db"
}

// sanitizeTableName will return the sanitized and correctly formatted table name
// formatting the table name as "example_com_port__year_month_day_hour_minute_sec"
// table name can't have '.','-' and can't start with numbers
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

	if unicode.IsNumber(rune(tableName[0])) {
		tableName = "_" + tableName
	}

	return tableName
}

// Done closes the connection to the database
func (p *DatabasePrinter) Done() {
	p.Conn.Close()
}

// Shutdown performs final cleanup for the printer.
func (p *DatabasePrinter) Shutdown(s *statistics.Statistics) {
	p.Done()
}

// PrintStart prints a message indicating that TCPing has started for the given hostname and port.
func (p *DatabasePrinter) PrintStart(s *statistics.Statistics) {
	fmt.Printf("TCPinging %s on port %d - saving the results to: %s\n", s.Hostname, s.Port, p.FilePath)
}

// PrintProbeSuccess satisfies the "printer" interface but does nothing in this implementation
func (p *DatabasePrinter) PrintProbeSuccess(s *statistics.Statistics) {
	if p.opt.ShowFailuresOnly {
		return
	}

	timestamp := ""
	if p.opt.ShowTimestamp {
		timestamp = s.StartTimeFormatted()
	}

	data := dbData{
		eventType:               ProbeEvent,
		success:                 "true",
		ongoingSuccessfulProbes: s.OngoingSuccessfulProbes,
	}

	if s.Hostname == s.IP.String() {
		data.destIsIP = "true"

		if timestamp == "" {
			if p.opt.ShowSourceAddress {
				data.ipAddr = s.IP.String()
				data.port = s.Port
				data.sourceAddr = s.SourceAddr()
				data.time = s.RTTStr()
			} else {
				data.ipAddr = s.IP.String()
				data.port = s.Port
				data.time = s.RTTStr()
			}
		} else {
			data.timestamp = timestamp

			if p.opt.ShowSourceAddress {
				data.ipAddr = s.IP.String()
				data.port = s.Port
				data.sourceAddr = s.SourceAddr()
				data.time = s.RTTStr()
			} else {
				data.ipAddr = s.IP.String()
				data.port = s.Port
				data.time = s.RTTStr()
			}
		}
	} else {
		data.destIsIP = "false"

		if timestamp == "" {
			if p.opt.ShowSourceAddress {
				data.hostname = s.Hostname
				data.ipAddr = s.IP.String()
				data.port = s.Port
				data.sourceAddr = s.SourceAddr()
				data.time = s.RTTStr()
			} else {
				data.hostname = s.Hostname
				data.ipAddr = s.IP.String()
				data.port = s.Port
				data.time = s.RTTStr()
			}
		} else {
			data.timestamp = timestamp

			if p.opt.ShowSourceAddress {
				data.ipAddr = s.IP.String()
				data.port = s.Port
				data.sourceAddr = s.SourceAddr()
				data.time = s.RTTStr()
			} else {
				data.ipAddr = s.IP.String()
				data.port = s.Port
				data.time = s.RTTStr()
			}
		}
	}

	if err := sqlitex.Execute(
		p.Conn,
		fmt.Sprintf(dataTableInsertSchema, p.probeTableName),
		&sqlitex.ExecOptions{Args: data.toArgs()},
	); err != nil {
		p.PrintError("Failed writing probe success data to database: %s\n", err)
	}
}

// PrintProbeFailure satisfies the "printer" interface but does nothing in this implementation
func (p *DatabasePrinter) PrintProbeFailure(s *statistics.Statistics) {
	timestamp := ""
	if p.opt.ShowTimestamp {
		timestamp = s.StartTimeFormatted()
	}

	data := dbData{
		eventType:                 ProbeEvent,
		success:                   "false",
		ongoingUnsuccessfulProbes: s.OngoingUnsuccessfulProbes,
	}

	if s.Hostname == s.IP.String() {
		data.destIsIP = "true"

		if timestamp == "" {
			data.ipAddr = s.IP.String()
			data.port = s.Port
		} else {
			data.timestamp = timestamp
			data.ipAddr = s.IP.String()
			data.port = s.Port
		}
	} else {
		data.destIsIP = "false"

		if timestamp == "" {
			data.hostname = s.Hostname
			data.ipAddr = s.IP.String()
			data.port = s.Port
		} else {
			data.timestamp = timestamp
			data.hostname = s.Hostname
			data.ipAddr = s.IP.String()
			data.port = s.Port
		}
	}

	if err := sqlitex.Execute(
		p.Conn,
		fmt.Sprintf(dataTableInsertSchema, p.probeTableName),
		&sqlitex.ExecOptions{Args: data.toArgs()},
	); err != nil {
		p.PrintError("Failed writing probe failure data to database: %s\n", err)
	}
}

// PrintError prints an error message to stderr and exits the program.
func (p *DatabasePrinter) PrintError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

// PrintRetryingToResolve prints a message indicating that the program is retrying to resolve the hostname.
func (p *DatabasePrinter) PrintRetryingToResolve(s *statistics.Statistics) {
	fmt.Printf("Retrying to resolve %s\n", s.Hostname)
}

// PrintStatistics saves TCPing statistics to the database.
// If an error occurs while saving, it logs the error.
func (p *DatabasePrinter) PrintStatistics(s *statistics.Statistics) {
	data := dbStats{
		eventType:                StatisticsEvent,
		timestamp:                time.Now().Format(time.DateTime),
		ipAddr:                   s.IP.String(),
		hostname:                 s.Hostname,
		port:                     s.Port,
		totalSuccessfulPackets:   s.TotalSuccessfulProbes,
		totalUnsuccessfulPackets: s.TotalUnsuccessfulProbes,
		startTimestamp:           s.StartTime.Format(time.DateTime),
		totalUptime:              statistics.DurationToString(s.TotalUptime),
		totalDowntime:            statistics.DurationToString(s.TotalDowntime),
		totalPackets:             s.TotalSuccessfulProbes + s.TotalUnsuccessfulProbes,
	}

	if len(s.HostnameChanges) > 1 {
		for i := 0; i < len(s.HostnameChanges)-1; i++ {
			if s.HostnameChanges[i].Addr.String() == "" {
				continue
			}

			data.hostnameChanges += fmt.Sprintf("from %s to %s at %v\n",
				s.HostnameChanges[i].Addr.String(),
				s.HostnameChanges[i+1].Addr.String(),
				s.HostnameChanges[i+1].When.Format(time.DateTime),
			)
		}
	}

	totalPackets := s.TotalSuccessfulProbes + s.TotalUnsuccessfulProbes
	packetLoss := (float32(s.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	data.totalPacketLossPercent = fmt.Sprintf("%.2f", packetLoss)

	if !s.LastSuccessfulProbe.IsZero() {
		data.lastSuccessfulProbe = s.LastSuccessfulProbe.Format(time.DateTime)
	}

	if !s.LastUnsuccessfulProbe.IsZero() {
		data.lastUnsuccessfulProbe = s.LastUnsuccessfulProbe.Format(time.DateTime)
	}

	if s.LongestUp.Duration != 0 {
		data.longestUptime = fmt.Sprintf("%.0f", s.LongestUp.Duration.Seconds())
		data.longestConsecutiveUptimeStart = s.LongestUp.Start.Format(time.DateTime)
		data.longestConsecutiveUptimeEnd = s.LongestUp.End.Format(time.DateTime)
	}

	if s.LongestDown.Duration != 0 {
		data.longestDowntime = fmt.Sprintf("%.0f", s.LongestDown.Duration.Seconds())
		data.longestConsecutiveDowntimeStart = s.LongestDown.Start.Format(time.DateTime)
		data.longestConsecutiveDowntimeEnd = s.LongestDown.End.Format(time.DateTime)
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
		data.endTimestamp = s.EndTime.Format(time.DateTime)
	}

	totalDuration := s.TotalDowntime + s.TotalUptime
	data.totalDuration = fmt.Sprintf("%.0f", totalDuration.Seconds())

	if err := sqlitex.Execute(
		p.Conn,
		fmt.Sprintf(statsTableInsertSchema, p.statsTableName),
		&sqlitex.ExecOptions{Args: data.toArgs()},
	); err != nil {
		p.PrintError("Failed writing statistics to database: %s\n", err)
	}

	fmt.Printf("\nProbe and statistics data for %q have been saved to the table %q and %q, respectively\n",
		s.Hostname,
		p.probeTableName,
		p.statsTableName,
	)
}

// PrintTotalDownTime satisfies the "printer" interface but does nothing in this implementation
func (p *DatabasePrinter) PrintTotalDownTime(_ *statistics.Statistics) {}
