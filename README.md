# TCPING

A cross-platform ping program for ```TCP``` ports similar to Linux's ping utility. This program will send ```TCP``` probes to an ```IP address``` or a ```hostname``` specified by you and prints the result. It works with both `IPv4` and `IPv6`.

It uses different `TCP sequence numbering` for successful and unsuccessful probes, so that when you look at the results after a while, and seeing for instance, a failed probe, understanding the total packet drops so far would be illustrative enough.

## Application

* Calculate packet loss.
* Assess latency of your network.
* Show min/avg/max probes latency.
* Monitor and audit your peers network.
* Calculate total up or downtime when conducting a maintenance.
* An alternative to `ping` in environments that `ICMP` is blocked.

## Images

![WindowsVersion](/Images/windowsVersion.png)

## Demo

[![asciicast](https://asciinema.org/a/bNMtJKmujGEpfEhvDiTeSvtO4.svg)](https://asciinema.org/a/bNMtJKmujGEpfEhvDiTeSvtO4)

## Download the executables for

* ### [Windows](https://github.com/pouriyajamshidi/tcping/releases/download/1.1.0/tcping_Windows.rar)

* ### [Linux](https://github.com/pouriyajamshidi/tcping/releases/download/1.1.0/tcping_Linux.rar)

* ### [macOS](https://github.com/pouriyajamshidi/tcping/releases/download/1.1.0/tcping_MacOS.rar)

In addition to downloading the executables, you can also compile the code yourself by following below instructions:

```bash
go get github.com/gookit/color
go env -w GO111MODULE=auto # for Go versions above 1.15
go build tcping
```

## Usage

Go to the directory/folder in which you have downloaded the application.

### On ```Linux``` and ```macOS```:

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

* While the program is running, upon pressing the ```enter``` key, the summary of all probes will be shown as depicted in the [demo](#Demo).

## Notes

This program is still in a ```beta``` stage. There are several shortcomings that I will rectify in the near future.
There are also some parts that might not make sense because this program was mainly designed to help me experiment with `Go`.

## Tested on

Windows, Linux and macOS.

## Contributing

Pull requests are welcome.

## License

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
