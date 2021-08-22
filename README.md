# TCPING

A cross-platform ping program for ```TCP``` ports inspired by the Linux's ping utility. This program will send ```TCP``` probes to an ```IP address``` or a ```hostname``` specified by you and prints the result. It works with both `IPv4` and `IPv6`.

It uses a different `TCP sequence numbering` for successful and unsuccessful probes, so that when you look at the results and spot a failed probe, understanding the total packet drops to that point would be illustrative enough.

## Features And Application

* Monitor your network connection.
* Print statistics on pressing the `Enter` key.
* Calculate packet loss.
* Assess the latency of your network.
* Show min/avg/max probes latency.
* Monitor and audit your peers network.
* Calculate the total uptime/downtime when conducting a maintenance.
* An alternative to `ping` in environments that `ICMP` is blocked.

## Images

![WindowsVersion](/Images/windowsVersion.png)

## Demo

[![asciicast](https://asciinema.org/a/AcdJc4EYY1b8Wtfi29YqcZlfg.svg)](https://asciinema.org/a/AcdJc4EYY1b8Wtfi29YqcZlfg)

## Download the executables for

* ### [Windows](https://github.com/pouriyajamshidi/tcping/releases/download/1.1.9/tcping_Windows.zip)

* ### [Linux](https://github.com/pouriyajamshidi/tcping/releases/download/1.1.9/tcping_Linux.zip)

* ### [macOS](https://github.com/pouriyajamshidi/tcping/releases/download/1.1.9/tcping_MacOS.zip)

In addition to downloading the executables, you can also compile the code yourself by running the `make` command:

```bash
make build
```

## Usage

Go to the directory/folder in which you have downloaded the application.

### On ```Linux``` and ```macOS```

```bash
sudo chmod +x tcping
```

For easier use, you can copy it to your system ```PATH``` like /bin/ or /usr/bin/

```bash
sudo cp tcping /bin/
```

Then run it like, `tcping <hostname/IP address> <port>`. For instance:

```bash
tcping www.example.com 443
```

OR

```bash
tcping 10.10.10.1 22
```

### On ```Windows```

I recommend ```Windows Terminal``` for the best experience and proper colorization.

For easier use, copy ```tcping.exe``` to your system ```PATH``` like C:\Windows\System32 or from your terminal application, go to the folder that contains the ```tcping.exe``` program.

Run it like:

```powershell
.\tcping www.example.com 443
```

OR

```powershell
tcping 10.10.10.1 22
```

**Please note, if you copy the program to your system ```PATH```, you don't need to specify ```.\``` to run the program anymore.**

## Tips

* While the program is running, upon pressing the `Enter` key, the summary of all probes will be shown as depicted in the [demo](#Demo).

## Notes

This program is still in a ```beta``` stage. There are several shortcomings that I will rectify in the near future.
There are also some parts that might not make sense because this program was mainly designed to help me experiment with `Go`.

## TODO

* [x] Convert to a proper Go project
* [x] If IP is given, skip the DNS lookup.
* [ ] Display the longest downtime
* [ ] Implement a better uptime timing calculation instead of a hardcoded one (default TCP connection timeout is 1 second).
* [ ] Ability to recheck DNS after a certain number of failed requests (for geographically sparse sites).
* [ ] Minutes to hour(s) conversion

## Tested on

Windows, Linux and mac OS.

## Contributing

Pull requests are welcome to solve bugs, add new features and also to help me with the [TODOs](#todo).

## License

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
