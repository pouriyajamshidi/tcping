package main

import (
	"fmt"
	"math"
	"net/netip"
	"strconv"
	"testing"
	"time"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func TestNewDBTableCreation(t *testing.T) {
	arg := []string{"localhost", "8001"}
	db := newDB(":memory:", arg)
	defer db.conn.Close()

	query := "SELECT name FROM sqlite_master WHERE type='table';"
	err := sqlitex.Execute(db.conn, query, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			Equals(t, stmt.ColumnCount(), 1)
			Equals(t, stmt.ColumnText(0), db.tableName)
			return nil
		},
	})

	isNil(t, err)
}

func TestDbSaveStats(t *testing.T) {
	// There are many fields, so many things could go wrong; that's why this elaborate test.
	arg := []string{"localhost", "8001"}
	db := newDB(":memory:", arg)
	t.Log(db.tableName)
	defer db.conn.Close()

	stat := mockStats()
	err := db.saveStats(stat)
	isNil(t, err)

	query := `SELECT
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
total_duration
FROM ` + fmt.Sprintf("%s WHERE event_type = '%s'", db.tableName, eventTypeStatistics)

	var (
		addr, sourceAddr, hostname, port               string
		hostNameResolveTries                           int
		totalSuccessfulProbes, totalUnsuccessfulProbes uint
		neverSucceedProbe, neverFailedProbe            bool
		lastSuccessfulProbe                            time.Time
		totalPackets                                   uint
		totalPacketsLoss                               float32
		totalUptime, totalDowntime                     string
		longestUptime                                  string
		longestUptimeStart, longestUptimeEnd           string
		longestDowntime                                string
		longestDowntimeStart, longestDowntimeEnd       string
		lMin, lAvg, lMax                               float32
		startTimestamp, endTimestamp                   time.Time
		totalDuration                                  string
	)

	resFunc := func(stmt *sqlite.Stmt) error {
		Equals(t, stmt.ColumnCount(), 27)
		var err error

		// addr
		addr = stmt.ColumnText(0)
		// source address
		sourceAddr = stmt.ColumnText(1)
		// hostname
		hostname = stmt.ColumnText(2)
		// port
		port = stmt.ColumnText(3)
		// hostname_resolve_retries
		hostNameResolveTries = stmt.ColumnInt(4)
		// total_successful_probes
		totalSuccessfulProbes = uint(stmt.ColumnInt(5))
		// total_unsuccessful_probes
		totalUnsuccessfulProbes = uint(stmt.ColumnInt(6))
		// never_succeed_probe
		neverSucceedProbe = stmt.ColumnBool(7)
		// never_failed_probe
		neverFailedProbe = stmt.ColumnBool(8)
		// last_successful_probe
		lastSuccessfulProbe, err = time.Parse(timeFormat, stmt.ColumnText(9))
		isNil(t, err)
		// last_unsuccessful_probe
		Equals(t, "", stmt.ColumnText(10)) // simulating never failed
		// isNil(t, err)
		// total_packets
		totalPackets = uint(stmt.ColumnInt(11))
		// total_packet_loss
		totalPacketsLoss = float32(stmt.ColumnFloat(12))
		// total_uptime
		totalUptime = stmt.ColumnText(13)
		// total_downtime
		totalDowntime = stmt.ColumnText(14)
		// longest_uptime
		longestUptime = stmt.ColumnText(15)
		// longest_uptime_start
		longestUptimeStart = stmt.ColumnText(16)
		// longest_uptime_end
		longestUptimeEnd = stmt.ColumnText(17)
		// longest_downtime
		longestDowntime = stmt.ColumnText(18)
		// longest_downtime_start
		longestDowntimeStart = stmt.ColumnText(19)
		// longest_downtime_end
		longestDowntimeEnd = stmt.ColumnText(20)
		// latency_min
		lMin = float32(stmt.ColumnFloat(21))
		// latency_avg
		lAvg = float32(stmt.ColumnFloat(22))
		// latency_max
		lMax = float32(stmt.ColumnFloat(23))
		// start_time
		startTimestamp, err = time.Parse(timeFormat, stmt.ColumnText(24))
		isNil(t, err)
		// end_time
		endTimestamp, err = time.Parse(timeFormat, stmt.ColumnText(25))
		isNil(t, err)
		// total_duration
		totalDuration = stmt.ColumnText(26)
		return nil
	}

	err = sqlitex.Execute(db.conn, query, &sqlitex.ExecOptions{
		ResultFunc: resFunc,
	})
	isNil(t, err)

	stat.rttResults.min = toFixedFloat(stat.rttResults.min, 3)
	stat.rttResults.average = toFixedFloat(stat.rttResults.average, 3)
	stat.rttResults.max = toFixedFloat(stat.rttResults.max, 3)

	Equals(t, addr, stat.userInput.ip.String())
	Equals(t, sourceAddr, "source address")
	Equals(t, hostname, stat.userInput.hostname)
	Equals(t, totalUnsuccessfulProbes, stat.totalUnsuccessfulProbes)
	Equals(t, totalSuccessfulProbes, stat.totalSuccessfulProbes)
	Equals(t, port, strconv.Itoa(int(stat.userInput.port)))

	Equals(t, hostNameResolveTries, int(stat.retriedHostnameLookups))
	packetLoss := (float32(stat.totalUnsuccessfulProbes) / float32(stat.totalSuccessfulProbes+stat.totalUnsuccessfulProbes)) * 100
	Equals(t, totalPacketsLoss, packetLoss)

	Equals(t, neverSucceedProbe, stat.lastSuccessfulProbe.IsZero())
	Equals(t, neverFailedProbe, stat.lastUnsuccessfulProbe.IsZero())

	Equals(t, lastSuccessfulProbe.Format(timeFormat), stat.lastSuccessfulProbe.Format(timeFormat))

	Equals(t, lMin, stat.rttResults.min)
	Equals(t, lAvg, stat.rttResults.average)
	Equals(t, lMax, stat.rttResults.max)
	Equals(t, startTimestamp.Format(timeFormat), stat.startTime.Format(timeFormat))
	Equals(t, endTimestamp.Format(timeFormat), stat.endTime.Format(timeFormat))

	actualDuration := stat.endTime.Sub(stat.startTime).String()
	Equals(t, totalDuration, actualDuration)
	Equals(t, totalUptime, stat.totalUptime.String())
	Equals(t, totalDowntime, stat.totalDowntime.String())
	Equals(t, totalPackets, stat.totalSuccessfulProbes+stat.totalUnsuccessfulProbes)

	Equals(t, longestUptime, stat.longestUptime.duration.String())
	Equals(t, longestUptimeStart, stat.longestUptime.start.Format(timeFormat))
	Equals(t, longestUptimeEnd, stat.longestUptime.end.Format(timeFormat))

	Equals(t, longestDowntime, stat.longestDowntime.duration.String())
	Equals(t, longestDowntimeStart, stat.longestDowntime.start.Format(timeFormat))
	Equals(t, longestDowntimeEnd, stat.longestDowntime.end.Format(timeFormat))
}

func TestSaveHostname(t *testing.T) {
	// There are many fields, so many things could go wrong; that's why this elaborate test.
	arg := []string{"local-.host", "8001"}
	db := newDB(":memory:", arg)
	defer db.conn.Close()

	// Test if hostName is sanitized correctly
	expectedTableName := fmt.Sprintf("%s_%s_%s", "local__host", "8001", time.Now().Format("15_04_05_01_02_2006"))
	Equals(t, db.tableName, expectedTableName)

	// Ensure the table is created correctly
	dropQuery := fmt.Sprintf("DROP TABLE IF EXISTS %s;", db.tableName)
	err := sqlitex.Execute(db.conn, dropQuery, nil)
	isNil(t, err)

	createQuery := fmt.Sprintf("CREATE TABLE %s (id INTEGER PRIMARY KEY, event_type TEXT NOT NULL, hostname_changed_to TEXT, hostname_change_time TEXT);", db.tableName)
	err = sqlitex.Execute(db.conn, createQuery, nil)
	isNil(t, err)

	stat := mockStats()

	err = db.saveHostNameChange(stat.hostnameChanges)
	isNil(t, err)

	// testing the host names if they are properly written
	query := `SELECT
		hostname_changed_to, hostname_change_time
		FROM ` + fmt.Sprintf("%s WHERE event_type IS '%s';", db.tableName, eventTypeHostnameChange)

	idx := 0
	err = sqlitex.Execute(db.conn, query, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			hostName := stmt.ColumnText(0)
			cTime := stmt.ColumnText(1)
			actualHost := stat.hostnameChanges[idx]
			idx++
			Equals(t, hostName, actualHost.Addr.String())
			Equals(t, cTime, actualHost.When.Format(timeFormat))

			return nil
		}})
	isNil(t, err)

	Equals(t, idx, len(stat.hostnameChanges))
}

func hostNameChange() []hostnameChange {
	ipAddresses := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334",
	}
	var hostNames []hostnameChange
	for i, ip := range ipAddresses {
		host := hostnameChange{
			Addr: netip.MustParseAddr(ip),
			When: time.Now().Add(time.Duration(i) * time.Minute),
		}
		hostNames = append(hostNames, host)
	}
	return hostNames
}

func mockStats() tcping {
	stat := tcping{
		startTime:           time.Now(),
		endTime:             time.Now().Add(10 * time.Minute),
		lastSuccessfulProbe: time.Now().Add(1 * time.Minute),
		// lastUnsuccessfulProbe is left with the default value "0" to simulate no probe failed
		retriedHostnameLookups: 10,
		longestUptime: longestTime{
			start:    time.Now().Add(20 * time.Second),
			end:      time.Now().Add(80 * time.Second),
			duration: time.Minute,
		},
		longestDowntime: longestTime{
			start:    time.Now().Add(20 * time.Second),
			end:      time.Now().Add(140 * time.Second),
			duration: time.Minute * 2,
		},
		userInput: userInput{
			ip:       netip.MustParseAddr("192.168.1.1"),
			hostname: "example.com",
			port:     1234,
		},
		totalUptime:             time.Second * 32,
		totalDowntime:           time.Second * 60,
		totalSuccessfulProbes:   201,
		totalUnsuccessfulProbes: 123,
		rttResults: rttResult{
			min:     2.832,
			average: 3.8123,
			max:     4.0932,
		},

		hostnameChanges: hostNameChange(),
	}

	return stat
}

// Equals compares two values.
// This is for avoiding code duplications.
func Equals[T comparable](t *testing.T, value, want T) {
	t.Helper()
	if want != value {
		t.Errorf("wanted %v; got %v", want, value)
		t.FailNow()
	}
}

// isNil compares a value to nil, in some cases you may need to do `Equals(t, value, nil)`
func isNil(t *testing.T, value any) {
	t.Helper()

	if value != nil {
		t.Logf(`expected "%v" to be nil`, value)
		t.FailNow()
	}
}

// toFixedFloat takes in a float and the precision number
// and will round it to that specified precision
// works for small numbers
//
// example: toFixedFloat(3.14159, 3) -> 3.142
func toFixedFloat(input float32, precision int) float32 {
	num := float64(input)
	round := func(num float64) int {
		return int(num + math.Copysign(0.5, num))
	}

	output := math.Pow(10, float64(precision))
	return float32(float64(round(num*output)) / output)
}
