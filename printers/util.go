package printers

import (
	"fmt"
	"math"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04:05"
	hourFormat = "15:04:05"
)

// durationToString creates a human-readable string for a given duration
func durationToString(duration time.Duration) string {
	hours := math.Floor(duration.Hours())
	if hours > 0 {
		duration -= time.Duration(hours * float64(time.Hour))
	}

	minutes := math.Floor(duration.Minutes())
	if minutes > 0 {
		duration -= time.Duration(minutes * float64(time.Minute))
	}

	seconds := duration.Seconds()

	switch {
	// Hours
	case hours >= 2:
		return fmt.Sprintf("%.0f hours %.0f minutes %.0f seconds", hours, minutes, seconds)
	case hours == 1 && minutes == 0 && seconds == 0:
		return fmt.Sprintf("%.0f hour", hours)
	case hours == 1:
		return fmt.Sprintf("%.0f hour %.0f minutes %.0f seconds", hours, minutes, seconds)

	// Minutes
	case minutes >= 2:
		return fmt.Sprintf("%.0f minutes %.0f seconds", minutes, seconds)
	case minutes == 1 && seconds == 0:
		return fmt.Sprintf("%.0f minute", minutes)
	case minutes == 1:
		return fmt.Sprintf("%.0f minute %.0f seconds", minutes, seconds)

	// Seconds
	case seconds == 0 || seconds == 1 || seconds >= 1 && seconds < 1.1:
		return fmt.Sprintf("%.0f second", seconds)
	case seconds < 1:
		return fmt.Sprintf("%.1f seconds", seconds)

	default:
		return fmt.Sprintf("%.0f seconds", seconds)
	}
}
