package main

import (
	"time"
)

type stats struct {
	startTime                 time.Time
	endTime                   time.Time
	startOfUptime             time.Time
	startOfDowntime           time.Time
	lastSuccessfulProbe       time.Time
	lastUnsuccessfulProbe     time.Time
	retryHostnameResolveAfter *uint // Retry resolving target's hostname after a certain number of failed requests
	ip                        string
	port                      string
	hostname                  string
	rtt                       []uint
	totalDowntime             time.Duration
	totalUptime               time.Duration
	longestDowntime           longestTime
	totalSuccessfulPkts       uint
	totalUnsuccessfulPkts     uint
	ongoingUnsuccessfulPkts   uint
	retriedHostnameResolves   uint
	longestUptime             longestTime
	wasDown                   bool // Used to determine the duration of a downtime
	isIP                      bool // If IP is provided instead of hostname, suppresses printing the IP information twice
	shouldRetryResolve        bool
	StatsPrinter
}
