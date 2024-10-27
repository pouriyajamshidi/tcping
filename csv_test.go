package main

import (
	"os"
	"testing"
)

func TestNewCSVPrinter(t *testing.T) {
	args := []string{"localhost", "8001"}
	filename := "test.csv"

	cp, err := newCSVPrinter(filename, args)
	if err != nil {
		t.Fatalf("error creating CSV printer: %v", err)
	}

	// Ensure file is closed even if the test fails
	defer func() {
		if cp != nil && cp.file != nil {
			cp.file.Close()
		}
		// Clean up the created file
		os.Remove(filename)
	}()

	// Check if file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("file %s was not created", filename)
	}

	if cp.filename != filename {
		t.Errorf("expected filename %q, got %q", filename, cp.filename)
	}

	if cp.headerDone {
		t.Errorf("expected headerDone to be false, got true")
	}

	// Check if writer is initialized
	if cp.writer == nil {
		t.Errorf("CSV writer was not initialized")
	}

}
