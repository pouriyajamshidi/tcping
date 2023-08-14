package main

import (
	"net/netip"
	"strconv"
	"testing"
	"time"
)

func TestNewDB(t *testing.T) {
	arg := []string{"localhost", "8001"}
	s := newDb(arg, ":memory:")
	// testing if the table names are
	rows, err := s.db.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		t.Error(err)
		return
	}
	defer rows.Close()
	defer s.db.Close()

	var tNames []string
	for rows.Next() {
		var tblName string
		if err := rows.Scan(&tblName); err != nil {
			t.Error(err)
			return
		}
		// this is automatically created by SQLite
		if tblName == "sqlite_sequence" {
			continue
		}
		tNames = append(tNames, tblName)
	}

	if len(tNames) != 2 {
		t.Errorf("expected 2 tables; got %d", len(tNames))
		return
	}

	tableName, hostnameChangeTable := s.tableName, s.hostnameChangeTable
	if tNames[0] != tableName {
		t.Errorf("expected %s; got %s", tableName, tNames[0])
	}

	if tNames[1] != hostnameChangeTable {
		t.Errorf("expected %s; got %s", hostnameChangeTable, tNames[1])
	}
}

func TestDbHostnameSave(t *testing.T) {
	arg := []string{"localhost", "8001"}
	s := newDb(arg, ":memory:")

	ipAddresses := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334",
	}

	var hostNames []hostnameChange
	for _, ip := range ipAddresses {
		host := hostnameChange{
			Addr: netip.MustParseAddr(ip),
			When: time.Now(),
		}
		hostNames = append(hostNames, host)
	}

	for _, host := range hostNames {
		s.saveHostNameChange(host)
	}

	var dbIp []string

	rows, err := s.db.Query("SELECT ip FROM " + s.hostnameChangeTable)
	if err != nil {
		t.Error(err)
		return
	}

	for rows.Next() {
		var ip string

		err = rows.Scan(&ip)
		if err != nil {
			t.Error(err)
			return
		}
		dbIp = append(dbIp, ip)
	}

	for i, ip := range hostNames {
		if dbIp[i] != ip.Addr.String() {
			t.Errorf("expected %q; got %q", ip.Addr.String(), dbIp[i])
			return
		}
	}

}

func TestDbPrintStart(t *testing.T) {
	arg := []string{"localhost", "8001"}
	s := newDb(arg, ":memory:")

	expectedHost := "localhost"
	exptectedPort := uint16(8001)
	s.printStart(expectedHost, exptectedPort)
	rows, err := s.db.Query("SELECT hostname, port, timestamp FROM " + s.tableName)

	if err != nil {
		t.Log(err)
		return
	}

	var hostname string
	var port int
	var timeStamp time.Time
	for rows.Next() {
		err = rows.Scan(&hostname, &port, &timeStamp)
		if err != nil {
			t.Error(err)
			return
		}
	}

	if expectedHost != hostname {
		t.Errorf("expected %q; got %q", expectedHost, hostname)
		return
	} else if int(exptectedPort) != port {
		t.Errorf("expected %q; got %q", exptectedPort, port)
		return
	} else if timeStamp.IsZero() {
		t.Error("timeStamp is empty")
	}
}

func TestDbPrintProbeSuccess(t *testing.T) {
	arg := []string{"localhost", "8001"}
	s := newDb(arg, ":memory:")

	expectedHost := "localhost"
	expectedIp := "127.0.0.1"
	exptectedPort := uint16(8001)
	expectedStreak := 100
	expectedRtt := float32(1.0001)

	s.printProbeSuccess(expectedHost, expectedIp, exptectedPort, uint(expectedStreak), expectedRtt)
	rows, err := s.db.Query("SELECT total_successful_probes, latency, timestamp FROM " + s.tableName)

	if err != nil {
		t.Log(err)
		return
	}

	var successfullProves int
	var latency float32
	var timeStamp time.Time
	for rows.Next() {
		err = rows.Scan(&successfullProves, &latency, &timeStamp)
		if err != nil {
			t.Error(err)
			return
		}
	}

	if successfullProves != expectedStreak {
		t.Errorf("expected %q; got %q", expectedHost, successfullProves)
		return
	} else if expectedRtt != latency {
		t.Errorf("expected %f; got %f", expectedRtt, latency)
		return
	} else if timeStamp.IsZero() {
		t.Error("timeStamp is empty")
	}
}

func TestDbPrintStatistics(t *testing.T) {
	arg := []string{"localhost", "8001"}
	s := newDb(arg, ":memory:")
	defer s.db.Close()

	// There are many fields, so many things could go wrong; that's why this elaborate test.
	stat := stats{
		startTime:             time.Now(),
		endTime:               time.Now().Add(10 * time.Minute),
		lastSuccessfulProbe:   time.Now().Add(1 * time.Minute),
		lastUnsuccessfulProbe: time.Now().Add(2 * time.Minute),
		printer:               s,
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
	}

	s.printStatistics(stat)

	// prepare the query
	query := `SELECT
	type,
	ip, hostname, port,
    total_successful_probes, total_unsuccessful_probes,
    last_successful_probe, last_unsuccessful_probe,
    latency_min, latency_avg, latency_max,
    start_timestamp, end_timestamp,
    total_duration, total_uptime, total_downtime,
	longest_uptime, longest_uptime_start, longest_uptime_end,
    longest_downtime, longest_downtime_start, longest_downtime_end,
    total_packets
	FROM ` + s.tableName

	rows, err := s.db.Query(query)
	Nil(t, err)

	if !rows.Next() {
		t.Error("rows can't be empty")
		return
	}

	var (
		typeOf, ip, hostname, port                     string
		totalSuccessfulProbes, totalUnsuccessfulProbes uint
		lastSuccessfulProbe, lastUnsuccessfulProbe     time.Time
		lMin, lAvg, lMax                               float32
		startTimestamp, endTimestamp                   time.Time
		totalDuration, totalUptime, totalDowntime      string
		longestUptime                                  string
		longestUptimeStart, longestUptimeEnd           time.Time
		longestDowntime                                string
		longestDowntimeStart, longestDowntimeEnd       time.Time
		totalPackets                                   uint
	)

	err = rows.Scan(
		&typeOf,
		&ip, &hostname, &port,
		&totalSuccessfulProbes, &totalUnsuccessfulProbes,
		&lastSuccessfulProbe, &lastUnsuccessfulProbe,
		&lMin, &lAvg, &lMax,
		&startTimestamp, &endTimestamp,
		&totalDuration, &totalUptime, &totalDowntime,
		&longestUptime, &longestUptimeStart, &longestUptimeEnd,
		&longestDowntime, &longestDowntimeStart, &longestDowntimeEnd,
		&totalPackets,
	)
	Nil(t, err)

	const timeFormat = "2006-01-02 15:04:05.999999999 -0700 -07"

	Equals(t, typeOf, string(statisticsEvent))
	Equals(t, ip, stat.userInput.ip.String())
	Equals(t, hostname, stat.userInput.hostname)
	Equals(t, totalSuccessfulProbes, stat.totalSuccessfulProbes)
	Equals(t, totalUnsuccessfulProbes, stat.totalUnsuccessfulProbes)
	Equals(t, port, strconv.Itoa(int(stat.userInput.port)))
	Equals(t, lMin, stat.rttResults.min)
	Equals(t, lAvg, stat.rttResults.average)
	Equals(t, lMax, stat.rttResults.max)
	Equals(t, startTimestamp.Format(timeFormat), stat.startTime.Format(timeFormat))
	Equals(t, endTimestamp.Format(timeFormat), stat.endTime.Format(timeFormat))

	actualDuration := stat.totalUptime + stat.totalDowntime
	Equals(t, totalDuration, actualDuration.String())
	Equals(t, totalUptime, stat.totalUptime.String())
	Equals(t, totalDowntime, stat.totalDowntime.String())
	Equals(t, totalPackets, stat.totalSuccessfulProbes+stat.totalUnsuccessfulProbes)

	Equals(t, longestUptime, stat.longestUptime.duration.String())
	Equals(t, longestUptimeStart.Format(timeFormat), stat.longestUptime.start.Format(timeFormat))
	Equals(t, longestUptimeEnd.Format(timeFormat), stat.longestUptime.end.Format(timeFormat))

	Equals(t, longestDowntime, stat.longestDowntime.duration.String())
	Equals(t, longestDowntimeStart.Format(timeFormat), stat.longestDowntime.start.Format(timeFormat))
	Equals(t, longestDowntimeEnd.Format(timeFormat), stat.longestDowntime.end.Format(timeFormat))
}

// Equals compares two values
func Equals[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if want != got {
		t.Errorf("wanted %v; got %v", want, got)
		t.FailNow()
	}
}

// Nil compares a value to nil, in some cases you may need to do `Equals(t, value, nil)`
func Nil(t *testing.T, value any) {
	t.Helper()

	if value != nil {
		t.Logf("expected '%v' to be nil", value)
		t.FailNow()
	}
}
