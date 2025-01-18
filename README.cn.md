<div align="center" style="width: 100%;">
 <img alt="tcping" src="Artwork/tcping_logo3.jpeg" style="width:70%;">
</div>

# TCPING

[![Go Report Card](https://goreportcard.com/badge/github.com/pouriyajamshidi/tcping)](https://goreportcard.com/report/github.com/pouriyajamshidi/tcping)
[![CodeFactor](https://www.codefactor.io/repository/github/pouriyajamshidi/tcping/badge)](https://www.codefactor.io/repository/github/pouriyajamshidi/tcping)
[![Go](https://github.com/pouriyajamshidi/tcping/actions/workflows/.github/workflows/codeql-analysis.yml/badge.svg)](https://github.com/pouriyajamshidi/tcping/actions/workflows/go.yml)
[![Docker container build](https://github.com/pouriyajamshidi/tcping/actions/workflows/container-publish.yml/badge.svg)](https://github.com/pouriyajamshidi/tcping/actions/workflows/container-publish.yml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pouriyajamshidi/tcping)
[![Go project version](https://badge.fury.io/go/github.com%2Fpouriyajamshidi%2Ftcping.svg)](https://badge.fury.io/go/github.com%2Fpouriyajamshidi%2Ftcping)
![Download](https://img.shields.io/github/downloads/pouriyajamshidi/tcping/total.svg?label=DOWNLOADS&logo=github)
![Docker Pulls](https://img.shields.io/docker/pulls/pouriyajamshidi/tcping)

这是一个跨平台的 `TCP` 端口 ping 程序，灵感来自 Linux 的 ping 工具。此程序将向您指定的 `IP 地址` 或 `主机名` 发送 `TCP` 探测，并打印结果。它支持 `IPv4` 和 `IPv6`。

**TCPING** 对 _成功_ 和 _不成功_ 的探测使用不同的 `TCP 序列号`，因此当您查看结果并发现探测失败时，可以很容易地推断出到该点为止的总丢包数。

以下是 **TCPING** 的一些功能：

- 在 `ICMP` 被阻止的环境中替代 `ping`。
- 监控您的网络连接。
- 确定丢包率。
- 分析网络延迟。
- 计算网络探测的 `最小`、`平均` 和 `最大` 延迟。
- 按下 `Enter` 键即可打印连接统计信息，而无需停止程序。
- 使用 `-r` 标志在预定次数的探测失败后重试主机名解析。适用于测试您的 `DNS` 负载均衡或全局服务器负载均衡器 `(GSLB)`。
- 强制使用 `IPv4` 或 `IPv6`。
- 显示遇到的最长 `停机时间` 和 `正常运行时间` 持续时间和时间。
- 监控和审计您的对等网络 (SLA)。
- 在进行维护时计算网络的总正常运行时间或停机时间。
- 提供彩色、纯文本、JSON、CSV 和 SQLite3 多种输出格式。

---

## 目录

- [TCPING](#tcping)
  - [目录](#目录)
  - [演示](#演示)
    - [基本用法](#基本用法)
    - [重试主机名查找 (`-r`)](#重试主机名查找--r)
    - [JSON 格式输出 (`-j --pretty`)](#json-格式输出--j---pretty)
  - [下载](#下载)
  - [用法](#用法)
    - [Linux - Debian 和 Ubuntu](#linux---debian-和-ubuntu)
    - [Linux、BSD 和 mac OS](#linuxbsd-和-mac-os)
    - [Windows](#windows)
    - [Docker](#docker)
  - [标志](#标志)
  - [提示](#提示)
  - [检查更新](#检查更新)
  - [贡献](#贡献)
  - [功能请求和问题](#功能请求和问题)
  - [测试平台](#测试平台)
  - [帮助项目](#帮助项目)
  - [许可证](#许可证)

---

## 演示

### 基本用法

![tcping](Images/gifs/tcping.gif)

---

### 重试主机名查找 (`-r`)

![tcping resolve example](Images/gifs/tcping_resolve.gif)

---

### JSON 格式输出 (`-j --pretty`)

![tcping json example](Images/gifs/tcping_json_pretty.gif)

---

## 下载

- ### [Windows](https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping_Windows.zip)

- ### [Linux](https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping_Linux.tar.gz) - 也可通过 `brew` 和 [.deb 软件包](#linux---debian-和-ubuntu) 获得

- ### [macOS](https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping_MacOS.tar.gz) - 也可通过 `brew` 获得

- ### [macOS M1 - ARM](https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping_MacOS_ARM.tar.gz) - 也可通过 `brew` 获得

- ### [FreeBSD](https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping_FreeBSD.tar.gz)

下载完成后，请转到[用法](#用法)部分。

**或者**，您可以：

- 使用 `Docker` 镜像：

  ```bash
  docker pull pouriyajamshidi/tcping:latest
  ```

  > 镜像也可以在 GitHub 容器注册表中找到：

  ```bash
  docker pull ghcr.io/pouriyajamshidi/tcping:latest
  ```

- 使用 `go install` 安装：

  Go 版本最低要求为 `1.23.1`

  ```bash
  go install github.com/pouriyajamshidi/tcping/v2@latest
  ```

- 使用 `brew` 安装：

  ```bash
  brew install pouriyajamshidi/tap/tcping
  ```

- [x tcping](https://x-cmd.com/pkg/tcping)
  
   在 x-cmd 中，无需安装即可**直接使用 tcping 命令**：

   ```bash
   x tcping bing.com 80
   ```

   或者，你也可以选择将 tcping 安装到用户空间，不需 root 特权，亦不影响全局依赖：

   ```bash
   x env use tcping
   tcping bing.com 80
   ```


- 或者通过在 `tcping` 目录中运行 `make` 命令来自行编译代码：

  ```bash
  make build
  ```

  这将在 `executables` 文件夹中为您提供一个压缩文件，其中包含所有受支持操作系统的可执行文件。

---

## 用法

请按照您操作系统的说明进行操作：

- [TCPING](#tcping)
  - [目录](#目录)
  - [演示](#演示)
    - [基本用法](#基本用法)
    - [重试主机名查找 (`-r`)](#重试主机名查找--r)
    - [JSON 格式输出 (`-j --pretty`)](#json-格式输出--j---pretty)
  - [下载](#下载)
  - [用法](#用法)
    - [Linux - Debian 和 Ubuntu](#linux---debian-和-ubuntu)
    - [Linux、BSD 和 mac OS](#linuxbsd-和-mac-os)
    - [Windows](#windows)
    - [Docker](#docker)
  - [标志](#标志)
  - [提示](#提示)
  - [检查更新](#检查更新)
  - [贡献](#贡献)
  - [功能请求和问题](#功能请求和问题)
  - [测试平台](#测试平台)
  - [帮助项目](#帮助项目)
  - [许可证](#许可证)

另请查看[此处提供的可用标志](#标志)。

### Linux - Debian 和 Ubuntu

在 **Debian** 及其衍生版本（如 **Ubuntu**）上，下载 `.deb` 软件包：

```bash
wget https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping_amd64.deb -O /tmp/tcping.deb
```

并安装它：

```bash
sudo apt install -y /tmp/tcping.deb
```

如果您使用的是其他 Linux 发行版，请继续阅读[本节](#linux-bsd-和-mac-os)。

### Linux、BSD 和 mac OS

解压缩文件：

```bash
tar -xvf tcping_Linux.tar.gz
#
# 或在 Mac OS 上
#
tar -xvf tcping_MacOS.tar.gz
#
# 在 Mac OS ARM 上
#
tar -xvf tcping_MacOS_ARM.tar.gz
#
# 在 BSD 上
#
tar -xvf tcping_FreeBSD.tar.gz
```

设置文件为可执行：

```bash
chmod +x tcping
```

将可执行文件复制到您的系统 `PATH` 中，例如 `/usr/local/bin/`：

```bash
sudo cp tcping /usr/local/bin/
```

运行：

```bash
tcping www.example.com 443
# 或
tcping 10.10.10.1 22
```

### Windows

我们建议使用 [Windows 终端](https://apps.microsoft.com/store/detail/windows-terminal/9N0DX20HK701) 以获得最佳体验和正确的颜色显示。

将 `tcping.exe` 复制到您的系统 [PATH](https://www.howtogeek.com/118594/how-to-edit-your-system-path-for-easy-command-line-access/) 中，例如 `C:\Windows\System32`，然后像这样运行它：

```powershell
tcping www.example.com 443
# 或提供 -r 标志以
# 在一定次数的失败后启用名称解析重试：
tcping www.example.com 443 -r 10
```

> TCPING 可能会被 Windows Defender 或某些反恶意软件错误地标记。这在 Go 程序中很常见。请查看 Go 的官方文档 [此处](https://go.dev/doc/faq#virus)。

### Docker

Docker 镜像可以像这样使用：

```bash
# 使用 Docker Hub
docker run -it pouriyajamshidi/tcping:latest example.com 443

# 使用 GitHub 容器注册表：
docker run -it ghcr.io/pouriyajamshidi/tcping:latest example.com 443
```

---

## 标志

以下标志可用于控制应用程序的行为：

| 标志                   | 描述                                                                             |
| ---------------------- | ------------------------------------------------------------------------------- |
| `-h`                   | 显示帮助                                                                          |
| `-4`                   | 仅使用 IPv4 地址                                                                  |
| `-6`                   | 仅使用 IPv6 地址                                                                  |
| `-r`                   | 在 `<n>` 次探测失败后重试解析目标主机名。例如，-r 10 表示在 10 次探测失败后重试            |
| `-c`                   | 在 `<n>` 次探测后停止，无论结果如何。默认情况下，不应用限制                              |
| `-t`                   | 等待响应的时间（以秒为单位）。允许使用实数。0 表示无限超时                                |
| `-D`                   | 在探测输出中显示日期和时间。类似于 Linux 的 ping 工具，但更易于阅读                      |
| `-i`                   | 发送探测之间的间隔                                                                 |
| `-I`                   | 用于发送探测的接口名称                                                              |
| `--no-color`           | 输出不带颜色                                                                      |
| `--csv`                | 以 CSV 格式输出到指定的文件路径                                                     |
| `-j`                   | 以 `JSON` 格式输出                                                                |
| `--pretty`             | 美化 `JSON` 输出                                                                 |
| `--db`                 | 用于存储 tcping 输出到 sqlite 数据库的路径和文件名。例如 `--db /tmp/tcping.db`         |
| `-v`                   | 打印版本                                                                         |
| `-u`                   | 检查更新                                                                         |
| `--show-failures-only` | 仅显示探测失败，并省略打印探测成功消息                                                |
| `--show-source-address` | 显示探测所用的来源IP地址及端口                                                      | 

> 如果未指定 `-4` 和 `-6` 标志，tcping 将根据 DNS 查找随机选择一个 IP 地址。

---

## 提示

- 在程序运行时按 `Enter` 键，可以在不终止程序的情况下查看所有探测的摘要，如[演示](#演示)部分所示。

---

## 检查更新

`TCPING` 正在不断改进，添加了许多新功能并修复了错误。请务必查看更新的版本。

```bash
tcping -u
```

## 贡献

欢迎提交拉取请求以解决错误、添加新功能以及帮助解决可以在[此处](https://github.com/pouriyajamshidi/tcping/issues)找到的未解决问题

1. 选择您觉得可以处理的任何问题。
1. Fork 存储库。
1. 创建一个分支。
1. 提交您的工作。
1. 如果可能，请添加测试。
1. 运行测试 `go test` 或 `make test` 并确保它们成功。
1. 创建一个拉取请求

当前未解决问题的数量：![GitHub issues](https://img.shields.io/github/issues/pouriyajamshidi/tcping.svg)。

请确保您的拉取请求**仅涵盖一个特定的问题/功能**，并且不处理两个或多个问题。这使我们更容易检查您的拉取请求，并有助于保持干净的 git 历史记录。

## 功能请求和问题

如果您需要新功能或发现错误，请随时[打开拉取请求](#贡献)或提交问题。

> 对于较大的功能/贡献，请确保在开始工作之前先在 `issue` 上进行沟通。

## 测试平台

Windows、Linux 和 macOS。

## 帮助项目

如果 tcping 对您有用，请考虑给它一个 ⭐ 以扩大其影响力并帮助其他人也从中受益。

此外，您可以使用以下链接支持该项目。

请我喝杯咖啡：[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/pouriyajamshidi)

GitHub 赞助：[![sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/pouriyajamshidi)

赞助总数：![GitHub Sponsor](https://img.shields.io/github/sponsors/pouriyajamshidi?label=Sponsor&logo=GitHub)

## 许可证

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
