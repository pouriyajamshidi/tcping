# Changelog

## v3.0.0 - Unreleased (Work in progress)

- build (breaking): remove sqlite3 support from Windows binaries
- build: bump stale days before close from 7 to 14
- build: ensure to fail on any linting issues
- build: simplify container publish workflow
- build: general cleanups
- build: support more container architectures
- improvement: make print statistics (when the **Enter** key is pressed) snappy. No more waiting when using high probe intervals
- improvement: when the `-I` flag is used, show the interface name on probe **failures** too
- refactor: drop `TimeFormat` constants in favor of stdlib's `time.DateTime`
- refactor: drop `HourFormat` constants in favor of stdlib's `time.TimeOnly`
- templates: improve pull and bug report templates
- docs: we have a new logo thanks to Gemini!
- refactor: modernize with `go fix`
- dependencies: replace `github.com/google/go-github` with Go's built-in HTTP library
- docs: grammar fix in the README.md thanks to @taiman724
- project structure: move Artwork and Images folders to the docs folder
- refactor (breaking): rewrite the JSON output events - each event now carries a minimal, consistent payload (`start`, `probe`, `retry`, `downtimeDuration`, `uptimeDuration`, `nameResolution`, `statistics`, `error`) instead of one large struct with inconsistent field names; the `info` event type is removed
- refactor (breaking): rewrite the sqlite3 schema - drop the generic `type` column and rename `success` to `reachable` and `time` to `latency`
- refactor (breaking): rename the CSV `Status` column to `Reachable` (values are now lowercase `true`/`false` instead of `Reply`/`No Reply`) and the stats file's `Metric` column to `Statistic`
- refactor (breaking): `-v` now turns on verbose output for HTTP(S) probes; the version is printed with `--version` instead
- fix: CSV probe rows no longer duplicate the RTT and connection-count values when `--show-source-address` is used
- fix: fix a bug that showed downtime as uptime
- fix: fix incorrect timestamp handling in the JSON printer
- feat: add HTTP(S) probing by giving a URL as the target, e.g. `tcping https://example.com/health`, with `-v` to show the HTTP and TLS details of every probe and `--skip-tls` to skip certificate verification
- feat: CSV output filenames now get a date/time suffix by default; add `--csv-no-timestamp` to reuse the same file across runs instead
- feat: add `--resolve-every-probe` flag to resolve the target's hostname before every probe instead of only at startup or on retry (`-r`)
- feat: add `--dns-timeout` flag to configure the DNS resolution timeout; also fixes a bug where it was silently ignored and the 2-second default was always used regardless of what was configured
- feat: show how long hostname resolution took - at startup, on every retry-resolve, and per-entry in the "IP address changes" summary
- feat: report the mean deviation of the latency (`mdev`) alongside the minimum, average and maximum, the same way `ping` does. It is carried by every printer: the summary line, the JSON `latencyMdev` field, the CSV `Latency Mdev` row, the sqlite3 `latency_mdev` column, Alloy's `tcping_rtt_milliseconds{stat="mdev"}` and InfluxDB's `rtt_mdev_ms`
- feat: add `--alloy` flag to send the results to a [Grafana Alloy](https://grafana.com/docs/alloy/latest/) OTLP HTTP endpoint as metrics instead of printing them, along with `--alloy-stats-interval` to control how often the statistics are sent
- feat: add `--influxdb` flag to write the results to an InfluxDB v2 or v3 server as line protocol, along with `--influxdb-org`, `--influxdb-bucket` and `--influxdb-stats-interval`. The API token is read from the `INFLUXDB_TOKEN` environment variable
- improvement: `-I` now keeps working correctly when the interface has both an IPv4 and an IPv6 address and the target's resolved address family changes mid-run
- improvement: DNS resolution is now sourced from `-I`'s interface too, matching what probes already did
- improvement: show how long the target was up right when it starts failing, mirroring the existing downtime message
- improvement: probing now starts immediately instead of waiting for the first interval to elapse
- fix: hostname retry-resolve (`-r`) now actually changes the address being probed instead of only updating what's displayed
- fix: the summary's `duration (HH:MM:SS)` now always matches the `started at`/`ended at` timestamps instead of silently drifting from them
- feat: add `--omit-stats` flag to skip printing the statistics when the program exits. Pressing the **Enter** key still shows them
- refactor: simplify RTT min/avg/max tracking into a running calculation instead of storing every sample
- refactor: consolidate uptime/downtime tracking and remove duplicated/unused `Statistics` fields

## v2.8.0 - 2026-05-11

- feat: add a _non-interactive_ mode through `--non-interactive` flag so that tcping can run in the background using `nohup` or `disown`
- feat: add support for `host:port` format in command arguments in [362](https://github.com/pouriyajamshidi/tcping/pull/362) thanks to @bingoohuang
- fix: omit printing the IP address twice when the given target is an IP address itself
- fix: add missing comma separator in no-color statistics output in [376](https://github.com/pouriyajamshidi/tcping/issues/376) thanks to @clarabennettdev
- fix: version typo resulting in erroneous update message raised in [313](https://github.com/pouriyajamshidi/tcping/issues/313)
- build: bump Golang base image to `1.26.3-alpine3.23`
- documents: fix typo in the Chinese README in [386](https://github.com/pouriyajamshidi/tcping/pull/386) thanks to @peeweep
- documents: clarify the difference between static and dynamic binaries in README raised in [357](https://github.com/pouriyajamshidi/tcping/issues/357)

## v2.7.1 - 2025-01-26

- release: add tcping to [WinGet](https://learn.microsoft.com/en-us/windows/package-manager/winget) [#113](https://github.com/pouriyajamshidi/tcping/issues/113)
- bug: fix name resolution in static builds with `-4` flag causing name resolution failures due to _IPv4-mapped IPv6 addresses_
- CI: apply **Revive** suggestions
- CI: add **Revive** to CI
- CI: add **Revive** config
- documents: revamp and simplify the README file
- documents: update the Chinese translation thanks to @edwinjhlee

## v2.7.0 - 2025-01-18

- new feature: implement **csv** output through `--csv <filename>` flag [#254](https://github.com/pouriyajamshidi/tcping/pull/254) thanks to @Ilhan-Personal
- new feature: implement plain (color-less) output through `--no-color` flag [#253](https://github.com/pouriyajamshidi/tcping/issues/253)
- new feature: implement display of source IP address and port used to create TCP connections through `--show-source-address` flag [#249](https://github.com/pouriyajamshidi/tcping/issues/249)
- refactor: rename `planePrinter` to `colorPrinter` to match the actual functionality of the function
- refactor: rename `localAddr` to `sourceAddr` throughout the code-base for better clarity
- refactor: complete rewrite of the **Makefile** thanks to @cyqsimon
- refactor: add containerization section in the **Makefile** thanks to @cyqsimon
- fix: crash on database writes when hostname includes a hyphen thanks to @pro0o
- documents: add Chinese translation thanks to @edwinjhlee
- documents: add link to [X CMD](https://x-cmd.com/pkg/tcping) thanks to @edwinjhlee
- tests: add new tests for `printProbeSuccess` and `printProbeFail` thanks to @basil-gray
- tests: add tests for `show-local-address` flag
- dependencies:
  - crypto v0.28.0 => v0.32.0
  - exp v0.0.0-20241004190924-225e2abe05e6 => v0.0.0-20250106191152-7588d65b2ba8
  - sys v0.26.0 => v0.29.0
  - modernc.org/libc v1.61.6 => v1.61.8
  - modernc.org/memory v1.8.0 => v1.8.2
  - modernc.org/sqlite v1.34.4 => v1.34.5

## v2.6.0 - 2024-10-05

- new feature: add `-D` flag to display date and time in probe output by @SYSHIL
- new feature: add `-h` flag to show available flags by @karimalzalek
- fix: display `second` instead of `seconds` on probe failures that convert to a value more than 1 and less than 1.1 second
- refactor: Makefile: Split build section into smaller, distinct targets by @iskiy

## v2.5.0 - 2024-01-13

- new feature: add `-show-failures-only` flag to omit printing successful probes
- new feature: re-add **static** Linux binary. Thanks to @daniql
- new feature: add support for Linux `arm64` in Makefile. Thanks to @ChrisClarke246
- fix: extra precision for seconds calculation when the value is under a second. Thanks to @daniql
- refactor: migrate to a pure-Go `sqlite` package. Thanks to @wizsk
- refactor: user flag handlers
- cleanup: user input functions. Thanks to @friday963
- chore: fix typos

## v2.4.0 - 2023-09-10

- new feature: add `-i` to specify the interval between sending probes. Thanks to @luca-patrignani
- new feature: add `-I` to specify the source interface to use for sending probes. Thanks to @wizsk
- new feature: add `-t` to specify a custom timeout for probes. Thanks to @luca-patrignani
- new feature: add `--db` to specify the path and file name to store tcping output to sqlite database. e.g. `--db /tmp/tcping.db`. Thanks to @wizsk
- fix: add `rtt` to JSON output
- fix: CI warning thanks to @wizsk
- refactor: remove unnecessary custom types
- refactor: memory align `structs`
- refactor: Debian packaging instructions

## v2.0.0 - 2023-08-05

- new feature: add `-c` or count flag to exit **TCPING** after a certain amount of probes specified by user thanks to @ravsii
- new feature: add **BSD** support
- new feature: add **Debian** package to make **TCPING** `apt installable`
- fix: packet loss `NaN` when program terminated too quickly thanks to @ravsii
- fix: random IP address selector index out of range bug
- fix: display format of IPv4 embedded in IPv6 addresses
- fix: time report bug. Everything is now accurate
- fix: Enter key detection for Windows machines
- refactor: complete overhaul of time calculation. **TCPING** now is hack-free when it comes to time handling thanks to @ravsii
- refactor: memory align `structs`
- refactor: improve code readability
- refactor: refactor `stats struct` and extract user input to a separate `struct`
- refactor: Enter key detection logic
- refactor: name resolution handling. The maximum allowed time to wait for DNS response is now 2 seconds
- refactor: and unify exit points thanks to @ravsii
- tests: add more test special thanks to @ravsii
- enhancement: add dependabot
- docs: improve documentation

## v1.22.1 - 2023-5-14

- new feature: implement JSON output thanks to @ravsii
- new feature: implement JSON output [prettifier](https://github.com/pouriyajamshidi/tcping/raw/master/Images/gifs/tcping_json_pretty.gif) thanks to @ravsii
- fix IP version selection bug when `-4` or `-6` flags are passed

## v1.21.2 - 2023-5-8

- make `stats` struct fields' names uniform
- add `|` separator to summary report for better visibility

## v1.21.1 - 2023-5-8

- fix retry resolve logic

## v1.21.0 - 2023-5-7

- add option to enforce the use of IPv4 `-4` or IPv6 `-6` addresses only
- instead of always picking the first, randomly pick an address from the list of resolved IP addresses

## v1.20.0 - 2023-4-22

- add hostname, IP and port number to summary output

## v1.19.2 - 2023-4-7

- display stats even if all the probes had failed update version
- update version
- incorporate sha256sum into Makefile

## v1.19.1 - 2023-3-4

- close `TCP` connections faster to lessen the resource utilization on target

## v1.19.0 - 2023-2-26

- implement sub-millisecond timing report to make it suitable for Data center and Cloud environments
- refactor `tcping` function and simplify it
- fix downtime report miscalculation
- fix picking of go version
- improve build process
- changed `ipAddress` type from string to `netip.Addr` thanks @segfault99
- fix `statsprinter` formats thanks @segfault99
- upgrade actions thanks @wutingfeng
- fix undeclared `statsPrinter` warning
- fix code scanning alert - Incorrect conversion between integer types #43
- add `stale` workflow
- add new logo
- add Linux brew section
- add docker demo recording
- restructure README file
- update dependencies and bump Go version
- improve Makefile
- fix tag detection on Actions workflow
- add `Go` version to `CodeQL`
- add `downloads` badge
- improve checkUpdate message
- fix go install guide
- fix bug report template
- create SECURITY.md
- improve pull request template
- improve stale workflow

## v1.12.0 - 2022-7-10

- add preliminary JSON output support thanks @icemint0828 for collaboration
- add Docker container images on Docker Hub and GitHub container registry
- add and optimize GitHub workflows
- small refactoring and cleanups
- add -v flag to show version
- improve code readability
- add logo thanks @code-hyker

## v1.9.0 - 2022-5-29

- Add `-r` flag to retry resolving the hostname after a certain amount of probe failures (thanks to @icemint0828)
- Show statistics if the RTT is less than 1ms (thanks to @icemint0828)
- Show longest uptime similar to longest downtime (thanks to @icemint0828)
- Improve time calculation and display time in reports (thanks to @icemint0828)
- Add initial test cases (thanks to @icemint0828)
- General refactoring, fixes and decrease of resource utilization
- Update dependencies
- Update `Makefile` to include `go fmt` command in `build`
- Update `Makefile` to include `go test` command in `build`

## v1.4.4 - 2022-2-26

- Improve time constants for better readability

## v1.4.3 - 2022-2-21

- Revert successful reply text color

## v1.4.2 - 2022-2-20

- Memory alignment for rttResults struct

## v1.4.1 - 2022-2-20

- Make hour format a constant

## v1.4.0 - 2022-2-19

- Remove sort function to increase performance
- General refactoring to make the code more readable

## v1.3.0 - 2022-2-9

- Fix longest downtime bug

## v1.2.0 - 2022-2-6

- Improve memory alignment
- Add display of longest downtime
- Add display of runtime duration
- Add display of last successful and unsuccessful probes
- General improvements and cleanup
