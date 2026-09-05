//go:build !windows

package printers

import (
	"fmt"
	"net/netip"
	"path/filepath"
	"regexp"
	"testing"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"

	"github.com/pouriyajamshidi/tcping/v3/stats"
)

func TestAddDBExtension(t *testing.T) {
	tests := map[string]string{
		":memory:":   ":memory:",
		"results.db": "results.db",
		"results":    "results.db",
	}

	for input, want := range tests {
		got := addDbExtension(input)
		if got != want {
			t.Fatalf("addDbExtension(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestBoolToInt(t *testing.T) {
	if got := boolToInt(true); got != 1 {
		t.Fatalf("boolToInt(true) = %d, want 1", got)
	}
	if got := boolToInt(false); got != 0 {
		t.Fatalf("boolToInt(false) = %d, want 0", got)
	}
}

func TestSanitizeTableName(t *testing.T) {
	name := sanitizeTableName("example.com-1", "443")
	if !regexp.MustCompile(`^example_com_1_443__[0-9]{4}_[0-9]{2}_[0-9]{2}_[0-9]{2}_[0-9]{2}_[0-9]{2}$`).MatchString(name) {
		t.Fatalf("unexpected table name %q", name)
	}

	name = sanitizeTableName("1.2.3.4", "443")
	if name[0] != '_' {
		t.Fatalf("table name %q should be prefixed when target starts with a number", name)
	}
}

func TestNewDatabasePrinterConfiguresSQLite(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "results.db")

	printer, err := NewDatabasePrinter(Config{Target: "example.com", Port: 443, OutputDBPath: dbPath})
	if err != nil {
		t.Fatal(err)
	}
	defer printer.done()

	var journalMode, synchronous, busyTimeout string
	err = sqlitex.Execute(printer.conn,
		"SELECT (SELECT journal_mode FROM pragma_journal_mode), (SELECT synchronous FROM pragma_synchronous), (SELECT timeout FROM pragma_busy_timeout)",
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				journalMode = stmt.ColumnText(0)
				synchronous = stmt.ColumnText(1)
				busyTimeout = stmt.ColumnText(2)
				return nil
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if journalMode != "wal" {
		t.Errorf("journal_mode = %q, want wal", journalMode)
	}
	if synchronous != "1" {
		t.Errorf("synchronous = %q, want 1 (NORMAL)", synchronous)
	}
	if busyTimeout != "5000" {
		t.Errorf("busy_timeout = %q, want 5000", busyTimeout)
	}
}

func TestNewDatabasePrinterSchema(t *testing.T) {
	printer, err := NewDatabasePrinter(Config{Target: "example.com", Port: 443, OutputDBPath: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	defer printer.done()

	assertColumns := func(table string, expected map[string]string) {
		t.Helper()

		got := make(map[string]string)
		err := sqlitex.Execute(printer.conn,
			"SELECT name, type FROM pragma_table_info(?)",
			&sqlitex.ExecOptions{
				Args: []any{table},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					got[stmt.ColumnText(0)] = stmt.ColumnText(1)
					return nil
				},
			},
		)
		if err != nil {
			t.Fatal(err)
		}

		for column, wantType := range expected {
			if got[column] != wantType {
				t.Fatalf("column %q has type %q, want %q", column, got[column], wantType)
			}
		}
	}

	assertColumns(printer.probeTableName, map[string]string{
		"reachable":                   "INTEGER",
		"destination_is_ip":           "INTEGER",
		"latency":                     "REAL",
		"ongoing_successful_probes":   "INTEGER",
		"ongoing_unsuccessful_probes": "INTEGER",
	})

	assertColumns(printer.statsTableName, map[string]string{
		"total_packet_loss_percent": "REAL",
		"latency_min":               "REAL",
		"latency_avg":               "REAL",
		"latency_max":               "REAL",
		"hostname_changes":          "TEXT",
	})
}

func TestInsertProbeStoresSQLiteTypes(t *testing.T) {
	printer, err := NewDatabasePrinter(Config{Target: "example.com", Port: 443, OutputDBPath: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	defer printer.done()

	ip := netip.MustParseAddr("192.0.2.10")

	probeStats := &stats.Statistics{
		Hostname:                "example.com",
		IP:                      ip,
		Port:                    443,
		DestIsIP:                true,
		OngoingSuccessfulProbes: 7,
	}
	printer.insertProbe(true, probeStats, "12.345", 7, 0)

	var (
		reachable        string
		destinationIsIP  string
		latency          string
		successfulProbes string
	)

	if err := sqlitex.Execute(printer.conn,
		fmt.Sprintf(`SELECT typeof(reachable), typeof(destination_is_ip), typeof(latency), typeof(ongoing_successful_probes) FROM %s`, printer.probeTableName),
		&sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				reachable = stmt.ColumnText(0)
				destinationIsIP = stmt.ColumnText(1)
				latency = stmt.ColumnText(2)
				successfulProbes = stmt.ColumnText(3)
				return nil
			},
		},
	); err != nil {
		t.Fatal(err)
	}

	if reachable != "integer" {
		t.Errorf("typeof(reachable) = %q, want integer", reachable)
	}
	if destinationIsIP != "integer" {
		t.Errorf("typeof(destination_is_ip) = %q, want integer", destinationIsIP)
	}
	if latency != "real" {
		t.Errorf("typeof(latency) = %q, want real", latency)
	}
	if successfulProbes != "integer" {
		t.Errorf("typeof(ongoing_successful_probes) = %q, want integer", successfulProbes)
	}
}
