package main

import (
	"encoding/csv"
	"os"
	"testing"
)

func setupTempCSVFile(t *testing.T) (*csvPrinter, func()) {
	t.Helper()
	tempFile, err := os.CreateTemp("", "test_csv_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	cp := &csvPrinter{
		file:   tempFile,
		writer: csv.NewWriter(tempFile),
	}

	cleanup := func() {
		cp.writer.Flush()
		if err := cp.writer.Error(); err != nil {
			t.Errorf("Error flushing CSV writer: %v", err)
		}
		if err := cp.file.Close(); err != nil {
			t.Errorf("Error closing file: %v", err)
		}
		if err := os.Remove(cp.file.Name()); err != nil {
			t.Errorf("Error removing temp file: %v", err)
		}
	}

	return cp, cleanup
}

func TestNewCSVPrinter(t *testing.T) {
	args := []string{"localhost", "8001"}
	filename := "test.csv"

	cp, err := newCSVPrinter(filename, args)
	if err != nil {
		t.Fatalf("error creating CSV printer: %v", err)
	}

	defer func() {
		if cp != nil && cp.file != nil {
			cp.file.Close()
		}
		os.Remove(filename)
	}()

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("file %s was not created", filename)
	}

	if cp.filename != filename {
		t.Errorf("expected filename %q, got %q", filename, cp.filename)
	}

	if cp.headerDone {
		t.Errorf("expected headerDone to be false, got true")
	}

	if cp.writer == nil {
		t.Errorf("CSV writer was not initialized")
	}
}
