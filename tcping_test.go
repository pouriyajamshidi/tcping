package main

import (
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// createTestStats should be used to create new stats structs.
// it uses localhost and 12345 to be tested using the testServerListen.
// it could fail if netip.ParseAddr has failed.
func createTestStats(t *testing.T) *stats {
	addr, err := netip.ParseAddr("127.0.0.1")
	s := stats{
		printer: &dummyPrinter{},
		ip:      addr,
		port:    12345,
		ticker:  time.NewTicker(time.Second),
	}
	if err != nil {
		t.Errorf("ip parse: %v", err)
	}

	return &s
}

// testServerListen creates a new listener
// on port 12345 and automatically starts it.
//
// Use t.Cleanup to close it after the test,
// so that other tests are not affected.
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

	expectedSuccessfull := 100

	for i := 0; i < expectedSuccessfull; i++ {
		tcping(stats)
	}

	assert.Equal(t, stats.totalSuccessfulProbes, uint(expectedSuccessfull))
	assert.Equal(t, stats.ongoingSuccessfulProbes, uint(expectedSuccessfull))

	// TODO: change when custom ping intervals will be introduced
	assert.Equal(t, stats.totalUptime, 100*time.Second)
}

func TestProbeFail(t *testing.T) {
	stats := createTestStats(t)
	stats.ticker = time.NewTicker(time.Nanosecond)

	expectedFailed := 100

	for i := 0; i < expectedFailed; i++ {
		tcping(stats)
	}

	assert.Equal(t, stats.totalUnsuccessfulProbes, uint(expectedFailed))
	assert.Equal(t, stats.ongoingUnsuccessfulProbes, uint(expectedFailed))

	// TODO: change when custom ping intervals will be introduced
	assert.Equal(t, stats.totalDowntime, 100*time.Second)
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
