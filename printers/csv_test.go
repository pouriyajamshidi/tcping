package printers

import (
	"encoding/csv"
	"net/netip"
	"os"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestNewCSVPrinter(t *testing.T) {
	probeFilename := "test_data.csv"

	cfg := PrinterConfig{OutputCSVPath: "test_data"}

	cp, err := NewCSVPrinter(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, cp)
	assert.Equal(t, probeFilename, cp.ProbeFile.Name())
	assert.Equal(t, probeFilename[:len(probeFilename)-4]+"_stats.csv", cp.StatsFile.Name())

	cp.Done()

	os.Remove(cp.ProbeFile.Name())
	os.Remove(cp.StatsFile.Name())
}

func TestWriteRecord(t *testing.T) {
	cfg := PrinterConfig{}

	cp, err := NewCSVPrinter(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, cp)

	file, err := os.Open(cp.ProbeFile.Name())
	assert.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	assert.NoError(t, err)
	assert.Equal(t, []string{"Status", "Hostname", "IP", "Port", "Connection", "Latency(ms)"}, headers)

	opts := types.Options{
		IP:       netip.MustParseAddr("127.0.0.1"),
		Hostname: "localhost",
		Port:     80,
	}

	cp.PrintProbeSuccess(time.Now(), "192.168.1.10:1234", opts, 1, "10.123")

	readRecord, err := reader.Read()
	assert.NoError(t, err)

	record := []string{"Reply", "localhost", "127.0.0.1", "80", "1", "10.123"}
	assert.Equal(t, record, readRecord)

	cp.Done()

	os.Remove(cp.ProbeFile.Name())
	os.Remove(cp.StatsFile.Name())
}

func TestWriteStatistics(t *testing.T) {
	probeFilename := "test_data.csv"

	cfg := PrinterConfig{
		OutputCSVPath: probeFilename,
		Target:        "localhost",
		Port:          "1234",
	}

	cp, err := NewCSVPrinter(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, cp)

	tcping := types.Tcping{
		Printer:                 cp,
		TotalSuccessfulProbes:   1,
		TotalUnsuccessfulProbes: 0,
		LastSuccessfulProbe:     time.Now(),
		StartTime:               time.Now(),
	}

	PrintStats(&tcping)

	statsFile, err := os.Open(cp.StatsFile.Name())
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

	os.Remove(cp.ProbeFile.Name())
	os.Remove(cp.StatsFile.Name())
}

func TestCleanup(t *testing.T) {
	probeFilename := "test_data.csv"

	cfg := PrinterConfig{
		OutputCSVPath: probeFilename,
		Target:        "localhost",
		Port:          "1234",
	}

	cp, err := NewCSVPrinter(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, cp)

	tcping := types.Tcping{
		Printer:                 cp,
		TotalSuccessfulProbes:   1,
		TotalUnsuccessfulProbes: 0,
		LastSuccessfulProbe:     time.Now(),
		StartTime:               time.Now(),
	}

	PrintStats(&tcping)

	cp.Done()

	_, err = os.Stat(probeFilename)
	assert.NoError(t, err)

	_, err = os.Stat(cp.StatsFile.Name())
	assert.NoError(t, err)

	os.Remove(cp.ProbeFile.Name())
	os.Remove(cp.StatsFile.Name())
}
