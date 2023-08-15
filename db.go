package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strings"
	"time"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
)

type saveDb struct {
	db        *sql.DB
	tableName string
}

const (
	dbLocation              = "tcping.db"
	eventTypeStatistics     = "statistics"
	eventTypeHostnameChange = "hostname change"
	tableSchema             = `
CREATE TABLE %s (
    id INTEGER PRIMARY KEY,
    event_type TEXT NOT NULL, -- for the data type
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

func newDb(args []string, dbPath string) saveDb {
	tableName := newTableNames(args)
	tableSchema := fmt.Sprintf(tableSchema, tableName)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(tableSchema)
	if err != nil {
		log.Fatal(err)
	}

	return saveDb{db, tableName}
}

func newTableNames(args []string) string {
	if len(args) < 2 {
		usage()
	}
	tableName := fmt.Sprintf("%s_%s_%s", strings.ReplaceAll(args[0], ".", "_"), args[1], time.Now().Format("15_04_05_01_02_2006"))
	if unicode.IsNumber(rune(tableName[0])) {
		tableName = "_" + tableName
	}
	return tableName
}

func (s saveDb) saveToDb(query string, arg ...any) {
	statement := fmt.Sprintf(query, s.tableName)
	if _, err := s.db.Exec(statement, arg...); err != nil {
		log.Fatal(err)
	}
}

func (s saveDb) saveHostNameChange(h []hostnameChange) {
	for _, host := range h {
		if host.Addr.String() == "" {
			continue
		}
		schema := `INSERT INTO %s
	(event_type, hostname_changed_to, hostname_change_time)
	VALUES (?, ?, ?)`

		s.saveToDb(schema, eventTypeHostnameChange, host.Addr.String(), host.When)
	}
}

func (s saveDb) printStatistics(stat stats) {
	schema := `INSERT INTO %s (event_type, timestamp,
		addr, hostname, port, hostname_resolve_tries,
		total_successful_probes, total_unsuccessful_probes,
		last_successful_probe, last_unsuccessful_probe,
		total_packets, total_packet_loss,
		total_uptime, total_downtime,
		longest_uptime, longest_uptime_end, longest_uptime_start,
		longest_downtime, longest_downtime_start, longest_downtime_end,
		latency_min, latency_avg, latency_max,
		start_timestamp, end_timestamp, total_duration)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	totalPackets := stat.totalSuccessfulProbes + stat.totalUnsuccessfulProbes
	packetLoss := (float32(stat.totalUnsuccessfulProbes) / float32(totalPackets)) * 100
	if math.IsNaN(float64(packetLoss)) {
		packetLoss = 0
	}

	s.saveToDb(schema,
		eventTypeStatistics, time.Now(),
		stat.userInput.ip.String(), stat.userInput.hostname, stat.userInput.port, stat.retriedHostnameLookups,
		stat.totalSuccessfulProbes, stat.totalUnsuccessfulProbes,
		stat.lastSuccessfulProbe, stat.lastUnsuccessfulProbe,
		totalPackets, packetLoss,
		stat.totalUptime.String(), stat.totalDowntime.String(),
		stat.longestUptime.duration.String(), stat.longestUptime.start, stat.longestUptime.end,
		stat.longestDowntime.duration.String(), stat.longestDowntime.start, stat.longestDowntime.end,
		stat.rttResults.min, stat.rttResults.average, stat.rttResults.max,
		stat.startTime, stat.endTime, stat.endTime.Sub(stat.startTime).String(),
	)

	s.saveHostNameChange(stat.hostnameChanges)

	colorYellow("\noutput has been written to %q\nIn the table %q\n", dbLocation, s.tableName)
}

func (s saveDb) printStart(hostname string, port uint16)                                      {}
func (s saveDb) printProbeSuccess(hostname, ip string, port uint16, streak uint, rtt float32) {}
func (s saveDb) printProbeFail(hostname, ip string, port uint16, streak uint)                 {}
func (s saveDb) printRetryingToResolve(hostname string)                                       {}
func (s saveDb) printTotalDownTime(downtime time.Duration)                                    {}
func (s saveDb) printVersion()                                                                {}
func (s saveDb) printInfo(format string, args ...any)                                         {}
func (s saveDb) printError(format string, args ...any)                                        {}
