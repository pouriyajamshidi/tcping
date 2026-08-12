package dns

import (
	"net/netip"
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/config"
	"github.com/pouriyajamshidi/tcping/v3/internal/probers"
)

// dummyPrinter is a fake test implementation
// of a printer that does nothing.
type dummyPrinter struct{}

func (fp *dummyPrinter) PrintStart(_ string, _ uint16) {}
func (fp *dummyPrinter) PrintProbeSuccess(_ time.Time, _ string, _ config.Config, _ uint, _ string) {
}
func (fp *dummyPrinter) PrintProbeFailure(_ time.Time, _ config.Config, _ uint) {}
func (fp *dummyPrinter) PrintRetryingToResolve(_ string)                        {}
func (fp *dummyPrinter) PrintTotalDownTime(_ time.Duration)                     {}
func (fp *dummyPrinter) PrintStatistics(_ probers.Tcping)                       {}
func (fp *dummyPrinter) PrintError(_ string, _ ...interface{})                  {}

// createTestStats should be used to create new stats structs.
// it uses "127.0.0.1:12345" as default values, because
// [testServerListen] use the same values.
// It'll call t.Errorf if netip.ParseAddr has failed.
func createTestStats(t *testing.T) *probers.Tcping {
	_, err := netip.ParseAddr("127.0.0.1")
	s := probers.Tcping{
		Ticker: time.NewTicker(time.Second),
	}
	if err != nil {
		t.Errorf("ip parse: %v", err)
	}

	return &s
}

func TestSelectResolvedIPv4(t *testing.T) {
	userInputV4 := config.Config{
		UseIPv4: true,
	}

	stats := createTestStats(t)
	stats.Options = userInputV4

	var (
		ip1 = netip.MustParseAddr("172.20.10.238")
		ip2 = netip.MustParseAddr("8.8.8.8")
	)

	t.Run("IPv4 Selection", func(t *testing.T) {
		actual, _ := selectRandomResolvedIP([]netip.Addr{ip1, ip2}, true, false)

		if !actual.IsValid() {
			t.Errorf("Expected an IP but got invalid address")
		}
		if actual != ip1 && actual != ip2 {
			t.Errorf("Expected an IP but got invalid address")
		}
	})
}

func TestSelectResolvedIPv6(t *testing.T) {
	userInputV6 := config.Config{
		UseIPv6: true,
	}

	stats := createTestStats(t)
	stats.Options = userInputV6

	var (
		ip1 = netip.MustParseAddr("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
		ip2 = netip.MustParseAddr("2001:4860:4860::8888")
	)

	t.Run("IPv6 Selection", func(t *testing.T) {
		actual, _ := selectRandomResolvedIP([]netip.Addr{ip1, ip2}, false, true)
		if !actual.IsValid() {
			t.Errorf("Expected an IP but got invalid address")
		}
		if actual != ip1 && actual != ip2 {
			t.Errorf("Expected an IP but got invalid address")
		}
	})
}
