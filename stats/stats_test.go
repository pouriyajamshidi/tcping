package stats

import (
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/config"
)

func TestRuntimeDuration_MatchesStartAndEndTime(t *testing.T) {
	s := &Statistics{
		StartTime: time.Date(2026, 9, 1, 12, 49, 25, 0, time.UTC),
		EndTime:   time.Date(2026, 9, 1, 12, 49, 28, 0, time.UTC),
		// Total up/downtime intentionally left smaller than EndTime-StartTime,
		// mirroring the gap before the first probe's result comes in.
		TotalUptime: 2 * time.Second,
	}

	if got, want := s.RuntimeDuration(), "00:00:03"; got != want {
		t.Errorf("RuntimeDuration() = %q, want %q (must match StartTime/EndTime, not TotalUptime+TotalDowntime)", got, want)
	}
}

func TestUptimeDuration(t *testing.T) {
	s := &Statistics{CurrentUptime: 90 * time.Second}

	if got, want := s.UptimeDuration(), "1 minute 30 seconds"; got != want {
		t.Errorf("UptimeDuration() = %q, want %q", got, want)
	}
}

func TestDurationToMilliseconds(t *testing.T) {
	if got, want := DurationToMilliseconds(21770*time.Microsecond), float32(21.770); got != want {
		t.Errorf("DurationToMilliseconds() = %v, want %v", got, want)
	}
}

func TestNameResolutionDurationStr(t *testing.T) {
	s := &Statistics{NameResolutionDuration: 23456 * time.Microsecond}

	if got, want := s.NameResolutionDurationStr(), "23.456"; got != want {
		t.Errorf("NameResolutionDurationStr() = %q, want %q", got, want)
	}
}

func TestHostnameChange_DurationStr(t *testing.T) {
	h := &HostnameChange{Duration: 1734 * time.Microsecond}

	if got, want := h.DurationStr(), "1.734"; got != want {
		t.Errorf("DurationStr() = %q, want %q", got, want)
	}
}

func TestRTTResultUpdate(t *testing.T) {
	var r RTTResult

	samples := []float32{15.5, 10.0, 20.0, 30.0, 1.1, 2.2, 3.3, 4.4}

	for i, s := range samples {
		r.Update(s, uint(i+1)) //nolint:gosec // i+1 is always positive and small
	}

	if r.Min != 1.1 {
		t.Errorf("Min = %v, want 1.1", r.Min)
	}
	if r.Max != 30 {
		t.Errorf("Max = %v, want 30", r.Max)
	}
	if got, want := r.Average, float32(10.8125); got < want-0.0001 || got > want+0.0001 {
		t.Errorf("Average = %v, want %v", got, want)
	}

	// The same number ping would print for these samples, which is the
	// deviation over the whole population and not over a sample of it.
	if got, want := r.Mdev, float32(9.625933); got < want-0.001 || got > want+0.001 {
		t.Errorf("Mdev = %v, want %v", got, want)
	}
}

// Probes that all take the same time have nothing to deviate by, and that
// has to come out as an exact zero rather than as rounding noise.
func TestRTTResultUpdate_IdenticalSamplesHaveNoDeviation(t *testing.T) {
	var r RTTResult

	for i := 1; i <= 5; i++ {
		r.Update(7.5, uint(i)) //nolint:gosec // i is always positive and small
	}

	if r.Mdev != 0 {
		t.Errorf("Mdev = %v, want 0", r.Mdev)
	}
}

func TestRTTResultUpdate_SingleSample(t *testing.T) {
	var r RTTResult

	r.Update(15, 1)

	if r.Min != 15 || r.Max != 15 || r.Average != 15 {
		t.Errorf("RTTResult = %+v, want Min=Max=Average=15", r)
	}

	if r.Mdev != 0 {
		t.Errorf("Mdev = %v, want 0, a single sample cannot deviate from itself", r.Mdev)
	}
}

// TestRuntimeDuration_AgreesWithUptime guards the mismatch where a 6.6
// second run reported "00:00:06" next to a "7 seconds" uptime, because
// RuntimeDuration truncated while durationToString rounded.
func TestRuntimeDuration_AgreesWithUptime(t *testing.T) {
	tests := []struct {
		name        string
		elapsed     time.Duration
		wantRuntime string
		wantUptime  string
	}{
		{"rounds down", 6200 * time.Millisecond, "00:00:06", "6 seconds"},
		{"half rounds up", 6500 * time.Millisecond, "00:00:07", "7 seconds"},
		{"rounds up", 6600 * time.Millisecond, "00:00:07", "7 seconds"},
		{"just under a minute", 59700 * time.Millisecond, "00:01:00", "1 minute"},
	}

	start := time.Date(2026, 9, 2, 10, 48, 20, 0, time.UTC)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Statistics{StartTime: start, EndTime: start.Add(tt.elapsed)}

			if got := s.RuntimeDuration(); got != tt.wantRuntime {
				t.Errorf("RuntimeDuration() = %q, want %q", got, tt.wantRuntime)
			}

			if got := durationToString(tt.elapsed); got != tt.wantUptime {
				t.Errorf("durationToString(%v) = %q, want %q", tt.elapsed, got, tt.wantUptime)
			}
		})
	}
}

// TestDurationToString_NoSixtySeconds makes sure the seconds part never
// carries over, which used to print 1m59.7s as "1 minute 60 seconds".
func TestDurationToString_NoSixtySeconds(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{119700 * time.Millisecond, "2 minutes 0 seconds"},
		{59600 * time.Millisecond, "1 minute"},
		{3599700 * time.Millisecond, "1 hour"},
		{0, "0 seconds"},
		{1500 * time.Millisecond, "2 seconds"},
	}

	for _, tt := range tests {
		if got := durationToString(tt.d); got != tt.want {
			t.Errorf("durationToString(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

// TestDurationToString_SubSecond guards against reporting a gap that really
// happened as "0 seconds", which is what a Ctrl+C in the middle of a probe
// used to produce on the "longest consecutive downtime" line.
func TestDurationToString_SubSecond(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "0 seconds"},
		{5 * time.Millisecond, "5.000 ms"},
		{40 * time.Millisecond, "40.000 ms"},
		{300 * time.Millisecond, "0.3 seconds"},
		{892 * time.Millisecond, "0.9 seconds"},
		{999 * time.Millisecond, "1.0 seconds"},
		{time.Second, "1 second"},
	}

	for _, tt := range tests {
		if got := durationToString(tt.d); got != tt.want {
			t.Errorf("durationToString(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestIsHTTPAndHasHTTPResponse(t *testing.T) {
	tests := []struct {
		name            string
		protocol        config.Protocol
		statusCode      int
		wantIsHTTP      bool
		wantHasResponse bool
	}{
		{"tcp target", config.TCP, 0, false, false},
		{"http target with a response", config.HTTP, 200, true, true},
		{"https target with a response", config.HTTPS, 503, true, true},
		{"https target that never connected", config.HTTPS, 0, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Statistics{Protocol: tt.protocol, HTTP: HTTPInfo{StatusCode: tt.statusCode}}

			if got := s.IsHTTP(); got != tt.wantIsHTTP {
				t.Errorf("IsHTTP() = %v, want %v", got, tt.wantIsHTTP)
			}

			if got := s.HasHTTPResponse(); got != tt.wantHasResponse {
				t.Errorf("HasHTTPResponse() = %v, want %v", got, tt.wantHasResponse)
			}
		})
	}
}

func TestCertHelpers(t *testing.T) {
	s := &Statistics{}

	if got := s.CertExpiryStr(); got != "" {
		t.Errorf("CertExpiryStr() with no certificate = %q, want an empty string", got)
	}

	if got := s.CertDaysRemaining(); got != 0 {
		t.Errorf("CertDaysRemaining() with no certificate = %d, want 0", got)
	}

	s.HTTP.CertExpiry = time.Now().Add(10*24*time.Hour + time.Hour)

	if got := s.CertDaysRemaining(); got != 10 {
		t.Errorf("CertDaysRemaining() = %d, want 10", got)
	}

	s.HTTP.CertExpiry = time.Now().Add(-48 * time.Hour)

	if got := s.CertDaysRemaining(); got >= 0 {
		t.Errorf("CertDaysRemaining() for an expired certificate = %d, want a negative number", got)
	}
}

func TestHTTPDurationStrings(t *testing.T) {
	s := &Statistics{HTTP: HTTPInfo{
		StatusCode:      404,
		ConnectDuration: 12500 * time.Microsecond,
		TLSDuration:     3200 * time.Microsecond,
		TimeToFirstByte: 41750 * time.Microsecond,
	}}

	for _, tt := range []struct {
		name string
		got  string
		want string
	}{
		{"StatusCodeStr", s.StatusCodeStr(), "404"},
		{"ConnectDurationStr", s.ConnectDurationStr(), "12.500"},
		{"TLSDurationStr", s.TLSDurationStr(), "3.200"},
		{"TimeToFirstByteStr", s.TimeToFirstByteStr(), "41.750"},
	} {
		if tt.got != tt.want {
			t.Errorf("%s() = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}
