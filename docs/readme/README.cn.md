<div align="center" style="width: 100%;">
 <img alt="tcping" src="../Artwork/tcping_avatar_v3.jpeg" style="width:70%;">
</div>

# TCPING

![Download](https://img.shields.io/github/downloads/pouriyajamshidi/tcping/total.svg?label=DOWNLOADS&logo=github)
![Docker Pulls](https://img.shields.io/docker/pulls/pouriyajamshidi/tcping)
![CodeFactor](https://www.codefactor.io/repository/github/pouriyajamshidi/tcping/badge)
![Go](https://github.com/pouriyajamshidi/tcping/actions/workflows/codeql-analysis.yml/badge.svg)
![Tests](https://github.com/pouriyajamshidi/tcping/actions/workflows/test.yml/badge.svg)
![Release](https://github.com/pouriyajamshidi/tcping/actions/workflows/release.yml/badge.svg)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pouriyajamshidi/tcping)
![Latest release](https://img.shields.io/github/v/release/pouriyajamshidi/tcping)

一个跨平台的 ping 程序，使用 `TCP`、`UDP` 或 `HTTP(S)` 代替 `ICMP`，灵感来自 Linux 的 ping 工具。

> [!TIP]
> 本文档也提供 [English](../../README.md)、[日本語](README.ja.md)、[한국어](README.ko.md)、[فارسی](README.fa.md)、[Deutsch](README.de.md)、[Français](README.fr.md)、[Nederlands](README.nl.md) 版本。

- 在 `ICMP` 被阻止的环境中替代 `ping`，通过 `TCP`、`HTTP(S)` 或 `UDP` 进行探测。
- 和 `ping` 一样报告丢包率，以及延迟的最小值、平均值、最大值和标准差，另外还有最长的正常运行时间和停机时间以及它们发生的时间。
- 按下 `Enter` 键即可随时打印统计信息，而无需停止程序。
- 支持 **彩色**、**纯文本**、**JSON**、**CSV** 或 **sqlite3** 输出格式，也可以把每次探测作为指标通过 **OTLP**（Grafana Alloy、OpenTelemetry Collector 等）或发送到 **InfluxDB**。
- 显示每次 `HTTP(S)` 探测的状态码、TLS 版本和加密套件、证书到期时间，以及连接、TLS 握手和首字节耗时。
- 在一定次数的失败后（`-r`）或每次探测前（`--resolve-every-probe`）重新解析目标主机名，并报告每次查询耗时。适用于测试您的 `DNS` 负载均衡或全局服务器负载均衡器 `(GSLB)`。
- 可以选择 **来源接口**、**超时**、**间隔** 和 **DNS 服务器**，并可强制使用 `IPv4` 或 `IPv6`。
- 分别为 _成功_ 和 _不成功_ 的探测编号，因此总数一目了然。

查看[演示](#演示)来了解 **tcping** 的外观和使用感受。

---

## 目录

- [演示](#演示)
- [下载与安装](#下载与安装)
- [用法](#用法)
- [标志](#标志)
- [贡献](#贡献)
- [帮助项目](#帮助项目)
- [许可证](#许可证)

---

## 演示

<details>
<summary>点击展开</summary>

### 基本用法

![tcping](../Images/gifs/tcping.gif)

---

### 带时间戳的 IPv4 探测（`-D -4`）标志

![tcping timestamp example](../Images/gifs/tcping_timestamp.gif)

---

### 纯文本输出，仅在按下 Enter 时显示统计信息（`--no-color --no-stats`）标志

![tcping plain output example](../Images/gifs/tcping_plain.gif)

---

### 重试主机名查找（`-r`）标志

![tcping resolve example](../Images/gifs/tcping_resolve.gif)

---

### JSON 输出（`-j --pretty`）标志

![tcping json example](../Images/gifs/tcping_json_pretty.gif)

---

### 将 JSON 输出流式发送到收集器（`--json-url --json-server`）标志

![tcping JSON streaming example](../Images/gifs/tcping_json_stream.gif)

---

### 主机名解析耗时（`--resolve-every-probe`）标志

![tcping resolution timing example](../Images/gifs/tcping_dns_timing.gif)

---

### 来源接口（`-I`）标志

![tcping interface example](../Images/gifs/tcping_interface.gif)

---

### HTTP(S) 探测

![tcping HTTP example](../Images/gifs/tcping_http.gif)

---

### HTTP(S) 探测详情（`-v`）标志

![tcping HTTP verbose example](../Images/gifs/tcping_http_verbose.gif)

---

### 跳过证书验证（`--insecure`）标志

![tcping skip TLS example](../Images/gifs/tcping_skip_tls.gif)

---

### 通过代理探测（`--proxy`）标志

![tcping proxy example](../Images/gifs/tcping_proxy.gif)

---

### 对 UDP 服务器进行 UDP 探测（`--udp-server`）标志

![tcping UDP example](../Images/gifs/tcping_udp.gif)

---

### 仅显示失败的探测（`--failures-only`）标志

![tcping failures only example](../Images/gifs/tcping_failures_only.gif)

</details>

---

## 下载与安装

| 平台                                    | 安装方式                                                         |
| --------------------------------------- | ---------------------------------------------------------------- |
| [Windows](#windows)                     | `winget install pj.tcping`                                       |
| [macOS](#macos)                         | `brew install pouriyajamshidi/tap/tcping`                        |
| [Debian、Ubuntu](#linux---软件源)       | 添加软件源后执行 `sudo apt install tcping`                       |
| [Fedora、RHEL、CentOS](#linux---软件源) | 添加软件源后执行 `sudo dnf install tcping`                       |
| [Alpine](#linux---软件源)               | 添加软件源后执行 `sudo apk add tcping`                           |
| [Arch、Manjaro](#linux---软件源)        | `yay -S tcping-bin`                                              |
| [任意 Linux](#linux---一键安装)         | [一键安装](#linux---一键安装)或[软件包文件](#linux---软件包文件) |
| [BSD](#bsd-和-linux---手动安装)         | [预编译的可执行文件](#bsd-和-linux---手动安装)                   |
| [Nix](#nix)                             | `nix profile install github:pouriyajamshidi/tcping`              |
| [Docker](#其他方式)                     | `docker run -it pouriyajamshidi/tcping example.com 443`          |
| [Go](#其他方式)                         | `go install github.com/pouriyajamshidi/tcping/v3@latest`         |

这些可执行文件是静态的，也就是说它们把所需的一切都放在一个文件里，不依赖您机器上是否存在任何库。它们位于
[发布页面](https://github.com/pouriyajamshidi/tcping/releases/latest/)，提供 _amd64_ 和 _arm64_ 版本。

安装完成后，请转到[用法](#用法)部分。

### Windows

在 Windows 上安装 **tcping** 的最佳方式是通过 _Windows 程序包管理器_，使用 [WinGet](https://learn.microsoft.com/zh-cn/windows/package-manager/winget/?ref=github.com%2Fpouriyajamshidi%2Ftcping)。自 2020 年 9 月起，几乎所有 Windows _10_ 和 _11_ 机器都默认带有它：

```powershell
winget install pj.tcping
```

若要手动安装，请解压下载的 zip 文件，并将 `tcping.exe` 复制到您的系统 [PATH](https://www.howtogeek.com/118594/how-to-edit-your-system-path-for-easy-command-line-access/) 中，例如 `C:\Windows\System32`。

> [!CAUTION]
> TCPING 可能会被 Windows Defender 或某些反恶意软件错误地标记。这在 Go 程序中很常见。请查看 Go 团队的官方说明 [此处](https://go.dev/doc/faq#virus)。

> [!WARNING]
> Windows 可执行文件不再提供 `--sqlite`（sqlite3）输出格式。其他所有标志均正常工作。

### macOS

```bash
brew install pouriyajamshidi/tap/tcping
```

### Linux - 一键安装

将以下内容粘贴到终端中，即可获取适合您架构的最新可执行文件并安装：

```bash
cd /tmp &&
ARCH=$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/') &&
curl -LO "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
tar -xf "tcping-linux-$ARCH.tar.gz" &&
sudo install tcping /usr/local/bin/ &&
tcping --version
```

请参阅 [Shell 补全](#shell-补全) 来设置 Tab 补全。

如果您没有 `curl`，请把它那一行换成 `wget`：

```bash
wget "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
```

### Linux - 软件源

添加一次软件源后，`apt`、`dnf` 或 `apk` 就会像处理其他软件包一样负责安装和升级。

**Debian**、**Ubuntu** 及其衍生版本：

```bash
sudo install -d /etc/apt/keyrings &&
  sudo curl -fsSL https://pouriyajamshidi.github.io/packages/keys/pouriyajamshidi.gpg -o /etc/apt/keyrings/pouriyajamshidi.gpg &&
  echo "deb [signed-by=/etc/apt/keyrings/pouriyajamshidi.gpg] https://pouriyajamshidi.github.io/packages/deb ./" | sudo tee /etc/apt/sources.list.d/pouriyajamshidi.list &&
  sudo apt update &&
  sudo apt install tcping
```

**Fedora**、**RHEL**、**CentOS** 及其衍生版本：

```bash
sudo curl -fsSL https://pouriyajamshidi.github.io/packages/rpm/pouriyajamshidi.repo -o /etc/yum.repos.d/pouriyajamshidi.repo &&
  sudo dnf install tcping
```

**Alpine**：

```bash
sudo curl -fsSL https://pouriyajamshidi.github.io/packages/keys/pouriyajamshidi.rsa.pub -o /etc/apk/keys/pouriyajamshidi.rsa.pub &&
  echo "https://pouriyajamshidi.github.io/packages/apk" | sudo tee -a /etc/apk/repositories &&
  sudo apk update &&
  sudo apk add tcping
```

**Arch**、**Manjaro**、**EndeavourOS** 及其衍生版本可以在
[AUR](https://aur.archlinux.org/packages/tcping-bin) 中找到，请使用您喜欢的辅助工具：

```bash
yay -S tcping-bin
```

### Linux - 软件包文件

如果不想添加软件源，可以从
[发布页面](https://github.com/pouriyajamshidi/tcping/releases/latest/)
下载适合您发行版的软件包并安装：

| 发行版               | 软件包                     | 安装                                                |
| -------------------- | -------------------------- | --------------------------------------------------- |
| Debian、Ubuntu       | `tcping-amd64.deb`         | `sudo apt install -y ./tcping-amd64.deb`            |
| Fedora、RHEL、CentOS | `tcping-amd64.rpm`         | `sudo dnf install -y ./tcping-amd64.rpm`            |
| Arch、Manjaro        | `tcping-amd64.pkg.tar.zst` | `sudo pacman -U ./tcping-amd64.pkg.tar.zst`         |
| Alpine               | `tcping-amd64.apk`         | `sudo apk add --allow-untrusted ./tcping-amd64.apk` |

在 ARM 机器上请把 `amd64` 换成 `arm64`。这四种软件包都会为您安装可执行文件和
[Shell 补全](#shell-补全)。没有 `dnf` 的机器可以改用 `yum`，而 Alpine 需要
`--allow-untrusted`，因为我们的软件包没有用 Alpine 密钥签名。

### BSD 和 Linux - 手动安装

下载适合您操作系统和架构的文件：

```bash
wget https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-freebsd-amd64.tar.gz
# 或者在 Linux ARM64 机器上使用 cURL
curl -LO https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-arm64.tar.gz
```

解压并将可执行文件复制到您的系统 `PATH` 中，例如 `/usr/local/bin/`：

```bash
tar -xvf tcping-freebsd-amd64.tar.gz &&
  chmod +x tcping &&
  sudo cp tcping /usr/local/bin/
```

请参阅 [Shell 补全](#shell-补全) 来设置 Tab 补全。

### Nix

**tcping** 提供了一个 flake，因此在任何装有 [Nix](https://nixos.org/download/) 的机器上，
您都可以不安装任何东西直接运行它：

```bash
nix run github:pouriyajamshidi/tcping -- example.com 80
```

或者永久安装：

```bash
nix profile install github:pouriyajamshidi/tcping
```

两种方式都会提供可执行文件和 [Shell 补全](#shell-补全)。该 flake 还有一个
`devShells.default`，包含开发 tcping 所需的一切，可通过 `nix develop` 获得。

### 其他方式

- `Docker` 镜像：

  ```bash
  docker pull pouriyajamshidi/tcping:latest
  # 或者
  docker pull ghcr.io/pouriyajamshidi/tcping:latest
  ```

- 使用 `go install`：

  > 这要求 go 版本至少为 `1.26.8`

  ```bash
  go install github.com/pouriyajamshidi/tcping/v3@latest
  ```

- [x tcping](https://x-cmd.com/pkg/tcping)：

  在 [x-cmd](https://www.x-cmd.com) 中，**无需安装即可直接使用**。

  ```bash
  x tcping example.com 80
  ```

  或者使用 x-cmd 将 `tcping` 安装到本地，不需要 root 权限，也不影响您的全局环境。

  ```bash
  x env use tcping
  tcping example.com 80
  ```

- 最后，您可以通过运行 `make` 命令自行编译代码：

  ```bash
  make build
  ```

  这将在 `target` 文件夹中生成可执行文件。

### Shell 补全

**tcping** 自带 `bash`、`zsh`、`fish` 和 `PowerShell` 的补全脚本，
可以用 `--completions <shell>` 打印出来。它们可以补全标志、`-I` 的接口名称，
以及 `--csv` 和 `--sqlite` 的文件名。

Linux 软件包、Homebrew、AUR 软件包、Nix 和
[一键安装](#linux---一键安装)会为您放好它们。否则，请保存适合您 Shell 的那一个：

- `bash`：

  ```bash
  tcping --completions bash | sudo tee /usr/share/bash-completion/completions/tcping >/dev/null
  ```

- `zsh`：

  ```bash
  tcping --completions zsh | sudo tee /usr/share/zsh/site-functions/_tcping >/dev/null
  ```

- `fish`：

  ```bash
  mkdir -p ~/.config/fish/completions &&
    tcping --completions fish > ~/.config/fish/completions/tcping.fish
  ```

- `PowerShell`，从您的配置文件中加载：

  ```powershell
  New-Item -ItemType Directory -Force (Split-Path $PROFILE) | Out-Null
  Add-Content $PROFILE "tcping --completions powershell | Out-String | Invoke-Expression"
  ```

之后请打开一个新的 Shell 使其生效。

---

## 用法

给 **tcping** 一个目标和一个端口，它就会开始探测：

```bash
tcping www.example.com 443
tcping www.example.com:443          # 也可以使用 主机:端口 格式
tcping 192.168.1.1:80
tcping '[2001:db8::1]:443'          # 使用引号，避免被 shell 解析
tcping https://www.example.com      # 通过 HTTP(S) 探测
tcping udp://127.0.0.1 53           # 通过 UDP 探测
tcping tcp://www.example.com 443    # tcp:// 也可以
```

您可以对一次运行提出的一些要求：

```bash
tcping www.example.com 443 -c 5            # 5 次探测后停止
tcping www.example.com 443 -i 2 -t 5       # 探测间隔 2 秒，超时 5 秒
tcping www.example.com 443 -I eth2         # 从指定接口发送探测
tcping www.example.com 443 -4              # 或 -6，强制使用某个地址族
tcping www.example.com 443 -D              # 为每次探测显示时间戳
tcping www.example.com 443 -r 5            # 5 次失败后重新解析主机名
```

以及它可以产生的输出格式，用来代替默认的彩色输出：

```bash
tcping www.example.com 443 -j              # JSON，加上 --pretty 可美化
tcping www.example.com 443 --no-color      # 纯文本，不带 ANSI 颜色
tcping www.example.com 443 --csv out.csv   # CSV 文件
tcping www.example.com 443 --sqlite out.db # sqlite3 数据库
```

Docker 镜像接受相同的目标和标志：

```bash
docker run -it pouriyajamshidi/tcping:latest example.com 443
# 或者，使用 GitHub 容器注册表
docker run -it ghcr.io/pouriyajamshidi/tcping:latest example.com:443
```

> [!TIP]
> 在程序运行时按 `Enter` 键，可以在不终止程序的情况下查看所有探测的摘要，如[演示](#演示)部分所示。

> [!NOTE]
> 查看[此处](#标志)的**可用标志**以了解更高级的用法。

### 通过 HTTP(S) 探测

给出 URL 而不是主机和端口，tcping 就会通过 HTTP(S) 探测它，
并报告状态码和服务器的响应耗时。它不会下载响应正文，
所以大页面不会拖慢探测：

```bash
tcping https://www.example.com/health
```

端口由协议决定，`80` 或 `443`，除非 URL 自带端口，或者您在其后另外给出端口：

```bash
tcping http://www.example.com:8080/health
# 效果相同
tcping http://www.example.com/health 8080
```

`-v` 会显示一次探测收集到的所有信息：HTTP 版本、TLS 版本和加密套件、
证书还剩多少天到期，以及连接、TLS 握手和首字节耗时。`--insecure` 会跳过证书检查，
这正是对付自签名或已过期证书时所需要的：

```bash
tcping https://www.example.com/health -v
tcping https://self-signed.example.com --insecure
```

### 通过代理探测

`--proxy` 让 TCP 和 HTTP(S) 探测经过代理。支持 `http://`、`https://` 和
`socks5://` 代理，登录信息写在 URL 里，例如
`socks5://user:password@127.0.0.1:1080`。tcping 不读取 `HTTPS_PROXY`、
`HTTP_PROXY` 和 `NO_PROXY` 环境变量，所以只有在你要求时，探测才会经过代理：

```bash
tcping https://www.example.com --proxy http://proxy.example.com:3128
tcping db.example.com 5432 --proxy socks5://127.0.0.1:1080
```

TCP 探测会请求代理连接目标，代理回复已连接时探测即成功。它的耗时包括到代理的
那段路和代理自己到目标的那段路。很多 HTTP 代理只允许连接 `443` 端口，所以即使
目标在线，探测其他端口也可能失败。如果你能通过 SSH 进入某个网络，
`ssh -D 1080 bastion.example.com` 会在 `127.0.0.1:1080` 上开一个 SOCKS5 代理。

代理会自己连接目标的主机名，所以源地址以及 HTTP(S) 探测的连接耗时都是到代理的
那段连接的，运行的第一行也会写明所用的代理。tcping 不会自己解析主机名，所以输出
中不显示它的地址，`-r` 和 `--resolve-every-probe` 也不能和代理一起使用。`-4`、
`-6` 和 `--dns-server` 改为作用于代理自己的主机名。UDP 探测不会经过代理。PAC
文件里设置的代理需要手动指定，因为 tcping 不读取 PAC 文件。

### 通过 UDP 探测

```bash
tcping udp://127.0.0.1 53
```

UDP 没有握手，所以只有对方回应时探测才算成功。
要得到回应，请在另一端把 tcping 作为 UDP 服务器运行，它会把每个数据报回送给发送者：

```bash
# 在被探测的机器上：
tcping --udp-server 127.0.0.1 9999
# 在进行探测的机器上：
tcping udp://127.0.0.1 9999
```

> [!NOTE]
> 被拒绝的 UDP 探测会报告 `port unreachable`（端口不可达），这意味着有东西在阻止我们。
> 完全没有回应的探测同样会报告为失败，但它无法区分是端口开放但保持沉默，
> 还是数据包在途中被丢弃。
>
> 在 Windows 上，被拒绝的探测总是显示为没有回应。Go 在 Windows 的 UDP 套接字上
> 关闭了“端口不可达”错误，所以 tcping 永远看不到它们。

每个 UDP 探测都会把自己的编号作为负载发送，服务器会原样回送。
加上 `-v` 会在每一行显示该编号，这样就能指出丢失的是哪个探测：

```text
Reply from 127.0.0.1 on port 9999 UDP_conn=4 time=1.276 ms
    reply echoed back probe 4
```

### 将 JSON 输出流式发送到服务器

`JSON` 输出不必打印出来。给 `--json-url` 一个地址，
每个事件发生时就会被 `POST` 到那里，每个请求一个事件，
这样进行探测的机器就可以向负责收集的机器报告：

```bash
tcping www.example.com 443 --json-url http://localhost:8000/tcping
```

这些事件及其字段与 `-j` 打印的完全一致，所以任何已经在读取管道输出的程序也能读取它们。
每个流式事件还会在 `source` 字段中写明发送它的机器，默认是该机器的主机名，
因此从多台机器收集事件的收集器可以分辨出它在看谁的运行结果。
使用 `--source-label` 可以自己命名：

```bash
tcping www.example.com 443 --json-url http://collector:8000/tcping --source-label brussels
```

tcping 也可以作为接收端，就像 `--udp-server` 回应 UDP 探测一样。
`--json-server` 不进行探测：它监听给定的主机和端口，并打印发送给它的每个事件（任何路径都可以），
所以负责收集的机器也不需要安装其他任何东西：

```bash
# 在负责收集的机器上：
tcping --json-server 0.0.0.0 8000

# 在进行探测的机器上：
tcping www.example.com 443 --json-url http://collector:8000/tcping
```

事件输出到标准输出，其他内容输出到标准错误，所以
`tcping --json-server 0.0.0.0 8000 > events.jsonl` 得到的文件只包含事件，
并且任何不是 `JSON` 的内容都会被拒绝而不是打印。

任何接受带 `JSON` 正文的 `POST` 的服务器都同样适用，例如：

```python
from fastapi import FastAPI

app = FastAPI()


@app.post("/tcping")
async def tcping(event: dict) -> None:
    print(event)
```

一次运行的所有事件都通过同一个连接发送，该连接在探测之间保持打开，
而不是每次都重新建立。服务器宕机或出问题不会中断探测：
tcping 只会提示一次，然后继续运行，丢弃无法送达的事件。

### 通过 OTLP 或发送到 InfluxDB

tcping 可以不打印每次探测，而是把它作为指标发送，这样一次运行就能变成图表，
也能让多台机器监视同一个目标。通过 OTLP 发送到任何接受它的服务，
例如以 `--web.enable-otlp-receiver` 启动的 [Prometheus](https://prometheus.io/)：

```bash
tcping www.example.com 443 --otlp http://localhost:9090/api/v1/otlp
```

或者发送到 [Grafana Alloy](https://grafana.com/docs/alloy/latest/) 和
OpenTelemetry Collector，它们可以把数据转发到多个地方：

```bash
tcping www.example.com 443 --otlp http://localhost:4318
```

托管的后端需要令牌，每家放在各自的请求头里。用 `--otlp-header` 给出完整的请求头，
或者放在 `OTLP_HEADER` 环境变量中，避免它出现在 shell 历史记录里：

```bash
export OTLP_HEADER="Authorization: Bearer your-api-token" && \
  tcping www.example.com 443 --otlp https://otlp.example.com
```

或者以行协议直接发送到 [InfluxDB](https://www.influxdata.com/) v2 或 v3 服务器：

```bash
export INFLUXDB_TOKEN=your-api-token && \
  tcping www.example.com 443 --influxdb http://localhost:8086 --influxdb-org home --influxdb-bucket tcping
```

每次探测都带有它的往返时间、是否成功、目标解析到的地址和计数器，
如果目标是 HTTP 或 UDP，还会带上相应的详细信息。
此外，通常在退出时看到的完整统计信息块每 10 秒发送一次，
所以没人盯着的运行也会报告它。使用 `--stats-interval` 可以更改这个间隔。

每台机器在指标中用 `source` 标签给自己命名，默认是它的主机名，
所以多台机器探测同一个目标时，数据会落在各自的序列里而不是互相叠加。
使用 `--source-label` 可以自己命名：

```bash
tcping www.example.com 443 --otlp http://localhost:4318 --source-label brussels
```

所有会被发送的内容、如何查询它们，以及现成的 Prometheus、Alloy、
OpenTelemetry Collector 和 InfluxDB 环境（包含 Grafana 和 tcping，一条命令即可启动，
无需先搭建服务器就能试用），都在
[docs/observability](../observability/README.md) 中。

---

## 标志

标志可以放在目标之前或之后，可以使用一个或两个短横线，
所以 `-c 5` 和 `--c 5` 是同一个标志。

### 通用

| 标志        | 默认值 | 描述               |
| ----------- | ------ | ------------------ |
| `-h`        |        | 显示可用标志并退出 |
| `--version` |        | 显示版本并退出     |
| `-u`        |        | 检查更新并退出     |

### 探测

| 标志            | 默认值 | 描述                                                         |
| --------------- | ------ | ------------------------------------------------------------ |
| `-c <n>`        | 无限制 | 在 `<n>` 次探测后停止，无论结果如何                          |
| `-i <seconds>`  | `1`    | 探测之间的间隔。允许使用实数，例如 `-i 0.5`                  |
| `-t <seconds>`  | `1`    | 等待响应的时间（以秒为单位）。允许使用实数。`0` 表示无限超时 |
| `-4`            |        | 仅使用 IPv4 地址                                             |
| `-6`            |        | 仅使用 IPv6 地址                                             |
| `-I <name\|IP>` |        | 用于发送探测和 DNS 查询的接口名称或 IP 地址                  |
| `--proxy <URL>` |        | 通过代理发送 TCP 和 `HTTP(S)` 探测，例如 `--proxy http://proxy.example.com:3128`。不使用 `HTTPS_PROXY`、`HTTP_PROXY` 和 `NO_PROXY` 环境变量 |

> [!TIP]
> 如果未指定 `-4` 和 `-6` 标志，tcping 将根据 DNS 查找随机选择一个 IP 地址。

### 名称解析

| 标志                      | 默认值   | 描述                                                                                      |
| ------------------------- | -------- | ----------------------------------------------------------------------------------------- |
| `-r <n>`                  | 从不     | 在 `<n>` 次探测失败后重新解析目标主机名，例如 `-r 10`                                     |
| `--resolve-every-probe`   |          | 在每一次探测前都解析目标主机名，而不是只在启动时解析。优先于 `-r`，当目标是 IP 地址时无效 |
| `--dns-server <IP>`       | 系统默认 | 使用自定义 DNS 服务器。允许使用 IP 和端口的组合，例如 `--dns-server 1.1.1.1:53`           |
| `--dns-timeout <seconds>` | `2`      | 等待 DNS 响应的时间（以秒为单位）。允许使用实数。`0` 表示无限超时                         |

### 终端输出

| 标志                    | 默认值 | 描述                                                                                                                                                                                               |
| ----------------------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-D`                    |        | 为每次探测显示时间戳                                                                                                                                                                               |
| `--no-color`            |        | 输出不带颜色                                                                                                                                                                                       |
| `--show-source-address` |        | 显示探测所用的来源 IP 地址及端口                                                                                                                                                                   |
| `--failures-only`       |        | 仅显示失败的探测。成功的探测仍会被计数                                                                                                                                                             |
| `--no-stats`            |        | 程序退出时不显示统计信息。按 **Enter** 键仍会显示。当输出不是到终端时无效                                                                                                                          |
| `-v`                    |        | 显示 HTTP(S) 探测收集到的所有信息：HTTP 版本、TLS 版本和加密套件、证书到期时间，以及连接、TLS 和首字节耗时。对于 UDP 目标，显示探测的编号以及回复是否带回了它，以便区分丢失的探测。对 TCP 目标无效 |

### 文件和机器可读输出

| 标志               | 默认值 | 描述                                                                                                                                                                           |
| ------------------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `-j`               |        | 以 `JSON` 格式输出                                                                                                                                                             |
| `--pretty`         |        | 美化 `JSON` 输出。没有 `-j` 时无效                                                                                                                                             |
| `--json-url <URL>` |        | 将 `JSON` 输出发送到 HTTP 服务器而不是打印，每个事件一次 `POST`，例如 `--json-url http://localhost:8000/tcping`。会自动开启 `JSON` 输出，因此不需要 `-j`                       |
| `--json-server`    |        | 不进行探测。监听给定的主机和端口，并打印发送给它的每个 `JSON` 事件，这样其他地方使用 `--json-url` 的 tcping 就有地方发送它的运行结果，例如 `tcping --json-server 0.0.0.0 8000` |
| `--csv <file>`     |        | 将输出存储到 `CSV` 文件中。统计信息会存到同名并加上 `_stats` 后缀的文件中                                                                                                      |
| `--csv-fixed-name` |        | 按原样使用 `--csv` 的文件名，不加日期/时间后缀，因此重复运行会覆盖同一个文件                                                                                                   |
| `--sqlite <file>`  |        | 将输出存储到 sqlite3 数据库中，例如 `--sqlite /tmp/tcping.db`。Windows 上不可用                                                                                                |

### 指标

| 标志                         | 默认值 | 描述                                                                                                                                                                          |
| ---------------------------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--otlp <URL>`               |        | 将结果作为指标发送到 OTLP HTTP 端点（例如 [Grafana Alloy](https://grafana.com/docs/alloy/latest/) 或 OpenTelemetry Collector）而不是打印，例如 `--otlp http://localhost:4318` |
| `--otlp-header <header>`     |        | 发送到 `--otlp` 端点的额外 HTTP 请求头，格式为 `"Name: value"`，例如 `"Authorization: Bearer <token>"`。也可以放在 `OTLP_HEADER` 环境变量中，避免它出现在 shell 历史记录里    |
| `--influxdb <URL>`           |        | 以行协议将结果写入 [InfluxDB](https://www.influxdata.com/) v2 或 v3 服务器而不是打印，例如 `--influxdb http://localhost:8086`                                                 |
| `--influxdb-org <org>`       |        | 要写入的 InfluxDB 组织。使用 `--influxdb` 时必需                                                                                                                              |
| `--influxdb-bucket <bucket>` |        | 要写入的 InfluxDB 存储桶。使用 `--influxdb` 时必需                                                                                                                            |
| `--influxdb-token <token>`   |        | InfluxDB API 令牌。使用 `--influxdb` 时必需。也可以放在 `INFLUXDB_TOKEN` 环境变量中，避免它出现在 shell 历史记录里                                                            |
| `--stats-interval <seconds>` | `10`   | 向 OTLP 端点或 InfluxDB 发送统计信息的频率。没有 `--otlp` 或 `--influxdb` 时无效                                                                                              |
| `--source-label <name>`      | 主机名 | 在 `JSON` 事件以及通过 OTLP 或发送到 InfluxDB 的指标中为这台机器命名，以便区分探测同一目标的多台机器。发送到其他地方的结果在未指定时带有主机名；留在本机的输出则不带任何标识  |

### HTTP(S) 和 UDP

| 标志            | 默认值 | 描述                                                                                                                                       |
| --------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `--insecure`    |        | 探测 `https://` 目标时不验证服务器证书。适用于自签名或已过期的证书                                                                         |
| `--udp-server`  |        | 不进行探测。监听给定的主机和端口，并把每个 UDP 数据报回送给发送者，这样指向这台机器的 UDP 探测就能得到回复                                 |

---

## 贡献

欢迎提交拉取请求来修复错误、添加新功能，以及帮助解决可以在
[此处](https://github.com/pouriyajamshidi/tcping/issues)找到的未解决问题。
当前未解决问题的数量：![GitHub issues](https://img.shields.io/github/issues/pouriyajamshidi/tcping.svg)。

1. 选择您觉得可以处理的任何问题。
1. Fork 存储库。
1. 创建一个分支。
1. 提交您的工作。
1. 添加测试。
1. 运行测试 `make test` 并确保它们成功。
1. 创建一个拉取请求

请确保您的拉取请求**仅涵盖一个特定的问题/功能**，并且不处理两个或多个问题。这使我们更容易检查您的拉取请求，并有助于保持干净的 git 历史记录。

除非您要修复的是一个非常小的问题，否则请在开始工作之前，先在一个 **issue** 上说明您的打算。

若要在糟糕的网络下试用您的更改，`tools/netcond.sh` 可以增加延迟、丢弃一部分数据包，
或者直接屏蔽某个目的地，并在之后撤销它添加的所有内容：

```bash
sudo ./tools/netcond.sh delay 1.1.1.1 800ms 200ms 30
sudo ./tools/netcond.sh loss 1.1.1.1 40
sudo ./tools/netcond.sh clear
```

## 帮助项目

如果 tcping 对您有用，请考虑与您的人脉分享它，以扩大其影响力并帮助其他人也从中受益。

此外，您可以使用以下链接支持该项目：

- 请我喝杯咖啡：[![Buy Me A Coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/pouriyajamshidi)
- GitHub 赞助：[![Sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/pouriyajamshidi)

## 许可证

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
