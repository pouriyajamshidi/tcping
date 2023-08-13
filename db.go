package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbLocation = "/tmp/tcping.db"

	tableSchema = `CREATE TABLE %s (
    id INTEGER PRIMARY KEY,
    type TEXT NOT NULL,
    message TEXT,
    timestamp DATETIME NOT NULL,
    ip TEXT,
    hostname TEXT,
    hostname_resolve_tries INTEGER,
    hostname_changes TEXT,
    is_ip INTEGER,
    port INTEGER,
    success INTEGER,
    latency REAL,
    latency_min TEXT,
    latency_avg TEXT,
    latency_max TEXT,
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
    total_packet_loss TEXT,
    total_packets INTEGER,
    total_successful_probes INTEGER,
    total_unsuccessful_probes INTEGER,
    total_uptime REAL,
    total_downtime REAL
);


CREATE TABLE %s (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip TEXT,
    time DATETIME
);`
)

type saveDb struct {
	db                  *sql.DB
	tableName           string
	hostnameChangeTable string
}

func (s saveDb) saveToDb(query string, arg ...any) {
	statement := fmt.Sprintf(query, s.tableName)
	if _, err := s.db.Exec(statement, arg...); err != nil {
		log.Fatal(err)
	}
}

func (s saveDb) saveHostNameChange(h hostnameChange) {
	statement := fmt.Sprintf("INSERT INTO %s (ip, time) VALUES (?, ?)", s.hostnameChangeTable)
	if _, err := s.db.Exec(statement, h.Addr.String(), h.When); err != nil {
		log.Fatal(err)
	}
}

func newDb(args []string, dbPath string) saveDb {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix | log.Lshortfile)
	if len(args) < 2 {
		usage()
	}

	tableName := fmt.Sprintf("%s_%s_%s", strings.ReplaceAll(args[0], ".", "_"), args[1], time.Now().Format("15_04_05_01_02_2006"))
	if unicode.IsNumber(rune(tableName[0])) {
		tableName = "_" + tableName
	}

	hostnameChangeTable := fmt.Sprintf("hostname_changes_%s", strings.TrimPrefix(tableName, "_"))
	schema := fmt.Sprintf(tableSchema, tableName, hostnameChangeTable)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}

	return saveDb{db, tableName, hostnameChangeTable}
}

func (s saveDb) printStart(hostname string, port uint16) {
	s.saveToDb("INSERT INTO %s (type, hostname, port, timestamp) VALUES (?, ?, ?, ?)", startEvent, hostname, port, time.Now())
}

func (s saveDb) printProbeSuccess(hostname, ip string, port uint16, streak uint, rtt float32) {
	success := 1 // true
	if hostname != "" {
		isIp := 0
		statement := "INSERT INTO %s (type, hostname, ip, port, is_ip, success, total_successful_probes, latency, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
		s.saveToDb(statement, probeEvent, hostname, ip, port, isIp, success, streak, rtt, time.Now())
	} else {
		isIp := 1
		statement := "INSERT INTO %s (type, ip, port, is_ip, success, total_successful_probes, latency, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
		s.saveToDb(statement, probeEvent, ip, port, isIp, success, streak, rtt, time.Now())
	}
}

func (s saveDb) printProbeFail(hostname, ip string, port uint16, streak uint) {
	success := 0 // false
	if hostname != "" {
		isIp := 0
		statement := "INSERT INTO %s (type, hostname, ip, port, is_ip, success, total_unsuccessful_probes, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
		s.saveToDb(statement, probeEvent, hostname, ip, port, isIp, success, streak, time.Now())
	} else {
		isIp := 1
		statement := "INSERT INTO %s (type, ip, port, is_ip, success, total_unsuccessful_probes, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)"
		s.saveToDb(statement, probeEvent, ip, port, isIp, success, streak, time.Now())
	}
}

// printRetryingToResolve should print a message with the hostname
// it is trying to resolve an ip for.
//
// This is only being printed when the -r flag is applied.
func (s saveDb) printRetryingToResolve(hostname string) {
	statement := "INSERT INTO %s (type, hostname, timestamp) VALUES (?, ?, ?)"
	s.saveToDb(statement, retryEvent, hostname, time.Now())
}

// printTotalDownTime should print a downtime duration.
//
// This is being called when host was unavailable for some time
// but the latest probe was successful (became available).
func (s saveDb) printTotalDownTime(downtime time.Duration) {
	statement := "INSERT INTO %s (type, total_downtime, timestamp) VALUES (?, ?, ?)"
	s.saveToDb(statement, retrySuccessEvent, downtime.Seconds(), time.Now())
}

/*
{
    "type": "statistics",
    "message": "stats for 10.0.0.101",
    "timestamp": "2023-08-13T18:32:39.271077641+06:00",
    "addr": "10.0.0.101", // ip
    "hostname": "10.0.0.101",
    "latency_min": "1.822",
    "latency_avg": "3.558",
    "latency_max": "7.529",
    "total_duration": "72",
    "start_timestamp": "2023-08-13T18:31:28.222860613+06:00",
    "end_timestamp": "2023-08-13T18:32:40.222860613+06:00",
    "last_successful_probe": "2023-08-13T18:32:39.223234255+06:00",
    "last_unsuccessful_probe": "2023-08-13T18:32:30.22447031+06:00",
    "longest_uptime": "62",
    "longest_uptime_end": "2023-08-13T18:32:30.223625011+06:00",
    "longest_uptime_start": "2023-08-13T18:31:28.223625011+06:00",
    "longest_downtime": "1",
    "longest_downtime_end": "2023-08-13T18:32:31.22447031+06:00",
    "longest_downtime_start": "2023-08-13T18:32:30.22447031+06:00",
    "total_packet_loss": "1.39",
    "total_packets": 72,
    "total_successful_probes": 71,
    "total_unsuccessful_probes": 1,
    "total_uptime": 71,
    "total_downtime": 1
}
*/

func (s saveDb) printStatistics(stat stats) {
	// this is rediculous
	schema := `INSERT INTO %s (type, message, ip, hostname, port,
        total_successful_probes, total_unsuccessful_probes,
        last_successful_probe, last_unsuccessful_probe,
        latency_min, latency_avg, latency_max,
        start_timestamp, end_timestamp,
        total_uptime, total_downtime,
		longest_uptime, longest_uptime_start, longest_uptime_end,
     	longest_downtime, longest_downtime_start, longest_downtime_end,
        total_packets,
		timestamp)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	s.saveToDb(schema,
		statisticsEvent, fmt.Sprintf("stats for %s", stat.userInput.hostname), stat.userInput.ip.String(), stat.userInput.hostname, stat.userInput.port,
		stat.totalSuccessfulProbes, stat.totalUnsuccessfulProbes,
		stat.lastSuccessfulProbe, stat.lastSuccessfulProbe,
		stat.rttResults.min, stat.rttResults.average, stat.rttResults.max,
		stat.startTime, stat.endTime,
		stat.totalUptime, stat.totalDowntime,
		stat.longestUptime.duration, stat.longestUptime.start, stat.longestUptime.end,
		stat.longestDowntime.duration, stat.longestDowntime.start, stat.longestUptime.end,
		stat.totalSuccessfulProbes+stat.totalUnsuccessfulProbes,
		time.Now(),
	)

	for _, host := range stat.hostnameChanges {
		s.saveHostNameChange(host)
	}

}

func (s saveDb) printVersion() {
	statement := "INSERT INTO %s (type, message, timestamp) VALUES (?, ?, ?, ?)"
	s.saveToDb(statement, retrySuccessEvent, "TCPING version "+version, time.Now())
}

func (s saveDb) printInfo(format string, args ...any) {
	statement := "INSERT INTO %s (type, message, timestamp) VALUES (?, ?, ?)"
	s.saveToDb(statement, infoEvent, fmt.Sprintf(format, args...), time.Now())
}

func (s saveDb) printError(format string, args ...any) {
	statement := "INSERT INTO %s (type, message, timestamp) VALUES (?, ?, ?)"
	s.saveToDb(statement, errorEvent, fmt.Sprintf(format, args...), time.Now())
}
