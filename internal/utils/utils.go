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
