package app

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pouriyajamshidi/tcping/v3"
	"github.com/pouriyajamshidi/tcping/v3/dns"
	"github.com/pouriyajamshidi/tcping/v3/pingers"
	"github.com/pouriyajamshidi/tcping/v3/statistics"
)

// Run executes the tcping application and returns an exit code
func Run() int {
	config, err := ProcessUserInput()
	if err != nil {
		return handleError(err, nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	ip, err := resolveTarget(ctx, config.Hostname, config.UseIPv4, config.UseIPv6)
	if err != nil {
		return handleError(err, nil)
	}

	pinger := buildPinger(ip, config)

	printer, err := tcping.NewPrinter(config.PrinterConfig)
	if err != nil {
		return handleError(err, nil)
	}

	prober := buildProber(pinger, printer, config, ip)

	probeCtx := setupSignalHandler(context.Background())

	stats, err := prober.Probe(probeCtx)
	if err != nil {
		return handleError(err, printer)
	}

	finalizeStatistics(&stats)

	printer.PrintStatistics(&stats)
	printer.Shutdown(&stats)

	return 0
}

func resolveTarget(ctx context.Context, hostname string, useIPv4, useIPv6 bool) (netip.Addr, error) {
	if ip, err := netip.ParseAddr(hostname); err == nil {
		return ip, nil
	}

	return dns.ResolveHostname(ctx, hostname, useIPv4, useIPv6)
}

func buildPinger(ip netip.Addr, config ProberConfig) *pingers.TCPPinger {
	if config.InterfaceDialer == nil {
		return pingers.NewTCPPinger(ip, config.Port, pingers.WithTimeout(config.Timeout))
	}

	config.InterfaceDialer.Timeout = config.Timeout
	return pingers.NewTCPPinger(ip, config.Port, pingers.WithDialer(config.InterfaceDialer))
}

func buildProber(pinger *pingers.TCPPinger, printer tcping.Printer, config ProberConfig, ip netip.Addr) *tcping.Prober {
	opts := []tcping.ProberOption{
		tcping.WithPrinter(printer),
		tcping.WithInterval(config.Interval),
		tcping.WithTimeout(config.Timeout),
		tcping.WithProbeCount(config.ProbeCountLimit),
		tcping.WithShowFailuresOnly(config.ShowFailuresOnly),
	}

	if config.Hostname != ip.String() {
		opts = append(opts, tcping.WithHostname(config.Hostname))
	}

	return tcping.NewProber(pinger, opts...)
}

func setupSignalHandler(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	return ctx
}

func finalizeStatistics(stats *statistics.Statistics) {
	stats.EndTime = time.Now()
	stats.RTTResults = statistics.CalcMinAvgMaxRttTime(stats.RTT)

	if stats.DestWasDown {
		statistics.SetLongestDuration(stats.StartOfDowntime, time.Since(stats.StartOfDowntime), &stats.LongestDown)
		return
	}

	statistics.SetLongestDuration(stats.StartOfUptime, time.Since(stats.StartOfUptime), &stats.LongestUp)
}

func handleError(err error, printer tcping.Printer) int {
	if err == nil {
		return 0
	}

	if errors.Is(err, ErrUsageRequested) {
		PrintUsage()
		return 1
	}

	if errors.Is(err, ErrVersionRequested) {
		PrintVersion()
		return 0
	}

	if errors.Is(err, ErrUpdateCheckRequested) {
		msg, checkErr := CheckForUpdates()
		if checkErr != nil {
			printError(checkErr, printer)
			return 1
		}
		fmt.Println(msg)
		return 0
	}

	printError(err, printer)
	return 1
}

func printError(err error, printer tcping.Printer) {
	if printer != nil {
		printer.PrintError("%v", err)
		return
	}

	fmt.Fprintf(os.Stderr, "error: %v\n", err)
}
