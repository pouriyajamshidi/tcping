package stats

import "testing"

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
}

func TestRTTResultUpdate_SingleSample(t *testing.T) {
	var r RTTResult

	r.Update(15, 1)

	if r.Min != 15 || r.Max != 15 || r.Average != 15 {
		t.Errorf("RTTResult = %+v, want Min=Max=Average=15", r)
	}
}
