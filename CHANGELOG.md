# Changelog

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
