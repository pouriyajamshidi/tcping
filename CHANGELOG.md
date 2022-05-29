# Changelog

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
