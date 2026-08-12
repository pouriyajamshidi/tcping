// Package consts provides constants and utility variables for this project,
// including version information, time formats, and color printing utilities.
package consts

// Version is set at compile time via the Makefile
var Version = ""

// Used when checking for updates
const (
	Owner = "pouriyajamshidi"
	Repo  = "tcping"
)

type Protocol string

const (
	TCP   Protocol = "TCP"
	UDP   Protocol = "UDP"
	HTTP  Protocol = "HTTP"
	HTTPS Protocol = "HTTPS"
	ICMP  Protocol = "ICMP"
)
