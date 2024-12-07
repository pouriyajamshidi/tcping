package main

import (
    "encoding/csv"
    "os"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

func TestNewCSVPrinter(t *testing.T) {
    dataFilename := "test_data.csv"
    showTimestamp := true
    showLocalAddress := true

    cp, err := newCSVPrinter(dataFilename, showTimestamp, showLocalAddress)
    assert.NoError(t, err)
    assert.NotNil(t, cp)
    assert.Equal(t, dataFilename, cp.dataFilename)
    assert.Equal(t, dataFilename[:len(dataFilename)-4]+"_stats.csv", cp.statsFilename)

    cp.cleanup()
    os.Remove(dataFilename)
    os.Remove(cp.statsFilename)
}

func TestWriteRecord(t *testing.T) {
    dataFilename := "test_data.csv"
    showTimestamp := false
    showLocalAddress := true

    cp, err := newCSVPrinter(dataFilename, showTimestamp, showLocalAddress)
    assert.NoError(t, err)
    assert.NotNil(t, cp)

    record := []string{"Success", "hostname", "127.0.0.1", "80", "1", "10.123", "localAddr"}
    err = cp.writeRecord(record)
    assert.NoError(t, err)

    // Verify the record is written
    file, err := os.Open(dataFilename)
    assert.NoError(t, err)
    defer file.Close()

    reader := csv.NewReader(file)
    headers, err := reader.Read()
    assert.NoError(t, err)
    assert.Equal(t, []string{"Status", "Hostname", "IP", "Port", "TCP_Conn", "Latency(ms)", "Local Address"}, headers)

    readRecord, err := reader.Read()
    assert.NoError(t, err)
    assert.Equal(t, record, readRecord)

    // Cleanup
    cp.cleanup()
    os.Remove(dataFilename)
    os.Remove(cp.statsFilename)
}

func TestWriteStatistics(t *testing.T) {
    dataFilename := "test_data.csv"
    showTimestamp := true
    showLocalAddress := false

    cp, err := newCSVPrinter(dataFilename, showTimestamp, showLocalAddress)
    assert.NoError(t, err)
    assert.NotNil(t, cp)

    tcping := tcping{
        totalSuccessfulProbes:   1,
        totalUnsuccessfulProbes: 0,
        lastSuccessfulProbe:     time.Now(),
        startTime:               time.Now(),
    }

    cp.printStatistics(tcping)

    statsFile, err := os.Open(cp.statsFilename)
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

    cp.cleanup()
    os.Remove(dataFilename)
    os.Remove(cp.statsFilename)
}

func TestCleanup(t *testing.T) {
    dataFilename := "test_data.csv"
    showTimestamp := true
    showLocalAddress := false

    cp, err := newCSVPrinter(dataFilename, showTimestamp, showLocalAddress)
    assert.NoError(t, err)
    assert.NotNil(t, cp)

    os.Create(dataFilename)
    os.Create(cp.statsFilename)

    cp.cleanup()

    _, err = os.Stat(dataFilename)
    assert.NoError(t, err)

    _, err = os.Stat(cp.statsFilename)
    assert.NoError(t, err)

    os.Remove(dataFilename)
    os.Remove(cp.statsFilename)
}
