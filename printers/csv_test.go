package printers

import (
	"encoding/csv"
	"os"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestNewCSVPrinter(t *testing.T) {
	dataFilename := "test_data.csv"
	// showTimestamp := true
	// showSourceAddress := true

	cfg := PrinterConfig{}

	cp, err := NewCSVPrinter(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, cp)
	assert.Equal(t, dataFilename, cp.ProbeFilename)
	assert.Equal(t, dataFilename[:len(dataFilename)-4]+"_stats.csv", cp.StatsFilename)

	cp.Done()
	os.Remove(dataFilename)
	os.Remove(cp.StatsFilename)
}

func TestWriteRecord(t *testing.T) {
	dataFilename := "test_data.csv"
	// showTimestamp := false
	// showSourceAddress := true

	cfg := PrinterConfig{}

	cp, err := NewCSVPrinter(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, cp)

	record := []string{"Success", "hostname", "127.0.0.1", "80", "1", "10.123", "sourceAddr"}
	assert.NoError(t, err)

	// Verify the record is written
	file, err := os.Open(dataFilename)
	assert.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	assert.NoError(t, err)
	assert.Equal(t, []string{"Status", "Hostname", "IP", "Port", "TCP_Conn", "Latency(ms)", "Source Address"}, headers)

	readRecord, err := reader.Read()
	assert.NoError(t, err)
	assert.Equal(t, record, readRecord)

	// Cleanup
	cp.Done()
	os.Remove(dataFilename)
	os.Remove(cp.StatsFilename)
}

func TestWriteStatistics(t *testing.T) {
	dataFilename := "test_data.csv"
	// showTimestamp := true
	// showSourceAddress := false
	cfg := PrinterConfig{}

	cp, err := NewCSVPrinter(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, cp)

	tcping := types.Tcping{
		TotalSuccessfulProbes:   1,
		TotalUnsuccessfulProbes: 0,
		LastSuccessfulProbe:     time.Now(),
		StartTime:               time.Now(),
	}

	PrintStats(&tcping)

	statsFile, err := os.Open(cp.StatsFilename)
	assert.NoError(t, err)
	defer statsFile.Close()

	reader := csv.NewReader(statsFile)
	headers, err := reader.Read()
	assert.NoError(t, err)
	assert.Equal(t, []string{"Metric", "Value"}, headers)

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		assert.NotEmpty(t, record)
	}

	cp.Done()
	os.Remove(dataFilename)
	os.Remove(cp.StatsFilename)
}

func TestCleanup(t *testing.T) {
	dataFilename := "test_data.csv"
	// showTimestamp := true
	// showSourceAddress := false

	cfg := PrinterConfig{}

	cp, err := NewCSVPrinter(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, cp)

	// Call printStatistics to ensure the stats file is created
	tcping := types.Tcping{
		TotalSuccessfulProbes:   1,
		TotalUnsuccessfulProbes: 0,
		LastSuccessfulProbe:     time.Now(),
		StartTime:               time.Now(),
	}

	PrintStats(&tcping)

	// Perform cleanup
	cp.Done()

	// Verify files are closed and flushed
	_, err = os.Stat(dataFilename)
	assert.NoError(t, err)

	_, err = os.Stat(cp.StatsFilename)
	assert.NoError(t, err)

	// Cleanup files
	os.Remove(dataFilename)
	os.Remove(cp.StatsFilename)
}
