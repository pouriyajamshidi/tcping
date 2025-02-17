// Package printers contains the logic for printing information
package printers

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/pouriyajamshidi/tcping/v2/consts"
	"github.com/pouriyajamshidi/tcping/v2/internal/utils"
	"github.com/pouriyajamshidi/tcping/v2/types"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

// EventType is a special type for each method
// in the printer interface so that automatic tools
// can understand what kind of an event they've received.
// For instance, probe vs statistics...
type EventType string

const (
	eventTypeProbe          EventType = "probe"
	eventTypeStatistics     EventType = "statistics"
	eventTypeHostnameChange EventType = "hostname change"
)

// TODO: These should be sorted. very unlikely it works as is

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
		hostname_changes TEXT,
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

func (d *dbData) toArgs() []interface{} {
	return []interface{}{
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
	hostnameChanges                 []types.HostnameChange
	lastSuccessfulProbe             string
	lastUnsuccessfulProbe           string
	longestConsecutiveUptimeStart   string
	longestConsecutiveUptimeEnd     string
	longestConsecutiveDowntimeStart string
	longestConsecutiveDowntimeEnd   string
	latency                         float32
	latencyMin                      string
	latencyAvg                      string
	latencyMax                      string
	startTimestamp                  string
	endTimestamp                    string
}

func (d *dbStats) toArgs() []interface{} {
	return []interface{}{
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
		d.latency,
		d.latencyMin,
		d.latencyAvg,
		d.latencyMax,
		d.startTimestamp,
		d.endTimestamp,
	}
}

// DatabasePrinter represents a SQLite database connection for storing TCPing results.
type DatabasePrinter struct {
	Conn      *sqlite.Conn
	TableName string
	cfg       PrinterConfig
}

// NewDatabasePrinter initializes a new sqlite3 Database instance, creates the data table, and returns a pointer to it.
// If any error occurs during database creation or table initialization, the function exits the program.
func NewDatabasePrinter(cfg PrinterConfig) *DatabasePrinter {
	cfg.OutputDBPath = addDbExtension(cfg.OutputDBPath)

	conn, err := sqlite.OpenConn(cfg.OutputDBPath, sqlite.OpenCreate, sqlite.OpenReadWrite)
	if err != nil {
		consts.ColorRed("\nError creating the database %q: %s\n", cfg.OutputDBPath, err)
		os.Exit(1)
	}

	tableName := sanitizeTableName(cfg.Target, cfg.Port)
	tableSchema := fmt.Sprintf(dataTableSchema, tableName)

	err = sqlitex.Execute(conn, tableSchema, &sqlitex.ExecOptions{})
	if err != nil {
		consts.ColorRed("\nError creating the data table: %s\n", err)
		os.Exit(1)
	}

	statsTableSchema := fmt.Sprintf(statsTableSchema, tableName+"stats")
	err = sqlitex.Execute(conn, statsTableSchema, &sqlitex.ExecOptions{})
	if err != nil {
		consts.ColorRed("\nError creating the statistics table: %s\n", err)
		os.Exit(1)
	}

	return &DatabasePrinter{Conn: conn, TableName: tableName, cfg: cfg}
}

func addDbExtension(filename string) string {
	if strings.HasSuffix(filename, ".db") {
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

	sanitizedTime := strings.ReplaceAll(time.Now().Format(consts.TimeFormat), "-", "_")
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

// PrintStart prints a message indicating that TCPing has started for the given hostname and port.
func (p *DatabasePrinter) PrintStart(hostname string, port uint16) {
	fmt.Printf("TCPinging %s on port %d - saving results to: %s\n", hostname, port, p.cfg.OutputDBPath)
}

// PrintProbeSuccess satisfies the "printer" interface but does nothing in this implementation
func (p *DatabasePrinter) PrintProbeSuccess(startTime time.Time, sourceAddr string, opts types.Options, streak uint, rtt string) {
	if p.cfg.ShowFailuresOnly {
		return
	}

	timestamp := ""
	if p.cfg.WithTimestamp {
		timestamp = startTime.Format(consts.TimeFormat)
	}

	data := dbData{
		eventType:               probeEvent,
		success:                 "true",
		ongoingSuccessfulProbes: streak,
	}

	if opts.Hostname == opts.IP.String() {
		data.destIsIP = "true"

		if timestamp == "" {
			if p.cfg.WithSourceAddress {
				data.ipAddr = opts.IP.String()
				data.port = opts.Port
				data.sourceAddr = sourceAddr
				data.time = rtt
			} else {
				data.ipAddr = opts.IP.String()
				data.port = opts.Port
				data.time = rtt
			}
		} else {
			data.timestamp = timestamp

			if p.cfg.WithSourceAddress {
				data.ipAddr = opts.IP.String()
				data.port = opts.Port
				data.sourceAddr = sourceAddr
				data.time = rtt
			} else {
				data.ipAddr = opts.IP.String()
				data.port = opts.Port
				data.time = rtt
			}
		}
	} else {
		data.destIsIP = "false"

		if timestamp == "" {
			if p.cfg.WithSourceAddress {
				data.hostname = opts.Hostname
				data.ipAddr = opts.IP.String()
				data.port = opts.Port
				data.sourceAddr = sourceAddr
				data.time = rtt
			} else {
				data.hostname = opts.Hostname
				data.ipAddr = opts.IP.String()
				data.port = opts.Port
				data.time = rtt
			}
		} else {
			data.timestamp = timestamp

			if p.cfg.WithSourceAddress {
				data.ipAddr = opts.IP.String()
				data.port = opts.Port
				data.sourceAddr = sourceAddr
				data.time = rtt
			} else {
				data.ipAddr = opts.IP.String()
				data.port = opts.Port
				data.time = rtt
			}
		}
	}

	if err := sqlitex.Execute(
		p.Conn,
		fmt.Sprintf(dataTableInsertSchema, p.TableName),
		&sqlitex.ExecOptions{Args: data.toArgs()},
	); err != nil {
		p.PrintError("Failed writing probe success data to database: %s\n", err)
	}
}

// PrintProbeFail satisfies the "printer" interface but does nothing in this implementation
func (p *DatabasePrinter) PrintProbeFail(startTime time.Time, opts types.Options, streak uint) {
	timestamp := ""
	if p.cfg.WithTimestamp {
		timestamp = startTime.Format(consts.TimeFormat)
	}

	data := dbData{
		eventType:                 probeEvent,
		success:                   "false",
		ongoingUnsuccessfulProbes: streak,
	}

	if opts.Hostname == opts.IP.String() {
		data.destIsIP = "true"

		if timestamp == "" {
			data.ipAddr = opts.IP.String()
			data.port = opts.Port
		} else {
			data.timestamp = timestamp
			data.ipAddr = opts.IP.String()
			data.port = opts.Port
		}
	} else {
		data.destIsIP = "false"

		if timestamp == "" {
			data.hostname = opts.Hostname
			data.ipAddr = opts.IP.String()
			data.port = opts.Port
		} else {
			data.timestamp = timestamp
			data.hostname = opts.Hostname
			data.ipAddr = opts.IP.String()
			data.port = opts.Port
		}
	}

	if err := sqlitex.Execute(
		p.Conn,
		fmt.Sprintf(dataTableInsertSchema, p.TableName),
		&sqlitex.ExecOptions{Args: data.toArgs()},
	); err != nil {
		p.PrintError("Failed writing probe failure data to database: %s\n", err)
	}
}

// PrintStatistics saves TCPing statistics to the database.
// If an error occurs while saving, it logs the error.
func (p *DatabasePrinter) PrintStatistics(t types.Tcping) {
	data := dbStats{
		eventType:                statisticsEvent,
		timestamp:                time.Now().Format(consts.TimeFormat),
		ipAddr:                   t.Options.IP.String(),
		hostname:                 t.Options.Hostname,
		port:                     t.Options.Port,
		totalSuccessfulPackets:   t.TotalSuccessfulProbes,
		totalUnsuccessfulPackets: t.TotalUnsuccessfulProbes,
		startTimestamp:           t.StartTime.Format(consts.TimeFormat),
		totalUptime:              utils.DurationToString(t.TotalUptime),
		totalDowntime:            utils.DurationToString(t.TotalDowntime),
		totalPackets:             t.TotalSuccessfulProbes + t.TotalUnsuccessfulProbes,
	}

	if len(t.HostnameChanges) > 1 {
		data.hostnameChanges = t.HostnameChanges
	}

	totalPackets := t.TotalSuccessfulProbes + t.TotalUnsuccessfulProbes
	packetLoss := (float32(t.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100

	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	data.totalPacketLossPercent = fmt.Sprintf("%.2f", packetLoss)

	if !t.LastSuccessfulProbe.IsZero() {
		data.lastSuccessfulProbe = t.LastSuccessfulProbe.Format(consts.TimeFormat)
	}

	if !t.LastUnsuccessfulProbe.IsZero() {
		data.lastUnsuccessfulProbe = t.LastUnsuccessfulProbe.Format(consts.TimeFormat)
	}

	if t.LongestUptime.Duration != 0 {
		data.longestUptime = fmt.Sprintf("%.0f", t.LongestUptime.Duration.Seconds())
		data.longestConsecutiveUptimeStart = t.LongestUptime.Start.Format(consts.TimeFormat)
		data.longestConsecutiveUptimeEnd = t.LongestUptime.End.Format(consts.TimeFormat)
	}

	if t.LongestDowntime.Duration != 0 {
		data.longestDowntime = fmt.Sprintf("%.0f", t.LongestDowntime.Duration.Seconds())
		data.longestConsecutiveDowntimeStart = t.LongestDowntime.Start.Format(consts.TimeFormat)
		data.longestConsecutiveDowntimeEnd = t.LongestDowntime.End.Format(consts.TimeFormat)
	}

	if !t.DestIsIP {
		data.hostnameResolveRetries = t.RetriedHostnameLookups
	}

	if t.RttResults.HasResults {
		data.latencyMin = fmt.Sprintf("%.3f", t.RttResults.Min)
		data.latencyAvg = fmt.Sprintf("%.3f", t.RttResults.Average)
		data.latencyMax = fmt.Sprintf("%.3f", t.RttResults.Max)
	}

	if !t.EndTime.IsZero() {
		data.endTimestamp = t.EndTime.Format(consts.TimeFormat)
	}

	totalDuration := t.TotalDowntime + t.TotalUptime
	data.totalDuration = fmt.Sprintf("%.0f", totalDuration.Seconds())

	if err := sqlitex.Execute(
		p.Conn,
		fmt.Sprintf(dataTableInsertSchema, p.TableName+"stats"),
		&sqlitex.ExecOptions{Args: data.toArgs()},
	); err != nil {
		p.PrintError("Failed writing statistics to database: %s\n", err)
	}

	consts.ColorYellow("\nStatistics for %q have been saved to %q in the table %q\n", t.Options.Hostname, p.cfg.OutputDBPath, p.TableName+"stats")
}

func (p *DatabasePrinter) saveStats(tcping types.Tcping) error {
	totalPackets := tcping.TotalSuccessfulProbes + tcping.TotalUnsuccessfulProbes
	packetLoss := (float32(tcping.TotalUnsuccessfulProbes) / float32(totalPackets)) * 100
	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	// If the time is zero, that means it never failed.
	// In this case, the time should be empty instead of "0001-01-01 00:00:00".
	// Rather, it should be left empty.
	lastSuccessfulProbe := tcping.LastSuccessfulProbe.Format(consts.TimeFormat)
	var neverSucceedProbe, neverFailedProbe bool
	if tcping.LastSuccessfulProbe.IsZero() {
		lastSuccessfulProbe = ""
		neverSucceedProbe = true
	}
	lastUnsuccessfulProbe := tcping.LastUnsuccessfulProbe.Format(consts.TimeFormat)
	if tcping.LastUnsuccessfulProbe.IsZero() {
		lastUnsuccessfulProbe = ""
		neverFailedProbe = true
	}

	// if the longest uptime is empty, then the column should also be empty
	var longestUptimeDuration, longestUptimeStart, longestUptimeEnd string
	var longestDowntimeDuration, longestDowntimeStart, longestDowntimeEnd string
	longestUptimeDuration = "0s"
	longestDowntimeDuration = "0s"

	if !tcping.LongestUptime.Start.IsZero() {
		longestUptimeDuration = tcping.LongestUptime.Duration.String()
		longestUptimeStart = tcping.LongestUptime.Start.Format(consts.TimeFormat)
		longestUptimeEnd = tcping.LongestUptime.End.Format(consts.TimeFormat)
	}

	if !tcping.LongestDowntime.Start.IsZero() {
		longestDowntimeDuration = tcping.LongestDowntime.Duration.String()
		longestDowntimeStart = tcping.LongestDowntime.Start.Format(consts.TimeFormat)
		longestDowntimeEnd = tcping.LongestDowntime.End.Format(consts.TimeFormat)
	}

	var totalDuration string
	if tcping.EndTime.IsZero() {
		totalDuration = time.Since(tcping.StartTime).String()
	} else {
		totalDuration = tcping.EndTime.Sub(tcping.StartTime).String()
	}

	// TODO: Find a clean way to include source address
	// other printers utilize printProbeSuccess which takes the net.Conn
	// whereas DB is having its own way
	args := []interface{}{
		eventTypeStatistics,
		time.Now().Format(consts.TimeFormat),
		tcping.Options.IP.String(),
		"source address",
		tcping.Options.Hostname,
		tcping.Options.Port,
		tcping.RetriedHostnameLookups,
		tcping.TotalSuccessfulProbes,
		tcping.TotalUnsuccessfulProbes,
		neverSucceedProbe,
		neverFailedProbe,
		lastSuccessfulProbe,
		lastUnsuccessfulProbe,
		totalPackets,
		packetLoss,
		tcping.TotalUptime.String(),
		tcping.TotalDowntime.String(),
		longestUptimeDuration,
		longestUptimeStart,
		longestUptimeEnd,
		longestDowntimeDuration,
		longestDowntimeStart,
		longestDowntimeEnd,
		fmt.Sprintf("%.3f", tcping.RttResults.Min),
		fmt.Sprintf("%.3f", tcping.RttResults.Average),
		fmt.Sprintf("%.3f", tcping.RttResults.Max),
		tcping.StartTime.Format(consts.TimeFormat),
		tcping.EndTime.Format(consts.TimeFormat),
		totalDuration,
	}

	return sqlitex.Execute(
		p.Conn,
		fmt.Sprintf(dataTableInsertSchema, p.TableName),
		&sqlitex.ExecOptions{Args: args},
	)
}

// saveHostNameChang saves the hostname changes
// in multiple rows with event_type = eventTypeHostnameChange
func (p *DatabasePrinter) saveHostNameChange(h []types.HostnameChange) error {
	// %s will be replaced by the table name
	schema := `INSERT INTO %s
	(event_type, hostname_changed_to, hostname_change_time)
	VALUES (?, ?, ?)`

	for _, host := range h {
		if host.Addr.String() == "" {
			continue
		}
		err := sqlitex.Execute(p.Conn, fmt.Sprintf(schema, p.TableName), &sqlitex.ExecOptions{
			Args: []interface{}{eventTypeHostnameChange, host.Addr.String(), host.When.Format(consts.TimeFormat)}})
		if err != nil {
			return err
		}
	}

	return nil
}

// PrintError prints an error message to stderr and exits the program.
func (p *DatabasePrinter) PrintError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

// PrintRetryingToResolve satisfies the "printer" interface but does nothing in this implementation
func (p *DatabasePrinter) PrintRetryingToResolve(_ string) {}

// PrintTotalDownTime satisfies the "printer" interface but does nothing in this implementation
func (p *DatabasePrinter) PrintTotalDownTime(_ time.Duration) {}

// PrintInfo satisfies the "printer" interface but does nothing in this implementation
func (p *DatabasePrinter) PrintInfo(_ string, _ ...any) {}
