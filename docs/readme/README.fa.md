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

یک برنامه‌ی ping چندسکویی که به جای `ICMP` از `TCP`، `UDP` یا `HTTP(S)` استفاده می‌کند و از ابزار ping در لینوکس الهام گرفته است.

> [!TIP]
> این سند به زبان‌های [English](../../README.md)، [中文](README.cn.md)، [日本語](README.ja.md)، [한국어](README.ko.md)، [Deutsch](README.de.md)، [Français](README.fr.md) و [Nederlands](README.nl.md) هم موجود است.

- جایگزینی برای `ping` در جاهایی که `ICMP` مسدود است و از طریق `TCP`، `HTTP(S)` یا `UDP` بررسی می‌کند.
- مانند `ping` درصد از دست رفتن بسته‌ها و کمینه، میانگین، بیشینه و انحراف معیار تأخیر را گزارش می‌دهد، به‌علاوه‌ی طولانی‌ترین زمان فعال بودن و قطعی و اینکه چه زمانی رخ داده‌اند.
- با فشردن کلید `Enter` در هر لحظه آمار را چاپ می‌کند، بدون اینکه برنامه متوقف شود.
- خروجی را به‌صورت **رنگی**، **ساده**، **JSON**، **CSV** یا **sqlite3** می‌دهد، یا هر probe را به‌شکل metric از طریق **OTLP** (مانند Grafana Alloy و OpenTelemetry Collector) یا به **InfluxDB** می‌فرستد.
- برای هر probe از نوع `HTTP(S)` کد وضعیت، نسخه‌ی TLS و cipher، تاریخ انقضای گواهی و زمان‌های اتصال، دست‌دهی TLS و اولین بایت را نشان می‌دهد.
- نام میزبان هدف را پس از تعدادی شکست (`-r`) یا پیش از هر probe (`--resolve-every-probe`) دوباره resolve می‌کند و گزارش می‌دهد هر جستجو چقدر طول کشیده است. برای آزمایش توزیع بار `DNS` یا Global Server Load Balancer `(GSLB)` مناسب است.
- به شما اجازه می‌دهد **رابط مبدأ**، **timeout**، **فاصله‌ی زمانی** و **سرور DNS** را انتخاب کنید و `IPv4` یا `IPv6` را اجباری کنید.
- probeهای _موفق_ و _ناموفق_ را جدا از هم شماره‌گذاری می‌کند، بنابراین جمع‌ها یک‌نگاهی دیده می‌شوند.

برای دیدن ظاهر و حس کار با **tcping** به [دموها](#دموها) سر بزنید.

---

## فهرست مطالب

- [دموها](#دموها)
- [دانلود و نصب](#دانلود-و-نصب)
- [استفاده](#استفاده)
- [پارامترها](#پارامترها)
- [مشارکت](#مشارکت)
- [حمایت از پروژه](#حمایت-از-پروژه)
- [مجوز](#مجوز)

---

## دموها

<details>
<summary>برای باز کردن کلیک کنید</summary>

### استفاده‌ی پایه

![tcping](../Images/gifs/tcping.gif)

---

### probeهای زمان‌دار روی IPv4 (پارامترهای `-D -4`)

![tcping timestamp example](../Images/gifs/tcping_timestamp.gif)

---

### خروجی ساده، آمار فقط با فشردن Enter (پارامترهای `--no-color --no-stats`)

![tcping plain output example](../Images/gifs/tcping_plain.gif)

---

### تلاش دوباره برای resolve کردن نام میزبان (پارامتر `-r`)

![tcping resolve example](../Images/gifs/tcping_resolve.gif)

---

### خروجی JSON (پارامترهای `-j --pretty`)

![tcping json example](../Images/gifs/tcping_json_pretty.gif)

---

### ارسال زنده‌ی خروجی JSON به یک collector (پارامترهای `--json-url --json-server`)

![tcping JSON streaming example](../Images/gifs/tcping_json_stream.gif)

---

### زمان resolve کردن نام میزبان (پارامتر `--resolve-every-probe`)

![tcping resolution timing example](../Images/gifs/tcping_dns_timing.gif)

---

### رابط مبدأ (پارامتر `-I`)

![tcping interface example](../Images/gifs/tcping_interface.gif)

---

### probeهای HTTP(S)

![tcping HTTP example](../Images/gifs/tcping_http.gif)

---

### جزئیات probeهای HTTP(S) (پارامتر `-v`)

![tcping HTTP verbose example](../Images/gifs/tcping_http_verbose.gif)

---

### نادیده گرفتن بررسی گواهی (پارامتر `--insecure`)

![tcping skip TLS example](../Images/gifs/tcping_skip_tls.gif)

---

### probeها از طریق proxy (پارامتر `--proxy`)

![tcping proxy example](../Images/gifs/tcping_proxy.gif)

---

### probeهای UDP به یک سرور UDP (پارامتر `--udp-server`)

![tcping UDP example](../Images/gifs/tcping_udp.gif)

---

### فقط probeهای ناموفق (پارامتر `--failures-only`)

![tcping failures only example](../Images/gifs/tcping_failures_only.gif)

</details>

---

## دانلود و نصب

| سکو                                         | نصب با                                                                   |
| ------------------------------------------- | ------------------------------------------------------------------------ |
| [Windows](#windows)                         | `winget install pj.tcping`                                               |
| [macOS](#macos)                             | `brew install pouriyajamshidi/tap/tcping`                                |
| [Debian, Ubuntu](#linux---مخازن-بسته)       | `sudo apt install tcping`، پس از افزودن مخزن                             |
| [Fedora, RHEL, CentOS](#linux---مخازن-بسته) | `sudo dnf install tcping`، پس از افزودن مخزن                             |
| [Alpine](#linux---مخازن-بسته)               | `sudo apk add tcping`، پس از افزودن مخزن                                 |
| [Arch, Manjaro](#linux---مخازن-بسته)        | `yay -S tcping-bin`                                                      |
| [هر لینوکسی](#linux---نصب-با-یک-خط)         | [نصب با یک خط](#linux---نصب-با-یک-خط) یا [فایل بسته](#linux---فایل-بسته) |
| [BSD](#bsd-و-linux---نصب-دستی)              | [فایل اجرایی از پیش ساخته شده](#bsd-و-linux---نصب-دستی)                  |
| [Nix](#nix)                                 | `nix profile install github:pouriyajamshidi/tcping`                      |
| [Docker](#موارد-دیگر)                       | `docker run -it pouriyajamshidi/tcping example.com 443`                  |
| [Go](#موارد-دیگر)                           | `go install github.com/pouriyajamshidi/tcping/v3@latest`                 |

فایل‌های اجرایی ایستا هستند، یعنی هر آنچه لازم دارند را در یک فایل با خود دارند
و به وجود هیچ کتابخانه‌ای روی دستگاه شما وابسته نیستند. آن‌ها در
[صفحه‌ی انتشارها](https://github.com/pouriyajamshidi/tcping/releases/latest/) برای
_amd64_ و _arm64_ قرار دارند.

پس از پایان نصب، به بخش [استفاده](#استفاده) بروید.

### Windows

بهترین راه نصب **tcping** روی ویندوز، _Windows Package Manager_ با [WinGet](https://learn.microsoft.com/fa-ir/windows/package-manager/winget/?ref=github.com%2Fpouriyajamshidi%2Ftcping) است که از سپتامبر ۲۰۲۰ به‌طور پیش‌فرض روی تقریباً همه‌ی دستگاه‌های ویندوز _10_ و _11_ وجود دارد:

```powershell
winget install pj.tcping
```

برای نصب دستی، فایل zip دانلودشده را باز کنید و `tcping.exe` را در [PATH](https://www.howtogeek.com/118594/how-to-edit-your-system-path-for-easy-command-line-access/) سیستم خود، مثلاً `C:\Windows\System32`، کپی کنید.

> [!CAUTION]
> ممکن است TCPING به‌اشتباه توسط Windows Defender یا برخی نرم‌افزارهای ضد بدافزار علامت‌گذاری شود. این برای برنامه‌های Go رایج است. بیانیه‌ی رسمی تیم Go را [اینجا](https://go.dev/doc/faq#virus) ببینید.

> [!WARNING]
> قالب خروجی `--sqlite` (sqlite3) دیگر در فایل‌های اجرایی ویندوز موجود نیست. همه‌ی پارامترهای دیگر همان‌طور که انتظار می‌رود کار می‌کنند.

### macOS

```bash
brew install pouriyajamshidi/tap/tcping
```

### Linux - نصب با یک خط

دستور زیر را در ترمینال خود بچسبانید تا آخرین فایل اجرایی مناسب معماری شما دریافت و نصب شود:

```bash
cd /tmp &&
ARCH=$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/') &&
curl -LO "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
tar -xf "tcping-linux-$ARCH.tar.gz" &&
sudo install tcping /usr/local/bin/ &&
tcping --version
```

برای راه‌اندازی تکمیل خودکار با کلید Tab بخش [تکمیل خودکار شل](#تکمیل-خودکار-شل) را ببینید.

اگر `curl` ندارید، خط آن را با `wget` عوض کنید:

```bash
wget "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
```

### Linux - مخازن بسته

با یک‌بار افزودن مخزن، `apt`، `dnf` یا `apk` از آن به بعد نصب و
به‌روزرسانی را مانند هر بسته‌ی دیگر انجام می‌دهد.

**Debian**، **Ubuntu** و مشتقات آن‌ها:

```bash
sudo install -d /etc/apt/keyrings &&
  sudo curl -fsSL https://pouriyajamshidi.github.io/packages/keys/pouriyajamshidi.gpg -o /etc/apt/keyrings/pouriyajamshidi.gpg &&
  echo "deb [signed-by=/etc/apt/keyrings/pouriyajamshidi.gpg] https://pouriyajamshidi.github.io/packages/deb ./" | sudo tee /etc/apt/sources.list.d/pouriyajamshidi.list &&
  sudo apt update &&
  sudo apt install tcping
```

**Fedora**، **RHEL**، **CentOS** و مشتقات آن‌ها:

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

**Arch**، **Manjaro**، **EndeavourOS** و مشتقات آن‌ها آن را در
[AUR](https://aur.archlinux.org/packages/tcping-bin) دارند، پس از ابزار کمکی دلخواه خود استفاده کنید:

```bash
yay -S tcping-bin
```

### Linux - فایل بسته

بدون افزودن مخزن، بسته‌ی مناسب توزیع خود را از
[صفحه‌ی انتشارها](https://github.com/pouriyajamshidi/tcping/releases/latest/) دانلود و
نصب کنید:

| توزیع                | بسته                       | نصب                                                 |
| -------------------- | -------------------------- | --------------------------------------------------- |
| Debian, Ubuntu       | `tcping-amd64.deb`         | `sudo apt install -y ./tcping-amd64.deb`            |
| Fedora, RHEL, CentOS | `tcping-amd64.rpm`         | `sudo dnf install -y ./tcping-amd64.rpm`            |
| Arch, Manjaro        | `tcping-amd64.pkg.tar.zst` | `sudo pacman -U ./tcping-amd64.pkg.tar.zst`         |
| Alpine               | `tcping-amd64.apk`         | `sudo apk add --allow-untrusted ./tcping-amd64.apk` |

روی دستگاه‌های ARM به جای `amd64` از `arm64` استفاده کنید. هر چهار مورد فایل اجرایی و
[تکمیل خودکار شل](#تکمیل-خودکار-شل) را برای شما نصب می‌کنند. دستگاه‌های بدون `dnf` می‌توانند
به جای آن از `yum` استفاده کنند، و Alpine به `--allow-untrusted` نیاز دارد چون بسته‌های ما
با کلید Alpine امضا نشده‌اند.

### BSD و Linux - نصب دستی

فایل مناسب سیستم‌عامل و معماری خود را دانلود کنید:

```bash
wget https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-freebsd-amd64.tar.gz
# یا برای دستگاه‌های Linux ARM64 و با استفاده از cURL
curl -LO https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-arm64.tar.gz
```

آن را باز کنید و فایل اجرایی را در `PATH` سیستم خود، مثلاً `/usr/local/bin/`، کپی کنید:

```bash
tar -xvf tcping-freebsd-amd64.tar.gz &&
  chmod +x tcping &&
  sudo cp tcping /usr/local/bin/
```

برای راه‌اندازی تکمیل خودکار با کلید Tab بخش [تکمیل خودکار شل](#تکمیل-خودکار-شل) را ببینید.

### Nix

**tcping** یک flake دارد، پس روی هر دستگاهی که [Nix](https://nixos.org/download/) دارد
می‌توانید آن را بدون نصب چیزی اجرا کنید:

```bash
nix run github:pouriyajamshidi/tcping -- example.com 80
```

یا برای همیشه نصبش کنید:

```bash
nix profile install github:pouriyajamshidi/tcping
```

هر دو، فایل اجرایی و [تکمیل خودکار شل](#تکمیل-خودکار-شل) را به شما می‌دهند. این
flake همچنین یک `devShells.default` دارد که هر آنچه برای کار روی tcping لازم است
در آن هست و با `nix develop` به دست می‌آید.

### موارد دیگر

- تصاویر `Docker`:

  ```bash
  docker pull pouriyajamshidi/tcping:latest
  # یا
  docker pull ghcr.io/pouriyajamshidi/tcping:latest
  ```

- با استفاده از `go install`:

  > این کار دست‌کم به نسخه‌ی `1.26.8` از go نیاز دارد

  ```bash
  go install github.com/pouriyajamshidi/tcping/v3@latest
  ```

- [x tcping](https://x-cmd.com/pkg/tcping):

  **مستقیم و بدون نصب** در [x-cmd](https://www.x-cmd.com).

  ```bash
  x tcping example.com 80
  ```

  یا `tcping` را با x-cmd به‌صورت محلی نصب کنید، بدون نیاز به دسترسی root و بدون اثر بر تنظیمات سراسری خود.

  ```bash
  x env use tcping
  tcping example.com 80
  ```

- در نهایت می‌توانید با اجرای دستور `make` کد را خودتان کامپایل کنید:

  ```bash
  make build
  ```

  این کار فایل اجرایی را در پوشه‌ی `target` قرار می‌دهد.

### تکمیل خودکار شل

**tcping** اسکریپت‌های تکمیل خودکار برای `bash`، `zsh`، `fish` و `PowerShell` دارد
و آن‌ها را با `--completions <shell>` چاپ می‌کند. این اسکریپت‌ها پارامترها،
نام رابط‌ها برای `-I` و نام فایل‌ها برای `--csv` و `--sqlite` را تکمیل می‌کنند.

بسته‌های لینوکس، Homebrew، بسته‌ی AUR، Nix و
[نصب با یک خط](#linux---نصب-با-یک-خط) آن‌ها را برای شما در جای خود می‌گذارند.
در غیر این صورت، اسکریپت شل خود را ذخیره کنید:

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

- `PowerShell`، با بارگذاری از پروفایل خود:

  ```powershell
  Add-Content $PROFILE "tcping --completions powershell | Out-String | Invoke-Expression"
  ```

پس از آن یک شل جدید باز کنید تا اعمال شوند.

---

## استفاده

به **tcping** یک هدف و یک پورت بدهید تا probe کردن را شروع کند:

```bash
tcping www.example.com 443
tcping www.example.com:443          # قالب میزبان:پورت هم کار می‌کند
tcping 192.168.1.1:80
tcping '[2001:db8::1]:443'          # داخل کوتیشن، تا شل کاری به آن نداشته باشد
tcping https://www.example.com      # از طریق HTTP(S) بررسی می‌شود
tcping udp://127.0.0.1 53           # از طریق UDP بررسی می‌شود
tcping tcp://www.example.com 443    # tcp:// هم کار می‌کند
```

چند نمونه از چیزهایی که می‌توانید از یک اجرا بخواهید:

```bash
tcping www.example.com 443 -c 5            # پس از ۵ probe متوقف شو
tcping www.example.com 443 -i 2 -t 5       # ۲ ثانیه بین probeها، ۵ ثانیه timeout
tcping www.example.com 443 -I eth2         # probeها را از یک رابط مشخص بفرست
tcping www.example.com 443 -4              # یا -6، برای اجباری کردن یک خانواده‌ی آدرس
tcping www.example.com 443 -D              # برای هر probe زمان نشان بده
tcping www.example.com 443 -r 5            # پس از ۵ شکست نام میزبان را دوباره resolve کن
```

و قالب‌های خروجی که می‌تواند به جای خروجی رنگی پیش‌فرض بسازد:

```bash
tcping www.example.com 443 -j              # JSON، برای خوانا شدن --pretty را اضافه کنید
tcping www.example.com 443 --no-color      # ساده، بدون رنگ ANSI
tcping www.example.com 443 --csv out.csv   # فایل CSV
tcping www.example.com 443 --sqlite out.db # پایگاه داده‌ی sqlite3
```

تصویر Docker همین هدف‌ها و پارامترها را می‌پذیرد:

```bash
docker run -it pouriyajamshidi/tcping:latest example.com 443
# یا از رجیستری کانتینر GitHub
docker run -it ghcr.io/pouriyajamshidi/tcping:latest example.com:443
```

> [!TIP]
> در حین اجرای برنامه کلید `Enter` را بفشارید تا خلاصه‌ی همه‌ی probeها را بدون متوقف کردن برنامه ببینید، همان‌طور که در بخش [دموها](#دموها) نشان داده شده است.

> [!NOTE]
> برای استفاده‌ی پیشرفته‌تر، **پارامترهای موجود** را [اینجا](#پارامترها) ببینید.

### بررسی از طریق HTTP(S)

به جای میزبان و پورت یک URL بدهید تا tcping آن را از طریق HTTP(S) بررسی کند و
کد وضعیت و مدت پاسخ دادن سرور را گزارش دهد. بدنه‌ی پاسخ
دانلود نمی‌شود، پس صفحه‌ی بزرگ probe را کند نمی‌کند:

```bash
tcping https://www.example.com/health
```

پورت از روی scheme تعیین می‌شود، `80` یا `443`، مگر اینکه URL پورت خودش را داشته باشد
یا شما بعد از آن پورتی بدهید:

```bash
tcping http://www.example.com:8080/health
# همان نتیجه
tcping http://www.example.com/health 8080
```

`-v` همه‌ی آنچه یک probe جمع کرده را نشان می‌دهد: نسخه‌ی HTTP، نسخه‌ی TLS و
cipher، چند روز تا انقضای گواهی مانده و زمان‌های اتصال، دست‌دهی TLS
و اولین بایت. `--insecure` بررسی گواهی را رد می‌کند،
که برای گواهی خودامضا یا منقضی همان چیزی است که می‌خواهید:

```bash
tcping https://www.example.com/health -v
tcping https://self-signed.example.com --insecure
```

### بررسی از طریق proxy

`--proxy` probeهای TCP و HTTP(S) را از طریق یک proxy می‌فرستد. proxyهای
`http://`، `https://` و `socks5://` کار می‌کنند، و نام کاربری و رمز در خود URL
می‌آیند، مثلاً `socks5://user:password@127.0.0.1:1080`. tcping متغیرهای محیطی
`HTTPS_PROXY`، `HTTP_PROXY` و `NO_PROXY` را نمی‌خواند، پس یک probe فقط وقتی از
proxy عبور می‌کند که خودتان بخواهید:

```bash
tcping https://www.example.com --proxy http://proxy.example.com:3128
tcping db.example.com 5432 --proxy socks5://127.0.0.1:1080
```

یک probe از نوع TCP از proxy می‌خواهد که به هدف وصل شود، و وقتی موفق است که
proxy بگوید وصل شده. پس زمان آن هم مسیر تا proxy را در بر می‌گیرد و هم مسیر خود
proxy تا هدف را. بسیاری از proxyهای HTTP فقط اتصال به پورت `443` را اجازه
می‌دهند، پس probe به پورتی دیگر ممکن است شکست بخورد در حالی که هدف در دسترس است.
برای یک proxy از نوع SOCKS5 به شبکه‌ای که با SSH به آن دسترسی دارید،
`ssh -D 1080 bastion.example.com` یکی روی `127.0.0.1:1080` باز می‌کند.

خود proxy به نام میزبان هدف وصل می‌شود، پس آدرس مبدأ، و زمان اتصال یک probe از
نوع HTTP(S)، مربوط به اتصال تا proxy هستند، و خط اول اجرا نام آن را می‌آورد.
tcping نام میزبان را خودش resolve نمی‌کند، پس خروجی آدرسی برای آن نشان نمی‌دهد،
و `-r` و `--resolve-every-probe` با proxy کار نمی‌کنند. `-4`، `-6` و
`--dns-server` به جای آن روی نام میزبان خود proxy اعمال می‌شوند. probeهای UDP
از proxy عبور نمی‌کنند. proxyای که در یک فایل PAC تنظیم شده باید دستی داده شود،
چون tcping فایل‌های PAC را نمی‌خواند.

### بررسی از طریق UDP

```bash
tcping udp://127.0.0.1 53
```

UDP دست‌دهی ندارد، پس یک probe فقط وقتی موفق است که طرف مقابل پاسخ دهد.
برای گرفتن پاسخ، tcping را در طرف مقابل به‌عنوان سرور UDP اجرا کنید که
هر datagram را به فرستنده‌اش برمی‌گرداند:

```bash
# روی دستگاهی که بررسی می‌شود:
tcping --udp-server 127.0.0.1 9999
# روی دستگاهی که بررسی می‌کند:
tcping udp://127.0.0.1 9999
```

> [!NOTE]
> یک probe‌ی UDP که رد شود `port unreachable` گزارش می‌دهد، یعنی
> چیزی جلوی ما را گرفته است. probe‌ای که اصلاً پاسخی نگیرد هم به‌عنوان
> شکست گزارش می‌شود، اما نمی‌تواند پورت بازی که ساکت مانده را از
> بسته‌ای که در راه حذف شده تشخیص دهد.
>
> روی ویندوز، probe‌ی ردشده همیشه به‌صورت «بدون پاسخ» دیده می‌شود. Go خطاهای
> "port unreachable" را روی سوکت‌های UDP ویندوز خاموش می‌کند، بنابراین tcping هرگز آن‌ها را نمی‌بیند.

هر probe‌ی UDP شماره‌ی خودش را به‌عنوان payload می‌فرستد و سرور آن را برمی‌گرداند.
با افزودن `-v` این شماره در هر خط نشان داده می‌شود، پس می‌توان probe گم‌شده را نام برد:

```text
Reply from 127.0.0.1 on port 9999 UDP_conn=4 time=1.276 ms
    reply echoed back probe 4
```

### ارسال زنده‌ی خروجی JSON به یک سرور

خروجی `JSON` لازم نیست چاپ شود. به `--json-url` یک آدرس بدهید تا
هر رویداد همان لحظه با `POST` به آنجا فرستاده شود، یک رویداد در هر درخواست، تا دستگاهی
که بررسی می‌کند بتواند به دستگاهی که جمع‌آوری می‌کند گزارش دهد:

```bash
tcping www.example.com 443 --json-url http://localhost:8000/tcping
```

رویدادها و فیلدهایشان دقیقاً همان‌هایی هستند که `-j` چاپ می‌کند، پس هر چیزی که
خروجی لوله‌شده را می‌خواند این‌ها را هم می‌خواند. هر رویداد ارسالی همچنین
دستگاه فرستنده را در فیلد `source` نام می‌برد که به‌طور پیش‌فرض نام میزبان
همان دستگاه است، تا collectorی که از چند دستگاه رویداد می‌گیرد بتواند بفهمد
اجرای چه کسی را می‌بیند. با `--source-label` خودتان نام بگذارید:

```bash
tcping www.example.com 443 --json-url http://collector:8000/tcping --source-label brussels
```

tcping می‌تواند سمت دریافت‌کننده هم باشد، همان‌طور که `--udp-server` به probeهای
UDP پاسخ می‌دهد. `--json-server` بررسی نمی‌کند: روی میزبان و پورت داده‌شده
گوش می‌دهد و هر رویدادی را که به آن فرستاده شود، روی هر مسیری، چاپ می‌کند، پس دستگاهی
که جمع‌آوری می‌کند هم به نصب چیزی نیاز ندارد:

```bash
# روی دستگاهی که جمع‌آوری می‌کند:
tcping --json-server 0.0.0.0 8000

# روی دستگاهی که بررسی می‌کند:
tcping www.example.com 443 --json-url http://collector:8000/tcping
```

رویدادها به خروجی استاندارد می‌روند و بقیه‌ی چیزها به خطای استاندارد، پس
`tcping --json-server 0.0.0.0 8000 > events.jsonl` فایلی می‌سازد که فقط
رویداد دارد، و هر چیزی که `JSON` نباشد به‌جای چاپ شدن رد می‌شود.

هر سروری که یک `POST` با بدنه‌ی `JSON` بپذیرد به همین خوبی کار می‌کند، مثلاً:

```python
from fastapi import FastAPI

app = FastAPI()


@app.post("/tcping")
async def tcping(event: dict) -> None:
    print(event)
```

همه‌ی رویدادهای یک اجرا از یک اتصال می‌روند که بین probeها باز نگه داشته می‌شود
و برای هر کدام دوباره برقرار نمی‌شود. سروری که خاموش است یا مشکل دارد
جلوی probe کردن را نمی‌گیرد: tcping یک بار اعلام می‌کند و ادامه می‌دهد و
رویدادهایی را که نمی‌تواند تحویل دهد دور می‌اندازد.

### ارسال نتایج از طریق OTLP یا به InfluxDB

tcping به جای چاپ هر probe می‌تواند آن را به‌شکل metric بفرستد، که یک
اجرا را به نمودار تبدیل می‌کند و اجازه می‌دهد چند دستگاه یک هدف را زیر نظر بگیرند. از طریق OTLP،
به هر چیزی که آن را بپذیرد، مثل [Prometheus](https://prometheus.io/)
که با `--web.enable-otlp-receiver` راه‌اندازی شده باشد:

```bash
tcping www.example.com 443 --otlp http://localhost:9090/api/v1/otlp
```

یا [Grafana Alloy](https://grafana.com/docs/alloy/latest/) و
OpenTelemetry Collector که می‌توانند آن را به چند جا ارسال کنند:

```bash
tcping www.example.com 443 --otlp http://localhost:4318
```

بک‌اندهای میزبانی‌شده توکن می‌خواهند، هر کدام در یک header مخصوص. کل
header را با `--otlp-header` بدهید، یا در متغیر محیطی `OTLP_HEADER` تا
از تاریخچه‌ی شل شما دور بماند:

```bash
export OTLP_HEADER="Authorization: Bearer your-api-token" && \
  tcping www.example.com 443 --otlp https://otlp.example.com
```

یا مستقیم به یک سرور [InfluxDB](https://www.influxdata.com/) نسخه‌ی ۲ یا ۳ به‌صورت
line protocol:

```bash
export INFLUXDB_TOKEN=your-api-token && \
  tcping www.example.com 443 --influxdb http://localhost:8086 --influxdb-org home --influxdb-bucket tcping
```

هر probe زمان رفت‌وبرگشت، موفق بودن یا نبودن، آدرسی که هدف به آن resolve شده
و شمارنده‌ها را با خود دارد، و اگر هدف HTTP یا UDP باشد جزئیات آن‌ها را هم.
کل بلوک آماری که معمولاً هنگام خروج می‌بینید علاوه بر این
هر ۱۰ ثانیه فرستاده می‌شود، پس اجرایی که کسی
تماشایش نمی‌کند هم آن را گزارش می‌دهد. برای تغییر این فاصله از `--stats-interval` استفاده کنید.

هر دستگاه در metricها با برچسب `source` خودش را نام می‌برد که به‌طور پیش‌فرض
نام میزبانش است، پس چند دستگاهی که یک هدف را بررسی می‌کنند در
سری‌های جداگانه‌ی خودشان می‌نشینند و روی هم نمی‌افتند. با `--source-label` خودتان
نام بگذارید:

```bash
tcping www.example.com 443 --otlp http://localhost:4318 --source-label brussels
```

هر آنچه فرستاده می‌شود، روش پرس‌وجو از آن، و محیط‌های آماده‌ی Prometheus، Alloy،
OpenTelemetry Collector و InfluxDB با Grafana و tcping که می‌توانید
با یک دستور راه بیندازید تا بدون راه‌اندازی سرور از قبل این را امتحان کنید، در
[docs/observability](../observability/README.md) هستند.

---

## پارامترها

پارامترها را می‌توان قبل یا بعد از هدف داد، با یک یا دو
خط تیره، پس `-c 5` و `--c 5` یک پارامتر هستند.

### عمومی

| پارامتر     | پیش‌فرض | توضیح                         |
| ----------- | ------- | ----------------------------- |
| `-h`        |         | نمایش پارامترهای موجود و خروج |
| `--version` |         | نمایش نسخه و خروج             |
| `-u`        |         | بررسی به‌روزرسانی و خروج      |

### بررسی

| پارامتر         | پیش‌فرض | توضیح                                                                           |
| --------------- | ------- | ------------------------------------------------------------------------------- |
| `-c <n>`        | بدون حد | پس از `<n>` probe، فارغ از نتیجه متوقف شو                                       |
| `-i <seconds>`  | `1`     | فاصله‌ی بین probeها. عدد اعشاری مجاز است، مثلاً `-i 0.5`                        |
| `-t <seconds>`  | `1`     | زمان انتظار برای پاسخ، به ثانیه. عدد اعشاری مجاز است. `0` یعنی timeout بی‌نهایت |
| `-4`            |         | فقط از آدرس‌های IPv4 استفاده کن                                                 |
| `-6`            |         | فقط از آدرس‌های IPv6 استفاده کن                                                 |
| `-I <name\|IP>` |         | نام رابط یا آدرس IP برای فرستادن probeها و جستجوهای DNS                         |
| `--proxy <URL>` |         | probeهای TCP و `HTTP(S)` را از طریق یک proxy بفرست، مثلاً `--proxy http://proxy.example.com:3128`. متغیرهای محیطی `HTTPS_PROXY`، `HTTP_PROXY` و `NO_PROXY` به کار نمی‌روند |

> [!TIP]
> بدون پارامترهای `-4` و `-6`، tcping بر پایه‌ی جستجوهای DNS یک آدرس IP را تصادفی انتخاب می‌کند.

### Name resolution

| پارامتر                   | پیش‌فرض      | توضیح                                                                                                                  |
| ------------------------- | ------------ | ---------------------------------------------------------------------------------------------------------------------- |
| `-r <n>`                  | هرگز         | پس از `<n>` probe‌ی ناموفق نام میزبان هدف را دوباره resolve کن، مثلاً `-r 10`                                          |
| `--resolve-every-probe`   |              | نام میزبان هدف را پیش از هر probe resolve کن، نه فقط هنگام شروع. بر `-r` اولویت دارد و اگر هدف آدرس IP باشد اثری ندارد |
| `--dns-server <IP>`       | سراسری سیستم | سرور DNS سفارشی. ترکیب IP و پورت مجاز است، مثلاً `--dns-server 1.1.1.1:53`                                             |
| `--dns-timeout <seconds>` | `2`          | زمان انتظار برای پاسخ DNS، به ثانیه. عدد اعشاری مجاز است. `0` یعنی timeout بی‌نهایت                                    |

### خروجی ترمینال

| پارامتر                 | پیش‌فرض | توضیح                                                                                                                                                                                                                                                                       |
| ----------------------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-D`                    |         | برای هر probe زمان نشان بده                                                                                                                                                                                                                                                 |
| `--no-color`            |         | خروجی را رنگی نکن                                                                                                                                                                                                                                                           |
| `--show-source-address` |         | آدرس IP مبدأ و پورتی را که برای probeها استفاده شده نشان بده                                                                                                                                                                                                                |
| `--failures-only`       |         | فقط probeهای ناموفق را نشان بده. probeهای موفق همچنان شمرده می‌شوند                                                                                                                                                                                                         |
| `--no-stats`            |         | هنگام خروج برنامه آمار را نشان نده. فشردن کلید **Enter** همچنان آن را نشان می‌دهد. وقتی خروجی به جایی غیر از ترمینال می‌رود اثری ندارد                                                                                                                                      |
| `-v`                    |         | همه‌ی آنچه یک probe‌ی HTTP(S) جمع کرده را نشان بده: نسخه‌ی HTTP، نسخه‌ی TLS و cipher، انقضای گواهی و زمان‌های اتصال، TLS و اولین بایت. برای هدف UDP شماره‌ی probe و اینکه پاسخ آن را برگردانده یا نه را نشان می‌دهد تا probe گم‌شده از بقیه جدا شود. روی هدف TCP اثری ندارد |

### خروجی فایل و قابل خواندن برای ماشین

| پارامتر            | پیش‌فرض | توضیح                                                                                                                                                                                                                   |
| ------------------ | ------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-j`               |         | خروجی با قالب `JSON`                                                                                                                                                                                                    |
| `--pretty`         |         | خروجی `JSON` را خوانا کن. بدون `-j` اثری ندارد                                                                                                                                                                          |
| `--json-url <URL>` |         | خروجی `JSON` را به جای چاپ به یک سرور HTTP بفرست، یک `POST` برای هر رویداد، مثلاً `--json-url http://localhost:8000/tcping`. خودش خروجی `JSON` را روشن می‌کند، پس `-j` لازم نیست                                        |
| `--json-server`    |         | بررسی نکن. روی میزبان و پورت داده‌شده گوش بده و هر رویداد `JSON` ارسالی به آن را چاپ کن، تا tcpingی که جای دیگری از `--json-url` استفاده می‌کند مقصدی برای اجرایش داشته باشد، مثلاً `tcping --json-server 0.0.0.0 8000` |
| `--csv <file>`     |         | خروجی را در یک فایل `CSV` ذخیره کن. آمار در همان نام با پسوند `_stats` ذخیره می‌شود                                                                                                                                     |
| `--csv-fixed-name` |         | نام فایل `--csv` را همان‌طور که هست و بدون پسوند تاریخ/زمان به کار ببر، پس اجراهای تکراری روی همان فایل می‌نویسند                                                                                                       |
| `--sqlite <file>`  |         | خروجی را در یک پایگاه داده‌ی sqlite3 ذخیره کن، مثلاً `--sqlite /tmp/tcping.db`. روی ویندوز موجود نیست                                                                                                                   |

### Metrics

| پارامتر                      | پیش‌فرض    | توضیح                                                                                                                                                                                                                                                                                                      |
| ---------------------------- | ---------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--otlp <URL>`               |            | نتایج را به جای چاپ، به‌شکل metric به یک endpoint از نوع OTLP HTTP مثل [Grafana Alloy](https://grafana.com/docs/alloy/latest/) یا OpenTelemetry Collector بفرست، مثلاً `--otlp http://localhost:4318`                                                                                                      |
| `--otlp-header <header>`     |            | header اضافه‌ی HTTP برای فرستادن به endpoint مربوط به `--otlp`، به‌شکل `"Name: value"`، مثلاً `"Authorization: Bearer <token>"`. می‌توان آن را در متغیر محیطی `OTLP_HEADER` هم داد که از تاریخچه‌ی شل شما دور می‌ماند                                                                                      |
| `--influxdb <URL>`           |            | نتایج را به جای چاپ، به‌صورت line protocol روی یک سرور [InfluxDB](https://www.influxdata.com/) نسخه‌ی ۲ یا ۳ بنویس، مثلاً `--influxdb http://localhost:8086`                                                                                                                                               |
| `--influxdb-org <org>`       |            | سازمان InfluxDB که باید در آن نوشت. با `--influxdb` الزامی است                                                                                                                                                                                                                                             |
| `--influxdb-bucket <bucket>` |            | bucket در InfluxDB که باید در آن نوشت. با `--influxdb` الزامی است                                                                                                                                                                                                                                          |
| `--influxdb-token <token>`   |            | توکن API در InfluxDB. با `--influxdb` الزامی است. می‌توان آن را در متغیر محیطی `INFLUXDB_TOKEN` هم داد که از تاریخچه‌ی شل شما دور می‌ماند                                                                                                                                                                  |
| `--stats-interval <seconds>` | `10`       | هر چند وقت آمار به endpoint مربوط به OTLP یا InfluxDB فرستاده شود. بدون `--otlp` یا `--influxdb` اثری ندارد                                                                                                                                                                                                |
| `--source-label <name>`      | نام میزبان | این دستگاه را در رویدادهای `JSON` و در metricهایی که از طریق OTLP یا به InfluxDB فرستاده می‌شوند نام ببر، تا چند دستگاهی که یک هدف را بررسی می‌کنند قابل تشخیص باشند. نتایجی که به جای دیگری فرستاده می‌شوند در صورت ندادن این پارامتر نام میزبان را دارند؛ خروجی‌ای که روی همان دستگاه می‌ماند چیزی ندارد |

### HTTP(S) و UDP

| پارامتر         | پیش‌فرض | توضیح                                                                                                                                                                       |
| --------------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--insecure`    |         | هنگام بررسی هدف `https://` گواهی سرور را تأیید نکن. برای گواهی‌های خودامضا یا منقضی مفید است                                                                                |
| `--udp-server`  |         | بررسی نکن. روی میزبان و پورت داده‌شده گوش بده و هر datagram‌ی UDP را به فرستنده‌اش برگردان، تا probe‌ی UDP که به این دستگاه می‌رسد پاسخ بگیرد                               |

---

## مشارکت

pull requestها برای رفع باگ، افزودن ویژگی‌های جدید و کمک به
issueهای بازی که [اینجا](https://github.com/pouriyajamshidi/tcping/issues) هستند خوش‌آمد هستند.
تعداد فعلی issueهای باز: ![GitHub issues](https://img.shields.io/github/issues/pouriyajamshidi/tcping.svg).

1. هر issueی را که با آن راحت هستید انتخاب کنید.
1. مخزن را Fork کنید.
1. یک شاخه بسازید.
1. کارتان را commit کنید.
1. تست اضافه کنید.
1. تست‌ها را با `make test` اجرا کنید و مطمئن شوید موفق هستند.
1. یک pull request بسازید.

لطفاً مطمئن شوید pull request شما **فقط یک issue یا ویژگی‌ی مشخص را پوشش می‌دهد** و دو یا چند مورد را با هم انجام نمی‌دهد. این کار بررسی pull request شما را برای ما آسان‌تر می‌کند و به تمیز ماندن تاریخچه‌ی git کمک می‌کند.

مگر اینکه مشکل واقعاً کوچکی را رفع می‌کنید، لطفاً پیش از شروع کار، ابتدا قصدتان را در یک **issue** اعلام کنید.

برای امتحان تغییرات خود روی یک شبکه‌ی بد، `tools/netcond.sh` می‌تواند تأخیر اضافه کند،
بخشی از بسته‌ها را دور بیندازد یا یک مقصد را کاملاً مسدود کند، و پس از آن
هر آنچه اضافه کرده را برگرداند:

```bash
sudo ./tools/netcond.sh delay 1.1.1.1 800ms 200ms 30
sudo ./tools/netcond.sh loss 1.1.1.1 40
sudo ./tools/netcond.sh clear
```

## حمایت از پروژه

اگر tcping برای شما مفید است، آن را با اطرافیان خود به اشتراک بگذارید تا دامنه‌اش گسترده‌تر شود و دیگران هم بتوانند از آن بهره ببرند.

همچنین می‌توانید با پیوندهای زیر از پروژه حمایت کنید:

- یک قهوه برایم بخرید: [![Buy Me A Coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/pouriyajamshidi)
- GitHub Sponsors: [![Sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/pouriyajamshidi)

## مجوز

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
