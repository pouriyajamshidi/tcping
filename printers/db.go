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
	"github.com/pouriyajamshidi/tcping/v2/types"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

const (
	eventTypeStatistics     = "statistics"
	eventTypeHostnameChange = "hostname change"
)

const (
	dataTableSchema = `CREATE TABLE %s (
    id INTEGER PRIMARY KEY,
    event_type TEXT NOT NULL, -- for the data type eg. statistics, hostname change
    timestamp DATETIME,
    ip_address TEXT,
    source_addr TEXT,
    hostname TEXT,
    port INTEGER,
    hostname_resolve_retries INTEGER,

    hostname_changed_to TEXT,
    hostname_change_time DATETIME,

    latency_min REAL,
    latency_avg REAL,
    latency_max REAL,

	total_duration TEXT,
    start_time DATETIME,
    end_time DATETIME,

	never_succeed_probe INTEGER, -- value will be 1 if a probe never succeeded
	never_failed_probe INTEGER, -- value will be 1 if a probe never failed
    last_successful_probe DATETIME,
    last_unsuccessful_probe DATETIME,

    longest_uptime TEXT,
    longest_uptime_start DATETIME,
    longest_uptime_end DATETIME,

    longest_downtime TEXT,
    longest_downtime_start DATETIME,
    longest_downtime_end DATETIME,

    total_packets INTEGER,
    total_packet_loss REAL,
    total_successful_probes INTEGER,
    total_unsuccessful_probes INTEGER,

    total_uptime TEXT,
    total_downtime TEXT
	);`

	// SQL statement for inserting statistics into the table
	statSaveSchema = `INSERT INTO %s (
	event_type,
	timestamp,
	ip_address,
	source_addr,
	hostname,
	port,
	hostname_resolve_retries,
	total_successful_probes,
	total_unsuccessful_probes,
	never_succeed_probe,
	never_failed_probe,
	last_successful_probe,
	last_unsuccessful_probe,
	total_packets,
	total_packet_loss,
	total_uptime,
	total_downtime,
	longest_uptime,
	longest_uptime_start,
	longest_uptime_end,
	longest_downtime,
	longest_downtime_start,
	longest_downtime_end,
	latency_min,
	latency_avg,
	latency_max,
	start_time,
	end_time,
	total_duration) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
)

// DatabasePrinter represents a SQLite database connection for storing TCPing results.
type DatabasePrinter struct {
	Conn      *sqlite.Conn
	DbPath    string
	TableName string
}

// NewDatabasePrinter initializes a new sqlite3 Database instance, creates the data table, and returns a pointer to it.
// If any error occurs during database creation or table initialization, the function exits the program.
func NewDatabasePrinter(cfg PrinterConfig) *DatabasePrinter {
	filename := addDbExtension(cfg.OutputDBPath)

	conn, err := sqlite.OpenConn(filename, sqlite.OpenCreate, sqlite.OpenReadWrite)
	if err != nil {
		consts.ColorRed("\nError creating the database %q: %s\n", filename, err)
		os.Exit(1)
	}

	tableName := sanitizeTableName(cfg.Target, cfg.Port)
	tableSchema := fmt.Sprintf(dataTableSchema, tableName)

	err = sqlitex.Execute(conn, tableSchema, &sqlitex.ExecOptions{})
	if err != nil {
		consts.ColorRed("\nError creating the data table: %s\n", err)
		os.Exit(1)
	}

	return &DatabasePrinter{conn, filename, tableName}
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
func (db *DatabasePrinter) PrintStart(hostname string, port uint16) {
	fmt.Printf("TCPinging %s on port %d - saving results to: %s\n", hostname, port, db.DbPath)
}

// PrintProbeSuccess satisfies the "printer" interface but does nothing in this implementation
func (db *DatabasePrinter) PrintProbeSuccess(_ time.Time, _ string, _ types.Options, _ uint, _ string) {
}

// saveStats saves stats to the database with proper formatting
func (db *DatabasePrinter) saveStats(tcping types.Tcping) error {
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
		db.Conn,
		fmt.Sprintf(statSaveSchema, db.TableName),
		&sqlitex.ExecOptions{Args: args},
	)
}

// saveHostNameChang saves the hostname changes
// in multiple rows with event_type = eventTypeHostnameChange
func (db *DatabasePrinter) saveHostNameChange(h []types.HostnameChange) error {
	// %s will be replaced by the table name
	schema := `INSERT INTO %s
	(event_type, hostname_changed_to, hostname_change_time)
	VALUES (?, ?, ?)`

	for _, host := range h {
		if host.Addr.String() == "" {
			continue
		}
		err := sqlitex.Execute(db.Conn, fmt.Sprintf(schema, db.TableName), &sqlitex.ExecOptions{
			Args: []interface{}{eventTypeHostnameChange, host.Addr.String(), host.When.Format(consts.TimeFormat)}})
		if err != nil {
			return err
		}
	}

	return nil
}

// PrintStatistics saves TCPing statistics to the database.
// If an error occurs while saving, it logs the error.
func (db *DatabasePrinter) PrintStatistics(tcping types.Tcping) {
	err := db.saveStats(tcping)
	if err != nil {
		db.PrintError("\nError while writing stats to the database %q\nerr: %s", db.DbPath, err)
	}

	// Hostname changes should be written during the final call.
	// If the endTime is 0, it indicates that this is not the last call.
	if !tcping.EndTime.IsZero() {
		err = db.saveHostNameChange(tcping.HostnameChanges)
		if err != nil {
			db.PrintError("\nError while writing hostname changes to the database %q\nerr: %s", db.DbPath, err)
		}
	}

	consts.ColorYellow("\nStatistics for %q have been saved to %q in the table %q\n", tcping.Options.Hostname, db.DbPath, db.TableName)
}

// PrintError prints an error message to stderr and exits the program.
func (db *DatabasePrinter) PrintError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

// PrintProbeFail satisfies the "printer" interface but does nothing in this implementation
func (db *DatabasePrinter) PrintProbeFail(_ time.Time, _ types.Options, _ uint) {}

// PrintRetryingToResolve satisfies the "printer" interface but does nothing in this implementation
func (db *DatabasePrinter) PrintRetryingToResolve(_ string) {}

// PrintTotalDownTime satisfies the "printer" interface but does nothing in this implementation
func (db *DatabasePrinter) PrintTotalDownTime(_ time.Duration) {}

// PrintInfo satisfies the "printer" interface but does nothing in this implementation
func (db *DatabasePrinter) PrintInfo(_ string, _ ...any) {}
