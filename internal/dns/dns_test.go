package dns

import (
	"net/netip"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/types"
)

// dummyPrinter is a fake test implementation
// of a printer that does nothing.
type dummyPrinter struct{}

func (fp *dummyPrinter) PrintStart(_ string, _ uint16)                                  {}
func (fp *dummyPrinter) PrintProbeSuccess(_ string, _ types.Options, _ uint, _ float32) {}
func (fp *dummyPrinter) PrintProbeFail(_ types.Options, _ uint)                         {}
func (fp *dummyPrinter) PrintRetryingToResolve(_ string)                                {}
func (fp *dummyPrinter) PrintTotalDownTime(_ time.Duration)                             {}
func (fp *dummyPrinter) PrintStatistics(_ types.Tcping)                                 {}
func (fp *dummyPrinter) PrintInfo(_ string, _ ...interface{})                           {}
func (fp *dummyPrinter) PrintError(_ string, _ ...interface{})                          {}

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

func TestSelectResolvedIPv4(t *testing.T) {
	userInputV4 := types.Options{
		UseIPv4: true,
	}

	stats := createTestStats(t)
	stats.Options = userInputV4

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
	userInputV6 := types.Options{
		UseIPv6: true,
	}

	stats := createTestStats(t)
	stats.Options = userInputV6

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
