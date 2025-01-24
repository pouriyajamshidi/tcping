<div align="center" style="width: 100%;">
 <img alt="tcping" src="Artwork/tcping_logo3.jpeg" style="width:70%;">
</div>

# TCPING

![Go Report Card](https://goreportcard.com/badge/github.com/pouriyajamshidi/tcping)
![CodeFactor](https://www.codefactor.io/repository/github/pouriyajamshidi/tcping/badge)
![Go](https://github.com/pouriyajamshidi/tcping/actions/workflows/.github/workflows/codeql-analysis.yml/badge.svg)
![Docker container build](https://github.com/pouriyajamshidi/tcping/actions/workflows/container-publish.yml/badge.svg)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pouriyajamshidi/tcping)
![Go project version](https://badge.fury.io/go/github.com%2Fpouriyajamshidi%2Ftcping.svg)
![Download](https://img.shields.io/github/downloads/pouriyajamshidi/tcping/total.svg?label=DOWNLOADS&logo=github)
![Docker Pulls](https://img.shields.io/docker/pulls/pouriyajamshidi/tcping)

A cross-platform ping program using `TCP` instead of `ICMP`, inspired by Linux's ping utility.

> [!TIP]
> This document is also available in [Chinese | 中文](README.cn.md).

Here are some of the features of **TCPING**:

- An alternative to `ping` in environments that `ICMP` is blocked.
- Outputs information in **colored**, **plain**, **JSON**, **CSV** and **sqlite3** formats.
- Monitor and audit your or your peers network latency, packet loss, and connection quality.
- Let's you specify the **source interface**, **timeout**, and **interval** between probes.
- Supports both `IPv4` or `IPv6` and lets you enforce using either.
- Prints total connection statistics by pressing the `Enter` key, without stopping the program.
- Reports the longest encountered `downtime` and `uptime` duration and time.
- Retries hostname resolution after a predetermined number of probe failures by using the `-r` flag . Suitable to test your `DNS` load balancing or Global Server Load Balancer `(GSLB)`.
- uses different `TCP sequence numbering` for _successful_ and _unsuccessful_ probes to infer the total failed or successful probes at a glance.

Check out the [demos](#demos) to get a look and feel of **tcping**.

---

## Table of Contents

- [TCPING](#tcping)
  - [Table of Contents](#table-of-contents)
  - [Download and Installation](#download-and-installation)
    - [Windows](#windows)
    - [macOS](#macos)
    - [Linux - Debian and Derivatives](#linux---debian-and-derivatives)
    - [BSD and Linux - Manual Way](#bsd-and-linux---manual-way)
    - [Alternative Ways](#alternative-ways)
  - [Usage](#usage)
  - [Flags](#flags)
  - [Demos](#demos)
    - [Basic usage](#basic-usage)
    - [Retry hostname lookup (`-r`) flag](#retry-hostname-lookup--r-flag)
    - [JSON output (`-j --pretty`) flag](#json-output--j---pretty-flag)
  - [Contributing](#contributing)
  - [Feature Requests and Issues](#feature-requests-and-issues)
  - [Help The Project](#help-the-project)
  - [License](#license)

---

## Download and Installation

We offer prebuilt binaries for various operating systems ([Windows](#windows), [Linux](#linux---debian-and-derivatives), [macOS](#macos), [Docker](#alternative-ways)) and architectures (_amd64_, _arm64_), which can be found on the [release page](https://github.com/pouriyajamshidi/tcping/releases/latest/).

Once you are done with the download and installation, head to the [usage](#usage) section.

### Windows

The best way to install **tcping** on Windows is through _Windows Package Manager_ by utilizing [WinGet](https://learn.microsoft.com/en-us/windows/package-manager/winget/?ref=github.com%2Fpouriyajamshidi%2Ftcping), which is available on practically all Windows _10_ and _11_ machines by default since September of 2020:

```powershell
winget install pj.tcping
```

> [!TIP]
> We recommend using [Windows Terminal](https://apps.microsoft.com/store/detail/windows-terminal/9N0DX20HK701) for the best experience and proper colorization.

If you wish to manually install **tcping**, extract the downloaded zip file and copy `tcping.exe` to your system [PATH](https://www.howtogeek.com/118594/how-to-edit-your-system-path-for-easy-command-line-access/) like `C:\Windows\System32`

> [!CAUTION]
> TCPING might falsely get flagged by Windows Defender or some anti-malware software. This is common among Go programs. Check out the official statement from the Go team [here](https://go.dev/doc/faq#virus).

### macOS

Install using `brew`:

```bash
brew install pouriyajamshidi/tap/tcping
```

You can also manually download and install **tcping** following the steps described in [this section](#bsd-and-linux---manual-way).

### Linux - Debian and Derivatives

On **Debian** and its flavors such as **Ubuntu**, download the `.deb` package:

```bash
wget https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-amd64.deb -O /tmp/tcping.deb
# Or for ARM64 machines
wget https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-arm64.deb -O /tmp/tcping.deb
```

And install it:

```bash
sudo apt install -y /tmp/tcping.deb
```

If you are using different Linux distros, proceed to [this section](#bsd-and-linux---manual-way).

### BSD and Linux - Manual Way

Download the file for your respective OS and architecture:

```bash
wget https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-freebsd-amd64-static.tar.gz
# Or for Linux ARM64 machines and using cURL
curl -LO https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-arm64-static.tar.gz
```

Extract the file:

```bash
tar -xvf tcping-freebsd-amd64-static.tar.gz
```

Make the file executable:

```bash
chmod +x tcping
```

Copy the executable to your system `PATH` like `/usr/local/bin/`:

```bash
sudo cp tcping /usr/local/bin/
```

> [!TIP]
> In case you have `brew` installed, you can install tcping using `brew install pouriyajamshidi/tap/tcping`

### Alternative Ways

These are some additional ways in which **tcping** can be installed:

- `Docker` images:

  ```bash
  docker pull pouriyajamshidi/tcping:latest
  # Or
  docker pull ghcr.io/pouriyajamshidi/tcping:latest
  ```

- Using `go install`:

  > This requires at least go version `1.23.1`

  ```bash
  go install github.com/pouriyajamshidi/tcping/v2@latest
  ```

- [x tcping](https://x-cmd.com/pkg/tcping):

  **Directly without installation** in [x-cmd](https://www.x-cmd.com).

  ```bash
  x tcping example.com 80
  ```

  Or install `tcping` locally using x-cmd, without needing root privileges or affecting your global setup.

  ```bash
  x env use tcping
  tcping example.com 80
  ```

- Finally, you can compile the code yourself by running the `make` command:

  ```bash
  make build
  ```

  This will place the executables in the `output` folder.

---

## Usage

**tcping** can run in various ways.

1. The simplest form is providing the target and the port number:

```bash
tcping www.example.com 443
```

2. Specify the interval between probes (2 seconds), the timeout (5 seconds) and source interface:

```bash
tcping www.example.com 443 -i 2 -t 5 -I eth2
```

3. Enforce using IPv4 or IPv6 only:

```bash
  tcping www.example.com 443 -4
  # Or
  tcping www.example.com 443 -6
```

4. Show timestamp of probes:

```bash
tcping www.example.com 443 -D
```

5. Retry resolving the hostname after 5 failures:

```bash
tcping www.example.com 443 -r 5

```

6. Stop after 5 probes:

```bash
tcping www.example.com 443 -c 5
```

7. Change the default output from colored to:

```bash
# Save the output in CSV format:
tcping www.example.com 443 --csv example.com.csv
# Save the output in sqlite3 format:
tcping www.example.com 443 --db example.com.db
# Show the output in JSON format:
tcping www.example.com 443 --json
# Show the output in JSON format - pretty:
tcping www.example.com 443 --json --pretty
# Show the output in plain (no ANSI colors):
tcping www.example.com 443 --no-color
```

> [!NOTE]
> Check the **available flags** [here](#flags) for a more advanced usage.

The Docker image can be used with the same set of flags, like:

```bash
# If downloaded from Docker Hub
docker run -it pouriyajamshidi/tcping:latest example.com 443

# If downloaded from GitHub container registry:
docker run -it ghcr.io/pouriyajamshidi/tcping:latest example.com 443
```

> [!TIP]
> Press the `Enter` key while the program is running to examine the summary of all probes without terminating the program, as shown in the [demos](#demos) section.

---

## Flags

The following flags are available to control the behavior of **tcping**:

| Flag                    | Description                                                                                                       |
| ----------------------- | ----------------------------------------------------------------------------------------------------------------- |
| `-h`                    | Show help                                                                                                         |
| `-4`                    | Only use IPv4 addresses                                                                                           |
| `-6`                    | Only use IPv6 addresses                                                                                           |
| `-r`                    | Retry resolving target's hostname after `<n>` number of failed probes. e.g. -r 10 to retry after 10 failed probes |
| `-c`                    | Stop after `<n>` probes, regardless of the result. By default, no limit will be applied                           |
| `-t`                    | Time to wait for a response, in seconds. Real number allowed. 0 means infinite timeout                            |
| `-D`                    | Display date and time in probe output. Similar to Linux's ping utility but human-readable                         |
| `-i`                    | Interval between sending probes                                                                                   |
| `-I`                    | Interface name to use for sending probes                                                                          |
| `--no-color`            | Do not colorize output                                                                                            |
| `--csv`                 | Path and file name to store tcping output in `CSV` format                                                         |
| `-j`                    | Output in `JSON` format                                                                                           |
| `--pretty`              | Prettify the `JSON` output                                                                                        |
| `--db`                  | Path and file name to store tcping output to sqlite database. e.g. `--db /tmp/tcping.db`                          |
| `-v`                    | Print version                                                                                                     |
| `-u`                    | Check for updates                                                                                                 |
| `--show-failures-only`  | Only show probe failures and omit printing probe success messages                                                 |
| `--show-source-address` | Show the source IP address and port used for probes                                                               |

> [!TIP]
> Without specifying the `-4` and `-6` flags, tcping will randomly select an IP address based on DNS lookups.

---

## Demos

### Basic usage

![tcping](Images/gifs/tcping.gif)

---

### Retry hostname lookup (`-r`) flag

![tcping resolve example](Images/gifs/tcping_resolve.gif)

---

### JSON output (`-j --pretty`) flag

![tcping json example](Images/gifs/tcping_json_pretty.gif)

---

## Contributing

Pull requests are welcome to solve bugs, add new features and to help with the open issues that can be found [here](https://github.com/pouriyajamshidi/tcping/issues)

1. Pick any issue that you feel comfortable with.
1. Fork the repository.
1. Create a branch.
1. Commit your work.
1. Add tests if possible.
1. Run the tests `go test` or `make test` and ensure they are successful.
1. Create a pull request

Current number of open issues: ![GitHub issues](https://img.shields.io/github/issues/pouriyajamshidi/tcping.svg).

Please make sure that your pull request **only covers one specific issue/feature** and doesn't handle two or more tickets. This makes it simpler for us to examine your pull request and helps keeping a clean git history.

## Feature Requests and Issues

Should you need a new feature or find a bug, please feel free to [open a pull request](#contributing) or submit an issue.

For larger features/contributions, please make sure to first communicate it on a **discussion** before starting your work.

## Help The Project

If tcping proves to be useful for you, consider sharing it with your network to extend its reach and help other people to also benefit from it.

Furthermore, you can support the project using the links below:

- Buy me a coffee: ["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)

- GitHub Sponsors: [sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)

- Total number of sponsors: ![GitHub Sponsor](https://img.shields.io/github/sponsors/pouriyajamshidi?label=Sponsor&logo=GitHub)

## License

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
