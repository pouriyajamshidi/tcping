//go:build !windows

package printers

import (
	"fmt"
	"testing"
	"unicode"

	"github.com/pouriyajamshidi/tcping/v3/internal/stats"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func TestAddDbExtension(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"InMemory", ":memory:", ":memory:"},
		{"With DB Extension", "test.db", "test.db"},
		{"Without Extension", "results", "results.db"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addDbExtension(tt.input)
			if got != tt.expected {
				t.Errorf("addDbExtension(%q) = %q; expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeTableName(t *testing.T) {
	tests := []struct {
		hostname string
		port     string
	}{
		{"example.com", "80"},
		{"192.168.1.1", "443"},
		{"my-host.internal", "8080"},
		{"123numeric", "22"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s:%s", tt.hostname, tt.port), func(t *testing.T) {
			got := sanitizeTableName(tt.hostname, tt.port)

			if unicode.IsNumber(rune(got[0])) {
				t.Errorf("Table name %q starts with a number", got)
			}
			for _, r := range got {
				if r == '.' || r == '-' || r == ' ' {
					t.Errorf("Table name %q contains invalid character %q", got, r)
				}
			}
		})
	}
}

func TestDatabasePrinter_InMemoryOperations(t *testing.T) {
	dbPrinter, err := NewDatabasePrinter("example.com", "80", ":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize DatabasePrinter in memory: %v", err)
	}
	defer dbPrinter.Done()

	dummyStats := &stats.Statistics{
		Hostname:          "example.com",
		Port:              80,
		WithTimestamp:     true,
		WithSourceAddress: true,
	}

	// Test probe records execution
	dbPrinter.PrintProbeSuccess(dummyStats)
	dbPrinter.PrintProbeFailure(dummyStats)

	// Verify insertion into probe table
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s;", dbPrinter.probeTableName)
	err = sqlitex.Execute(dbPrinter.Conn, query, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			count = stmt.ColumnInt(0)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Failed querying probe table count: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 records in probe table, found %d", count)
	}

	// Test statistics write execution
	dbPrinter.PrintStatistics(dummyStats)

	// Verify insertion into stats table
	var statsCount int
	statsQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s;", dbPrinter.statsTableName)
	err = sqlitex.Execute(dbPrinter.Conn, statsQuery, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			statsCount = stmt.ColumnInt(0)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Failed querying stats table count: %v", err)
	}

	if statsCount != 1 {
		t.Errorf("Expected 1 record in stats table, found %d", statsCount)
	}
}
