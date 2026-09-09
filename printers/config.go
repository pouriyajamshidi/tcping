package printers

import (
	"io"
	"time"
)

// Config holds all configuration options for Printer creation
type Config struct {
	// Where the output goes. Nil means os.Stdout, which is what a real run
	// uses. Setting it lets a test read what a printer wrote, and leaves
	// room for a destination that is not the terminal, e.g. a socket.
	Writer io.Writer

	OutputJSON        bool
	PrettyJSON        bool
	NoColor           bool
	WithTimestamp     bool
	WithSourceAddress bool
	OmitStatistics    bool // Do not show the statistics. Only available for terminal printers
	Verbose           bool // Show everything an HTTP(S) probe learned, not just the status.
	OutputSQLitePath  string
	OutputCSVPath     string
	CSVNoTimestamp    bool // Omit the date/time suffix from CSV filenames, using OutputCSVPath as-is.

	Target string
	Port   uint16

	AlloyURL string // Address of a Grafana Alloy OTLP HTTP endpoint. Empty unless -alloy was given.

	InfluxDBURL    string // Address of an InfluxDB server. Empty unless -influxdb was given.
	InfluxDBOrg    string // InfluxDB organization to write to.
	InfluxDBBucket string // InfluxDB bucket to write to.
	InfluxDBToken  string // InfluxDB API token, from the -influxdb-token flag or the INFLUXDB_TOKEN environment variable.

	StatsInterval time.Duration // How often the run summary is sent to Alloy or InfluxDB.
	SourceLabel   string        // Names the machine tcping runs on in the metrics sent to Alloy and InfluxDB. Defaults to the hostname.
}
