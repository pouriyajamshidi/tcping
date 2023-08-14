package main

import (
	"net/netip"
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

	tableName, hostnameChangeTable := newTableNames(arg)
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
	}

	if tNames[0] != tableName {
		t.Errorf("expected %s; got %s", tableName, tNames[0])
	}

	if tNames[1] != hostnameChangeTable {
		t.Errorf("expected %s; got %s", hostnameChangeTable, tNames[1])
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

	s.printStatistics(*createTestStats(t))
}

func TestDbHostnameSave(t *testing.T) {
	arg := []string{"localhost", "8001"}
	s := newDb(arg, ":memory:")

	ipAddresses := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334", // IPv6 address
		// Add more IP addresses as needed
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
