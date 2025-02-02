// Package consts provides constants and utility variables for this project,
// including version information, time formats, and color printing utilities.
package consts

import (
	"time"

	"github.com/gookit/color"
)

// Version is set at compile time
var Version = ""

// Used when checking for updates
const (
	Owner = "pouriyajamshidi"
	Repo  = "tcping"
)

// DNSTimeout is the accepted duration when doing hostname resolution
const DNSTimeout = 2 * time.Second

// Date and Time formats
const (
	TimeFormat = "2006-01-02 15:04:05"
	HourFormat = "15:04:05"
)

// Color functions used when printing information
var (
	ColorYellow      = color.Yellow.Printf
	ColorGreen       = color.Green.Printf
	ColorRed         = color.Red.Printf
	ColorCyan        = color.Cyan.Printf
	ColorLightYellow = color.LightYellow.Printf
	ColorLightBlue   = color.FgLightBlue.Printf
	ColorLightGreen  = color.LightGreen.Printf
	ColorLightCyan   = color.LightCyan.Printf
)
