// db.go outputs data in sqlite3 format
package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"
	"unicode"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type database struct {
	conn      *sqlite.Conn
	dbPath    string
	tableName string
}

const (
	eventTypeStatistics     = "statistics"
	eventTypeHostnameChange = "hostname change"

	tableSchema = `
CREATE TABLE %s (
    id INTEGER PRIMARY KEY,
    event_type TEXT NOT NULL, -- for the data type eg. statistics, hostname change
    timestamp DATETIME,
    addr TEXT,
    sourceAddr TEXT,
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

	// %s will be replaced by the table name
	statSaveSchema = `INSERT INTO %s (
	event_type,
	timestamp,
	addr,
	sourceAddr,
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

// newDB creates a newDB with the given path and returns a pointer to the `database` struct
func newDB(dbPath string, args []string) *database {
	tableName := newTableName(args)
	tableSchema := fmt.Sprintf(tableSchema, tableName)

	conn, err := sqlite.OpenConn(dbPath, sqlite.OpenCreate, sqlite.OpenReadWrite)
	if err != nil {
		colorRed("\nError while creating the database %q: %s\n", dbPath, err)
		os.Exit(1)
	}

	err = sqlitex.Execute(conn, tableSchema, &sqlitex.ExecOptions{})
	if err != nil {
		colorRed("\nError writing to the database %q \nerr: %s\n", dbPath, err)
		os.Exit(1)
	}
	return &database{conn, dbPath, tableName}
}

// newTableName will return correctly formatted table name
// formatting the table name as "example_com_port_hour_minute_sec_day_month_year"
// table name can't have '.','-' and can't start with numbers
func newTableName(args []string) string {
	sanitizedHost := strings.ReplaceAll(args[0], ".", "_")
	sanitizedHost = strings.ReplaceAll(sanitizedHost, "-", "_")
	tableName := fmt.Sprintf("%s_%s_%s", sanitizedHost, args[1], time.Now().Format("15_04_05_01_02_2006"))

	if unicode.IsNumber(rune(tableName[0])) {
		tableName = "_" + tableName
	}

	return tableName
}

// saveStats saves stats to the database with proper formatting
func (db *database) saveStats(tcping tcping) error {
	totalPackets := tcping.totalSuccessfulProbes + tcping.totalUnsuccessfulProbes
	packetLoss := (float32(tcping.totalUnsuccessfulProbes) / float32(totalPackets)) * 100
	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	// If the time is zero, that means it never failed.
	// In this case, the time should be empty instead of "0001-01-01 00:00:00".
	// Rather, it should be left empty.
	lastSuccessfulProbe := tcping.lastSuccessfulProbe.Format(timeFormat)
	var neverSucceedProbe, neverFailedProbe bool
	if tcping.lastSuccessfulProbe.IsZero() {
		lastSuccessfulProbe = ""
		neverSucceedProbe = true
	}
	lastUnsuccessfulProbe := tcping.lastUnsuccessfulProbe.Format(timeFormat)
	if tcping.lastUnsuccessfulProbe.IsZero() {
		lastUnsuccessfulProbe = ""
		neverFailedProbe = true
	}

	// if the longest uptime is empty, then the column should also be empty
	var longestUptimeDuration, longestUptimeStart, longestUptimeEnd string
	var longestDowntimeDuration, longestDowntimeStart, longestDowntimeEnd string
	longestUptimeDuration = "0s"
	longestDowntimeDuration = "0s"

	if !tcping.longestUptime.start.IsZero() {
		longestUptimeDuration = tcping.longestUptime.duration.String()
		longestUptimeStart = tcping.longestUptime.start.Format(timeFormat)
		longestUptimeEnd = tcping.longestUptime.end.Format(timeFormat)
	}

	if !tcping.longestDowntime.start.IsZero() {
		longestDowntimeDuration = tcping.longestDowntime.duration.String()
		longestDowntimeStart = tcping.longestDowntime.start.Format(timeFormat)
		longestDowntimeEnd = tcping.longestDowntime.end.Format(timeFormat)
	}

	var totalDuration string
	if tcping.endTime.IsZero() {
		totalDuration = time.Since(tcping.startTime).String()
	} else {
		totalDuration = tcping.endTime.Sub(tcping.startTime).String()
	}

	// TODO: Find a clean way to include source address
	// other printers utilize printProbeSuccess which takes the net.Conn
	// whereas DB is having its own way
	args := []interface{}{
		eventTypeStatistics,
		time.Now().Format(timeFormat),
		tcping.userInput.ip.String(),
		"source address",
		tcping.userInput.hostname,
		tcping.userInput.port,
		tcping.retriedHostnameLookups,
		tcping.totalSuccessfulProbes,
		tcping.totalUnsuccessfulProbes,
		neverSucceedProbe,
		neverFailedProbe,
		lastSuccessfulProbe,
		lastUnsuccessfulProbe,
		totalPackets,
		packetLoss,
		tcping.totalUptime.String(),
		tcping.totalDowntime.String(),
		longestUptimeDuration,
		longestUptimeStart,
		longestUptimeEnd,
		longestDowntimeDuration,
		longestDowntimeStart,
		longestDowntimeEnd,
		fmt.Sprintf("%.3f", tcping.rttResults.min),
		fmt.Sprintf("%.3f", tcping.rttResults.average),
		fmt.Sprintf("%.3f", tcping.rttResults.max),
		tcping.startTime.Format(timeFormat),
		tcping.endTime.Format(timeFormat),
		totalDuration,
	}

	return sqlitex.Execute(
		db.conn,
		fmt.Sprintf(statSaveSchema, db.tableName),
		&sqlitex.ExecOptions{Args: args},
	)
}

// saveHostNameChang saves the hostname changes
// in multiple rows with event_type = eventTypeHostnameChange
func (db *database) saveHostNameChange(h []hostnameChange) error {
	// %s will be replaced by the table name
	schema := `INSERT INTO %s
	(event_type, hostname_changed_to, hostname_change_time)
	VALUES (?, ?, ?)`

	for _, host := range h {
		if host.Addr.String() == "" {
			continue
		}
		err := sqlitex.Execute(db.conn, fmt.Sprintf(schema, db.tableName), &sqlitex.ExecOptions{
			Args: []interface{}{eventTypeHostnameChange, host.Addr.String(), host.When.Format(timeFormat)}})
		if err != nil {
			return err
		}
	}

	return nil
}

// printStart will let the user know the program is running by
// printing a msg with the hostname, and port number to stdout
func (db *database) printStart(hostname string, port uint16) {
	fmt.Printf("TCPinging %s on port %d\n", hostname, port)
}

// printStatistics saves the statistics to the given database
// calls stat.printer.printError() on err
func (db *database) printStatistics(tcping tcping) {
	err := db.saveStats(tcping)
	if err != nil {
		db.printError("\nError while writing stats to the database %q\nerr: %s", db.dbPath, err)
	}

	// Hostname changes should be written during the final call.
	// If the endTime is 0, it indicates that this is not the last call.
	if !tcping.endTime.IsZero() {
		err = db.saveHostNameChange(tcping.hostnameChanges)
		if err != nil {
			db.printError("\nError while writing hostname changes to the database %q\nerr: %s", db.dbPath, err)
		}
	}

	colorYellow("\nStatistics for %q have been saved to %q in the table %q\n", tcping.userInput.hostname, db.dbPath, db.tableName)
}

// printError prints the err to the stderr and exits with status code 1
func (db *database) printError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

// Satisfying the "printer" interface.
func (db *database) printProbeSuccess(_ string, _ userInput, _ uint, _ float32) {}
func (db *database) printProbeFail(_ userInput, _ uint)                         {}
func (db *database) printRetryingToResolve(_ string)                            {}
func (db *database) printTotalDownTime(_ time.Duration)                         {}
func (db *database) printVersion()                                              {}
func (db *database) printInfo(_ string, _ ...any)                               {}
