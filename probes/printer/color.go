package printer

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pouriyajamshidi/tcping/v2/probes/statistics"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
)

type Printer interface {
	Print(s string)
	PrintStatistics(statistics statistics.Statistics)
	PrintError(err error)
}

type ColorPrinter struct {
	Color bool
	w     io.Writer
}

func NewColor() *ColorPrinter {
	return &ColorPrinter{
		Color: true,
		w:     os.Stdout,
	}
}

func (p *ColorPrinter) PrintStatistics(statistics statistics.Statistics) {
	if p.Color {
		p.w.Write(fmt.Appendf(nil, "%s%s%s\n", Green, "Statistics:", Reset))
		p.w.Write(fmt.Appendf(nil, "%s%s%s %s\n", Green, "Start Time:", Reset, statistics.StartTime.Format(time.RFC3339)))
		p.w.Write(fmt.Appendf(nil, "%s%s%s %s\n", Green, "End Time:", Reset, statistics.EndTime.Format(time.RFC3339)))
		p.w.Write(fmt.Appendf(nil, "%s%s%s %s\n", Green, "Up Time:", Reset, statistics.UpTime))
		p.w.Write(fmt.Appendf(nil, "%s%s%s %d\n", Green, "Total Pings:", Reset, len(statistics.RTT)))
		p.w.Write(fmt.Appendf(nil, "%s%s%s %s\n", Green, "Average RTT:", Reset, statistics.AverageRTT()))
		p.w.Write(fmt.Appendf(nil, "%s%s%s %s\n", Green, "Max RTT:", Reset, statistics.MaxRTT()))
		p.w.Write(fmt.Appendf(nil, "%s%s%s %s\n", Green, "Min RTT:", Reset, statistics.MinRTT()))
		p.w.Write(fmt.Appendf(nil, "%s%s%s %s\n", Green, "Longest Up:", Reset, statistics.LongestUp))
		p.w.Write(fmt.Appendf(nil, "%s%s%s %s\n", Green, "Longest Down:", Reset, statistics.LongestDown))
		p.w.Write(fmt.Appendf(nil, "%s%s%s %d\n", Green, "Successful Pings:", Reset, statistics.Successful))
		p.w.Write(fmt.Appendf(nil, "%s%s%s %d\n", Green, "Failed Pings:", Reset, statistics.Failed))
		return
	}
	p.w.Write(fmt.Appendf(nil, "Statistics:\n"))
	p.w.Write(fmt.Appendf(nil, "Start Time: %s\n", statistics.StartTime.Format(time.RFC3339)))
	p.w.Write(fmt.Appendf(nil, "End Time: %s\n", statistics.EndTime.Format(time.RFC3339)))
	p.w.Write(fmt.Appendf(nil, "Up Time: %s\n", statistics.UpTime))
	p.w.Write(fmt.Appendf(nil, "Total Pings: %d\n", len(statistics.RTT)))
	p.w.Write(fmt.Appendf(nil, "Average RTT: %s\n", statistics.AverageRTT()))
	p.w.Write(fmt.Appendf(nil, "Max RTT: %s\n", statistics.MaxRTT()))
	p.w.Write(fmt.Appendf(nil, "Min RTT: %s\n", statistics.MinRTT()))
}

func (p *ColorPrinter) Print(s string) {
	p.w.Write(fmt.Appendf(nil, "%s%s%s\n", Blue, s, Reset))
}

func (p *ColorPrinter) PrintError(err error) {
	p.w.Write(fmt.Appendf(nil, "%s%s%s\n", Red, err.Error(), Reset))
}
