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

	rows.Next()
	var tblName string
	err = rows.Scan(&tblName)
	Nil(t, err)
	Equals(t, tblName, s.tableName)
}

func TestDbPrintStatistics(t *testing.T) {
	arg := []string{"localhost", "8001"}
	s := newDb(arg, ":memory:")
	defer s.db.Close()

	// There are many fields, so many things could go wrong; that's why this elaborate test.
	stat := stats{
		startTime:              time.Now(),
		endTime:                time.Now().Add(10 * time.Minute),
		lastSuccessfulProbe:    time.Now().Add(1 * time.Minute),
		lastUnsuccessfulProbe:  time.Now().Add(2 * time.Minute),
		retriedHostnameLookups: 10,
		printer:                s,
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
		addr, hostname, port, hostname_resolve_tries,
		total_successful_probes, total_unsuccessful_probes,
		last_successful_probe, last_unsuccessful_probe,
		total_packets, total_packet_loss,
		total_uptime, total_downtime,
		longest_uptime, longest_uptime_end, longest_uptime_start,
		longest_downtime, longest_downtime_start, longest_downtime_end,
		latency_min, latency_avg, latency_max,
		start_timestamp, end_timestamp, total_duration
	FROM ` + s.tableName

	rows, err := s.db.Query(query)
	Nil(t, err)

	if !rows.Next() {
		t.Error("rows can't be empty")
		return
	}

	var (
		typeOf, addr, hostname, port                   string
		hostNameResolveTries                           uint
		totalSuccessfulProbes, totalUnsuccessfulProbes uint
		lastSuccessfulProbe, lastUnsuccessfulProbe     time.Time
		totalPackets                                   uint
		totalPacketsLoss                               float32
		totalUptime, totalDowntime                     string
		longestUptime                                  string
		longestUptimeStart, longestUptimeEnd           time.Time
		longestDowntime                                string
		longestDowntimeStart, longestDowntimeEnd       time.Time
		lMin, lAvg, lMax                               float32
		startTimestamp, endTimestamp                   time.Time
		totalDuration                                  string
	)

	err = rows.Scan(
		&typeOf,
		&addr, &hostname, &port, &hostNameResolveTries,
		&totalSuccessfulProbes, &totalUnsuccessfulProbes,
		&lastSuccessfulProbe, &lastUnsuccessfulProbe,
		&totalPackets, &totalPacketsLoss,
		&totalUptime, &totalDowntime,
		&longestUptime, &longestUptimeStart, &longestUptimeEnd,
		&longestDowntime, &longestDowntimeStart, &longestDowntimeEnd,
		&lMin, &lAvg, &lMax,
		&startTimestamp, &endTimestamp, &totalDuration,
	)

	Nil(t, err)

	const timeFormat = "2006-01-02 15:04:05.999999999 -0700 -07"

	Equals(t, typeOf, string(statisticsEvent))
	Equals(t, addr, stat.userInput.ip.String())
	Equals(t, hostname, stat.userInput.hostname)
	Equals(t, totalSuccessfulProbes, stat.totalSuccessfulProbes)
	Equals(t, totalUnsuccessfulProbes, stat.totalUnsuccessfulProbes)
	Equals(t, port, strconv.Itoa(int(stat.userInput.port)))
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
	Equals(t, longestUptimeStart.Format(timeFormat), stat.longestUptime.start.Format(timeFormat))
	Equals(t, longestUptimeEnd.Format(timeFormat), stat.longestUptime.end.Format(timeFormat))

	Equals(t, longestDowntime, stat.longestDowntime.duration.String())
	Equals(t, longestDowntimeStart.Format(timeFormat), stat.longestDowntime.start.Format(timeFormat))
	Equals(t, longestDowntimeEnd.Format(timeFormat), stat.longestDowntime.end.Format(timeFormat))
}

// Equals compares two values
func Equals[T comparable](t *testing.T, value, want T) {
	t.Helper()
	if want != value {
		t.Errorf("wanted %v; got %v", want, value)
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
