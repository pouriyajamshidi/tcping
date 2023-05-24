package main

import (
	"testing"
	"time"
)

func TestDurationToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "1 second",
			duration: time.Duration(time.Second),
			want:     "1 second",
		},
		{
			name:     "59 seconds",
			duration: time.Duration(59 * time.Second),
			want:     "59 seconds",
		},
		{
			name:     "1 minute",
			duration: time.Duration(time.Minute),
			want:     "1 minute",
		},
		{
			name:     "59 minutes 0 seconds",
			duration: time.Duration(59 * time.Minute),
			want:     "59 minutes 0 seconds",
		},
		{
			name:     "1 minute 5 seconds",
			duration: time.Duration(time.Minute + 5*time.Second),
			want:     "1 minute 5 seconds",
		},
		{
			name:     "59 minutes 5 seconds",
			duration: time.Duration(59*time.Minute + 5*time.Second),
			want:     "59 minutes 5 seconds",
		},
		{
			name:     "1 hour",
			duration: time.Duration(time.Hour),
			want:     "1 hour",
		},
		{
			name:     "1 hour 10 minutes 5 seconds",
			duration: time.Duration(time.Hour + 10*time.Minute + 5*time.Second),
			want:     "1 hour 10 minutes 5 seconds",
		},
		{
			name:     "59 hours 0 minutes 0 seconds",
			duration: time.Duration(59 * time.Hour),
			want:     "59 hours 0 minutes 0 seconds",
		},
		{
			name:     "59 hours 10 minutes 5 seconds",
			duration: time.Duration(59*time.Hour + 10*time.Minute + 5*time.Second),
			want:     "59 hours 10 minutes 5 seconds",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := durationToString(tt.duration); got != tt.want {
				t.Errorf("calcTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
