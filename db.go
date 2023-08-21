package main

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
)

type database struct {
	db        *sql.DB
	dbPath    string
	tableName string
}

const (
	eventTypeStatistics     = "statistics"
	eventTypeHostnameChange = "hostname change"
	tableSchema             = `
-- Organized row names together for better readability
CREATE TABLE %s (
    id INTEGER PRIMARY KEY,
    event_type TEXT NOT NULL, -- for the data type eg. statistics, hostname change
    timestamp DATETIME,
    addr TEXT,
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
)

// newDb creates a newDb with the given path and returns `database` struct
func newDb(args []string, dbPath string) database {
	tableName := newTableName(args)
	tableSchema := fmt.Sprintf(tableSchema, tableName)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		colorRed("\nError while creating the database %q: %s\n", dbPath, err)
		os.Exit(1)
	}

	_, err = db.Exec(tableSchema)
	if err != nil {
		colorRed("\nError while writing to the database %q \nerr: %s\n", dbPath, err)
		os.Exit(1)
	}

	return database{db, dbPath, tableName}
}

// newTableName will return correctly formatted table name
// formatting the table name as "example_com_port_hour_minute_sec_day_month_year"
// table name can't have '.' and can't start with numbers
func newTableName(args []string) string {
	tableName := fmt.Sprintf("%s_%s_%s", strings.ReplaceAll(args[0], ".", "_"), args[1], time.Now().Format("15_04_05_01_02_2006"))

	if unicode.IsNumber(rune(tableName[0])) {
		tableName = "_" + tableName
	}

	return tableName
}

// save will insert the table name and
// saves the args to the database
func (s database) save(query string, args ...any) error {
	// inserting the table name
	statement := fmt.Sprintf(query, s.tableName)

	// saving to the db
	_, err := s.db.Exec(statement, args...)

	return err
}

// saveStats saves stats to the dedbase with proper fomatting
func (s database) saveStats(stat stats) error {
	// %s will be replaced by the table name
	schema := `INSERT INTO %s (event_type, timestamp,
		addr, hostname, port, hostname_resolve_retries,
		total_successful_probes, total_unsuccessful_probes,
		never_succeed_probe, never_failed_probe,
		last_successful_probe, last_unsuccessful_probe,
		total_packets, total_packet_loss,
		total_uptime, total_downtime,
		longest_uptime, longest_uptime_end, longest_uptime_start,
		longest_downtime, longest_downtime_start, longest_downtime_end,
		latency_min, latency_avg, latency_max,
		start_time, end_time, total_duration)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	totalPackets := stat.totalSuccessfulProbes + stat.totalUnsuccessfulProbes
	packetLoss := (float32(stat.totalUnsuccessfulProbes) / float32(totalPackets)) * 100
	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	// If the time is zero, that means it never failed.
	// In this case, the time should be empty instead of "0001-01-01 00:00:00".
	// Rather, it should be left empty.
	lastSuccessfulProbe := stat.lastSuccessfulProbe.Format(timeFormat)
	var neverSucceedProbe, neverFailedProbe bool
	if stat.lastSuccessfulProbe.IsZero() {
		lastSuccessfulProbe = ""
		neverSucceedProbe = true
	}
	lastUnsuccessfulProbe := stat.lastUnsuccessfulProbe.Format(timeFormat)
	if stat.lastUnsuccessfulProbe.IsZero() {
		lastUnsuccessfulProbe = ""
		neverFailedProbe = true
	}

	var totalDuration string
	if stat.endTime.IsZero() {
		totalDuration = time.Since(stat.startTime).String()
	} else {
		totalDuration = stat.endTime.Sub(stat.startTime).String()
	}

	err := s.save(schema,
		eventTypeStatistics, time.Now().Format(timeFormat),
		stat.userInput.ip.String(), stat.userInput.hostname, stat.userInput.port, stat.retriedHostnameLookups,
		stat.totalSuccessfulProbes, stat.totalUnsuccessfulProbes,
		neverSucceedProbe, neverFailedProbe,
		lastSuccessfulProbe, lastUnsuccessfulProbe,
		totalPackets, packetLoss,
		stat.totalUptime.String(), stat.totalDowntime.String(),
		stat.longestUptime.duration.String(), stat.longestUptime.start.Format(timeFormat), stat.longestUptime.end.Format(timeFormat),
		stat.longestDowntime.duration.String(), stat.longestDowntime.start.Format(timeFormat), stat.longestDowntime.end.Format(timeFormat),
		stat.rttResults.min, stat.rttResults.average, stat.rttResults.max,
		stat.startTime.Format(timeFormat), stat.endTime.Format(timeFormat), totalDuration,
	)

	return err
}

// saveHostNameChang saves the hostname changes
// in multiple rows with event_type = eventTypeHostnameChange
func (s database) saveHostNameChange(h []hostnameChange) error {
	// %s will be replaced by the table name
	schema := `INSERT INTO %s
	(event_type, hostname_changed_to, hostname_change_time)
	VALUES (?, ?, ?)`

	for _, host := range h {
		if host.Addr.String() == "" {
			continue
		}
		err := s.save(schema, eventTypeHostnameChange, host.Addr.String(), host.When.Format(timeFormat))
		if err != nil {
			return err
		}
	}

	return nil
}

// printStart will let the user know the program is running by
// printing a msg with the hostname, and port number to stdout
func (s database) printStart(hostname string, port uint16) {
	fmt.Printf("TCPinging %s on port %d\n", hostname, port)
}

// printStatistics saves the statistics to the given database
// calls stat.printer.printError() on err
func (s database) printStatistics(stat stats) {
	err := s.saveStats(stat)
	if err != nil {
		s.printError("\nError while writing stats to the database %q\nerr: %s", s.dbPath, err)
	}

	// Hostname changes should be written during the final call.
	// If the endtime is 0, it indicates that this is not the last call.
	if !stat.endTime.IsZero() {
		err = s.saveHostNameChange(stat.hostnameChanges)
		if err != nil {
			s.printError("\nError while writing hostname changes to the database %q\nerr: %s", s.dbPath, err)
		}

	}

	colorYellow("\nStatistics for %q have been saved to %q in the table %q\n", stat.userInput.hostname, s.dbPath, s.tableName)
}

// printError prints the err to the stderr and exits with status code 1
func (s database) printError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

// Satisfying the "printer" interface.
func (s database) printProbeSuccess(hostname, ip string, port uint16, streak uint, rtt float32) {}
func (s database) printProbeFail(hostname, ip string, port uint16, streak uint)                 {}
func (s database) printRetryingToResolve(hostname string)                                       {}
func (s database) printTotalDownTime(downtime time.Duration)                                    {}
func (s database) printVersion()                                                                {}
func (s database) printInfo(format string, args ...any)                                         {}
