package statistics

import (
	"time"

	"github.com/pouriyajamshidi/tcping/v2/types"
)

type Statistics struct {
	StartTime   time.Time
	EndTime     time.Time
	UpTime      time.Duration
	DownTime    time.Duration
	Successful  int
	Failed      int
	LongestUp   types.LongestTime
	LongestDown types.LongestTime
	RTT         []time.Duration
	HostChanges []types.HostnameChange
	HasResults  bool
}

func (s *Statistics) AverageRTT() time.Duration {
	if len(s.RTT) == 0 {
		return 0
	}
	var total time.Duration
	for _, rtt := range s.RTT {
		total += rtt
	}
	return total / time.Duration(len(s.RTT))
}

func (s *Statistics) MaxRTT() time.Duration {
	if len(s.RTT) == 0 {
		return 0
	}
	max := s.RTT[0]
	for _, rtt := range s.RTT {
		if rtt > max {
			max = rtt
		}
	}
	return max
}

func (s *Statistics) MinRTT() time.Duration {
	if len(s.RTT) == 0 {
		return 0
	}
	min := s.RTT[0]
	for _, rtt := range s.RTT {
		if rtt < min {
			min = rtt
		}
	}
	return min
}
