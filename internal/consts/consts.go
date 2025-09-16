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

// Color functions used when printing information
var (
	ColorCyan        = color.Cyan.Printf
	ColorLightCyan   = color.LightCyan.Printf
	ColorGreen       = color.Green.Printf
	ColorLightGreen  = color.LightGreen.Printf
	ColorYellow      = color.Yellow.Printf
	ColorLightYellow = color.LightYellow.Printf
	ColorRed         = color.Red.Printf
	ColorLightBlue   = color.FgLightBlue.Printf
)
