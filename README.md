<div align="center" style="width: 100%;">
 <img alt="tcping" src="Artwork/tcping_logo.png" style="width:50%;">
</div>

# TCPING

[![Go Report Card](https://goreportcard.com/badge/github.com/pouriyajamshidi/tcping)](https://goreportcard.com/report/github.com/pouriyajamshidi/tcping)
[![CodeFactor](https://www.codefactor.io/repository/github/pouriyajamshidi/tcping/badge)](https://www.codefactor.io/repository/github/pouriyajamshidi/tcping)
[![Go](https://github.com/pouriyajamshidi/tcping/actions/workflows/.github/workflows/codeql-analysis.yml/badge.svg)](https://github.com/pouriyajamshidi/tcping/actions/workflows/go.yml)
[![Docker container build](https://github.com/pouriyajamshidi/tcping/actions/workflows/container-publish.yml/badge.svg)](https://github.com/pouriyajamshidi/tcping/actions/workflows/container-publish.yml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pouriyajamshidi/tcping)
[![Go project version](https://badge.fury.io/go/github.com%2Fpouriyajamshidi%2Ftcping.svg)](https://badge.fury.io/go/github.com%2Fpouriyajamshidi%2Ftcping)
![Download](https://img.shields.io/github/downloads/pouriyajamshidi/tcping/total.svg?label=DOWNLOADS&logo=github)

A cross-platform ping program for `TCP` ports inspired by the Linux's ping utility. This program will send `TCP` probes to an `IP address` or a `hostname` specified by you and prints the result. It works with both `IPv4` and `IPv6`.

TCPING uses different `TCP sequence numbering` for successful and unsuccessful probes, so that when you look at the results and spot a failed probe, understanding the total packet drops to that point would be illustrative enough.

- Monitor your network connection.
- Determine packet loss.
- Analyze the network's latency.
- Show `min`/`avg`/`max` probes latency.
- Use the `-r` flag to retry hostname resolution after a predetermined number of ping failures. If you want to test your `DNS` load balancing or Global Server Load Balancer `(GSLB)`, you should utilize this option..
- Print connection statistics on `Enter` key press.
- Display the longest encountered downtime and uptime duration and time.
- Monitor and audit your peers network.
- Calculate the total uptime/downtime when conducting a maintenance.
- An alternative to `ping` in environments that `ICMP` is blocked.

---

## Table of Contents

- [TCPING](#tcping)
  - [Table of Contents](#table-of-contents)
  - [Demos](#demos)
    - [Vanilla usage](#vanilla-usage)
    - [Retry resolve (`-r`) flag](#retry-resolve--r-flag)
  - [Download the executables](#download-the-executables)
  - [Usage](#usage)
    - [On `Linux` and `macOS`](#on-linux-and-macos)
    - [On `Windows`](#on-windows)
    - [Using Docker](#using-docker)
  - [Flags](#flags)
  - [Tips](#tips)
  - [Notes](#notes)
  - [Contributing](#contributing)
  - [Tested on](#tested-on)
  - [Sponsor us](#sponsor-us)
  - [Contact me](#contact-me)
  - [License](#license)

---

## Demos

### Vanilla usage

![tcping](Images/tcping.gif)

---

### Retry resolve (`-r`) flag

![tcping](Images/tcpingrflag.gif)

---

## Download the executables

- ### [Windows](https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping_Windows.zip)

- ### [Linux](https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping_Linux.zip)

- ### [macOS](https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping_MacOS.zip) - also available through `brew`

  ```bash
  brew install pouriyajamshidi/tap/tcping
  ```

When the download is complete, head to the [usage](#usage) section.

**Alternatively**, you can:

- Install using `go install`:

  ```bash
  go install github.com/pouriyajamshidi/tcping@latest
  ```

- Use the `Docker` images:

  ```bash
  docker pull pouriyajamshidi/tcping:latest
  ```

  > Image is also available on GitHub container registry:

  ```bash
  docker pull ghcr.io/pouriyajamshidi/tcping:latest
  ```

- Or compile the code yourself by running the `make` command in the `tcping` directory:

  ```bash
  make build
  ```

  This will give you a compressed file with executables for all the supported operating systems inside the `executables` folder.

---

## Usage

If you have decided to download the executables using the [aforementioned links](#download-the-executables), go to the folder containing the file and extract it. Then, depending on your operating system, follow the instructions below:

- [Linux and macOS](#on-linux-and-macos)
- [Windows](#on-windows)
- [Docker images](#using-docker)

### On `Linux` and `macOS`

Make the file executable:

```bash
chmod +x tcping
```

For easier use, copy the executable to your system `PATH` like `/usr/bin/`:

```bash
sudo cp tcping /usr/local/bin/
```

Then run it like, `tcping <hostname/IP address> <port>`. For instance:

```bash
tcping www.example.com 443
# OR
tcping 10.10.10.1 22
```

Specifying the `-r` option will cause a name resolution retry after a certain number of failures. For instance:

```bash
tcping www.example.com 443 -r 10
# OR
tcping -r 10 www.example.com 443
```

> The `-r 10` in the command above will result in a retry of name resolution after 10 probe failures.

### On `Windows`

We recommend [Windows Terminal](apps.microsoft.com/store/detail/windows-terminal/9N0DX20HK701) for the best experience and proper colorization.

For easier use, copy `tcping.exe` to your system [PATH](https://www.howtogeek.com/118594/how-to-edit-your-system-path-for-easy-command-line-access/) like `C:\Windows\System32` and run it like:

```powershell
tcping www.example.com 443

# OR provide the -r flag to
# enable name resolution retries after a certain number of failures:
tcping www.example.com 443 -r 10
```

> If you prefer not to add the executable to your `PATH`, go to the folder that contains the `tcping.exe` open up the terminal and run the following command:

```powershell
.\tcping.exe 10.10.10.1 22
```

**Please note, if you copy the program to your system `PATH`, you don't need to specify `.\` and the `.exe` extension to run the program anymore.**

### Using Docker

The Docker image can be used like:

```bash
# Using Docker Hub
docker run -it pouriyajamshidi/tcping:latest example.com 443

# Using GitHub container registry:
docker run -it ghcr.io/pouriyajamshidi/tcping:latest example.com 443
```

---

## Flags

The following flags are available to control the behavior of application:

| Flag | Description                                               |
| ---- | --------------------------------------------------------- |
| `-r` | Retry resolving a hostname after `<n>` number of failures |
| `-j` | Output in JSON format                                     |
| `-u` | Check for updates                                         |
| `-v` | Print version                                             |

---

## Tips

- Press the `Enter` key while the program is running to examine the summary of all probes without shutting it down, as shown in the [demos](#demos) section.

---

## Notes

`TCPING` is constantly being improved, adding numerous new features and fixing bugs. Be sure to look for updated versions..

```bash
tcping -u
```

## Contributing

Pull requests are welcome to solve bugs, add new features and also to help me with the open issues that can be found here ![GitHub issues](https://img.shields.io/github/issues/pouriyajamshidi/tcping.svg).

1. Pick any issue that you feel comfortable with.
2. Fork the repository.
3. Create a branch.
4. Commit your work.
5. Add tests if possible.
6. Run the tests `go test` or `make test`.
7. Create a pull request

Please make sure that your pull request only works on one specific issue and doesn't handle two or more tickets. This makes it simpler for us to examine your pull request and helps keep the git history clean.

## Tested on

Windows, Linux and macOS.

## Sponsor us

[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/pouriyajamshidi)  
[![sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/pouriyajamshidi)  
![GitHub Sponsor](https://img.shields.io/github/sponsors/pouriyajamshidi?label=Sponsor&logo=GitHub)

## Contact me

[![LinkedIn](https://img.shields.io/badge/LinkedIn-0077B5?style=for-the-badge&logo=linkedin&logoColor=white)](https://www.linkedin.com/in/pouriya-jamshidi/)

## License

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
