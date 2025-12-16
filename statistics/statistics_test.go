package statistics_test

import (
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/statistics"
)

func TestCalcMinAvgMaxRttTime(t *testing.T) {
	tests := []struct {
		name  string
		input []float32
		want  statistics.RttResult
	}{
		{
			name:  "empty array returns zero result",
			input: nil,
			want:  statistics.RttResult{},
		},
		{
			name:  "empty slice returns zero result",
			input: []float32{},
			want:  statistics.RttResult{},
		},
		{
			name:  "single value",
			input: []float32{5.5},
			want: statistics.RttResult{
				Min:        5.5,
				Max:        5.5,
				Average:    5.5,
				HasResults: true,
			},
		},
		{
			name:  "multiple values",
			input: []float32{1.0, 2.0, 3.0, 4.0, 5.0},
			want: statistics.RttResult{
				Min:        1.0,
				Max:        5.0,
				Average:    3.0,
				HasResults: true,
			},
		},
		{
			name:  "same values",
			input: []float32{10.0, 10.0, 10.0},
			want: statistics.RttResult{
				Min:        10.0,
				Max:        10.0,
				Average:    10.0,
				HasResults: true,
			},
		},
		{
			name:  "decimal values",
			input: []float32{1.5, 2.5, 3.5},
			want: statistics.RttResult{
				Min:        1.5,
				Max:        3.5,
				Average:    2.5,
				HasResults: true,
			},
		},
		{
			name:  "unordered values",
			input: []float32{5.0, 1.0, 3.0, 2.0, 4.0},
			want: statistics.RttResult{
				Min:        1.0,
				Max:        5.0,
				Average:    3.0,
				HasResults: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statistics.CalcMinAvgMaxRttTime(tt.input)

			if got.Min != tt.want.Min {
				t.Errorf("Min = %v, want %v", got.Min, tt.want.Min)
			}
			if got.Max != tt.want.Max {
				t.Errorf("Max = %v, want %v", got.Max, tt.want.Max)
			}
			if got.Average != tt.want.Average {
				t.Errorf("Average = %v, want %v", got.Average, tt.want.Average)
			}
			if got.HasResults != tt.want.HasResults {
				t.Errorf("HasResults = %v, want %v", got.HasResults, tt.want.HasResults)
			}
		})
	}
}

func TestNanoToMillisecond(t *testing.T) {
	tests := []struct {
		name  string
		nano  int64
		want  float32
	}{
		{
			name: "zero",
			nano: 0,
			want: 0,
		},
		{
			name: "one millisecond",
			nano: int64(time.Millisecond),
			want: 1.0,
		},
		{
			name: "one second",
			nano: int64(time.Second),
			want: 1000.0,
		},
		{
			name: "half millisecond",
			nano: int64(500 * time.Microsecond),
			want: 0.5,
		},
		{
			name: "100 milliseconds",
			nano: int64(100 * time.Millisecond),
			want: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statistics.NanoToMillisecond(tt.nano)
			if got != tt.want {
				t.Errorf("NanoToMillisecond(%d) = %v, want %v", tt.nano, got, tt.want)
			}
		})
	}
}

func TestSecondsToDuration(t *testing.T) {
	tests := []struct {
		name    string
		seconds float64
		want    time.Duration
	}{
		{
			name:    "zero",
			seconds: 0,
			want:    0,
		},
		{
			name:    "one second",
			seconds: 1.0,
			want:    time.Second,
		},
		{
			name:    "half second",
			seconds: 0.5,
			want:    500 * time.Millisecond,
		},
		{
			name:    "two and a half seconds",
			seconds: 2.5,
			want:    2500 * time.Millisecond,
		},
		{
			name:    "ten seconds",
			seconds: 10.0,
			want:    10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statistics.SecondsToDuration(tt.seconds)
			if got != tt.want {
				t.Errorf("SecondsToDuration(%v) = %v, want %v", tt.seconds, got, tt.want)
			}
		})
	}
}

func TestDurationToString(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "zero seconds",
			duration: 0,
			want:     "0 second",
		},
		{
			name:     "one second",
			duration: time.Second,
			want:     "1 second",
		},
		{
			name:     "two seconds",
			duration: 2 * time.Second,
			want:     "2 seconds",
		},
		{
			name:     "one minute",
			duration: time.Minute,
			want:     "1 minute",
		},
		{
			name:     "one minute thirty seconds",
			duration: time.Minute + 30*time.Second,
			want:     "1 minute 30 seconds",
		},
		{
			name:     "two minutes",
			duration: 2 * time.Minute,
			want:     "2 minutes 0 seconds",
		},
		{
			name:     "one hour",
			duration: time.Hour,
			want:     "1 hour",
		},
		{
			name:     "one hour thirty minutes",
			duration: time.Hour + 30*time.Minute,
			want:     "1 hour 30 minutes 0 seconds",
		},
		{
			name:     "two hours",
			duration: 2 * time.Hour,
			want:     "2 hours 0 minutes 0 seconds",
		},
		{
			name:     "sub-second",
			duration: 500 * time.Millisecond,
			want:     "0.5 seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statistics.DurationToString(tt.duration)
			if got != tt.want {
				t.Errorf("DurationToString(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestSetLongestDuration(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		start      time.Time
		duration   time.Duration
		existing   statistics.LongestTime
		wantUpdate bool
	}{
		{
			name:       "zero start time does nothing",
			start:      time.Time{},
			duration:   time.Hour,
			existing:   statistics.LongestTime{},
			wantUpdate: false,
		},
		{
			name:       "zero duration does nothing",
			start:      baseTime,
			duration:   0,
			existing:   statistics.LongestTime{},
			wantUpdate: false,
		},
		{
			name:       "first duration sets value",
			start:      baseTime,
			duration:   time.Hour,
			existing:   statistics.LongestTime{},
			wantUpdate: true,
		},
		{
			name:     "longer duration updates",
			start:    baseTime,
			duration: 2 * time.Hour,
			existing: statistics.LongestTime{
				Start:    baseTime.Add(-time.Hour),
				End:      baseTime,
				Duration: time.Hour,
			},
			wantUpdate: true,
		},
		{
			name:     "shorter duration does not update",
			start:    baseTime,
			duration: 30 * time.Minute,
			existing: statistics.LongestTime{
				Start:    baseTime.Add(-time.Hour),
				End:      baseTime,
				Duration: time.Hour,
			},
			wantUpdate: false,
		},
		{
			name:     "equal duration updates",
			start:    baseTime,
			duration: time.Hour,
			existing: statistics.LongestTime{
				Start:    baseTime.Add(-2 * time.Hour),
				End:      baseTime.Add(-time.Hour),
				Duration: time.Hour,
			},
			wantUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			longest := tt.existing
			originalDuration := longest.Duration

			statistics.SetLongestDuration(tt.start, tt.duration, &longest)

			if tt.wantUpdate {
				if longest.Duration != tt.duration {
					t.Errorf("Duration = %v, want %v", longest.Duration, tt.duration)
				}
				if longest.Start != tt.start {
					t.Errorf("Start = %v, want %v", longest.Start, tt.start)
				}
			} else {
				if longest.Duration != originalDuration {
					t.Errorf("Duration changed to %v, should remain %v", longest.Duration, originalDuration)
				}
			}
		})
	}
}

func TestNewLongestTime(t *testing.T) {
	start := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	duration := time.Hour

	got := statistics.NewLongestTime(start, duration)

	if got.Start != start {
		t.Errorf("Start = %v, want %v", got.Start, start)
	}
	if got.Duration != duration {
		t.Errorf("Duration = %v, want %v", got.Duration, duration)
	}
	expectedEnd := start.Add(duration)
	if got.End != expectedEnd {
		t.Errorf("End = %v, want %v", got.End, expectedEnd)
	}
}

func TestStatistics_IP_String(t *testing.T) {
	s := statistics.Statistics{}

	// zero value
	got := s.IP.String()
	if got != "invalid IP" {
		t.Errorf("IP.String() for zero value = %q, want %q", got, "invalid IP")
	}
}

func TestStatistics_PortStr(t *testing.T) {
	tests := []struct {
		name string
		port uint16
		want string
	}{
		{name: "port 80", port: 80, want: "80"},
		{name: "port 443", port: 443, want: "443"},
		{name: "port 8080", port: 8080, want: "8080"},
		{name: "port 0", port: 0, want: "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := statistics.Statistics{Port: tt.port}
			got := s.PortStr()
			if got != tt.want {
				t.Errorf("PortStr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStatistics_RTTStr(t *testing.T) {
	tests := []struct {
		name      string
		latestRTT float32
		want      string
	}{
		{name: "zero", latestRTT: 0, want: "0.000"},
		{name: "integer", latestRTT: 10.0, want: "10.000"},
		{name: "decimal", latestRTT: 1.234, want: "1.234"},
		{name: "small", latestRTT: 0.001, want: "0.001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := statistics.Statistics{LatestRTT: tt.latestRTT}
			got := s.RTTStr()
			if got != tt.want {
				t.Errorf("RTTStr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStatistics_TimeFormatted(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 14, 30, 45, 0, time.UTC)

	s := statistics.Statistics{
		StartTime: testTime,
		EndTime:   testTime.Add(time.Hour),
	}

	startFormatted := s.StartTimeFormatted()
	if startFormatted != "2024-06-15 14:30:45" {
		t.Errorf("StartTimeFormatted() = %q, want %q", startFormatted, "2024-06-15 14:30:45")
	}

	endFormatted := s.EndTimeFormatted()
	if endFormatted != "2024-06-15 15:30:45" {
		t.Errorf("EndTimeFormatted() = %q, want %q", endFormatted, "2024-06-15 15:30:45")
	}
}

func TestStatistics_ProtocolStr(t *testing.T) {
	s := statistics.Statistics{}
	// Protocol is unexported type, but ProtocolStr returns string representation
	got := s.ProtocolStr()
	// default zero value
	if got != "" {
		t.Errorf("ProtocolStr() for zero value = %q, want empty string", got)
	}
}
