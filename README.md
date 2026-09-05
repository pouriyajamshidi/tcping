<div align="center" style="width: 100%;">
 <img alt="tcping" src="docs/Artwork/tcping_logo_v3.png" style="width:70%;">
</div>

# TCPING

![Download](https://img.shields.io/github/downloads/pouriyajamshidi/tcping/total.svg?label=DOWNLOADS&logo=github)
![Docker Pulls](https://img.shields.io/docker/pulls/pouriyajamshidi/tcping)
![CodeFactor](https://www.codefactor.io/repository/github/pouriyajamshidi/tcping/badge)
![Go](https://github.com/pouriyajamshidi/tcping/actions/workflows/codeql-analysis.yml/badge.svg)
![Tests](https://github.com/pouriyajamshidi/tcping/actions/workflows/test.yml/badge.svg)
![Docker container build](https://github.com/pouriyajamshidi/tcping/actions/workflows/container-publish.yml/badge.svg)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pouriyajamshidi/tcping)
![Go project version](https://badge.fury.io/go/github.com%2Fpouriyajamshidi%2Ftcping.svg)

**NOTE: You are viewing a broken and a work in progress branch. head to the [MAIN BRANCH](https://github.com/pouriyajamshidi/tcping/tree/master) for the latest release.**
**This branch has become the default for now to ensure our contributors are not basing their work on version 2 mistakenly.**

A cross-platform ping program using `TCP` instead of `ICMP`, inspired by Linux's ping utility.

> [!TIP]
> This document is also available in [中文](README.cn.md).

Here are some of the features of **TCPING**:

- An alternative to `ping` in environments that `ICMP` is blocked.
- Probes over `TCP`, `HTTP(S)` and `UDP`, picked from the target you give it.
- Outputs information in **colored**, **plain**, **JSON**, **CSV** and **sqlite3** formats, or sends it to **Grafana Alloy** and **InfluxDB** as metrics.
- Monitor and audit your or your peers network latency, packet loss, and connection quality.
- Lets you specify the **source interface**, **timeout**, and **interval** between probes.
- Supports both `IPv4` or `IPv6` and lets you enforce using either.
- Prints total connection statistics by pressing the `Enter` key, without stopping the program.
- Reports the longest encountered `downtime` and `uptime` duration and time.
- Reports the minimum, average, maximum and mean deviation of the latency, the same way `ping` does.
- Retries hostname resolution after a predetermined number of probe failures using the `-r` flag, or before every single probe using `--resolve-every-probe`. Suitable to test your `DNS` load balancing or Global Server Load Balancer `(GSLB)`.
- Reports how long hostname resolution itself took, at startup and on every retry.
- Uses different `sequence numbering` for _successful_ and _unsuccessful_ probes to infer the total failed or successful probes at a glance.

Check out the [demos](#demos) to get a look and feel of **tcping**.

---

## Table of Contents

- [TCPING](#tcping)
  - [Table of Contents](#table-of-contents)
  - [Demos](#demos)
    - [Basic usage](#basic-usage)
    - [Retry hostname lookup (`-r`) flag](#retry-hostname-lookup--r-flag)
    - [JSON output (`-j --pretty`) flag](#json-output--j---pretty-flag)
    - [Hostname resolution timing (`--resolve-every-probe`) flag](#hostname-resolution-timing---resolve-every-probe-flag)
    - [Source interface (`-I`) flag](#source-interface--i-flag)
    - [HTTP(S) probes](#https-probes)
    - [HTTP(S) probe details (`-v`) flag](#https-probe-details--v-flag)
    - [Skipping certificate verification (`--insecure`) flag](#skipping-certificate-verification---insecure-flag)
  - [Download and Installation](#download-and-installation)
    - [Windows](#windows)
    - [macOS](#macos)
    - [Linux - Debian and Derivatives](#linux---debian-and-derivatives)
    - [BSD and Linux - Manual Way](#bsd-and-linux---manual-way)
    - [Alternative Ways](#alternative-ways)
  - [Usage](#usage)
    - [Probing over HTTP(S)](#probing-over-https)
    - [Probing over UDP](#probing-over-udp)
    - [Sending the results to Grafana Alloy](#sending-the-results-to-grafana-alloy)
    - [Sending the results to InfluxDB](#sending-the-results-to-influxdb)
  - [Flags](#flags)
    - [General](#general)
    - [Probing](#probing)
    - [Name resolution](#name-resolution)
    - [Terminal output](#terminal-output)
    - [File and machine-readable output](#file-and-machine-readable-output)
    - [Metrics](#metrics)
    - [HTTP(S) and UDP](#https-and-udp)
  - [Contributing](#contributing)
  - [Feature Requests and Issues](#feature-requests-and-issues)
  - [Help The Project](#help-the-project)
  - [License](#license)

---

## Demos

<details>
<summary>Click to expand</summary>

### Basic usage

![tcping](docs/Images/gifs/tcping.gif)

---

### Retry hostname lookup (`-r`) flag

![tcping resolve example](docs/Images/gifs/tcping_resolve.gif)

---

### JSON output (`-j --pretty`) flag

![tcping json example](docs/Images/gifs/tcping_json_pretty.gif)

---

### Hostname resolution timing (`--resolve-every-probe`) flag

![tcping resolution timing example](docs/Images/gifs/tcping_dns_timing.gif)

---

### Source interface (`-I`) flag

![tcping interface example](docs/Images/gifs/tcping_interface.gif)

---

### HTTP(S) probes

![tcping HTTP example](docs/Images/gifs/tcping_http.gif)

---

### HTTP(S) probe details (`-v`) flag

![tcping HTTP verbose example](docs/Images/gifs/tcping_http_verbose.gif)

---

### Skipping certificate verification (`--insecure`) flag

![tcping skip TLS example](docs/Images/gifs/tcping_skip_tls.gif)

</details>

---

## Download and Installation

We offer prebuilt binaries for various operating systems ([Windows](#windows), [Linux](#linux---debian-and-derivatives), [macOS](#macos), [Docker](#alternative-ways)) and architectures (_amd64_, _arm64_), which can be found on the [release page](https://github.com/pouriyajamshidi/tcping/releases/latest/).

There are static and dynamic versions available. In simple terms, static binaries include all needed code inside one file, while dynamic binaries load some code from shared operating system libraries when they run.

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

> [!WARNING]
> The `--db` (sqlite3) output format is not available on Windows binaries. All other flags, including `--csv`, work as expected.

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

  > This requires at least go version `1.26.7`

  ```bash
  go install github.com/pouriyajamshidi/tcping/v3@latest
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

2. You can also use the `host:port` format:

```bash
tcping www.example.com:443
# Or with an IP address
tcping 192.168.1.1:80
# IPv6 addresses (use quotes to prevent shell interpretation)
tcping '[2001:db8::1]:443'
```

3. Specify the interval between probes (2 seconds), the timeout (5 seconds) and source interface:

```bash
tcping www.example.com 443 -i 2 -t 5 -I eth2
```

4. Enforce using IPv4 or IPv6 only:

```bash
  tcping www.example.com 443 -4
  # Or
  tcping www.example.com 443 -6
```

5. Show timestamp of probes:

```bash
tcping www.example.com 443 -D
```

6. Retry resolving the hostname after 5 failures:

```bash
tcping www.example.com 443 -r 5

```

7. Stop after 5 probes:

```bash
tcping www.example.com 443 -c 5
```

8. Change the default output from colored to:

```bash
# Save the output in CSV format:
tcping www.example.com 443 --csv example.com.csv
# Save the output in sqlite3 format:
tcping www.example.com 443 --db example.com.db
# Show the output in JSON format:
tcping www.example.com 443 -j
# Show the output in JSON format - pretty:
tcping www.example.com 443 -j --pretty
# Show the output in plain (no ANSI colors):
tcping www.example.com 443 --no-color
```

> [!NOTE]
> Check the **available flags** [here](#flags) for a more advanced usage.

The Docker image can be used with the same set of flags, like:

```bash
# If downloaded from Docker Hub
docker run -it pouriyajamshidi/tcping:latest example.com 443
# Or using host:port format
docker run -it pouriyajamshidi/tcping:latest example.com:443

# If downloaded from GitHub container registry:
docker run -it ghcr.io/pouriyajamshidi/tcping:latest example.com 443
# Or using host:port format
docker run -it ghcr.io/pouriyajamshidi/tcping:latest example.com:443
```

> [!TIP]
> Press the `Enter` key while the program is running to see the summary of all probes without stopping the program, as shown in the [demos](#demos) section.

### Probing over HTTP(S)

Give a URL instead of a host and a port and tcping probes it over HTTP(S),
reporting the status code and how long the whole request took:

```bash
tcping https://www.example.com/health
```

The port comes from the scheme, `80` or `443`, unless the URL carries its own
or you pass one after it:

```bash
tcping http://www.example.com:8080/health
# Same thing
tcping http://www.example.com/health 8080
```

`-v` shows everything a probe collected: the HTTP version, the TLS version and
cipher, how many days are left on the certificate and the connect, TLS
handshake and first-byte timings. `--insecure` skips the certificate check,
which is what you want against a self-signed or an expired one:

```bash
tcping https://www.example.com/health -v
tcping https://self-signed.example.com --insecure
```

### Probing over UDP

```bash
tcping udp://127.0.0.1 53
```

UDP has no handshake, so a probe only succeeds when the other end answers it.
To get an answer, run tcping as a UDP server on the other end, which echoes
every datagram back to its sender:

```bash
# on the machine being probed:
tcping --udp-server 127.0.0.1 9999
# on the machine probing it:
tcping udp://127.0.0.1 9999
```

> [!NOTE]
> A UDP probe that is refused reports `port unreachable`, which means
> something is blocking us. A probe that gets no answer at all is reported as
> a failure too, but it cannot tell an open port that stays quiet from a
> packet that was dropped on the way.

Every UDP probe sends its own number as the payload, which the server echoes
back. Adding `-v` shows that number on each line, so a lost probe can be named:

```text
Reply from 127.0.0.1 on port 9999 UDP_conn=4 time=1.276 ms
    reply echoed back probe 4
```

### Sending the results to Grafana Alloy

```bash
tcping www.example.com 443 --alloy http://localhost:4318
```

Instead of printing each probe, tcping sends it to Alloy over OTLP, which can
forward it to Prometheus and turn a run into a graph. Every probe sends
`tcping_probe_success`, `tcping_probe_rtt_milliseconds` and
`tcping_probes_total`, labelled with the source, target, port and protocol. An
HTTP(S) target also sends the status code, the connect, TLS handshake and
first-byte timings, and the days left on the certificate. A UDP target sends
whether the reply was echoed back, whether the port refused us and how big the
reply was.

The address the target resolved to is sent on its own as
`tcping_target_address`, which is always 1 and carries the address as a label.
It is kept off the probe metrics because a label is part of what identifies a
series: with `-r` or `--resolve-every-probe` a hostname that resolves somewhere
else mid-run would leave the old series behind and start a new one, which
breaks a graph into pieces and makes the counters add up wrong. Query it on its
own to see which addresses a target has been answering from:

```promql
tcping_target_address{target="www.example.com"}
```

If you want the address alongside the probes, join to it, keeping in mind that
this only works while the target has one address at a time. A round-robin
hostname has several of them live at once and the join has nothing to pick
between them:

```promql
tcping_probe_rtt_milliseconds * on (source, target, port) group_left (ip) tcping_target_address
```

The `source` label says which machine the probe was sent from. It defaults to
that machine's hostname, so several machines probing the same target land in
their own series instead of on top of each other. Use `--source-label` to name
them yourself:

```bash
tcping www.example.com 443 --alloy http://localhost:4318 --source-label paris
```

The whole statistics block you would normally see on exit is sent every 10
seconds, so a run that nobody is watching still reports it: the packet loss,
the minimum, average, maximum and mean deviation of the latency, the total
uptime and downtime, the longest streak of each and when it ran from and to,
when the last successful and unsuccessful probes landed, how many times the
hostname had to be looked up again and how often it answered from a different
address, and when the run started, how long it has been going and when it
ended. Times are sent as milliseconds since the epoch, since a metric can only
carry a number. Use `--stats-interval` to change the interval.

On the Alloy side you need an OTLP receiver pointed at Prometheus. A working
one, along with a Prometheus, an InfluxDB and a Grafana dashboard you can
start in one command, is in
[docs/observability](docs/observability/README.md).

### Sending the results to InfluxDB

```bash
export INFLUXDB_TOKEN=your-api-token
tcping www.example.com 443 --influxdb http://localhost:8086 --influxdb-org home --influxdb-bucket tcping
```

Instead of printing each probe, tcping writes it to InfluxDB v2 or v3 as line
protocol. Every probe writes one point, named after what was probed:
`tcping_tcp`, `tcping_udp` or `tcping_http`, tagged with the source, target,
port and protocol. All three hold `success`, `rtt_ms`, the address the target
resolved to in the `ip` field, and the successful and unsuccessful probe
counts. The address is a field rather than a tag because tags identify a
series: with `-r` or `--resolve-every-probe` a hostname that resolves somewhere
else mid-run would otherwise leave the old series behind and start a new one. A `tcping_http` point also carries the status code,
the connect, TLS handshake and first-byte timings and the days left on the
certificate, and a `tcping_udp` point carries the probe number, the size of
the reply and whether it was echoed back or refused.

The whole statistics block you would normally see on exit is written to
`tcping_statistics` every 10 seconds, so a run that nobody is watching still
reports it: the packet loss, the minimum, average, maximum and mean deviation
of the latency, the total uptime and downtime, the longest streak of each and
when it ran from and to, when the last successful and unsuccessful probes
landed, how many times the hostname had to be looked up again and how often it
answered from a different address, and when the run started, how long it has
been going and when it ended. Times are written as milliseconds since the
epoch, since a string field cannot be graphed. Use
`--stats-interval` to change the interval.

The `source` tag says which machine the probe was sent from. It defaults to
that machine's hostname, so several machines writing to the same bucket land
in their own series instead of on top of each other. Use `--source-label` to
name them yourself:

```bash
tcping www.example.com 443 --influxdb http://localhost:8086 \
  --influxdb-org home --influxdb-bucket tcping --source-label paris
```

The API token can be given with `--influxdb-token`, or in the `INFLUXDB_TOKEN`
environment variable, which keeps it out of your shell history. The flag wins
if both are set.

To try this without setting a server up first, there is a ready made stack in
[docs/observability](docs/observability/README.md).

---

## Flags

Flags can be given before or after the target, and with either one or two
dashes, so `-c 5` and `--c 5` are the same flag.

### General

| Flag | Default | Description |
| --- | --- | --- |
| `-h` | | Show the available flags and exit |
| `--version` | | Show the version and exit |
| `-u` | | Check for updates and exit |

### Probing

| Flag | Default | Description |
| --- | --- | --- |
| `-c <n>` | no limit | Stop after `<n>` probes, regardless of the result |
| `-i <seconds>` | `1` | Interval between probes. Real number allowed, e.g. `-i 0.5` |
| `-t <seconds>` | `1` | Time to wait for a response, in seconds. Real number allowed. `0` means infinite timeout |
| `-4` | | Only use IPv4 addresses |
| `-6` | | Only use IPv6 addresses |
| `-I <name\|IP>` | | Interface name or IP address to send the probes and the DNS lookups from |

> [!TIP]
> Without specifying the `-4` and `-6` flags, tcping will randomly select an IP address based on DNS lookups.

### Name resolution

| Flag | Default | Description |
| --- | --- | --- |
| `-r <n>` | never | Retry resolving the target's hostname after `<n>` failed probes, e.g. `-r 10` |
| `--resolve-every-probe` | | Resolve the target's hostname before every single probe instead of only at startup. Takes precedence over `-r` and has no effect when the target is an IP address |
| `--dns-server <IP>` | system-wide | Custom DNS server to use. An IP and port combination is allowed, e.g. `--dns-server 1.1.1.1:53` |
| `--dns-timeout <seconds>` | `2` | Time to wait for a DNS response, in seconds. Real number allowed. `0` means infinite timeout |

### Terminal output

| Flag | Default | Description |
| --- | --- | --- |
| `-D` | | Show a timestamp for each probe |
| `--no-color` | | Do not colorize the output |
| `--show-source-address` | | Show the source IP address and port used for the probes |
| `--failures-only` | | Only show the failed probes. The successful ones are still counted |
| `--no-stats` | | Do not show the statistics when the program exits. Pressing the **Enter** key still shows them. No effect when the output goes elsewhere than the terminal |
| `-v` | | Show everything an HTTP(S) probe collected: the HTTP version, the TLS version and cipher, the certificate expiry and the connect, TLS and first-byte timings. For a UDP target, shows the probe's number and whether the reply carried it back, so a lost probe can be told apart from the rest. No effect on TCP targets |

### File and machine-readable output

| Flag | Default | Description |
| --- | --- | --- |
| `-j` | | Output in `JSON` format |
| `--pretty` | | Prettify the `JSON` output. No effect without `-j` |
| `--csv <file>` | | Store the output in a `CSV` file. The statistics go to the same name with a `_stats` suffix |
| `--csv-fixed-name` | | Use the `--csv` filename as it is, without a date/time suffix, so repeated runs overwrite the same file |
| `--db <file>` | | Store the output in a sqlite3 database, e.g. `--db /tmp/tcping.db`. Not available on Windows |

### Metrics

| Flag | Default | Description |
| --- | --- | --- |
| `--alloy <URL>` | | Send the results to a [Grafana Alloy](https://grafana.com/docs/alloy/latest/) OTLP HTTP endpoint as metrics instead of printing them, e.g. `--alloy http://localhost:4318` |
| `--influxdb <URL>` | | Write the results to an [InfluxDB](https://www.influxdata.com/) v2 or v3 server as line protocol instead of printing them, e.g. `--influxdb http://localhost:8086` |
| `--influxdb-org <org>` | | InfluxDB organization to write to. Required with `--influxdb` |
| `--influxdb-bucket <bucket>` | | InfluxDB bucket to write to. Required with `--influxdb` |
| `--influxdb-token <token>` | | InfluxDB API token. Required with `--influxdb`. Can also be given in the `INFLUXDB_TOKEN` environment variable, which keeps it out of your shell history |
| `--stats-interval <seconds>` | `10` | How often to send the statistics to Alloy or InfluxDB. No effect without `--alloy` or `--influxdb` |
| `--source-label <name>` | hostname | Name this machine in the metrics sent to Alloy or InfluxDB, so that several machines probing the same target can be told apart |

### HTTP(S) and UDP

| Flag | Default | Description |
| --- | --- | --- |
| `--insecure` | | Do not verify the server certificate when probing an `https://` target. Useful for self-signed or expired certificates |
| `--udp-server` | | Do not probe. Listen on the given host and port and echo every UDP datagram back to its sender, so a UDP probe pointed at this machine gets a reply |

---

## Contributing

Pull requests are welcome to solve bugs, add new features and to help with the open issues that can be found [here](https://github.com/pouriyajamshidi/tcping/issues)

1. Pick any issue that you feel comfortable with.
1. Fork the repository.
1. Create a branch.
1. Commit your work.
1. Add tests.
1. Run the tests `go test ./...` or `make test` and ensure they are successful.
1. Create a pull request

Current number of open issues: ![GitHub issues](https://img.shields.io/github/issues/pouriyajamshidi/tcping.svg).

Please make sure that your pull request **only covers one specific issue/feature** and doesn't handle two or more issues. This makes it simpler for us to review your pull request and helps keeping a clean git history.

To try your changes against a bad network, `tools/netcond.sh` can add latency,
drop a share of the packets or block one destination outright, and undo
everything it added afterwards:

```bash
sudo ./tools/netcond.sh delay 1.1.1.1 800ms 200ms 30
sudo ./tools/netcond.sh loss 1.1.1.1 40
sudo ./tools/netcond.sh clear
```

## Feature Requests and Issues

Do you wish that tcping could do more? Or maybe you have faced a bug?

Please feel free to open an issue and if you can, you are welcome to [open a pull request](#contributing) to contribute.

Although, keep in mind that unless you are fixing a really tiny issue, please ensure to first communicate your intention on an **issue** before starting your work.

## Help The Project

If tcping is useful for you, consider sharing it with your network to extend its reach and help other people to also benefit from it.

Furthermore, you can support the project using the links below:

- Buy me a coffee: ["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/docs/custom_images/orange_img.png)

- GitHub Sponsors: [sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)

- Total number of sponsors: ![GitHub Sponsor](https://img.shields.io/github/sponsors/pouriyajamshidi?label=Sponsor&logo=GitHub)

## License

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
