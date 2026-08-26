package utils

import (
	"fmt"
	"time"
)

// NanoToMillisecond returns an amount of milliseconds from nanoseconds.
// Using duration.Milliseconds() is not an option, because it drops
// decimal points, returning an int.
func NanoToMillisecond(nano int64) float32 {
	return float32(nano) / float32(time.Millisecond)
}

// SecondsToDuration returns the corresponding duration from seconds expressed with a float.
func SecondsToDuration(seconds float64) time.Duration {
	return time.Duration(1000*seconds) * time.Millisecond
}

// MaxDuration is the implementation of the math.Max function for time.Duration models.
// returns the longest duration of x or y.
// We need this so that if the probe took 600ms to complete while have a specified 500ms context,
// the 500ms is used in the calculations.
// TODO: Maybe I should get rid of this hack. Need some more testing.
func MaxDuration(x, y time.Duration) time.Duration {
	if x > y {
		return x
	}

	return y
}

// DurationToString creates a human-readable string for a given duration
func DurationToString(d time.Duration) string {
	hours := d / time.Hour
	d %= time.Hour

	minutes := d / time.Minute
	d %= time.Minute

	seconds := d.Seconds()

	switch {
	case hours >= 2:
		return fmt.Sprintf("%d hours %d minutes %.0f seconds", hours, minutes, seconds)

	case hours == 1:
		if minutes == 0 && seconds == 0 {
			return "1 hour"
		}
		return fmt.Sprintf("1 hour %d minutes %.0f seconds", minutes, seconds)

	case minutes >= 2:
		return fmt.Sprintf("%d minutes %.0f seconds", minutes, seconds)

	case minutes == 1:
		if seconds == 0 {
			return "1 minute"
		}
		return fmt.Sprintf("1 minute %.0f seconds", seconds)

	case seconds == 0:
		return "0 seconds"

	case seconds < 1.1:
		return "1 second"

	case seconds < 2:
		return fmt.Sprintf("%.1f seconds", seconds)

	default:
		return fmt.Sprintf("%.0f seconds", seconds)
	}
}
