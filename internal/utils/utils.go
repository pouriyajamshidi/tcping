package utils

import (
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
