package main

import (
	"fmt"
	"net/netip"
	"strconv"
	"testing"
	"time"
)

func TestNewDB(t *testing.T) {
	arg := []string{"localhost", "8001"}
	s := newDb(arg, ":memory:")
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
	isNil(t, err)
	Equals(t, tblName, s.tableName)
}

func TestDbSaveStats(t *testing.T) {
	// There are many fields, so many things could go wrong; that's why this elaborate test.
	arg := []string{"localhost", "8001"}
	s := newDb(arg, ":memory:")
	defer s.db.Close()

	stat := mockStats()
	err := s.saveStats(stat)
	isNil(t, err)

	query := `SELECT
		addr, hostname, port, hostname_resolve_retries,
		total_successful_probes, total_unsuccessful_probes,
		never_succeed_probe, never_failed_probe,
		last_successful_probe, last_unsuccessful_probe,
		total_packets, total_packet_loss,
		total_uptime, total_downtime,
		longest_uptime, longest_uptime_end, longest_uptime_start,
		longest_downtime, longest_downtime_start, longest_downtime_end,
		latency_min, latency_avg, latency_max,
		start_time, end_time, total_duration
	FROM ` + fmt.Sprintf("%s WHERE event_type = '%s'", s.tableName, eventTypeStatistics)

	rows, err := s.db.Query(query)
	isNil(t, err)

	if !rows.Next() {
		t.Error("rows are empty; expted 1 row")
		return
	}

	var (
		addr, hostname, port                           string
		hostNameResolveTries                           uint
		totalSuccessfulProbes, totalUnsuccessfulProbes uint
		neverSucceedProbe, neverFailedProbe            bool
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
		&addr, &hostname, &port, &hostNameResolveTries,
		&totalSuccessfulProbes, &totalUnsuccessfulProbes,
		&neverSucceedProbe, &neverFailedProbe,
		&lastSuccessfulProbe, &lastUnsuccessfulProbe,
		&totalPackets, &totalPacketsLoss,
		&totalUptime, &totalDowntime,
		&longestUptime, &longestUptimeStart, &longestUptimeEnd,
		&longestDowntime, &longestDowntimeStart, &longestDowntimeEnd,
		&lMin, &lAvg, &lMax,
		&startTimestamp, &endTimestamp, &totalDuration,
	)

	isNil(t, err)
	rows.Close()

	t.Log("the line number will tell you where the error happend")
	Equals(t, addr, stat.userInput.ip.String())
	Equals(t, hostname, stat.userInput.hostname)
	Equals(t, totalSuccessfulProbes, stat.totalSuccessfulProbes)
	Equals(t, totalUnsuccessfulProbes, stat.totalUnsuccessfulProbes)
	Equals(t, port, strconv.Itoa(int(stat.userInput.port)))

	Equals(t, neverSucceedProbe, stat.lastSuccessfulProbe.IsZero())
	Equals(t, neverFailedProbe, stat.lastUnsuccessfulProbe.IsZero())

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

func TestSaveHostname(t *testing.T) {
	// There are many fields, so many things could go wrong; that's why this elaborate test.
	arg := []string{"localhost", "8001"}
	s := newDb(arg, ":memory:")
	defer s.db.Close()
	stat := mockStats()

	err := s.saveHostNameChange(stat.hostnameChanges)
	isNil(t, err)
	// testing the host names if they are properly written
	query := `SELECT
	hostname_changed_to, hostname_change_time
	FROM ` + fmt.Sprintf("%s WHERE event_type IS '%s';", s.tableName, eventTypeHostnameChange)

	rows, err := s.db.Query(query)
	isNil(t, err)

	idx := 0
	for rows.Next() {
		var hostName string
		var cTime time.Time
		err = rows.Scan(&hostName, &cTime)
		isNil(t, err)

		actualHost := stat.hostnameChanges[idx]
		idx++
		Equals(t, hostName, actualHost.Addr.String())
		Equals(t, cTime.Format(timeFormat), actualHost.When.Format(timeFormat))
	}
	Equals(t, idx, len(stat.hostnameChanges))
	rows.Close()
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

func mockStats() stats {
	stat := stats{
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
