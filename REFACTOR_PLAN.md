# tcping v3 Refactoring Plan

## Phase 1: Critical Bugs (Must Fix)

### 1.1 Fix CSV Duplicate Column Bug
**File:** `printers/csv.go:199`

**Issue:** When `ShowSourceAddress` is true, columns are appended twice.

**Fix:**
```go
// Current (wrong):
if p.opt.ShowSourceAddress {
    record = append(record, s.SourceAddr(), strconv.FormatUint(uint64(s.OngoingSuccessfulProbes), 10), s.RTTStr())
}
record = append(record, strconv.FormatUint(uint64(s.OngoingSuccessfulProbes), 10), s.RTTStr())

// Should be:
if p.opt.ShowSourceAddress {
    record = append(record, s.SourceAddr())
}
record = append(record, strconv.FormatUint(uint64(s.OngoingSuccessfulProbes), 10), s.RTTStr())
```

**Estimated Impact:** Low effort, fixes correctness issue

---

## Phase 2: Error Handling & Exit Behavior (High Priority)

### 2.1 Remove os.Exit() from PrintError Implementations
**Files:**
- `printers/db.go:449-452`
- Other printers for consistency

**Issue:** `DatabasePrinter.PrintError()` calls `os.Exit(1)`, violating separation of concerns and making testing impossible.

**Fix:**
- Remove `os.Exit(1)` from all `PrintError` implementations
- Let the caller (application layer) decide whether to exit
- Document in `Printer` interface that `PrintError` should not exit

### 2.2 Refactor Input Validation to Return Errors
**Files:**
- `internal/app/input.go:177-190` (`convertAndValidatePort`)
- `internal/app/input.go:192-219` (`convertAndValidateRetry`)
- Any other validation functions calling `os.Exit()`

**Issue:** Validation functions call `os.Exit()` directly, making them untestable.

**Fix:**
```go
// Before:
func convertAndValidatePort(portStr string) uint16 {
    port, err := strconv.ParseUint(portStr, 10, 16)
    if err != nil {
        fmt.Printf("Invalid port number: %s\n", portStr)
        os.Exit(1)
    }
    return uint16(port)
}

// After:
func convertAndValidatePort(portStr string) (uint16, error) {
    port, err := strconv.ParseUint(portStr, 10, 16)
    if err != nil {
        return 0, fmt.Errorf("invalid port number: %s: %w", portStr, err)
    }
    if port == 0 {
        return 0, fmt.Errorf("port number must be greater than 0")
    }
    return uint16(port), nil
}
```

**Update call sites** in `internal/app/input.go` to handle returned errors.

---

## Phase 3: Code Duplication Elimination (Medium Priority)

### 3.1 Extract Finalization Logic in Prober.Probe()
**File:** `prober.go:102-227`

**Issue:** Uptime/downtime finalization block is duplicated 3 times (lines 118-130, 135-147, 209-221).

**Fix:** Extract to method:
```go
func (p *Prober) finalizeStatistics() {
    p.Statistics.EndTime = time.Now()
    p.Statistics.UpTime = p.Statistics.EndTime.Sub(p.Statistics.StartTime)

    if p.Statistics.DestWasDown {
        downDuration := p.Statistics.EndTime.Sub(p.Statistics.StartOfDowntime)
        p.Statistics.TotalDowntime += downDuration
        statistics.SetLongestDuration(p.Statistics.StartOfDowntime, downDuration, &p.Statistics.LongestDown)
    } else if !p.Statistics.StartOfUptime.IsZero() {
        upDuration := p.Statistics.EndTime.Sub(p.Statistics.StartOfUptime)
        p.Statistics.TotalUptime += upDuration
        statistics.SetLongestDuration(p.Statistics.StartOfUptime, upDuration, &p.Statistics.LongestUp)
    }
}
```

Replace all 3 duplicated blocks with `p.finalizeStatistics()`.

### 3.2 Reduce Printer Code Duplication (Optional - Larger Refactor)
**Files:** All files in `printers/`

**Issue:** Each printer has nearly identical 8-way branching logic for formatting probe results.

**Approach:** Create a message formatter:
```go
type ProbeMessage struct {
    Timestamp string
    Target    string  // "hostname (ip)" or just "ip"
    Port      uint16
    Source    string  // optional, empty if not shown
    RTT       string
    Streak    uint
    Success   bool
}

func FormatProbeResult(s *Statistics, opts FormatOptions) ProbeMessage {
    // Build canonical message structure
}
```

Each printer then formats this message in its own style (colored text, JSON, CSV row, etc.).

**Note:** This is a larger refactor. Consider doing it after Phase 1-3 are complete.

---

## Phase 4: Domain Model Improvements (Medium Priority)

### 4.1 Fix Timeout Semantics
**File:** `prober.go:39-41`

**Issue:**
```go
func WithTimeout(timeout time.Duration) ProberOption {
    return func(p *Prober) {
        p.Timeout = timeout + p.Interval  // Why add interval?
    }
}
```

User expects `-t 5` to mean 5 seconds, not `5 + interval`.

**Fix:**
```go
func WithTimeout(timeout time.Duration) ProberOption {
    return func(p *Prober) {
        p.Timeout = timeout
    }
}
```

**Verify:** Check if there's a legitimate reason for adding `p.Interval`. If the intent is "time between probes", that should be a separate concept/option.

### 4.2 Separate Display Concerns from Statistics
**File:** `statistics/statistics.go:22-72`

**Issue:** The `Statistics` struct mixes:
- Domain data (IP, Port, Protocol)
- Display options (ShowFailuresOnly, WithTimestamp, WithSourceAddress)
- Presentation helpers (DownTime "for printing")

**Current State:**
```go
type Statistics struct {
    // Target info
    IP       netip.Addr
    Port     uint16
    Hostname string
    Protocol protocol

    // Display options (shouldn't be here!)
    WithTimestamp      bool
    WithSourceAddress  bool
    ShowFailuresOnly   bool

    // ... 20+ other fields
}
```

**Proposed Structure:**
```go
// Domain: What we're probing
type Target struct {
    IP       netip.Addr
    Port     uint16
    Hostname string
    Protocol string  // or enum if needed
    LocalAddr net.Addr
}

// Domain: Result of a single probe
type ProbeResult struct {
    Success   bool
    RTT       time.Duration
    Timestamp time.Time
    Error     error
}

// Domain: Aggregated statistics
type ProbeStatistics struct {
    Target Target

    // Counters
    TotalSuccessfulProbes   uint
    TotalUnsuccessfulProbes uint
    OngoingSuccessfulProbes uint
    OngoingFailedProbes     uint

    // RTT stats
    TotalRTT    time.Duration
    AverageRTT  time.Duration
    MinRTT      time.Duration
    MaxRTT      time.Duration
    LastRTT     time.Duration

    // Time tracking
    StartTime        time.Time
    EndTime          time.Time
    UpTime           time.Duration
    TotalUptime      time.Duration
    TotalDowntime    time.Duration
    StartOfUptime    time.Time
    StartOfDowntime  time.Time
    LongestUpStart   time.Time
    LongestUp        time.Duration
    LongestDownStart time.Time
    LongestDown      time.Duration

    // State
    DestWasDown bool
}

// Display configuration (lives in printer packages or config)
type DisplayOptions struct {
    WithTimestamp      bool
    WithSourceAddress  bool
    ShowFailuresOnly   bool
}
```

**Migration Strategy:**
1. Create new types in `statistics/` package
2. Add methods to convert from old `Statistics` to new structure
3. Update one printer at a time to use new structure
4. Remove old `Statistics` struct once all printers migrated

**Note:** This is a significant refactor. Consider doing it in a separate branch/PR.

---

## Phase 5: Cleanup (Low Priority)

### 5.1 Remove Unused Protocol Constants
**File:** `statistics/statistics.go:14-20`

**Issue:** Only `TCP` is used. `UDP`, `HTTP`, `HTTPS`, `ICMP` are premature generalization.

**Options:**
1. Delete unused constants if no plans to implement them soon
2. Keep them if you're planning to add support in v3
3. Add TODO comments with issue links if planned

### 5.2 Implement or Remove DNS Retry Feature
**File:** `internal/app/input.go:51`

**Issue:** `RetryResolveAfter` field is set but never used in probe loop.

**Options:**
1. Implement the DNS re-resolution feature
2. Remove the unused field if not planned
3. Add TODO comment with issue number if planned for future

### 5.3 Review DNS Randomization Behavior
**File:** `dns/dns.go:125`

**Issue:**
```go
return ipAddrs[rand.Intn(len(ipAddrs))], nil
```

Uses `math/rand` without explicit seeding (auto-seeds in Go 1.20+).

**Question:** Should IP selection be deterministic (always first) or random? Consider user expectations.

**Options:**
1. Keep random selection (current)
2. Change to deterministic (first address)
3. Make it configurable via option

---

## Implementation Order Recommendation

1. **Phase 1** - Quick wins, fixes bugs
2. **Phase 2** - Improves testability and architecture
3. **Phase 3.1** - Small duplication fix
4. **Phase 4.1** - Clarifies timeout behavior
5. **Phase 5** - Cleanup
6. **Phase 3.2** (Optional) - Large printer refactor
7. **Phase 4.2** (Optional) - Large statistics refactor

**Phases 3.2 and 4.2 are optional larger refactors.** The codebase is functional without them. Consider doing them only if:
- You're planning to add more printers (3.2)
- You're planning to add more protocols/features (4.2)
- You find the current code hard to maintain

---

## Testing Strategy

After each phase:
1. Run existing test suite: `go test ./...`
2. Manual testing of main scenarios:
   - Basic TCP ping: `tcping example.com 80`
   - With timeout: `tcping -t 5 example.com 80`
   - With count: `tcping -c 10 example.com 80`
   - Different output formats: JSON, CSV, DB (if applicable)
3. Test error cases (invalid port, unreachable host, etc.)

---

## Questions to Consider

1. **Phase 4.1 (Timeout):** Is there a legitimate reason `Interval` was added to `Timeout`? Check git history or original requirements.

2. **Phase 4.2 (Statistics refactor):** How much future expansion is planned? This refactor makes sense if you're adding UDP/ICMP support soon.

3. **Phase 5.3 (DNS randomization):** What should the behavior be when multiple IPs are returned? Random load balancing or deterministic?
