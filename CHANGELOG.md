# Changelog

## v2.x.x - Unreleased

- release: add tcping to [WinGet](https://learn.microsoft.com/en-us/windows/package-manager/winget) [#113](https://github.com/pouriyajamshidi/tcping/issues/113)
- bug: fix name resolution in static builds with `-4` flag causing name resolution failures due to _IPv4-mapped IPv6 addresses_
- refactor: rename plane to plain printer
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
