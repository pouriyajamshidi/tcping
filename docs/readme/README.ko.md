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

`ICMP` 대신 `TCP`, `UDP`, `HTTP(S)`를 사용하는 크로스 플랫폼 ping 프로그램으로, Linux의 ping 유틸리티에서 영감을 받았습니다.

> [!TIP]
> 이 문서는 [English](../../README.md), [中文](README.cn.md), [日本語](README.ja.md), [فارسی](README.fa.md), [Deutsch](README.de.md), [Français](README.fr.md), [Nederlands](README.nl.md)로도 볼 수 있습니다.

- `ICMP`가 차단된 환경에서 `ping`을 대신해 `TCP`, `HTTP(S)`, `UDP`로 연결을 확인합니다.
- `ping`과 같은 방식으로 패킷 손실률과 지연 시간의 최솟값, 평균값, 최댓값, 표준 편차를 보고하며, 가장 긴 가동 시간과 중단 시간, 그리고 그것이 언제 일어났는지도 알려 줍니다.
- `Enter` 키를 누르면 프로그램을 멈추지 않고 언제든지 통계를 출력합니다.
- **컬러**, **일반 텍스트**, **JSON**, **CSV**, **sqlite3** 형식으로 출력하며, 모든 프로브를 메트릭으로 **OTLP**(Grafana Alloy, OpenTelemetry Collector 등)나 **InfluxDB**로 보낼 수도 있습니다.
- 모든 `HTTP(S)` 프로브의 상태 코드, TLS 버전과 암호 스위트, 인증서 만료일, 그리고 연결, TLS 핸드셰이크, 첫 바이트까지의 시간을 보여 줍니다.
- 일정 횟수 실패한 뒤(`-r`)나 모든 프로브 전에(`--resolve-every-probe`) 대상의 호스트 이름을 다시 해석하고, 각 조회에 걸린 시간을 보고합니다. `DNS` 로드 밸런싱이나 글로벌 서버 로드 밸런서 `(GSLB)`를 테스트하는 데 알맞습니다.
- **출발 인터페이스**, **타임아웃**, **간격**, **DNS 서버**를 고를 수 있고, `IPv4`나 `IPv6`를 강제할 수 있습니다.
- _성공한_ 프로브와 _실패한_ 프로브를 따로 세므로 합계를 한눈에 볼 수 있습니다.

**tcping**의 모습과 사용감은 [데모](#데모)에서 확인해 보세요.

---

## 목차

- [데모](#데모)
- [다운로드 및 설치](#다운로드-및-설치)
- [사용법](#사용법)
- [플래그](#플래그)
- [기여하기](#기여하기)
- [프로젝트 돕기](#프로젝트-돕기)
- [라이선스](#라이선스)

---

## 데모

<details>
<summary>클릭하여 펼치기</summary>

### 기본 사용법

![tcping](../Images/gifs/tcping.gif)

---

### 타임스탬프가 있는 IPv4 프로브(`-D -4`) 플래그

![tcping timestamp example](../Images/gifs/tcping_timestamp.gif)

---

### 일반 텍스트 출력, Enter를 누를 때만 통계 표시(`--no-color --no-stats`) 플래그

![tcping plain output example](../Images/gifs/tcping_plain.gif)

---

### 호스트 이름 조회 재시도(`-r`) 플래그

![tcping resolve example](../Images/gifs/tcping_resolve.gif)

---

### JSON 출력(`-j --pretty`) 플래그

![tcping json example](../Images/gifs/tcping_json_pretty.gif)

---

### JSON 출력을 수집기로 스트리밍(`--json-url --json-server`) 플래그

![tcping JSON streaming example](../Images/gifs/tcping_json_stream.gif)

---

### 호스트 이름 해석 시간(`--resolve-every-probe`) 플래그

![tcping resolution timing example](../Images/gifs/tcping_dns_timing.gif)

---

### 출발 인터페이스(`-I`) 플래그

![tcping interface example](../Images/gifs/tcping_interface.gif)

---

### HTTP(S) 프로브

![tcping HTTP example](../Images/gifs/tcping_http.gif)

---

### HTTP(S) 프로브 상세 정보(`-v`) 플래그

![tcping HTTP verbose example](../Images/gifs/tcping_http_verbose.gif)

---

### 인증서 검증 건너뛰기(`--insecure`) 플래그

![tcping skip TLS example](../Images/gifs/tcping_skip_tls.gif)

---

### 프록시를 통한 프로브(`--proxy`) 플래그

![tcping proxy example](../Images/gifs/tcping_proxy.gif)

---

### UDP 서버에 대한 UDP 프로브(`--udp-server`) 플래그

![tcping UDP example](../Images/gifs/tcping_udp.gif)

---

### 실패한 프로브만 표시(`--failures-only`) 플래그

![tcping failures only example](../Images/gifs/tcping_failures_only.gif)

</details>

---

## 다운로드 및 설치

| 플랫폼                                         | 설치 방법                                                                  |
| ---------------------------------------------- | -------------------------------------------------------------------------- |
| [Windows](#windows)                            | `winget install pj.tcping`                                                 |
| [macOS](#macos)                                | `brew install pouriyajamshidi/tap/tcping`                                  |
| [Debian, Ubuntu](#linux---패키지-저장소)       | 저장소를 추가한 뒤 `sudo apt install tcping`                               |
| [Fedora, RHEL, CentOS](#linux---패키지-저장소) | 저장소를 추가한 뒤 `sudo dnf install tcping`                               |
| [Alpine](#linux---패키지-저장소)               | 저장소를 추가한 뒤 `sudo apk add tcping`                                   |
| [Arch, Manjaro](#linux---패키지-저장소)        | `yay -S tcping-bin`                                                        |
| [모든 Linux](#linux---한-줄-설치)              | [한 줄 설치](#linux---한-줄-설치) 또는 [패키지 파일](#linux---패키지-파일) |
| [BSD](#bsd와-linux---수동-설치)                | [빌드된 바이너리](#bsd와-linux---수동-설치)                                |
| [Nix](#nix)                                    | `nix profile install github:pouriyajamshidi/tcping`                        |
| [Docker](#기타-방법)                           | `docker run -it pouriyajamshidi/tcping example.com 443`                    |
| [Go](#기타-방법)                               | `go install github.com/pouriyajamshidi/tcping/v3@latest`                   |

바이너리는 정적 빌드입니다. 필요한 모든 것이 파일 하나에 들어 있어서, 사용하는 컴퓨터에 어떤 라이브러리가 있는지와 상관없이 동작합니다.
[릴리스 페이지](https://github.com/pouriyajamshidi/tcping/releases/latest/)에서 *amd64*와 *arm64*용을 받을 수 있습니다.

설치를 마쳤다면 [사용법](#사용법) 섹션으로 넘어가세요.

### Windows

Windows에 **tcping**을 설치하는 가장 좋은 방법은 [WinGet](https://learn.microsoft.com/ko-kr/windows/package-manager/winget/?ref=github.com%2Fpouriyajamshidi%2Ftcping)을 통한 *Windows 패키지 관리자*입니다. 2020년 9월부터 거의 모든 Windows *10*과 _11_ 컴퓨터에 기본으로 들어 있습니다.

```powershell
winget install pj.tcping
```

직접 설치하려면, 내려받은 zip 파일의 압축을 풀고 `tcping.exe`를 `C:\Windows\System32`와 같은 시스템 [PATH](https://www.howtogeek.com/118594/how-to-edit-your-system-path-for-easy-command-line-access/)에 복사하세요.

> [!CAUTION]
> TCPING이 Windows Defender나 일부 악성코드 방지 소프트웨어에 의해 잘못 탐지될 수 있습니다. Go 프로그램에서는 흔한 일입니다. Go 팀의 공식 설명은 [여기](https://go.dev/doc/faq#virus)에서 확인하세요.

> [!WARNING]
> Windows 바이너리에는 더 이상 `--sqlite`(sqlite3) 출력 형식이 없습니다. 다른 모든 플래그는 정상적으로 동작합니다.

### macOS

```bash
brew install pouriyajamshidi/tap/tcping
```

### Linux - 한 줄 설치

다음을 터미널에 붙여 넣으면 사용 중인 아키텍처에 맞는 최신 바이너리를 받아서 설치합니다.

```bash
cd /tmp &&
ARCH=$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/') &&
curl -LO "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
tar -xf "tcping-linux-$ARCH.tar.gz" &&
sudo install tcping /usr/local/bin/ &&
tcping --version
```

탭 자동 완성을 설정하려면 [Shell 자동 완성](#shell-자동-완성)을 참고하세요.

`curl`이 없다면 그 줄을 `wget`으로 바꾸세요.

```bash
wget "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
```

### Linux - 패키지 저장소

저장소를 한 번 추가하면 그 뒤로는 다른 패키지와 똑같이 `apt`, `dnf`, `apk`가 설치와 업그레이드를 맡아 줍니다.

**Debian**, **Ubuntu** 및 그 파생 배포판:

```bash
sudo install -d /etc/apt/keyrings &&
  sudo curl -fsSL https://pouriyajamshidi.github.io/packages/keys/pouriyajamshidi.gpg -o /etc/apt/keyrings/pouriyajamshidi.gpg &&
  echo "deb [signed-by=/etc/apt/keyrings/pouriyajamshidi.gpg] https://pouriyajamshidi.github.io/packages/deb ./" | sudo tee /etc/apt/sources.list.d/pouriyajamshidi.list &&
  sudo apt update &&
  sudo apt install tcping
```

**Fedora**, **RHEL**, **CentOS** 및 그 파생 배포판:

```bash
sudo curl -fsSL https://pouriyajamshidi.github.io/packages/rpm/pouriyajamshidi.repo -o /etc/yum.repos.d/pouriyajamshidi.repo &&
  sudo dnf install tcping
```

**Alpine**:

```bash
sudo curl -fsSL https://pouriyajamshidi.github.io/packages/keys/pouriyajamshidi.rsa.pub -o /etc/apk/keys/pouriyajamshidi.rsa.pub &&
  echo "https://pouriyajamshidi.github.io/packages/apk" | sudo tee -a /etc/apk/repositories &&
  sudo apk update &&
  sudo apk add tcping
```

**Arch**, **Manjaro**, **EndeavourOS** 및 그 파생 배포판은
[AUR](https://aur.archlinux.org/packages/tcping-bin)에 있으니 선호하는 헬퍼를 사용하세요.

```bash
yay -S tcping-bin
```

### Linux - 패키지 파일

저장소를 추가하지 않으려면
[릴리스 페이지](https://github.com/pouriyajamshidi/tcping/releases/latest/)에서
사용 중인 배포판의 패키지를 내려받아 설치하세요.

| 배포판               | 패키지                     | 설치                                                |
| -------------------- | -------------------------- | --------------------------------------------------- |
| Debian, Ubuntu       | `tcping-amd64.deb`         | `sudo apt install -y ./tcping-amd64.deb`            |
| Fedora, RHEL, CentOS | `tcping-amd64.rpm`         | `sudo dnf install -y ./tcping-amd64.rpm`            |
| Arch, Manjaro        | `tcping-amd64.pkg.tar.zst` | `sudo pacman -U ./tcping-amd64.pkg.tar.zst`         |
| Alpine               | `tcping-amd64.apk`         | `sudo apk add --allow-untrusted ./tcping-amd64.apk` |

ARM 컴퓨터에서는 `amd64`를 `arm64`로 바꾸세요. 네 가지 모두 바이너리와
[Shell 자동 완성](#shell-자동-완성)을 함께 설치합니다. `dnf`가 없는 컴퓨터에서는
`yum`을 쓸 수 있고, Alpine에서 `--allow-untrusted`가 필요한 이유는 우리 패키지가
Alpine 키로 서명되어 있지 않기 때문입니다.

### BSD와 Linux - 수동 설치

사용 중인 OS와 아키텍처에 맞는 파일을 내려받으세요.

```bash
wget https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-freebsd-amd64.tar.gz
# 또는 Linux ARM64 컴퓨터에서 cURL을 사용하는 경우
curl -LO https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-arm64.tar.gz
```

압축을 풀고 실행 파일을 `/usr/local/bin/` 같은 시스템 `PATH`에 복사하세요.

```bash
tar -xvf tcping-freebsd-amd64.tar.gz &&
  chmod +x tcping &&
  sudo cp tcping /usr/local/bin/
```

탭 자동 완성을 설정하려면 [Shell 자동 완성](#shell-자동-완성)을 참고하세요.

### Nix

**tcping**은 flake를 제공하므로, [Nix](https://nixos.org/download/)가 있는 컴퓨터라면
아무것도 설치하지 않고 바로 실행할 수 있습니다.

```bash
nix run github:pouriyajamshidi/tcping -- example.com 80
```

또는 계속 쓰려면 설치합니다.

```bash
nix profile install github:pouriyajamshidi/tcping
```

두 방법 모두 바이너리와 [Shell 자동 완성](#shell-자동-완성)을 제공합니다. 이 flake에는
tcping 개발에 필요한 모든 것이 들어 있는 `devShells.default`도 있으며,
`nix develop`으로 사용할 수 있습니다.

### 기타 방법

- `Docker` 이미지:

  ```bash
  docker pull pouriyajamshidi/tcping:latest
  # 또는
  docker pull ghcr.io/pouriyajamshidi/tcping:latest
  ```

- `go install` 사용:

  > go 버전 `1.26.8` 이상이 필요합니다

  ```bash
  go install github.com/pouriyajamshidi/tcping/v3@latest
  ```

- [x tcping](https://x-cmd.com/pkg/tcping):

  [x-cmd](https://www.x-cmd.com)에서는 **설치 없이 바로** 쓸 수 있습니다.

  ```bash
  x tcping example.com 80
  ```

  또는 x-cmd로 root 권한 없이, 전역 환경에 영향을 주지 않고 `tcping`을 로컬에 설치할 수도 있습니다.

  ```bash
  x env use tcping
  tcping example.com 80
  ```

- 마지막으로, `make` 명령을 실행해 직접 코드를 컴파일할 수 있습니다.

  ```bash
  make build
  ```

  실행 파일은 `target` 폴더에 만들어집니다.

### Shell 자동 완성

**tcping**에는 `bash`, `zsh`, `fish`, `PowerShell`용 자동 완성 스크립트가 들어 있으며,
`--completions <shell>`로 출력할 수 있습니다. 플래그, `-I`의 인터페이스 이름,
`--csv`와 `--sqlite`의 파일 이름을 완성해 줍니다.

Linux 패키지, Homebrew, AUR 패키지, Nix,
[한 줄 설치](#linux---한-줄-설치)는 이 스크립트를 알아서 제자리에 놓아 줍니다.
그렇지 않다면 사용 중인 shell에 맞는 것을 저장하세요.

- `bash`:

  ```bash
  tcping --completions bash | sudo tee /usr/share/bash-completion/completions/tcping >/dev/null
  ```

- `zsh`:

  ```bash
  tcping --completions zsh | sudo tee /usr/share/zsh/site-functions/_tcping >/dev/null
  ```

- `fish`:

  ```bash
  mkdir -p ~/.config/fish/completions &&
    tcping --completions fish > ~/.config/fish/completions/tcping.fish
  ```

- `PowerShell`, 프로필에서 불러오기:

  ```powershell
  Add-Content $PROFILE "tcping --completions powershell | Out-String | Invoke-Expression"
  ```

적용하려면 그 뒤에 새 shell을 시작하세요.

---

## 사용법

**tcping**에 대상과 포트를 주면 프로브를 시작합니다.

```bash
tcping www.example.com 443
tcping www.example.com:443          # 호스트:포트 형식도 됩니다
tcping 192.168.1.1:80
tcping '[2001:db8::1]:443'          # shell이 건드리지 않도록 따옴표로 감쌉니다
tcping https://www.example.com      # HTTP(S)로 프로브합니다
tcping udp://127.0.0.1 53           # UDP로 프로브합니다
```

한 번의 실행에 요청할 수 있는 몇 가지 예입니다.

```bash
tcping www.example.com 443 -c 5            # 프로브 5번 후 중지
tcping www.example.com 443 -i 2 -t 5       # 프로브 간격 2초, 타임아웃 5초
tcping www.example.com 443 -I eth2         # 지정한 인터페이스에서 프로브 전송
tcping www.example.com 443 -4              # 또는 -6, 주소 체계를 강제
tcping www.example.com 443 -D              # 각 프로브에 타임스탬프 표시
tcping www.example.com 443 -r 5            # 5번 실패하면 호스트 이름을 다시 해석
```

기본 컬러 출력 대신 쓸 수 있는 출력 형식입니다.

```bash
tcping www.example.com 443 -j              # JSON, --pretty를 붙이면 보기 좋게 정리
tcping www.example.com 443 --no-color      # ANSI 색상이 없는 일반 텍스트
tcping www.example.com 443 --csv out.csv   # CSV 파일
tcping www.example.com 443 --sqlite out.db # sqlite3 데이터베이스
```

Docker 이미지도 같은 대상과 플래그를 받습니다.

```bash
docker run -it pouriyajamshidi/tcping:latest example.com 443
# 또는 GitHub 컨테이너 레지스트리에서
docker run -it ghcr.io/pouriyajamshidi/tcping:latest example.com:443
```

> [!TIP]
> 프로그램이 실행 중일 때 `Enter` 키를 누르면, [데모](#데모) 섹션에서 보듯이 프로그램을 멈추지 않고 모든 프로브의 요약을 볼 수 있습니다.

> [!NOTE]
> 더 고급 사용법은 [여기](#플래그)에서 **사용 가능한 플래그**를 확인하세요.

### HTTP(S)로 프로브하기

호스트와 포트 대신 URL을 주면 tcping이 HTTP(S)로 프로브하고,
상태 코드와 서버가 응답하는 데 걸린 시간을 보고합니다. 본문은 내려받지 않으므로
큰 페이지도 프로브를 느리게 하지 않습니다.

```bash
tcping https://www.example.com/health
```

포트는 스킴에서 정해지며 `80` 또는 `443`입니다. URL에 포트가 들어 있거나
그 뒤에 따로 포트를 주면 그 값을 씁니다.

```bash
tcping http://www.example.com:8080/health
# 같은 결과
tcping http://www.example.com/health 8080
```

`-v`는 프로브가 모은 모든 것을 보여 줍니다. HTTP 버전, TLS 버전과 암호 스위트,
인증서 만료까지 남은 일수, 그리고 연결, TLS 핸드셰이크, 첫 바이트까지의 시간입니다.
`--insecure`는 인증서 검사를 건너뛰며, 자체 서명 인증서나 만료된 인증서를 상대할 때 필요합니다.

```bash
tcping https://www.example.com/health -v
tcping https://self-signed.example.com --insecure
```

### 프록시를 통해 프로브하기

`--proxy`는 TCP와 HTTP(S) 프로브를 프록시를 통해 보냅니다. `http://`, `https://`,
`socks5://` 프록시를 쓸 수 있고, 로그인 정보는 URL에 넣습니다. 예:
`socks5://user:password@127.0.0.1:1080`. tcping은 `HTTPS_PROXY`, `HTTP_PROXY`,
`NO_PROXY` 환경 변수를 읽지 않으므로, 프로브는 직접 지정했을 때만 프록시를 거칩니다.

```bash
tcping https://www.example.com --proxy http://proxy.example.com:3128
tcping db.example.com 5432 --proxy socks5://127.0.0.1:1080
```

TCP 프로브는 프록시에게 대상에 연결해 달라고 요청하고, 프록시가 연결했다고 답하면
성공합니다. 그래서 시간에는 프록시까지 가는 길과 프록시에서 대상까지 가는 길이 모두
포함됩니다. 많은 HTTP 프록시는 `443` 포트로의 연결만 허용하므로, 대상이 살아 있어도
다른 포트로의 프로브는 실패할 수 있습니다. SSH로 들어갈 수 있는 네트워크로 가는
SOCKS5 프록시는 `ssh -D 1080 bastion.example.com`으로 `127.0.0.1:1080`에 열 수
있습니다.

프록시가 대상의 호스트 이름으로 직접 연결하므로, 출발지 주소와 HTTP(S) 프로브의
연결 시간은 프록시까지의 연결에 대한 것이 되고, 실행의 첫 줄에 그 프록시가
표시됩니다. tcping은 호스트 이름을 직접 확인하지 않으므로 출력에 그 주소가 표시되지
않으며, `-r`과 `--resolve-every-probe`는 프록시와 함께 쓸 수 없습니다. `-4`, `-6`,
`--dns-server`는 대신 프록시 자신의 호스트 이름에 적용됩니다. UDP 프로브는 프록시를
거치지 않습니다. PAC 파일로 설정된 프록시는 직접 지정해야 합니다. tcping은 PAC
파일을 읽지 않기 때문입니다.

### UDP로 프로브하기

```bash
tcping udp://127.0.0.1 53
```

UDP에는 핸드셰이크가 없으므로 상대가 응답해야만 프로브가 성공합니다.
응답을 받으려면 반대편에서 tcping을 UDP 서버로 실행하세요.
받은 데이터그램을 모두 보낸 쪽으로 그대로 돌려보냅니다.

```bash
# 프로브를 받는 컴퓨터에서:
tcping --udp-server 127.0.0.1 9999
# 프로브를 보내는 컴퓨터에서:
tcping udp://127.0.0.1 9999
```

> [!NOTE]
> 거부된 UDP 프로브는 `port unreachable`(포트에 도달할 수 없음)이라고 보고하며,
> 무언가가 우리를 막고 있다는 뜻입니다. 응답이 전혀 없는 프로브도 실패로 보고되지만,
> 열려 있는데 조용한 포트인지, 중간에 버려진 패킷인지는 구별할 수 없습니다.
>
> Windows에서는 거부된 프로브가 항상 응답 없음으로 표시됩니다. Go가 Windows UDP 소켓의
> "port unreachable" 오류를 꺼 두기 때문에 tcping은 그것을 볼 수 없습니다.

모든 UDP 프로브는 자신의 번호를 페이로드로 보내고, 서버는 그것을 그대로 돌려보냅니다.
`-v`를 붙이면 각 줄에 그 번호가 표시되므로 어느 프로브가 유실되었는지 알 수 있습니다.

```text
Reply from 127.0.0.1 on port 9999 UDP_conn=4 time=1.276 ms
    reply echoed back probe 4
```

### JSON 출력을 서버로 스트리밍하기

`JSON` 출력을 꼭 화면에 찍을 필요는 없습니다. `--json-url`에 주소를 주면
이벤트가 발생하는 대로, 요청 하나에 이벤트 하나씩 그곳으로 `POST`됩니다.
프로브하는 컴퓨터가 수집하는 컴퓨터에 보고할 수 있습니다.

```bash
tcping www.example.com 443 --json-url http://localhost:8000/tcping
```

이벤트와 그 필드는 `-j`가 출력하는 것과 똑같으므로, 파이프로 받은 출력을 이미 읽는
프로그램이라면 이것도 읽을 수 있습니다. 스트리밍되는 모든 이벤트에는 보낸 컴퓨터가
`source` 필드에 적히며, 기본값은 그 컴퓨터의 호스트 이름입니다. 그래서 여러 컴퓨터에서
이벤트를 받는 수집기가 누구의 실행 결과인지 구별할 수 있습니다.
`--source-label`로 직접 이름을 붙일 수 있습니다.

```bash
tcping www.example.com 443 --json-url http://collector:8000/tcping --source-label brussels
```

tcping은 받는 쪽도 될 수 있습니다. `--udp-server`가 UDP 프로브에 응답하는 것과 같습니다.
`--json-server`는 프로브하지 않습니다. 주어진 호스트와 포트에서 기다리다가
어떤 경로로든 전달된 모든 이벤트를 출력하므로, 수집하는 컴퓨터에도 따로 설치할 것이 없습니다.

```bash
# 수집하는 컴퓨터에서:
tcping --json-server 0.0.0.0 8000

# 프로브하는 컴퓨터에서:
tcping www.example.com 443 --json-url http://collector:8000/tcping
```

이벤트는 표준 출력으로, 나머지는 표준 오류로 나가므로
`tcping --json-server 0.0.0.0 8000 > events.jsonl`은 이벤트만 담긴 파일을 남기고,
`JSON`이 아닌 것은 출력되지 않고 거부됩니다.

`JSON` 본문이 있는 `POST`를 받는 서버라면 무엇이든 똑같이 동작합니다. 예를 들면 다음과 같습니다.

```python
from fastapi import FastAPI

app = FastAPI()


@app.post("/tcping")
async def tcping(event: dict) -> None:
    print(event)
```

한 번의 실행에서 나오는 이벤트는 모두 하나의 연결로 전송되며, 이 연결은 프로브마다
새로 맺지 않고 계속 열려 있습니다. 서버가 다운되었거나 상태가 나빠도 프로브는 멈추지 않습니다.
tcping은 한 번만 알리고 계속 진행하며, 전달할 수 없는 이벤트는 버립니다.

### OTLP 또는 InfluxDB로 결과 보내기

각 프로브를 출력하는 대신 tcping은 이를 메트릭으로 보낼 수 있습니다. 그러면 실행 결과를
그래프로 볼 수 있고, 여러 컴퓨터가 같은 대상을 감시할 수도 있습니다. OTLP를 받는 것이면
무엇이든 보낼 수 있는데, 예를 들어 `--web.enable-otlp-receiver`로 시작한
[Prometheus](https://prometheus.io/)가 있습니다.

```bash
tcping www.example.com 443 --otlp http://localhost:9090/api/v1/otlp
```

또는 여러 곳으로 전달할 수 있는 [Grafana Alloy](https://grafana.com/docs/alloy/latest/)와
OpenTelemetry Collector로 보냅니다.

```bash
tcping www.example.com 443 --otlp http://localhost:4318
```

호스팅 백엔드는 토큰을 요구하며, 서비스마다 헤더가 다릅니다. `--otlp-header`에 헤더 전체를 주거나,
shell 기록에 남지 않도록 `OTLP_HEADER` 환경 변수에 넣으세요.

```bash
export OTLP_HEADER="Authorization: Bearer your-api-token" && \
  tcping www.example.com 443 --otlp https://otlp.example.com
```

또는 라인 프로토콜로 [InfluxDB](https://www.influxdata.com/) v2나 v3 서버에 바로 보냅니다.

```bash
export INFLUXDB_TOKEN=your-api-token && \
  tcping www.example.com 443 --influxdb http://localhost:8086 --influxdb-org home --influxdb-bucket tcping
```

모든 프로브에는 왕복 시간, 성공 여부, 대상이 해석된 주소, 카운터가 담기고,
대상이 HTTP나 UDP라면 그 상세 정보도 함께 담깁니다. 그에 더해, 평소 종료할 때 보게 되는
전체 통계 블록이 10초마다 전송되므로 아무도 보고 있지 않은 실행도 통계를 보고합니다.
이 간격은 `--stats-interval`로 바꿀 수 있습니다.

각 컴퓨터는 메트릭의 `source` 레이블로 자신의 이름을 알리며, 기본값은 호스트 이름입니다.
그래서 여러 컴퓨터가 같은 대상을 프로브해도 서로 겹치지 않고 각자의 시리즈에 쌓입니다.
`--source-label`로 직접 이름을 붙일 수 있습니다.

```bash
tcping www.example.com 443 --otlp http://localhost:4318 --source-label brussels
```

전송되는 모든 내용, 조회하는 방법, 그리고 서버를 먼저 준비하지 않고도 한 번의 명령으로 시작해
써 볼 수 있는, Grafana와 tcping이 포함된 미리 만들어진 Prometheus, Alloy,
OpenTelemetry Collector, InfluxDB 환경은
[docs/observability](../observability/README.md)에 있습니다.

---

## 플래그

플래그는 대상의 앞이나 뒤에 줄 수 있고, 대시는 하나든 둘이든 상관없습니다.
그래서 `-c 5`와 `--c 5`는 같은 플래그입니다.

### 일반

| 플래그      | 기본값 | 설명                                |
| ----------- | ------ | ----------------------------------- |
| `-h`        |        | 사용 가능한 플래그를 보여 주고 종료 |
| `--version` |        | 버전을 보여 주고 종료               |
| `-u`        |        | 업데이트를 확인하고 종료            |

### 프로브

| 플래그          | 기본값    | 설명                                                     |
| --------------- | --------- | -------------------------------------------------------- |
| `-c <n>`        | 제한 없음 | 결과와 상관없이 `<n>`번 프로브한 뒤 중지                 |
| `-i <seconds>`  | `1`       | 프로브 사이의 간격. 실수 허용, 예: `-i 0.5`              |
| `-t <seconds>`  | `1`       | 응답을 기다리는 시간(초). 실수 허용. `0`은 타임아웃 없음 |
| `-4`            |           | IPv4 주소만 사용                                         |
| `-6`            |           | IPv6 주소만 사용                                         |
| `-I <name\|IP>` |           | 프로브와 DNS 조회를 보낼 인터페이스 이름 또는 IP 주소    |
| `--proxy <URL>` |        | TCP와 `HTTP(S)` 프로브를 프록시를 통해 보냄, 예: `--proxy http://proxy.example.com:3128`. `HTTPS_PROXY`, `HTTP_PROXY`, `NO_PROXY` 환경 변수는 사용하지 않음 |

> [!TIP]
> `-4`와 `-6` 플래그를 지정하지 않으면 tcping은 DNS 조회 결과에서 IP 주소를 무작위로 고릅니다.

### 이름 해석

| 플래그                    | 기본값        | 설명                                                                                                             |
| ------------------------- | ------------- | ---------------------------------------------------------------------------------------------------------------- |
| `-r <n>`                  | 안 함         | `<n>`번 프로브에 실패하면 대상의 호스트 이름을 다시 해석, 예: `-r 10`                                            |
| `--resolve-every-probe`   |               | 시작할 때만이 아니라 모든 프로브 전에 대상의 호스트 이름을 해석. `-r`보다 우선하며, 대상이 IP 주소이면 효과 없음 |
| `--dns-server <IP>`       | 시스템 설정값 | 사용할 사용자 지정 DNS 서버. IP와 포트 조합 허용, 예: `--dns-server 1.1.1.1:53`                                  |
| `--dns-timeout <seconds>` | `2`           | DNS 응답을 기다리는 시간(초). 실수 허용. `0`은 타임아웃 없음                                                     |

### 터미널 출력

| 플래그                  | 기본값 | 설명                                                                                                                                                                                                                                                   |
| ----------------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `-D`                    |        | 각 프로브에 타임스탬프 표시                                                                                                                                                                                                                            |
| `--no-color`            |        | 출력에 색상을 입히지 않음                                                                                                                                                                                                                              |
| `--show-source-address` |        | 프로브에 사용한 출발지 IP 주소와 포트 표시                                                                                                                                                                                                             |
| `--failures-only`       |        | 실패한 프로브만 표시. 성공한 프로브도 계속 집계됨                                                                                                                                                                                                      |
| `--no-stats`            |        | 프로그램이 종료될 때 통계를 표시하지 않음. **Enter** 키를 누르면 여전히 표시됨. 출력이 터미널이 아닌 곳으로 갈 때는 효과 없음                                                                                                                          |
| `-v`                    |        | HTTP(S) 프로브가 모은 모든 것을 표시: HTTP 버전, TLS 버전과 암호 스위트, 인증서 만료일, 연결, TLS, 첫 바이트까지의 시간. UDP 대상에서는 프로브의 번호와 응답이 그것을 돌려보냈는지를 표시하여 유실된 프로브를 구별할 수 있음. TCP 대상에서는 효과 없음 |

### 파일 및 기계 판독용 출력

| 플래그             | 기본값 | 설명                                                                                                                                                                                                      |
| ------------------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-j`               |        | `JSON` 형식으로 출력                                                                                                                                                                                      |
| `--pretty`         |        | `JSON` 출력을 보기 좋게 정리. `-j` 없이는 효과 없음                                                                                                                                                       |
| `--json-url <URL>` |        | `JSON` 출력을 출력하는 대신 HTTP 서버로 전송, 이벤트마다 `POST` 한 번, 예: `--json-url http://localhost:8000/tcping`. `JSON` 출력이 자동으로 켜지므로 `-j`는 필요 없음                                    |
| `--json-server`    |        | 프로브하지 않음. 주어진 호스트와 포트에서 기다리다가 전달된 모든 `JSON` 이벤트를 출력하여, 다른 곳에서 `--json-url`을 쓰는 tcping이 실행 결과를 보낼 곳을 마련함, 예: `tcping --json-server 0.0.0.0 8000` |
| `--csv <file>`     |        | 출력을 `CSV` 파일에 저장. 통계는 같은 이름에 `_stats` 접미사를 붙인 파일에 저장됨                                                                                                                         |
| `--csv-fixed-name` |        | `--csv` 파일 이름을 날짜/시간 접미사 없이 그대로 사용하므로, 반복 실행하면 같은 파일을 덮어씀                                                                                                             |
| `--sqlite <file>`  |        | 출력을 sqlite3 데이터베이스에 저장, 예: `--sqlite /tmp/tcping.db`. Windows에서는 사용 불가                                                                                                                |

### 메트릭

| 플래그                       | 기본값      | 설명                                                                                                                                                                                                                                                   |
| ---------------------------- | ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `--otlp <URL>`               |             | 결과를 출력하는 대신 메트릭으로 OTLP HTTP 엔드포인트([Grafana Alloy](https://grafana.com/docs/alloy/latest/)나 OpenTelemetry Collector 등)로 전송, 예: `--otlp http://localhost:4318`                                                                  |
| `--otlp-header <header>`     |             | `--otlp` 엔드포인트로 보낼 추가 HTTP 헤더, `"Name: value"` 형식, 예: `"Authorization: Bearer <token>"`. shell 기록에 남지 않도록 `OTLP_HEADER` 환경 변수로 줄 수도 있음                                                                                |
| `--influxdb <URL>`           |             | 결과를 출력하는 대신 라인 프로토콜로 [InfluxDB](https://www.influxdata.com/) v2 또는 v3 서버에 기록, 예: `--influxdb http://localhost:8086`                                                                                                            |
| `--influxdb-org <org>`       |             | 기록할 InfluxDB 조직. `--influxdb`와 함께 쓸 때 필수                                                                                                                                                                                                   |
| `--influxdb-bucket <bucket>` |             | 기록할 InfluxDB 버킷. `--influxdb`와 함께 쓸 때 필수                                                                                                                                                                                                   |
| `--influxdb-token <token>`   |             | InfluxDB API 토큰. `--influxdb`와 함께 쓸 때 필수. shell 기록에 남지 않도록 `INFLUXDB_TOKEN` 환경 변수로 줄 수도 있음                                                                                                                                  |
| `--stats-interval <seconds>` | `10`        | OTLP 엔드포인트나 InfluxDB로 통계를 보내는 주기. `--otlp`나 `--influxdb` 없이는 효과 없음                                                                                                                                                              |
| `--source-label <name>`      | 호스트 이름 | `JSON` 이벤트와 OTLP 또는 InfluxDB로 보내는 메트릭에서 이 컴퓨터의 이름을 지정하여, 같은 대상을 프로브하는 여러 컴퓨터를 구별할 수 있게 함. 다른 곳으로 보내는 결과는 지정하지 않으면 호스트 이름이 붙고, 그 컴퓨터에 남는 출력에는 아무것도 붙지 않음 |

### HTTP(S)와 UDP

| 플래그          | 기본값 | 설명                                                                                                                                                           |
| --------------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--insecure`    |        | `https://` 대상을 프로브할 때 서버 인증서를 검증하지 않음. 자체 서명 또는 만료된 인증서에 유용                                                                 |
| `--udp-server`  |        | 프로브하지 않음. 주어진 호스트와 포트에서 기다리다가 모든 UDP 데이터그램을 보낸 쪽으로 돌려보내, 이 컴퓨터를 향한 UDP 프로브가 응답을 받게 함                  |

---

## 기여하기

버그를 해결하고, 새 기능을 추가하고, [여기](https://github.com/pouriyajamshidi/tcping/issues)에서 볼 수 있는
열린 이슈를 돕는 풀 리퀘스트를 환영합니다.
현재 열린 이슈 수: ![GitHub issues](https://img.shields.io/github/issues/pouriyajamshidi/tcping.svg).

1. 편하게 다룰 수 있다고 느끼는 이슈를 고르세요.
1. 저장소를 Fork하세요.
1. 브랜치를 만드세요.
1. 작업을 커밋하세요.
1. 테스트를 추가하세요.
1. `make test`로 테스트를 실행하고 성공하는지 확인하세요.
1. 풀 리퀘스트를 만드세요.

풀 리퀘스트는 **하나의 특정 이슈/기능만 다루고** 둘 이상의 이슈를 함께 처리하지 않도록 해 주세요.
그래야 검토하기 쉽고 git 기록을 깔끔하게 유지하는 데 도움이 됩니다.

아주 작은 문제를 고치는 것이 아니라면, 작업을 시작하기 전에 먼저 **이슈**에서 의도를 알려 주세요.

나쁜 네트워크 환경에서 변경 사항을 시험해 보려면 `tools/netcond.sh`를 쓸 수 있습니다.
지연 시간을 더하고, 패킷 일부를 버리고, 한 목적지를 아예 차단할 수 있으며,
나중에 추가한 모든 것을 되돌려 줍니다.

```bash
sudo ./tools/netcond.sh delay 1.1.1.1 800ms 200ms 30
sudo ./tools/netcond.sh loss 1.1.1.1 40
sudo ./tools/netcond.sh clear
```

## 프로젝트 돕기

tcping이 유용하다면 주변 사람들과 공유해서 더 많은 사람에게 알리고 다른 사람들도 도움을 받을 수 있게 해 주세요.

또한 아래 링크로 프로젝트를 후원할 수 있습니다.

- 커피 한 잔 사 주기: [![Buy Me A Coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/pouriyajamshidi)
- GitHub 스폰서: [![Sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/pouriyajamshidi)

## 라이선스

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
