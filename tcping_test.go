package main

import (
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// createTestStats should be used to create new stats structs.
// it uses "127.0.0.1:12345" as default values, because
// [testServerListen] use the same values.
// It'll call t.Errorf if netip.ParseAddr has failed.
func createTestStats(t *testing.T) *tcping {
	addr, err := netip.ParseAddr("127.0.0.1")
	s := tcping{
		printer: &dummyPrinter{},
		userInput: userInput{
			ip:                    addr,
			port:                  12345,
			intervalBetweenProbes: time.Second,
			timeout:               time.Second,
		},
		ticker: time.NewTicker(time.Second),
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
	stats.ticker = time.NewTicker(time.Nanosecond)
	srv := testServerListen(t)
	t.Cleanup(func() {
		if err := srv.Close(); err != nil {
			t.Errorf("srv close: %v", err)
		}
	})

	expectedSuccessful := 100

	for i := 0; i < expectedSuccessful; i++ {
		tcpProbe(stats)
	}

	assert.Equal(t, stats.totalSuccessfulProbes, uint(expectedSuccessful))
	assert.Equal(t, stats.ongoingSuccessfulProbes, uint(expectedSuccessful))

	assert.Equal(t, stats.totalUptime, 100*time.Second)
}

func TestProbeSuccessInterval(t *testing.T) {
	stats := createTestStats(t)
	stats.userInput.intervalBetweenProbes = 10 * time.Second
	stats.ticker = time.NewTicker(time.Nanosecond)
	srv := testServerListen(t)
	t.Cleanup(func() {
		if err := srv.Close(); err != nil {
			t.Errorf("srv close: %v", err)
		}
	})

	expectedSuccessful := 100

	for i := 0; i < expectedSuccessful; i++ {
		tcpProbe(stats)
	}

	assert.Equal(t, stats.totalSuccessfulProbes, uint(expectedSuccessful))
	assert.Equal(t, stats.ongoingSuccessfulProbes, uint(expectedSuccessful))

	assert.Equal(t, stats.totalUptime, 16*time.Minute+40*time.Second)
}

func TestProbeFail(t *testing.T) {
	stats := createTestStats(t)
	stats.ticker = time.NewTicker(time.Nanosecond)

	expectedFailed := 100

	for i := 0; i < expectedFailed; i++ {
		tcpProbe(stats)
	}

	assert.Equal(t, stats.totalUnsuccessfulProbes, uint(expectedFailed))
	assert.Equal(t, stats.ongoingUnsuccessfulProbes, uint(expectedFailed))

	assert.Equal(t, stats.totalDowntime, 100*time.Second)
}

func TestProbeFailInterval(t *testing.T) {
	stats := createTestStats(t)
	stats.userInput.intervalBetweenProbes = 10 * time.Second
	stats.ticker = time.NewTicker(time.Nanosecond)

	expectedFailed := 100

	for i := 0; i < expectedFailed; i++ {
		tcpProbe(stats)
	}

	assert.Equal(t, stats.totalUnsuccessfulProbes, uint(expectedFailed))
	assert.Equal(t, stats.ongoingUnsuccessfulProbes, uint(expectedFailed))

	assert.Equal(t, stats.totalDowntime, 16*time.Minute+40*time.Second)
}

func TestPermuteArgs(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"host/ip before option",
			args{args: []string{"127.0.0.1", "8080", "-r", "3"}},
			[]string{"-r", "3", "127.0.0.1", "8080"},
		},
		{
			"host/ip after option",
			args{args: []string{"-r", "3", "127.0.0.1", "8080"}},
			[]string{"-r", "3", "127.0.0.1", "8080"},
		},
		{
			"check for updates",
			args{args: []string{"-u"}},
			[]string{"-u"},
		},
		/**
		 * cases in which the value of the option does not exist are not listed.
		 * they call directly usage() and exit with code 1.
		 */
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permuteArgs(tt.args.args)
			assert.Equal(t, tt.want, tt.args.args)
		})
	}
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
			got := nanoToMillisecond(tt.d.Nanoseconds())
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSelectResolvedIPv4(t *testing.T) {
	userInputV4 := userInput{
		useIPv4: true,
	}

	stats := createTestStats(t)
	stats.userInput = userInputV4

	var (
		ip1 = netip.MustParseAddr("172.20.10.238")
		ip2 = netip.MustParseAddr("8.8.8.8")
	)

	t.Run("IPv4 Selection", func(t *testing.T) {
		actual := selectResolvedIP(stats, []netip.Addr{ip1, ip2})

		if !actual.IsValid() {
			t.Errorf("Expected an IP but got invalid address")
		}
		if actual != ip1 && actual != ip2 {
			t.Errorf("Expected an IP but got invalid address")
		}
	})
}

func TestSelectResolvedIPv6(t *testing.T) {
	userInputV6 := userInput{
		useIPv6: true,
	}

	stats := createTestStats(t)
	stats.userInput = userInputV6

	var (
		ip1 = netip.MustParseAddr("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
		ip2 = netip.MustParseAddr("2001:4860:4860::8888")
	)

	t.Run("IPv6 Selection", func(t *testing.T) {
		actual := selectResolvedIP(stats, []netip.Addr{ip1, ip2})
		if !actual.IsValid() {
			t.Errorf("Expected an IP but got invalid address")
		}
		if actual != ip1 && actual != ip2 {
			t.Errorf("Expected an IP but got invalid address")
		}
	})
}

func TestSecondsToDuration(t *testing.T) {
	tests := []struct {
		name     string
		seconds  float64
		duration time.Duration
	}{
		{
			name:     "positive integer",
			seconds:  2,
			duration: 2 * time.Second,
		},
		{
			name:     "positive float",
			seconds:  1.5, // 1.5 = 3 / 2
			duration: time.Second * 3 / 2,
		},
		{
			name:     "negative integer",
			seconds:  -3,
			duration: -3 * time.Second,
		},
		{
			name:     "negative float",
			seconds:  -2.5, // -2.5 = -5 / 2
			duration: time.Second * -5 / 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.duration, secondsToDuration(tt.seconds))
		})
	}
}
