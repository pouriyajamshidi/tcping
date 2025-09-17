package probes

import (
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
	"github.com/pouriyajamshidi/tcping/v3/types"
	"github.com/stretchr/testify/assert"
)

// dummyPrinter is a fake test implementation
// of a printer that does nothing.
type dummyPrinter struct{}

func (fp *dummyPrinter) PrintStart(_ string, _ uint16)                                              {}
func (fp *dummyPrinter) PrintProbeSuccess(_ time.Time, _ string, _ types.Options, _ uint, _ string) {}
func (fp *dummyPrinter) PrintProbeFailure(_ time.Time, _ types.Options, _ uint)                     {}
func (fp *dummyPrinter) PrintRetryingToResolve(_ string)                                            {}
func (fp *dummyPrinter) PrintTotalDownTime(_ time.Duration)                                         {}
func (fp *dummyPrinter) PrintStatistics(_ types.Tcping)                                             {}
func (fp *dummyPrinter) PrintError(_ string, _ ...interface{})                                      {}

// createTestStats should be used to create new stats structs.
// it uses "127.0.0.1:12345" as default values, because
// [testServerListen] use the same values.
// It'll call t.Errorf if netip.ParseAddr has failed.
func createTestStats(t *testing.T) *types.Tcping {
	addr, err := netip.ParseAddr("127.0.0.1")
	s := types.Tcping{
		Printer: &dummyPrinter{},
		Options: types.Options{
			IP:                    addr,
			Port:                  12345,
			IntervalBetweenProbes: time.Second,
			Timeout:               time.Second,
		},
		Ticker: time.NewTicker(time.Second),
	}
	if err != nil {
		t.Errorf("ip parse: %v", err)
	}

	return &s
}

// testServerListen creates a new listener
// on port 12345 and automatically starts it.
//
// Use t.Cleanup with srv.Close() to close it after
// the test, so that other tests are not affected.
//
// It could fail if net.Listen or Accept has failed.
func testServerListen(t *testing.T) net.Listener {
	srv, err := net.Listen("tcp", ":12345")
	if err != nil {
		t.Errorf("test server: %v", err)
	}

	go func() {
		for {
			c, err := srv.Accept()
			if err != nil {
				return
			}

			c.Close()
		}
	}()

	return srv
}

func TestProbeSuccess(t *testing.T) {
	stats := createTestStats(t)
	stats.Ticker = time.NewTicker(time.Nanosecond)
	srv := testServerListen(t)
	t.Cleanup(func() {
		if err := srv.Close(); err != nil {
			t.Errorf("srv close: %v", err)
		}
	})

	expectedSuccessful := 100

	for i := 0; i < expectedSuccessful; i++ {
		Ping(stats)
	}

	assert.Equal(t, stats.TotalSuccessfulProbes, uint(expectedSuccessful))
	assert.Equal(t, stats.OngoingSuccessfulProbes, uint(expectedSuccessful))

	assert.Equal(t, stats.TotalUptime, 100*time.Second)
}

func TestProbeSuccessInterval(t *testing.T) {
	stats := createTestStats(t)
	stats.Options.IntervalBetweenProbes = 10 * time.Second
	stats.Ticker = time.NewTicker(time.Nanosecond)
	srv := testServerListen(t)
	t.Cleanup(func() {
		if err := srv.Close(); err != nil {
			t.Errorf("srv close: %v", err)
		}
	})

	expectedSuccessful := 100

	for i := 0; i < expectedSuccessful; i++ {
		Ping(stats)
	}

	assert.Equal(t, stats.TotalSuccessfulProbes, uint(expectedSuccessful))
	assert.Equal(t, stats.OngoingSuccessfulProbes, uint(expectedSuccessful))

	assert.Equal(t, stats.TotalUptime, 16*time.Minute+40*time.Second)
}

func TestProbeFail(t *testing.T) {
	stats := createTestStats(t)
	stats.Ticker = time.NewTicker(time.Nanosecond)

	expectedFailed := 100

	for i := 0; i < expectedFailed; i++ {
		Ping(stats)
	}

	assert.Equal(t, stats.TotalUnsuccessfulProbes, uint(expectedFailed))
	assert.Equal(t, stats.OngoingUnsuccessfulProbes, uint(expectedFailed))

	assert.Equal(t, stats.TotalDowntime, 100*time.Second)
}

func TestProbeFailInterval(t *testing.T) {
	stats := createTestStats(t)
	stats.Options.IntervalBetweenProbes = 10 * time.Second
	stats.Ticker = time.NewTicker(time.Nanosecond)

	expectedFailed := 100

	for i := 0; i < expectedFailed; i++ {
		Ping(stats)
	}

	assert.Equal(t, stats.TotalUnsuccessfulProbes, uint(expectedFailed))
	assert.Equal(t, stats.OngoingUnsuccessfulProbes, uint(expectedFailed))

	assert.Equal(t, stats.TotalDowntime, 16*time.Minute+40*time.Second)
}

func TestNanoToMilliseconds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		d    time.Duration
		want float32
	}{
		{d: time.Millisecond, want: 1},
		{d: 100*time.Millisecond + 123*time.Nanosecond, want: 100.000123},
		{d: time.Second, want: 1000},
		{d: time.Second + 100*time.Nanosecond, want: 1000.000123},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.d.String(), func(t *testing.T) {
			t.Parallel()
			got := utils.NanoToMillisecond(tt.d.Nanoseconds())
			assert.Equal(t, tt.want, got)
		})
	}
}
