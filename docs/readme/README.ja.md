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

`ICMP` の代わりに `TCP`、`UDP`、`HTTP(S)` を使う、Linux の ping ツールにヒントを得たクロスプラットフォームの ping プログラムです。

> [!TIP]
> このドキュメントは [English](../../README.md)、[中文](README.cn.md)、[한국어](README.ko.md)、[فارسی](README.fa.md)、[Deutsch](README.de.md)、[Français](README.fr.md)、[Nederlands](README.nl.md)、[Español](README.es.md) でも読めます。

- `ICMP` がブロックされている環境で `ping` の代わりになり、`TCP`、`HTTP(S)`、`UDP` で疎通を確認します。
- `ping` と同じように、パケットロス率と、遅延の最小値、平均値、最大値、標準偏差を報告します。さらに、最も長い稼働時間とダウンタイム、そしてそれらがいつ起きたかも報告します。
- `Enter` キーを押すと、プログラムを止めずにいつでも統計情報を表示できます。
- **カラー**、**プレーンテキスト**、**JSON**、**CSV**、**sqlite3** の出力形式に対応しています。各プローブをメトリクスとして **OTLP**（Grafana Alloy、OpenTelemetry Collector など）や **InfluxDB** に送ることもできます。
- `HTTP(S)` プローブごとに、ステータスコード、TLS のバージョンと暗号スイート、証明書の有効期限、そして接続、TLS ハンドシェイク、最初のバイトまでの時間を表示します。
- 一定回数の失敗の後（`-r`）や、プローブのたびに（`--resolve-every-probe`）ターゲットのホスト名を再解決し、毎回の問い合わせにかかった時間を報告します。`DNS` ロードバランシングやグローバルサーバーロードバランサー `(GSLB)` のテストに便利です。
- **送信元インターフェース**、**タイムアウト**、**間隔**、**DNS サーバー**を選べます。`IPv4` と `IPv6` のどちらかを強制することもできます。
- _成功_ したプローブと _失敗した_ プローブを別々に数えるので、合計がひと目で分かります。

**tcping** の見た目と使い心地は[デモ](#デモ)で確認できます。

---

## 目次

- [デモ](#デモ)
- [ダウンロードとインストール](#ダウンロードとインストール)
- [使い方](#使い方)
- [フラグ](#フラグ)
- [コントリビュート](#コントリビュート)
- [プロジェクトを支援する](#プロジェクトを支援する)
- [ライセンス](#ライセンス)

---

## デモ

### 基本的な使い方

![tcping](../Images/gifs/tcping.gif)

<details>
<summary><b>クリックして他のデモを見る</b></summary>

### タイムスタンプ付きの IPv4 プローブ（`-D -4`）フラグ

![tcping timestamp example](../Images/gifs/tcping_timestamp.gif)

---

### プレーンテキスト出力で、Enter を押したときだけ統計情報を表示（`--no-color --no-stats`）フラグ

![tcping plain output example](../Images/gifs/tcping_plain.gif)

---

### ホスト名の再解決（`-r`）フラグ

![tcping resolve example](../Images/gifs/tcping_resolve.gif)

---

### JSON 出力（`-j --pretty`）フラグ

![tcping json example](../Images/gifs/tcping_json_pretty.gif)

---

### JSON 出力をコレクターへストリーミング（`--json-url --json-server`）フラグ

![tcping JSON streaming example](../Images/gifs/tcping_json_stream.gif)

---

### ホスト名の解決にかかった時間（`--resolve-every-probe`）フラグ

![tcping resolution timing example](../Images/gifs/tcping_dns_timing.gif)

---

### 送信元インターフェース（`-I`）フラグ

![tcping interface example](../Images/gifs/tcping_interface.gif)

---

### HTTP(S) プローブ

![tcping HTTP example](../Images/gifs/tcping_http.gif)

---

### HTTP(S) プローブの詳細（`-v`）フラグ

![tcping HTTP verbose example](../Images/gifs/tcping_http_verbose.gif)

---

### 証明書の検証をスキップ（`--insecure`）フラグ

![tcping skip TLS example](../Images/gifs/tcping_skip_tls.gif)

---

### プロキシ経由のプローブ（`--proxy`）フラグ

![tcping proxy example](../Images/gifs/tcping_proxy.gif)

---

### UDP サーバーへの UDP プローブ（`--udp-server`）フラグ

![tcping UDP example](../Images/gifs/tcping_udp.gif)

---

### 失敗したプローブだけを表示（`--failures-only`）フラグ

![tcping failures only example](../Images/gifs/tcping_failures_only.gif)

</details>

---

## ダウンロードとインストール

| プラットフォーム                                  | インストール方法                                                                                                 |
| ------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| [Windows](#windows)                               | `winget install pj.tcping`                                                                                       |
| [macOS](#macos)                                   | `brew install pouriyajamshidi/tap/tcping`                                                                        |
| [Debian、Ubuntu](#linux---リポジトリ)             | リポジトリを追加してから `sudo apt install tcping`                                                               |
| [Fedora、RHEL、CentOS](#linux---リポジトリ)       | リポジトリを追加してから `sudo dnf install tcping`                                                               |
| [Alpine](#linux---リポジトリ)                     | リポジトリを追加してから `sudo apk add tcping`                                                                   |
| [Arch、Manjaro](#linux---リポジトリ)              | `yay -S pj.tcping`、まだ利用できません                                                                           |
| [任意の Linux](#linux---ワンコマンドインストール) | [ワンコマンドインストール](#linux---ワンコマンドインストール)か[パッケージファイル](#linux---パッケージファイル) |
| [BSD](#bsd-と-linux---手動インストール)           | [ビルド済みバイナリ](#bsd-と-linux---手動インストール)                                                           |
| [Nix](#nix)                                       | `nix profile install github:pouriyajamshidi/tcping`                                                              |
| [Docker](#その他の方法)                           | `docker run -it pouriyajamshidi/tcping example.com 443`                                                          |
| [Go](#その他の方法)                               | `go install github.com/pouriyajamshidi/tcping/v3@latest`                                                         |

バイナリは静的ビルドです。必要なものがすべて 1 つのファイルに入っていて、お使いのマシンにライブラリがあるかどうかに左右されません。
[リリースページ](https://github.com/pouriyajamshidi/tcping/releases/latest/)にあり、_amd64_ 版と _arm64_ 版があります。

インストールが終わったら、[使い方](#使い方)に進んでください。

### Windows

Windows に **tcping** を入れるいちばん良い方法は、[WinGet](https://learn.microsoft.com/ja-jp/windows/package-manager/winget/?ref=github.com%2Fpouriyajamshidi%2Ftcping) を使った _Windows パッケージ マネージャー_ です。2020 年 9 月以降、ほぼすべての Windows _10_ と _11_ に最初から入っています。

```powershell
winget install pj.tcping
```

手動でインストールするには、ダウンロードした zip ファイルを展開し、`tcping.exe` をシステムの [PATH](https://www.howtogeek.com/118594/how-to-edit-your-system-path-for-easy-command-line-access/) 上のフォルダ（たとえば `C:\Windows\System32`）にコピーします。

> [!CAUTION]
> TCPING が Windows Defender や一部のマルウェア対策ソフトに誤検知されることがあります。Go のプログラムではよくあることです。Go チームの公式の説明は[こちら](https://go.dev/doc/faq#virus)です。

> [!WARNING]
> Windows のバイナリには `--sqlite`（sqlite3）出力形式はありません。ほかのフラグはすべて使えます。

### macOS

```bash
brew install pouriyajamshidi/tap/tcping
```

### Linux - ワンコマンドインストール

次をターミナルに貼り付けると、お使いのアーキテクチャに合った最新のバイナリを取得してインストールします。

```bash
cd /tmp &&
ARCH=$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/') &&
curl -LO "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
tar -xf "tcping-linux-$ARCH.tar.gz" &&
sudo install tcping /usr/local/bin/ &&
tcping --version
```

Tab 補完の設定は [Shell 補完](#shell-補完)を見てください。

`curl` がない場合は、その行を `wget` に置き換えます。

```bash
wget "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
```

### Linux - リポジトリ

リポジトリを一度追加すれば、あとは `apt`、`dnf`、`apk` が、ほかのパッケージと同じようにインストールとアップグレードをしてくれます。

**Debian**、**Ubuntu** とその派生ディストリビューション:

```bash
sudo install -d /etc/apt/keyrings &&
  sudo curl -fsSL https://packages.thegraynode.io/keys/pouriyajamshidi.gpg -o /etc/apt/keyrings/pouriyajamshidi.gpg &&
  echo "deb [signed-by=/etc/apt/keyrings/pouriyajamshidi.gpg] https://packages.thegraynode.io/deb ./" | sudo tee /etc/apt/sources.list.d/pouriyajamshidi.list &&
  sudo apt update &&
  sudo apt install tcping
```

**Fedora**、**RHEL**、**CentOS** とその派生ディストリビューション:

```bash
sudo curl -fsSL https://packages.thegraynode.io/rpm/pouriyajamshidi.repo -o /etc/yum.repos.d/pouriyajamshidi.repo &&
  sudo dnf install tcping
```

**Alpine**:

```bash
sudo curl -fsSL https://packages.thegraynode.io/keys/pouriyajamshidi.rsa.pub -o /etc/apk/keys/pouriyajamshidi.rsa.pub &&
  echo "https://packages.thegraynode.io/apk" | sudo tee -a /etc/apk/repositories &&
  sudo apk update &&
  sudo apk add tcping
```

**Arch**、**Manjaro**、**EndeavourOS** とその派生ディストリビューションでは、
[AUR](https://aur.archlinux.org/packages/pj.tcping) からお好みのヘルパーで入れられます。

```bash
yay -S pj.tcping
```

> [!NOTE]
> AUR パッケージはまだ利用できません。AUR の新規登録が現在停止しているためです。それまでは [Arch のパッケージファイル](#linux---パッケージファイル)か[ワンコマンドインストール](#linux---ワンコマンドインストール)を使ってください。

### Linux - パッケージファイル

リポジトリを追加したくない場合は、
[リリースページ](https://github.com/pouriyajamshidi/tcping/releases/latest/)
からお使いのディストリビューション用のパッケージをダウンロードしてインストールします。

| ディストリビューション | パッケージ                 | インストール                                        |
| ---------------------- | -------------------------- | --------------------------------------------------- |
| Debian、Ubuntu         | `tcping-amd64.deb`         | `sudo apt install -y ./tcping-amd64.deb`            |
| Fedora、RHEL、CentOS   | `tcping-amd64.rpm`         | `sudo dnf install -y ./tcping-amd64.rpm`            |
| Arch、Manjaro          | `tcping-amd64.pkg.tar.zst` | `sudo pacman -U ./tcping-amd64.pkg.tar.zst`         |
| Alpine                 | `tcping-amd64.apk`         | `sudo apk add --allow-untrusted ./tcping-amd64.apk` |

ARM のマシンでは `amd64` を `arm64` に置き換えてください。4 種類のパッケージはどれも、バイナリと
[Shell 補完](#shell-補完)をインストールします。`dnf` がないマシンでは `yum` でも構いません。
Alpine で `--allow-untrusted` が必要なのは、パッケージに Alpine の鍵で署名していないためです。

### BSD と Linux - 手動インストール

お使いの OS とアーキテクチャに合ったファイルをダウンロードします。

```bash
wget https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-freebsd-amd64.tar.gz
# または、Linux ARM64 のマシンで cURL を使う場合
curl -LO https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-arm64.tar.gz
```

展開して、バイナリをシステムの `PATH` 上のフォルダ（たとえば `/usr/local/bin/`）にコピーします。

```bash
tar -xvf tcping-freebsd-amd64.tar.gz &&
  chmod +x tcping &&
  sudo cp tcping /usr/local/bin/
```

Tab 補完の設定は [Shell 補完](#shell-補完)を見てください。

### Nix

**tcping** には flake が付いているので、[Nix](https://nixos.org/download/) が入っているマシンなら、
何もインストールせずにそのまま実行できます。

```bash
nix run github:pouriyajamshidi/tcping -- example.com 80
```

ずっと使うならインストールします。

```bash
nix profile install github:pouriyajamshidi/tcping
```

どちらの方法でも、バイナリと [Shell 補完](#shell-補完)が使えます。この flake には、
tcping の開発に必要なものがすべて入った `devShells.default` もあり、`nix develop` で使えます。

### その他の方法

- `Docker` イメージ:

  ```bash
  docker pull pouriyajamshidi/tcping:latest
  # または
  docker pull ghcr.io/pouriyajamshidi/tcping:latest
  ```

- `go install` を使う:

  > go のバージョンが `1.27.3` 以上である必要があります

  ```bash
  go install github.com/pouriyajamshidi/tcping/v3@latest
  ```

- [x tcping](https://x-cmd.com/pkg/tcping):

  [x-cmd](https://www.x-cmd.com) なら、**インストールせずにそのまま使えます**。

  ```bash
  x tcping example.com 80
  ```

  x-cmd を使って、root 権限なし、グローバル環境を汚さずに `tcping` をローカルへインストールすることもできます。

  ```bash
  x env use tcping
  tcping example.com 80
  ```

- 最後に、`make` コマンドを実行して自分でコードをビルドすることもできます。

  ```bash
  make build
  ```

  バイナリは `target` フォルダに作られます。

### Shell 補完

**tcping** には `bash`、`zsh`、`fish`、`PowerShell` 用の補完スクリプトが付いていて、
`--completions <shell>` で出力できます。フラグ、`-I` のインターフェース名、
`--csv` と `--sqlite` のファイル名を補完します。

Linux のパッケージ、Homebrew、AUR パッケージ、Nix、
[ワンコマンドインストール](#linux---ワンコマンドインストール)は、これらを正しい場所に置いてくれます。
そうでない場合は、お使いの Shell 用のものを保存してください。

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

- `PowerShell`（プロファイルから読み込む）:

  ```powershell
  New-Item -ItemType Directory -Force (Split-Path $PROFILE) | Out-Null
  Add-Content $PROFILE "tcping --completions powershell | Out-String | Invoke-Expression"
  ```

  `-i` のようなハイフン 1 つのフラグには PowerShell 7 が必要です。
  Windows PowerShell 5.1 は `--` で始まるフラグしか補完しません。
  PowerShell では `-I` のインターフェース名は補完されません。

その後、新しい Shell を開くと有効になります。

---

## 使い方

**tcping** にターゲットとポートを渡すと、プローブを始めます。

```bash
tcping www.example.com 443
tcping www.example.com:443          # ホスト:ポート の形式も使えます
tcping 192.168.1.1:80
tcping '[2001:db8::1]:443'          # シェルに解釈されないよう引用符で囲みます
tcping https://www.example.com      # HTTP(S) でプローブします
tcping udp://127.0.0.1 53           # UDP でプローブします
tcping tcp://www.example.com 443    # tcp:// も使えます
```

1 回の実行に対してできる指定の例です。

```bash
tcping www.example.com 443 -c 5            # 5 回プローブして停止
tcping www.example.com 443 -i 2 -t 5       # 間隔 2 秒、タイムアウト 5 秒
tcping www.example.com 443 -I eth2         # 指定したインターフェースからプローブを送信
tcping www.example.com 443 -4              # または -6 で、アドレスファミリーを強制
tcping www.example.com 443 -D              # プローブごとにタイムスタンプを表示
tcping www.example.com 443 -r 5            # 5 回失敗したらホスト名を再解決
```

デフォルトのカラー出力の代わりに使える出力形式です。

```bash
tcping www.example.com 443 -j              # JSON。--pretty で整形
tcping www.example.com 443 --no-color      # ANSI カラーなしのプレーンテキスト
tcping www.example.com 443 --csv out.csv   # CSV ファイル
tcping www.example.com 443 --sqlite out.db # sqlite3 データベース
```

Docker イメージにも同じターゲットとフラグを渡せます。

```bash
docker run -it pouriyajamshidi/tcping:latest example.com 443
# または、GitHub Container Registry を使う場合
docker run -it ghcr.io/pouriyajamshidi/tcping:latest example.com:443
```

> [!TIP]
> 実行中に `Enter` キーを押すと、[デモ](#デモ)のように、プログラムを終了せずにすべてのプローブの概要を確認できます。

> [!NOTE]
> より高度な使い方は、[こちら](#フラグ)の**利用できるフラグ**を見てください。

### HTTP(S) でプローブする

ホストとポートの代わりに URL を渡すと、tcping は HTTP(S) でプローブし、
ステータスコードとサーバーの応答にかかった時間を報告します。
レスポンスの本文はダウンロードしないので、大きなページでもプローブが遅くなりません。

```bash
tcping https://www.example.com/health
```

ポートはスキームで決まり、`80` か `443` です。URL にポートが含まれている場合や、
その後ろに別のポートを渡した場合はそちらが使われます。

```bash
tcping http://www.example.com:8080/health
# 同じ意味です
tcping http://www.example.com/health 8080
```

`-v` を付けると、1 回のプローブで集めたすべての情報を表示します。HTTP のバージョン、
TLS のバージョンと暗号スイート、証明書の有効期限まで何日残っているか、
そして接続、TLS ハンドシェイク、最初のバイトまでの時間です。`--insecure` は証明書のチェックを
省略します。自己署名の証明書や期限切れの証明書を相手にするときに必要です。

```bash
tcping https://www.example.com/health -v
tcping https://self-signed.example.com --insecure
```

### プロキシ経由でプローブする

`--proxy` は TCP と HTTP(S) のプローブをプロキシ経由で送ります。`http://`、
`https://`、`socks5://` のプロキシが使え、ログイン情報は URL に入れます。例:
`socks5://user:password@127.0.0.1:1080`。tcping は `HTTPS_PROXY`、`HTTP_PROXY`、
`NO_PROXY` 環境変数を読まないので、プローブがプロキシを通るのは指定したときだけです。

```bash
tcping https://www.example.com --proxy http://proxy.example.com:3128
tcping db.example.com 5432 --proxy socks5://127.0.0.1:1080
```

TCP のプローブはプロキシにターゲットへの接続を頼み、プロキシが接続できたと答えれば
成功です。そのため時間には、プロキシまでの道のりと、プロキシからターゲットまでの
道のりの両方が含まれます。多くの HTTP プロキシは `443` 番ポートへの接続しか許さない
ので、ターゲットが動いていても別のポートへのプローブは失敗することがあります。SSH で
入れるネットワークへの SOCKS5 プロキシは、`ssh -D 1080 bastion.example.com` で
`127.0.0.1:1080` に開けます。

プロキシはターゲットのホスト名へ自分で接続します。そのため送信元アドレスと、HTTP(S)
プローブの接続時間はプロキシまでの接続のものになり、実行の最初の行にそのプロキシが
表示されます。tcping はホスト名を自分で解決しないので、出力にはそのアドレスが表示
されず、`-r` と `--resolve-every-probe` はプロキシと一緒には使えません。`-4`、`-6`、
`--dns-server` は代わりにプロキシ自身のホスト名に使われます。UDP のプローブは
プロキシを通りません。PAC ファイルで設定されたプロキシは手で指定する必要があります。
tcping は PAC ファイルを読まないためです。

### UDP でプローブする

```bash
tcping udp://127.0.0.1 53
```

UDP にはハンドシェイクがないので、相手が応答したときだけプローブが成功します。
応答を得るには、相手側で tcping を UDP サーバーとして動かします。受け取ったデータグラムをそのまま送信元へ返します。

```bash
# プローブされる側のマシンで:
tcping --udp-server 127.0.0.1 9999
# プローブする側のマシンで:
tcping udp://127.0.0.1 9999
```

> [!NOTE]
> 拒否された UDP プローブは `port unreachable`（ポートに到達できない）と報告されます。
> 何かが私たちを止めているということです。まったく応答がないプローブも失敗として報告されますが、
> ポートは開いているのに黙っているのか、途中でパケットが捨てられたのかは区別できません。
>
> Windows では、拒否されたプローブは常に応答なしとして表示されます。Go が Windows の
> UDP ソケットで「ポートに到達できない」エラーを無効にしているため、tcping はそれを見ることができません。

UDP プローブはそれぞれ、自分の番号をペイロードとして送り、サーバーはそれをそのまま返します。
`-v` を付けると各行にその番号が表示されるので、どのプローブが失われたかが分かります。

```text
Reply from 127.0.0.1 on port 9999 UDP_conn=4 time=1.276 ms
    reply echoed back probe 4
```

### JSON 出力をサーバーへストリーミングする

`JSON` 出力は画面に出さなくても構いません。`--json-url` にアドレスを渡すと、
イベントが起きるたびに、1 リクエストにつき 1 イベントとしてそこへ `POST` されます。
プローブするマシンから、収集するマシンへ報告できます。

```bash
tcping www.example.com 443 --json-url http://localhost:8000/tcping
```

イベントとそのフィールドは `-j` で表示されるものとまったく同じなので、
パイプの出力をすでに読んでいるプログラムなら、これも読めます。
ストリーミングされるイベントには、送信したマシンが `source` フィールドに書かれます。
デフォルトはそのマシンのホスト名です。そのため、複数のマシンからイベントを集めるコレクターは、
どの実行の結果かを区別できます。`--source-label` で好きな名前を付けられます。

```bash
tcping www.example.com 443 --json-url http://collector:8000/tcping --source-label brussels
```

tcping は受け取る側にもなれます。`--udp-server` が UDP プローブに応答するのと同じです。
`--json-server` はプローブをしません。指定したホストとポートで待ち受け、送られてきたイベントを
すべて表示します（パスは何でも構いません）。収集するマシンには、ほかに何もインストールする必要がありません。

```bash
# 収集するマシンで:
tcping --json-server 0.0.0.0 8000

# プローブするマシンで:
tcping www.example.com 443 --json-url http://collector:8000/tcping
```

イベントは標準出力に、それ以外は標準エラーに出力されます。そのため
`tcping --json-server 0.0.0.0 8000 > events.jsonl` で得られるファイルにはイベントだけが入り、
`JSON` ではないものは表示されずに拒否されます。

`JSON` の本文を持つ `POST` を受け付けるサーバーなら何でも使えます。たとえば次のようなものです。

```python
from fastapi import FastAPI

app = FastAPI()


@app.post("/tcping")
async def tcping(event: dict) -> None:
    print(event)
```

1 回の実行のイベントはすべて同じ接続で送られます。この接続はプローブの合間も開いたままで、
毎回つなぎ直しません。サーバーが落ちたり調子が悪かったりしても、プローブは止まりません。
tcping は一度だけ知らせて実行を続け、届けられないイベントは捨てます。

### OTLP や InfluxDB に送る

tcping は、プローブを 1 つずつ表示する代わりに、メトリクスとして送れます。
これで 1 回の実行をグラフにでき、複数のマシンで同じターゲットを監視することもできます。
OTLP を受け付けるサービスなら何でも送れます。たとえば
`--web.enable-otlp-receiver` で起動した [Prometheus](https://prometheus.io/) です。

```bash
tcping www.example.com 443 --otlp http://localhost:9090/api/v1/otlp
```

または、複数の宛先へ転送できる [Grafana Alloy](https://grafana.com/docs/alloy/latest/) や
OpenTelemetry Collector に送ります。

```bash
tcping www.example.com 443 --otlp http://localhost:4318
```

マネージドのバックエンドにはトークンが必要で、サービスごとに入れるヘッダーが違います。
`--otlp-header` にヘッダー全体を渡すか、シェルの履歴に残らないよう `OTLP_HEADER` 環境変数に入れます。

```bash
export OTLP_HEADER="Authorization: Bearer your-api-token" && \
  tcping www.example.com 443 --otlp https://otlp.example.com
```

または、ラインプロトコルで [InfluxDB](https://www.influxdata.com/) v2 や v3 のサーバーへ直接送ります。

```bash
export INFLUXDB_TOKEN=your-api-token && \
  tcping www.example.com 443 --influxdb http://localhost:8086 --influxdb-org home --influxdb-bucket tcping
```

各プローブには、往復時間、成功したかどうか、ターゲットが解決されたアドレス、カウンターが付きます。
ターゲットが HTTP や UDP の場合は、その詳細も付きます。
さらに、通常は終了時に見る完全な統計情報のブロックを 10 秒ごとに送るので、
誰も見ていない実行でも報告されます。この間隔は `--stats-interval` で変えられます。

各マシンは、メトリクスの `source` ラベルで自分の名前を伝えます。デフォルトはホスト名です。
同じターゲットを複数のマシンでプローブしても、データは重なり合わずにそれぞれの系列に入ります。
`--source-label` で好きな名前を付けられます。

```bash
tcping www.example.com 443 --otlp http://localhost:4318 --source-label brussels
```

送られるすべての内容、その問い合わせ方、そして Grafana と tcping を含み 1 つのコマンドで起動できる、
すぐ使える Prometheus、Alloy、OpenTelemetry Collector、InfluxDB の環境
（先にサーバーを用意しなくても試せます）は、
[docs/observability](../observability/README.md) にあります。

---

## フラグ

フラグはターゲットの前でも後ろでも置けます。ハイフンは 1 つでも 2 つでも構いません。
つまり `-c 5` と `--c 5` は同じフラグです。

### 全般

| フラグ      | デフォルト | 説明                           |
| ----------- | ---------- | ------------------------------ |
| `-h`        |            | 利用できるフラグを表示して終了 |
| `--version` |            | バージョンを表示して終了       |
| `-u`        |            | アップデートを確認して終了     |

### プローブ

| フラグ          | デフォルト | 説明                                                                  |
| --------------- | ---------- | --------------------------------------------------------------------- |
| `-c <n>`        | 無制限     | 結果にかかわらず、`<n>` 回プローブしたら停止                          |
| `-i <seconds>`  | `1`        | プローブの間隔。実数も使えます。例: `-i 0.5`                          |
| `-t <seconds>`  | `1`        | 応答を待つ時間（秒）。実数も使えます。`0` はタイムアウトなし。`-t` を指定しない場合、HTTP(S) のターゲットは `3` 秒待ちます |
| `-4`            |            | IPv4 アドレスだけを使う                                               |
| `-6`            |            | IPv6 アドレスだけを使う                                               |
| `-I <name\|IP>` |            | プローブと DNS クエリの送信に使うインターフェース名または IP アドレス |
| `--proxy <URL>` |            | TCP と `HTTP(S)` のプローブをプロキシ経由で送る。例: `--proxy http://proxy.example.com:3128`。`HTTPS_PROXY`、`HTTP_PROXY`、`NO_PROXY` 環境変数は使わない |

> [!TIP]
> `-4` も `-6` も指定しない場合、tcping は DNS の検索結果からランダムに 1 つの IP アドレスを選びます。

### 名前解決

| フラグ                    | デフォルト           | 説明                                                                                                                          |
| ------------------------- | -------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| `-r <n>`                  | なし                 | `<n>` 回プローブに失敗したら、ターゲットのホスト名を再解決する。例: `-r 10`                                                   |
| `--resolve-every-probe`   |                      | 起動時だけでなく、プローブのたびにターゲットのホスト名を解決する。`-r` より優先され、ターゲットが IP アドレスのときは効果なし |
| `--dns-server <IP>`       | システムのデフォルト | カスタムの DNS サーバーを使う。IP とポートの組み合わせも使えます。例: `--dns-server 1.1.1.1:53`                               |
| `--dns-timeout <seconds>` | `2`                  | DNS の応答を待つ時間（秒）。実数も使えます。`0` はタイムアウトなし                                                            |

### ターミナル出力

| フラグ                  | デフォルト | 説明                                                                                                                                                                                                                                                                                 |
| ----------------------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `-D`                    |            | プローブごとにタイムスタンプを表示する                                                                                                                                                                                                                                               |
| `--no-color`            |            | 色なしで出力する                                                                                                                                                                                                                                                                     |
| `--show-source-address` |            | プローブに使った送信元の IP アドレスとポートを表示する                                                                                                                                                                                                                               |
| `--failures-only`       |            | 失敗したプローブだけを表示する。成功したプローブも数には入る                                                                                                                                                                                                                         |
| `--no-stats`            |            | プログラムの終了時に統計情報を表示しない。**Enter** キーを押せば表示される。出力先がターミナルでないときは効果なし                                                                                                                                                                   |
| `-v`                    |            | HTTP(S) プローブで集めたすべての情報を表示する。HTTP のバージョン、TLS のバージョンと暗号スイート、証明書の有効期限、接続、TLS、最初のバイトまでの時間。UDP ターゲットでは、プローブの番号と、応答がそれを返したかを表示し、失われたプローブを区別できる。TCP ターゲットでは効果なし |

### ファイルと機械可読な出力

| フラグ             | デフォルト | 説明                                                                                                                                                                                                                      |
| ------------------ | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-j`               |            | `JSON` 形式で出力する                                                                                                                                                                                                     |
| `--pretty`         |            | `JSON` 出力を整形する。`-j` がないと効果なし                                                                                                                                                                              |
| `--json-url <URL>` |            | `JSON` 出力を表示する代わりに HTTP サーバーへ送る。イベントごとに 1 回 `POST` する。例: `--json-url http://localhost:8000/tcping`。`JSON` 出力が自動で有効になるので `-j` は不要                                          |
| `--json-server`    |            | プローブをしない。指定したホストとポートで待ち受け、送られてきた `JSON` イベントをすべて表示する。これにより、別の場所で `--json-url` を使う tcping に、実行結果の送り先ができる。例: `tcping --json-server 0.0.0.0 8000` |
| `--csv <file>`     |            | 出力を `CSV` ファイルに保存する。統計情報は、同じ名前に `_stats` を付けたファイルに保存される                                                                                                                             |
| `--csv-fixed-name` |            | `--csv` のファイル名をそのまま使い、日付や時刻の接尾辞を付けない。そのため、繰り返し実行すると同じファイルを上書きする                                                                                                    |
| `--sqlite <file>`  |            | 出力を sqlite3 データベースに保存する。例: `--sqlite /tmp/tcping.db`。Windows では使えない                                                                                                                                |

### メトリクス

| フラグ                       | デフォルト | 説明                                                                                                                                                                                                                               |
| ---------------------------- | ---------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--otlp <URL>`               |            | 結果を表示する代わりに、メトリクスとして OTLP HTTP エンドポイント（[Grafana Alloy](https://grafana.com/docs/alloy/latest/) や OpenTelemetry Collector など）へ送る。例: `--otlp http://localhost:4318`                             |
| `--otlp-header <header>`     |            | `--otlp` のエンドポイントへ送る追加の HTTP ヘッダー。`"Name: value"` の形式。例: `"Authorization: Bearer <token>"`。シェルの履歴に残らないよう、`OTLP_HEADER` 環境変数に入れることもできる                                         |
| `--influxdb <URL>`           |            | 結果を表示する代わりに、ラインプロトコルで [InfluxDB](https://www.influxdata.com/) v2 または v3 のサーバーへ書き込む。例: `--influxdb http://localhost:8086`                                                                       |
| `--influxdb-org <org>`       |            | 書き込み先の InfluxDB の organization。`--influxdb` を使うときは必須                                                                                                                                                               |
| `--influxdb-bucket <bucket>` |            | 書き込み先の InfluxDB の bucket。`--influxdb` を使うときは必須                                                                                                                                                                     |
| `--influxdb-token <token>`   |            | InfluxDB の API トークン。`--influxdb` を使うときは必須。シェルの履歴に残らないよう、`INFLUXDB_TOKEN` 環境変数に入れることもできる                                                                                                 |
| `--stats-interval <seconds>` | `10`       | OTLP エンドポイントや InfluxDB へ統計情報を送る頻度。`--otlp` も `--influxdb` もないと効果なし                                                                                                                                     |
| `--source-label <name>`      | ホスト名   | `JSON` イベントと、OTLP や InfluxDB へ送るメトリクスの中で、このマシンの名前を決める。同じターゲットをプローブする複数のマシンを区別できる。ほかへ送る結果では、指定がなければホスト名が付く。そのマシンに残る出力には何も付かない |

### HTTP(S) と UDP

| フラグ          | デフォルト | 説明                                                                                                                                                           |
| --------------- | ---------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--insecure`    |            | `https://` のターゲットをプローブするとき、サーバー証明書を検証しない。自己署名や期限切れの証明書に使える                                                      |
| `--udp-server`  |            | プローブをしない。指定したホストとポートで待ち受け、受け取った UDP データグラムをすべて送信元へ返す。このマシンへの UDP プローブが応答を得られる               |

---

## コントリビュート

バグの修正、新機能の追加、そして
[こちら](https://github.com/pouriyajamshidi/tcping/issues)にある未解決の issue の解決を助けるプルリクエストを歓迎します。
現在の未解決の issue の数: ![GitHub issues](https://img.shields.io/github/issues/pouriyajamshidi/tcping.svg)。

1. 取り組めそうな issue を選びます。
1. リポジトリを Fork します。
1. ブランチを作ります。
1. 作業をコミットします。
1. テストを追加します。
1. `make test` でテストを実行し、成功することを確認します。
1. プルリクエストを作ります。

プルリクエストは**特定の 1 つの issue や機能だけを扱い**、2 つ以上の問題を一緒に扱わないようにしてください。
そのほうがレビューしやすく、git の履歴もきれいに保てます。

ごく小さな修正でない限り、作業を始める前に **issue** で何をするつもりかを伝えてください。

ネットワークの状態が悪いときの変更を試すには、`tools/netcond.sh` が使えます。遅延を加える、
一部のパケットを捨てる、特定の宛先をブロックする、といったことができ、後で追加したものをすべて元に戻せます。

```bash
sudo ./tools/netcond.sh delay 1.1.1.1 800ms 200ms 30
sudo ./tools/netcond.sh loss 1.1.1.1 40
sudo ./tools/netcond.sh clear
```

## プロジェクトを支援する

tcping が役に立ったら、身近な人たちに紹介して、より多くの人に届け、ほかの人の役にも立つようにしてください。

また、次のリンクからプロジェクトを支援することもできます。

- コーヒーをおごる: [![Buy Me A Coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/pouriyajamshidi)
- GitHub スポンサー: [![Sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/pouriyajamshidi)

## ライセンス

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
