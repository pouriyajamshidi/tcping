//go:build exclude

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

type saveDb struct {
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
    hostname_resolve_tries INTEGER,

    hostname_changed_to TEXT,
    hostname_change_time DATETIME,

    latency_min REAL,
    latency_avg REAL,
    latency_max REAL,

	total_duration TEXT,
    start_timestamp DATETIME,
    end_timestamp DATETIME,

	never_succeed_probe INTEGER, -- value will be 1 if a prove never succeeded
	never_failed_probe INTEGER, -- value will be 1 if a prove never failed
    last_successful_probe DATETIME,
    last_unsuccessful_probe DATETIME,

    longest_uptime TEXT,
    longest_uptime_end DATETIME,
    longest_uptime_start DATETIME,

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

// newDb creas a newDb to the given path and returns `saveDb` struct
func newDb(args []string, dbPath string) saveDb {
	tableName := newTableName(args)
	tableSchema := fmt.Sprintf(tableSchema, tableName)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		colorRed("\nwhile creating the database %q: %s\n", dbPath, err)
		os.Exit(1)
	}

	_, err = db.Exec(tableSchema)
	if err != nil {
		colorRed("\nwhile writing to the detabase %q \nerr: %s\n", dbPath, err)
		os.Exit(1)
	}

	return saveDb{db, dbPath, tableName}
}

// newTableName will return correctly formatted table name
func newTableName(args []string) string {
	// table name can't have '.'
	// formating the table name "example_com_hour_minute_sec_day_month_year"
	tableName := fmt.Sprintf("%s_%s_%s", strings.ReplaceAll(args[0], ".", "_"), args[1], time.Now().Format("15_04_05_01_02_2006"))

	// table name can't start with numbers
	if unicode.IsNumber(rune(tableName[0])) {
		tableName = "_" + tableName
	}

	return tableName
}

// this will insert the table name and
// save the ags to the database
func (s saveDb) save(query string, args ...any) error {
	// inserting the table name
	statement := fmt.Sprintf(query, s.tableName)

	// saving to the db
	_, err := s.db.Exec(statement, args...)

	return err
}

// saves stats to the dedbase with proper fomatting
func (s saveDb) saveStats(stat stats) error {
	// %s will be replaced by the table name
	schema := `INSERT INTO %s (event_type, timestamp,
		addr, hostname, port, hostname_resolve_tries,
		total_successful_probes, total_unsuccessful_probes,
		never_succeed_probe, never_failed_probe,
		last_successful_probe, last_unsuccessful_probe,
		total_packets, total_packet_loss,
		total_uptime, total_downtime,
		longest_uptime, longest_uptime_end, longest_uptime_start,
		longest_downtime, longest_downtime_start, longest_downtime_end,
		latency_min, latency_avg, latency_max,
		start_timestamp, end_timestamp, total_duration)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	totalPackets := stat.totalSuccessfulProbes + stat.totalUnsuccessfulProbes
	packetLoss := (float32(stat.totalUnsuccessfulProbes) / float32(totalPackets)) * 100
	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	// if the time is zero that means never failsed
	// then the time should be emty not be "0001-01-01 00:00:00"
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
		stat.startTime.Format(timeFormat), stat.endTime.Format(timeFormat), stat.endTime.Sub(stat.startTime).String(),
	)

	return err
}

// As hostname changes as it's an array,
// it will be saved after statistics within multiple
// with event_type = eventTypeHostnameChange
func (s saveDb) saveHostNameChange(h []hostnameChange) error {
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

// it will let the user know the program is running by
// printing a msg with the hostname, and port number to stdout
func (s saveDb) printStart(hostname string, port uint16) {
	fmt.Printf("TCPinging %s on port %d\n", hostname, port)
}

// saves the statistics to the given database
// calls stat.printer.printError() on err
func (s saveDb) printStatistics(stat stats) {
	err := s.saveStats(stat)
	if err != nil {
		s.printError("\nwhile writing stats to the database %q\nerr: %s", s.dbPath, err)
	}
	err = s.saveHostNameChange(stat.hostnameChanges)
	if err != nil {
		s.printError("\nwhile writing hostname changes to the database %q\nerr: %s", s.dbPath, err)
	}

	fmt.Printf("\nStatistics for %q have been saved to %q in the table %q\n", stat.userInput.hostname, s.dbPath, s.tableName)
}

// prints the err to the stderr and exits with status code 1
func (s saveDb) printError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

// they are only here to satisfy the "printer" interface.
func (s saveDb) printProbeSuccess(hostname, ip string, port uint16, streak uint, rtt float32) {}
func (s saveDb) printProbeFail(hostname, ip string, port uint16, streak uint)                 {}
func (s saveDb) printRetryingToResolve(hostname string)                                       {}
func (s saveDb) printTotalDownTime(downtime time.Duration)                                    {}
func (s saveDb) printVersion()                                                                {}
func (s saveDb) printInfo(format string, args ...any)                                         {}
