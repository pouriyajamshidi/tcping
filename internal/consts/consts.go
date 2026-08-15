// Package consts provides constants and utility variables for this project,
// including version information, time formats, and color printing utilities.
package consts

type Protocol string

const (
	TCP   Protocol = "TCP"
	UDP   Protocol = "UDP"
	HTTP  Protocol = "HTTP"
	HTTPS Protocol = "HTTPS"
	ICMP  Protocol = "ICMP"
)
