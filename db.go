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

const (
	dbLocation  = "tcping.db"
	tableSchema = `
CREATE TABLE %s (
    id INTEGER PRIMARY KEY,
    type TEXT NOT NULL, -- for the data type
    timestamp DATETIME NOT NULL,

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

type saveDb struct {
	db        *sql.DB
	tableName string
}

func (s saveDb) saveToDb(query string, arg ...any) {
	statement := fmt.Sprintf(query, s.tableName)
	if _, err := s.db.Exec(statement, arg...); err != nil {
		log.Fatal(err)
	}
}

func (s saveDb) saveHostNameChange(h hostnameChange) {
	statement := fmt.Sprintf("INSERT INTO %s (ip, time) VALUES (?, ?)", s.tableName)
	if _, err := s.db.Exec(statement, h.Addr.String(), h.When); err != nil {
		log.Fatal(err)
	}
}

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

func (s saveDb) printStart(hostname string, port uint16) {
	// s.saveToDb("INSERT INTO %s (type, hostname, port, timestamp) VALUES (?, ?, ?, ?)", startEvent, hostname, port, time.Now())
}

func (s saveDb) printProbeSuccess(hostname, ip string, port uint16, streak uint, rtt float32) {
	// success := 1 // true
	// if hostname != "" {
	// 	isIp := 0
	// 	statement := "INSERT INTO %s (type, hostname, ip, port, is_ip, success, total_successful_probes, latency, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	// 	s.saveToDb(statement, probeEvent, hostname, ip, port, isIp, success, streak, rtt, time.Now())
	// } else {
	// 	isIp := 1
	// 	statement := "INSERT INTO %s (type, ip, port, is_ip, success, total_successful_probes, latency, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	// 	s.saveToDb(statement, probeEvent, ip, port, isIp, success, streak, rtt, time.Now())
	// }
}

func (s saveDb) printProbeFail(hostname, ip string, port uint16, streak uint) {
	// success := 0 // false
	// if hostname != "" {
	// 	isIp := 0
	// 	statement := "INSERT INTO %s (type, hostname, ip, port, is_ip, success, total_unsuccessful_probes, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	// 	s.saveToDb(statement, probeEvent, hostname, ip, port, isIp, success, streak, time.Now())
	// } else {
	// 	isIp := 1
	// 	statement := "INSERT INTO %s (type, ip, port, is_ip, success, total_unsuccessful_probes, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)"
	// 	s.saveToDb(statement, probeEvent, ip, port, isIp, success, streak, time.Now())
	// }
}

func (s saveDb) printRetryingToResolve(hostname string) {
	// statement := "INSERT INTO %s (type, hostname, timestamp) VALUES (?, ?, ?)"
	// s.saveToDb(statement, retryEvent, hostname, time.Now())
}
func (s saveDb) printTotalDownTime(downtime time.Duration) {
	// statement := "INSERT INTO %s (type, total_downtime, timestamp) VALUES (?, ?, ?)"
	// s.saveToDb(statement, retrySuccessEvent, downtime.Seconds(), time.Now())
}

const (
	typeStatistics     = "statistics"
	typeHostnameChange = "hostname change"
)

func (s saveDb) printStatistics(stat stats) {
	// hostname_changes,
	// hostname_changed_to,
	// hostname_change_time,
	schema := `INSERT INTO %s (type, timestamp,
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
		typeStatistics, time.Now(),
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

	// for _, host := range stat.hostnameChanges {
	// 	s.saveHostNameChange(host)
	// }

}

func (s saveDb) printVersion() {
	// statement := "INSERT INTO %s (type, message, timestamp) VALUES (?, ?, ?, ?)"
	// s.saveToDb(statement, retrySuccessEvent, "TCPING version "+version, time.Now())
}

func (s saveDb) printInfo(format string, args ...any) {
	// statement := "INSERT INTO %s (type, message, timestamp) VALUES (?, ?, ?)"
	// s.saveToDb(statement, infoEvent, fmt.Sprintf(format, args...), time.Now())
}

func (s saveDb) printError(format string, args ...any) {
	// statement := "INSERT INTO %s (type, message, timestamp) VALUES (?, ?, ?)"
	// s.saveToDb(statement, errorEvent, fmt.Sprintf(format, args...), time.Now())
}
