package printers

import "time"

// Config holds all configuration options for Printer creation
type Config struct {
	OutputJSON        bool
	PrettyJSON        bool
	NoColor           bool
	WithTimestamp     bool
	WithSourceAddress bool
	OmitStatistics    bool // Do not show the statistics. Only available for terminal printers
	Verbose           bool // Show everything an HTTP(S) probe learned, not just the status.
	OutputDBPath      string
	OutputCSVPath     string
	CSVNoTimestamp    bool // Omit the date/time suffix from CSV filenames, using OutputCSVPath as-is.

	Target string
	Port   uint16

	AlloyURL           string        // Address of a Grafana Alloy OTLP HTTP endpoint. Empty unless -alloy was given.
	AlloyStatsInterval time.Duration // How often the run summary is sent to Alloy.

	InfluxDBURL           string        // Address of an InfluxDB server. Empty unless -influxdb was given.
	InfluxDBOrg           string        // InfluxDB organization to write to.
	InfluxDBBucket        string        // InfluxDB bucket to write to.
	InfluxDBToken         string        // InfluxDB API token, from the -influxdb-token flag or the INFLUXDB_TOKEN environment variable.
	InfluxDBStatsInterval time.Duration // How often the run summary is written to InfluxDB.

	SourceLabel string // Names the machine tcping runs on in the metrics sent to Alloy and InfluxDB. Defaults to the hostname.
}
