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

Een platformonafhankelijk ping-programma dat `TCP`, `UDP` of `HTTP(S)` gebruikt in plaats van `ICMP`, geïnspireerd door het ping-hulpprogramma van Linux.

> [!TIP]
> Dit document is ook beschikbaar in het [English](../../README.md), [中文](README.cn.md), [日本語](README.ja.md), [한국어](README.ko.md), [فارسی](README.fa.md), [Deutsch](README.de.md) en [Français](README.fr.md).

- Een alternatief voor `ping` waar `ICMP` geblokkeerd is: het test via `TCP`, `HTTP(S)` of `UDP`.
- Meldt het pakketverlies en het minimum, het gemiddelde, het maximum en de standaardafwijking van de latentie, zoals `ping` dat doet, plus de langste uptime en downtime en wanneer die optraden.
- Toont de statistieken op elk moment als u op de `Enter`-toets drukt, zonder het programma te stoppen.
- Geeft uitvoer in **kleur**, **platte tekst**, **JSON**, **CSV** of **sqlite3**, of stuurt elke probe als metrics via **OTLP** (Grafana Alloy, de OpenTelemetry Collector en andere) of naar **InfluxDB**.
- Toont de statuscode, de TLS-versie en cipher, de vervaldatum van het certificaat en de tijden voor verbinden, TLS-handshake en eerste byte van elke `HTTP(S)`-probe.
- Zoekt de hostnaam van het doel opnieuw op na een aantal mislukkingen (`-r`) of vóór elke probe (`--resolve-every-probe`) en meldt hoe lang elke opzoeking duurde. Handig om uw `DNS`-loadbalancing of Global Server Load Balancer `(GSLB)` te testen.
- Laat u de **bronnetwerkinterface**, de **timeout**, het **interval** en de **DNS-server** kiezen en `IPv4` of `IPv6` afdwingen.
- Nummert _geslaagde_ en _mislukte_ probes apart, zodat de totalen in één oogopslag zichtbaar zijn.

Bekijk de [demo's](#demos) voor een indruk van het uiterlijk en het gebruik van **tcping**.

---

## Inhoudsopgave

- [Demo's](#demos)
- [Downloaden en installeren](#downloaden-en-installeren)
- [Gebruik](#gebruik)
- [Flags](#flags)
- [Bijdragen](#bijdragen)
- [Het project helpen](#het-project-helpen)
- [Licentie](#licentie)

---

## Demos

<details>
<summary>Klik om uit te klappen</summary>

### Basisgebruik

![tcping](../Images/gifs/tcping.gif)

---

### Probes met tijdstempel via IPv4 (`-D -4`) flags

![tcping timestamp example](../Images/gifs/tcping_timestamp.gif)

---

### Platte uitvoer, statistieken alleen bij Enter (`--no-color --no-stats`) flags

![tcping plain output example](../Images/gifs/tcping_plain.gif)

---

### Hostnaam opnieuw opzoeken (`-r`) flag

![tcping resolve example](../Images/gifs/tcping_resolve.gif)

---

### JSON-uitvoer (`-j --pretty`) flag

![tcping json example](../Images/gifs/tcping_json_pretty.gif)

---

### De JSON-uitvoer naar een collector streamen (`--json-url --json-server`) flags

![tcping JSON streaming example](../Images/gifs/tcping_json_stream.gif)

---

### Tijd voor het opzoeken van de hostnaam (`--resolve-every-probe`) flag

![tcping resolution timing example](../Images/gifs/tcping_dns_timing.gif)

---

### Bronnetwerkinterface (`-I`) flag

![tcping interface example](../Images/gifs/tcping_interface.gif)

---

### HTTP(S)-probes

![tcping HTTP example](../Images/gifs/tcping_http.gif)

---

### Details van HTTP(S)-probes (`-v`) flag

![tcping HTTP verbose example](../Images/gifs/tcping_http_verbose.gif)

---

### Certificaatcontrole overslaan (`--insecure`) flag

![tcping skip TLS example](../Images/gifs/tcping_skip_tls.gif)

---

### UDP-probes tegen een UDP-server (`--udp-server`) flag

![tcping UDP example](../Images/gifs/tcping_udp.gif)

---

### Alleen de mislukte probes (`--failures-only`) flag

![tcping failures only example](../Images/gifs/tcping_failures_only.gif)

</details>

---

## Downloaden en installeren

| Platform                                        | Installeren met                                                                                                    |
| ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| [Windows](#windows)                             | `winget install pj.tcping`                                                                                         |
| [macOS](#macos)                                 | `brew install pouriyajamshidi/tap/tcping`                                                                          |
| [Debian, Ubuntu](#linux---pakketbronnen)        | `sudo apt install tcping`, nadat u de pakketbron hebt toegevoegd                                                   |
| [Fedora, RHEL, CentOS](#linux---pakketbronnen)  | `sudo dnf install tcping`, nadat u de pakketbron hebt toegevoegd                                                   |
| [Alpine](#linux---pakketbronnen)                | `sudo apk add tcping`, nadat u de pakketbron hebt toegevoegd                                                       |
| [Arch, Manjaro](#linux---pakketbronnen)         | `yay -S tcping-bin`                                                                                                |
| [Elke Linux](#linux---installatie-in-één-regel) | een [installatie in één regel](#linux---installatie-in-één-regel) of een [pakketbestand](#linux---pakketbestanden) |
| [BSD](#bsd-en-linux---handmatig)                | het [vooraf gebouwde bestand](#bsd-en-linux---handmatig)                                                           |
| [Nix](#nix)                                     | `nix profile install github:pouriyajamshidi/tcping`                                                                |
| [Docker](#andere-manieren)                      | `docker run -it pouriyajamshidi/tcping example.com 443`                                                            |
| [Go](#andere-manieren)                          | `go install github.com/pouriyajamshidi/tcping/v3@latest`                                                           |

De uitvoerbare bestanden zijn statisch: ze bevatten alles wat ze nodig hebben in één bestand
en zijn niet afhankelijk van een bibliotheek op uw computer. Ze staan op de
[releasepagina](https://github.com/pouriyajamshidi/tcping/releases/latest/), voor
_amd64_ en _arm64_.

Als de installatie klaar is, ga dan naar het onderdeel [gebruik](#gebruik).

### Windows

De beste manier om **tcping** op Windows te installeren is via de _Windows Package Manager_ met [WinGet](https://learn.microsoft.com/nl-nl/windows/package-manager/winget/?ref=github.com%2Fpouriyajamshidi%2Ftcping), die sinds september 2020 standaard op vrijwel alle Windows-_10_- en -_11_-computers aanwezig is:

```powershell
winget install pj.tcping
```

Om het met de hand te installeren, pakt u het gedownloade zipbestand uit en kopieert u `tcping.exe` naar het [PATH](https://www.howtogeek.com/118594/how-to-edit-your-system-path-for-easy-command-line-access/) van uw systeem, bijvoorbeeld `C:\Windows\System32`.

> [!CAUTION]
> TCPING kan ten onrechte worden gemeld door Windows Defender of andere antimalwaresoftware. Dit komt vaak voor bij Go-programma's. Bekijk de officiële verklaring van het Go-team [hier](https://go.dev/doc/faq#virus).

> [!WARNING]
> Het uitvoerformaat `--sqlite` (sqlite3) is niet meer beschikbaar in de Windows-bestanden. Alle andere flags werken zoals verwacht.

### macOS

```bash
brew install pouriyajamshidi/tap/tcping
```

### Linux - Installatie in één regel

Plak het volgende in uw terminal om het nieuwste bestand voor uw architectuur op te halen en te installeren:

```bash
cd /tmp &&
ARCH=$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/') &&
curl -LO "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
tar -xf "tcping-linux-$ARCH.tar.gz" &&
sudo install tcping /usr/local/bin/ &&
tcping --version
```

Zie [Shell-aanvulling](#shell-aanvulling) om aanvullen met de Tab-toets in te stellen.

Als u geen `curl` heeft, vervang dan die regel door `wget`:

```bash
wget "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
```

### Linux - Pakketbronnen

Als u de pakketbron één keer toevoegt, regelt `apt`, `dnf` of `apk` vanaf dan de installatie en
updates, zoals bij elk ander pakket.

**Debian**, **Ubuntu** en hun varianten:

```bash
sudo install -d /etc/apt/keyrings &&
  sudo curl -fsSL https://pouriyajamshidi.github.io/packages/keys/pouriyajamshidi.gpg -o /etc/apt/keyrings/pouriyajamshidi.gpg &&
  echo "deb [signed-by=/etc/apt/keyrings/pouriyajamshidi.gpg] https://pouriyajamshidi.github.io/packages/deb ./" | sudo tee /etc/apt/sources.list.d/pouriyajamshidi.list &&
  sudo apt update &&
  sudo apt install tcping
```

**Fedora**, **RHEL**, **CentOS** en hun varianten:

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

**Arch**, **Manjaro**, **EndeavourOS** en hun varianten vinden het in de
[AUR](https://aur.archlinux.org/packages/tcping-bin), dus gebruik uw favoriete helper:

```bash
yay -S tcping-bin
```

### Linux - Pakketbestanden

Zonder een pakketbron toe te voegen, downloadt u het pakket voor uw distributie van de
[releasepagina](https://github.com/pouriyajamshidi/tcping/releases/latest/) en
installeert u het:

| Distributie          | Pakket                     | Installeren                                         |
| -------------------- | -------------------------- | --------------------------------------------------- |
| Debian, Ubuntu       | `tcping-amd64.deb`         | `sudo apt install -y ./tcping-amd64.deb`            |
| Fedora, RHEL, CentOS | `tcping-amd64.rpm`         | `sudo dnf install -y ./tcping-amd64.rpm`            |
| Arch, Manjaro        | `tcping-amd64.pkg.tar.zst` | `sudo pacman -U ./tcping-amd64.pkg.tar.zst`         |
| Alpine               | `tcping-amd64.apk`         | `sudo apk add --allow-untrusted ./tcping-amd64.apk` |

Vervang `amd64` door `arm64` op ARM-computers. Alle vier installeren het uitvoerbare bestand en de
[shell-aanvulling](#shell-aanvulling) voor u. Computers zonder `dnf` kunnen
`yum` gebruiken, en Alpine heeft `--allow-untrusted` nodig omdat onze pakketten niet
met een Alpine-sleutel zijn ondertekend.

### BSD en Linux - Handmatig

Download het bestand voor uw besturingssysteem en architectuur:

```bash
wget https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-freebsd-amd64.tar.gz
# Of voor Linux-ARM64-computers en met cURL
curl -LO https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-arm64.tar.gz
```

Pak het uit en kopieer het uitvoerbare bestand naar het `PATH` van uw systeem, bijvoorbeeld `/usr/local/bin/`:

```bash
tar -xvf tcping-freebsd-amd64.tar.gz &&
  chmod +x tcping &&
  sudo cp tcping /usr/local/bin/
```

Zie [Shell-aanvulling](#shell-aanvulling) om aanvullen met de Tab-toets in te stellen.

### Nix

**tcping** levert een flake mee, dus op elke computer met [Nix](https://nixos.org/download/)
kunt u het draaien zonder iets te installeren:

```bash
nix run github:pouriyajamshidi/tcping -- example.com 80
```

Of installeer het blijvend:

```bash
nix profile install github:pouriyajamshidi/tcping
```

Beide geven u het uitvoerbare bestand en de [shell-aanvulling](#shell-aanvulling). De
flake heeft ook een `devShells.default` met alles wat nodig is om aan tcping te werken,
die u krijgt met `nix develop`.

### Andere manieren

- `Docker`-images:

  ```bash
  docker pull pouriyajamshidi/tcping:latest
  # Of
  docker pull ghcr.io/pouriyajamshidi/tcping:latest
  ```

- Met `go install`:

  > Hiervoor is minstens go-versie `1.26.8` nodig

  ```bash
  go install github.com/pouriyajamshidi/tcping/v3@latest
  ```

- [x tcping](https://x-cmd.com/pkg/tcping):

  **Direct zonder installatie** in [x-cmd](https://www.x-cmd.com).

  ```bash
  x tcping example.com 80
  ```

  Of installeer `tcping` lokaal met x-cmd, zonder rootrechten en zonder uw globale instellingen aan te raken.

  ```bash
  x env use tcping
  tcping example.com 80
  ```

- Ten slotte kunt u de code zelf compileren met het commando `make`:

  ```bash
  make build
  ```

  Hiermee komt het uitvoerbare bestand in de map `target` te staan.

### Shell-aanvulling

**tcping** heeft aanvullingsscripts voor `bash`, `zsh`, `fish` en `PowerShell`
en toont ze met `--completions <shell>`. Ze vullen de flags aan, de
interfacenamen voor `-I` en de bestandsnamen voor `--csv` en `--sqlite`.

De Linux-pakketten, Homebrew, het AUR-pakket, Nix en de
[installatie in één regel](#linux---installatie-in-één-regel) zetten ze voor u op hun plek.
Anders slaat u het script voor uw shell op:

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

- `PowerShell`, door het te laden vanuit uw profiel:

  ```powershell
  Add-Content $PROFILE "tcping --completions powershell | Out-String | Invoke-Expression"
  ```

Start daarna een nieuwe shell om ze te activeren.

---

## Gebruik

Geef **tcping** een doel en een poort, en het begint te testen:

```bash
tcping www.example.com 443
tcping www.example.com:443          # host:poort werkt ook
tcping 192.168.1.1:80
tcping '[2001:db8::1]:443'          # tussen aanhalingstekens, zodat de shell er afblijft
tcping https://www.example.com      # getest via HTTP(S)
tcping udp://127.0.0.1 53           # getest via UDP
```

Enkele dingen die u van een run kunt vragen:

```bash
tcping www.example.com 443 -c 5            # stoppen na 5 probes
tcping www.example.com 443 -i 2 -t 5       # 2 seconden tussen de probes, 5 seconden timeout
tcping www.example.com 443 -I eth2         # de probes vanaf een bepaalde interface sturen
tcping www.example.com 443 -4              # of -6, om een adresfamilie af te dwingen
tcping www.example.com 443 -D              # een tijdstempel tonen bij elke probe
tcping www.example.com 443 -r 5            # de hostnaam na 5 mislukkingen opnieuw opzoeken
```

En de uitvoerformaten die het kan maken, in plaats van de standaard gekleurde uitvoer:

```bash
tcping www.example.com 443 -j              # JSON, voeg --pretty toe voor nettere opmaak
tcping www.example.com 443 --no-color      # platte tekst, zonder ANSI-kleuren
tcping www.example.com 443 --csv out.csv   # CSV-bestand
tcping www.example.com 443 --sqlite out.db # sqlite3-database
```

Het Docker-image accepteert dezelfde doelen en flags:

```bash
docker run -it pouriyajamshidi/tcping:latest example.com 443
# Of, via het GitHub-containerregister
docker run -it ghcr.io/pouriyajamshidi/tcping:latest example.com:443
```

> [!TIP]
> Druk op de `Enter`-toets terwijl het programma draait om de samenvatting van alle probes te zien zonder het programma te stoppen, zoals getoond in het onderdeel [demo's](#demos).

> [!NOTE]
> Bekijk de **beschikbare flags** [hier](#flags) voor geavanceerder gebruik.

### Testen via HTTP(S)

Geef een URL in plaats van een host en poort, en tcping test die via HTTP(S)
en meldt de statuscode en hoe lang de server over het antwoord deed. De body
wordt niet gedownload, dus een grote pagina vertraagt de probe niet:

```bash
tcping https://www.example.com/health
```

De poort komt uit het schema, `80` of `443`, tenzij de URL een eigen poort bevat
of u er een achter opgeeft:

```bash
tcping http://www.example.com:8080/health
# Hetzelfde
tcping http://www.example.com/health 8080
```

`-v` toont alles wat een probe heeft verzameld: de HTTP-versie, de TLS-versie en
cipher, hoeveel dagen het certificaat nog geldig is en de tijden voor verbinden, TLS-handshake
en eerste byte. `--insecure` slaat de certificaatcontrole over,
wat u wilt bij een zelfondertekend of verlopen certificaat:

```bash
tcping https://www.example.com/health -v
tcping https://self-signed.example.com --insecure
```

### Testen via UDP

```bash
tcping udp://127.0.0.1 53
```

UDP heeft geen handshake, dus een probe slaagt alleen als de andere kant antwoordt.
Om een antwoord te krijgen, draait u tcping aan de andere kant als UDP-server, die
elk datagram terugstuurt naar de afzender:

```bash
# op de computer die getest wordt:
tcping --udp-server 127.0.0.1 9999
# op de computer die test:
tcping udp://127.0.0.1 9999
```

> [!NOTE]
> Een geweigerde UDP-probe meldt `port unreachable`, wat betekent dat
> iets ons blokkeert. Een probe die helemaal geen antwoord krijgt, wordt ook als
> mislukking gemeld, maar kan een open poort die stil blijft niet onderscheiden van een
> pakket dat onderweg is weggevallen.
>
> Op Windows verschijnt een geweigerde probe altijd als geen antwoord. Go zet de
> "port unreachable"-fouten uit op UDP-sockets van Windows, dus tcping ziet ze nooit.

Elke UDP-probe stuurt zijn eigen nummer als payload mee, dat de server terugstuurt.
Met `-v` verschijnt dat nummer op elke regel, zodat een verloren probe bij naam genoemd kan worden:

```text
Reply from 127.0.0.1 on port 9999 UDP_conn=4 time=1.276 ms
    reply echoed back probe 4
```

### De JSON-uitvoer naar een server streamen

De `JSON`-uitvoer hoeft niet te worden afgedrukt. Geef `--json-url` een adres en
elke gebeurtenis wordt direct erheen gestuurd met een `POST`, één gebeurtenis per verzoek, zodat een computer
die test kan rapporteren aan een computer die verzamelt:

```bash
tcping www.example.com 443 --json-url http://localhost:8000/tcping
```

De gebeurtenissen en hun velden zijn precies die welke `-j` afdrukt, dus alles wat
de doorgesluisde uitvoer al leest, leest ook deze. Elke gestreamde gebeurtenis
noemt ook de computer die hem verstuurde in het veld `source`, dat standaard de hostnaam
van die computer is, zodat een collector die gebeurtenissen van meerdere computers ontvangt kan zien
van wie de run is die hij bekijkt. Gebruik `--source-label` om ze zelf een naam te geven:

```bash
tcping www.example.com 443 --json-url http://collector:8000/tcping --source-label brussels
```

tcping kan ook de ontvangende kant zijn, zoals `--udp-server` UDP-probes
beantwoordt. `--json-server` test niet: het luistert op de opgegeven host en
poort en drukt elke gebeurtenis af die erheen wordt gestuurd, op elk pad, dus de computer
die verzamelt hoeft ook niets te installeren:

```bash
# op de computer die verzamelt:
tcping --json-server 0.0.0.0 8000

# op de computer die test:
tcping www.example.com 443 --json-url http://collector:8000/tcping
```

De gebeurtenissen gaan naar de standaarduitvoer en al het andere naar de standaardfoutuitvoer, dus
`tcping --json-server 0.0.0.0 8000 > events.jsonl` levert een bestand op met alleen
gebeurtenissen, en alles wat geen `JSON` is, wordt geweigerd in plaats van afgedrukt.

Elke server die een `POST` met een `JSON`-body accepteert werkt even goed, bijvoorbeeld:

```python
from fastapi import FastAPI

app = FastAPI()


@app.post("/tcping")
async def tcping(event: dict) -> None:
    print(event)
```

Alle gebeurtenissen van een run gaan over één verbinding, die tussen de
probes open blijft in plaats van voor elke probe opnieuw te worden opgezet. Een server die uit staat of problemen heeft
stopt het testen niet: tcping meldt het één keer en gaat door, en laat de
gebeurtenissen vallen die het niet kan afleveren.

### De resultaten versturen via OTLP of naar InfluxDB

In plaats van elke probe af te drukken kan tcping hem als metric versturen, waardoor een
run een grafiek wordt en meerdere computers hetzelfde doel kunnen bewaken. Via OTLP,
naar alles wat het accepteert, zoals [Prometheus](https://prometheus.io/)
gestart met `--web.enable-otlp-receiver`:

```bash
tcping www.example.com 443 --otlp http://localhost:9090/api/v1/otlp
```

Of [Grafana Alloy](https://grafana.com/docs/alloy/latest/) en de
OpenTelemetry Collector, die het naar meerdere plekken kunnen doorsturen:

```bash
tcping www.example.com 443 --otlp http://localhost:4318
```

Gehoste backends willen een token, elk in een eigen header. Geef de hele
header met `--otlp-header`, of in de omgevingsvariabele `OTLP_HEADER` om hem
uit uw shell-geschiedenis te houden:

```bash
export OTLP_HEADER="Authorization: Bearer your-api-token" && \
  tcping www.example.com 443 --otlp https://otlp.example.com
```

Of rechtstreeks naar een [InfluxDB](https://www.influxdata.com/)-v2- of -v3-server als
line protocol:

```bash
export INFLUXDB_TOKEN=your-api-token && \
  tcping www.example.com 443 --influxdb http://localhost:8086 --influxdb-org home --influxdb-bucket tcping
```

Elke probe bevat zijn round-triptijd, of hij slaagde, het adres waarnaar het
doel werd omgezet en de tellers, samen met de HTTP- of UDP-details als het doel
een van die twee is. Het hele statistiekenblok dat u normaal bij het afsluiten ziet,
wordt daarnaast elke 10 seconden verstuurd, zodat ook een run waar niemand
naar kijkt het meldt. Gebruik `--stats-interval` om dat interval te wijzigen.

Elke computer geeft zichzelf in de metrics een naam met het label `source`, dat
standaard zijn hostnaam is, zodat meerdere computers die hetzelfde doel testen in
hun eigen reeks terechtkomen in plaats van over elkaar heen. Gebruik `--source-label` om
ze zelf een naam te geven:

```bash
tcping www.example.com 443 --otlp http://localhost:4318 --source-label brussels
```

Alles wat wordt verstuurd, hoe u het opvraagt, en kant-en-klare Prometheus-, Alloy-,
OpenTelemetry-Collector- en InfluxDB-omgevingen, met Grafana en tcping, die u
met één commando kunt starten om dit uit te proberen zonder eerst een server op te zetten, staan in
[docs/observability](../observability/README.md).

---

## Flags

Flags kunnen voor of na het doel worden opgegeven, met één of twee
streepjes, dus `-c 5` en `--c 5` zijn dezelfde flag.

### Algemeen

| Flag        | Standaard | Beschrijving                            |
| ----------- | --------- | --------------------------------------- |
| `-h`        |           | De beschikbare flags tonen en afsluiten |
| `--version` |           | De versie tonen en afsluiten            |
| `-u`        |           | Op updates controleren en afsluiten     |

### Testen

| Flag            | Standaard   | Beschrijving                                                                                  |
| --------------- | ----------- | --------------------------------------------------------------------------------------------- |
| `-c <n>`        | geen limiet | Stoppen na `<n>` probes, ongeacht het resultaat                                               |
| `-i <seconds>`  | `1`         | Interval tussen probes. Kommagetal toegestaan, bijv. `-i 0.5`                                 |
| `-t <seconds>`  | `1`         | Wachttijd op een antwoord, in seconden. Kommagetal toegestaan. `0` betekent oneindige timeout |
| `-4`            |             | Alleen IPv4-adressen gebruiken                                                                |
| `-6`            |             | Alleen IPv6-adressen gebruiken                                                                |
| `-I <name\|IP>` |             | Interfacenaam of IP-adres om de probes en de DNS-opzoekingen vanaf te sturen                  |

> [!TIP]
> Zonder de flags `-4` en `-6` kiest tcping willekeurig een IP-adres op basis van de DNS-opzoekingen.

### Naamomzetting

| Flag                      | Standaard       | Beschrijving                                                                                                                                             |
| ------------------------- | --------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-r <n>`                  | nooit           | De hostnaam van het doel opnieuw opzoeken na `<n>` mislukte probes, bijv. `-r 10`                                                                        |
| `--resolve-every-probe`   |                 | De hostnaam van het doel vóór elke probe opzoeken in plaats van alleen bij het starten. Gaat voor `-r` en heeft geen effect als het doel een IP-adres is |
| `--dns-server <IP>`       | van het systeem | Te gebruiken eigen DNS-server. Een combinatie van IP en poort is toegestaan, bijv. `--dns-server 1.1.1.1:53`                                             |
| `--dns-timeout <seconds>` | `2`             | Wachttijd op een DNS-antwoord, in seconden. Kommagetal toegestaan. `0` betekent oneindige timeout                                                        |

### Terminaluitvoer

| Flag                    | Standaard | Beschrijving                                                                                                                                                                                                                                                                                                                              |
| ----------------------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-D`                    |           | Een tijdstempel tonen bij elke probe                                                                                                                                                                                                                                                                                                      |
| `--no-color`            |           | De uitvoer niet kleuren                                                                                                                                                                                                                                                                                                                   |
| `--show-source-address` |           | Het bron-IP-adres en de poort tonen die voor de probes zijn gebruikt                                                                                                                                                                                                                                                                      |
| `--failures-only`       |           | Alleen de mislukte probes tonen. De geslaagde worden nog steeds geteld                                                                                                                                                                                                                                                                    |
| `--no-stats`            |           | De statistieken niet tonen als het programma afsluit. Op de **Enter**-toets drukken toont ze nog steeds. Geen effect als de uitvoer niet naar de terminal gaat                                                                                                                                                                            |
| `-v`                    |           | Alles tonen wat een HTTP(S)-probe heeft verzameld: de HTTP-versie, de TLS-versie en cipher, de vervaldatum van het certificaat en de tijden voor verbinden, TLS en eerste byte. Bij een UDP-doel het nummer van de probe en of het antwoord het heeft teruggebracht, zodat een verloren probe te herkennen is. Geen effect bij TCP-doelen |

### Bestands- en machineleesbare uitvoer

| Flag               | Standaard | Beschrijving                                                                                                                                                                                                                                   |
| ------------------ | --------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-j`               |           | Uitvoer in `JSON`-formaat                                                                                                                                                                                                                      |
| `--pretty`         |           | De `JSON`-uitvoer netter opmaken. Geen effect zonder `-j`                                                                                                                                                                                      |
| `--json-url <URL>` |           | De `JSON`-uitvoer naar een HTTP-server sturen in plaats van afdrukken, één `POST` per gebeurtenis, bijv. `--json-url http://localhost:8000/tcping`. Zet `JSON`-uitvoer zelf aan, dus `-j` is niet nodig                                        |
| `--json-server`    |           | Niet testen. Luisteren op de opgegeven host en poort en elke daarheen gestuurde `JSON`-gebeurtenis afdrukken, zodat een tcping met `--json-url` elders een plek heeft om zijn run naartoe te sturen, bijv. `tcping --json-server 0.0.0.0 8000` |
| `--csv <file>`     |           | De uitvoer opslaan in een `CSV`-bestand. De statistieken gaan naar dezelfde naam met het achtervoegsel `_stats`                                                                                                                                |
| `--csv-fixed-name` |           | De bestandsnaam van `--csv` ongewijzigd gebruiken, zonder datum-/tijdachtervoegsel, zodat herhaalde runs hetzelfde bestand overschrijven                                                                                                       |
| `--sqlite <file>`  |           | De uitvoer opslaan in een sqlite3-database, bijv. `--sqlite /tmp/tcping.db`. Niet beschikbaar op Windows                                                                                                                                       |

### Metrics

| Flag                         | Standaard | Beschrijving                                                                                                                                                                                                                                                                                                                    |
| ---------------------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--otlp <URL>`               |           | De resultaten als metrics naar een OTLP-HTTP-eindpunt sturen, zoals [Grafana Alloy](https://grafana.com/docs/alloy/latest/) of de OpenTelemetry Collector, in plaats van ze af te drukken, bijv. `--otlp http://localhost:4318`                                                                                                 |
| `--otlp-header <header>`     |           | Extra HTTP-header om naar het `--otlp`-eindpunt te sturen, als `"Name: value"`, bijv. `"Authorization: Bearer <token>"`. Kan ook in de omgevingsvariabele `OTLP_HEADER` worden gegeven, wat hem uit uw shell-geschiedenis houdt                                                                                                 |
| `--influxdb <URL>`           |           | De resultaten als line protocol naar een [InfluxDB](https://www.influxdata.com/)-v2- of -v3-server schrijven in plaats van ze af te drukken, bijv. `--influxdb http://localhost:8086`                                                                                                                                           |
| `--influxdb-org <org>`       |           | InfluxDB-organisatie om naar te schrijven. Verplicht bij `--influxdb`                                                                                                                                                                                                                                                           |
| `--influxdb-bucket <bucket>` |           | InfluxDB-bucket om naar te schrijven. Verplicht bij `--influxdb`                                                                                                                                                                                                                                                                |
| `--influxdb-token <token>`   |           | InfluxDB-API-token. Verplicht bij `--influxdb`. Kan ook in de omgevingsvariabele `INFLUXDB_TOKEN` worden gegeven, wat het uit uw shell-geschiedenis houdt                                                                                                                                                                       |
| `--stats-interval <seconds>` | `10`      | Hoe vaak de statistieken naar het OTLP-eindpunt of InfluxDB worden gestuurd. Geen effect zonder `--otlp` of `--influxdb`                                                                                                                                                                                                        |
| `--source-label <name>`      | hostnaam  | Deze computer een naam geven in de `JSON`-gebeurtenissen en in de metrics die via OTLP of naar InfluxDB worden gestuurd, zodat meerdere computers die hetzelfde doel testen te onderscheiden zijn. Resultaten die elders heen gaan dragen de hostnaam als dit niet is opgegeven; uitvoer die op de computer blijft draagt niets |

### HTTP(S) en UDP

| Flag           | Standaard | Beschrijving                                                                                                                                                      |
| -------------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--insecure`   |           | Het servercertificaat niet verifiëren bij het testen van een `https://`-doel. Handig voor zelfondertekende of verlopen certificaten                               |
| `--udp-server` |           | Niet testen. Luisteren op de opgegeven host en poort en elk UDP-datagram terugsturen naar de afzender, zodat een UDP-probe naar deze computer een antwoord krijgt |

---

## Bijdragen

Pull requests zijn welkom om bugs op te lossen, nieuwe functies toe te voegen en te helpen met de
openstaande issues die u [hier](https://github.com/pouriyajamshidi/tcping/issues) kunt vinden.
Huidig aantal openstaande issues: ![GitHub issues](https://img.shields.io/github/issues/pouriyajamshidi/tcping.svg).

1. Kies een issue waar u zich prettig bij voelt.
1. Fork de repository.
1. Maak een branch.
1. Commit uw werk.
1. Voeg tests toe.
1. Voer de tests uit met `make test` en zorg dat ze slagen.
1. Maak een pull request.

Zorg ervoor dat uw pull request **maar één specifiek issue of één functie behandelt** en niet twee of meer. Dat maakt het voor ons makkelijker om uw pull request te beoordelen en helpt de git-geschiedenis netjes te houden.

Tenzij u een heel klein probleem oplost, geef dan eerst uw bedoeling aan in een **issue** voordat u aan het werk begint.

Om uw wijzigingen op een slecht netwerk uit te proberen, kan `tools/netcond.sh` vertraging toevoegen,
een deel van de pakketten laten vallen of één bestemming helemaal blokkeren, en daarna
alles ongedaan maken wat het heeft toegevoegd:

```bash
sudo ./tools/netcond.sh delay 1.1.1.1 800ms 200ms 30
sudo ./tools/netcond.sh loss 1.1.1.1 40
sudo ./tools/netcond.sh clear
```

## Het project helpen

Als tcping nuttig voor u is, overweeg dan het te delen met uw netwerk om het bereik te vergroten en ook andere mensen ervan te laten profiteren.

Daarnaast kunt u het project steunen via de onderstaande links:

- Trakteer me op een kop koffie: [![Buy Me A Coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/pouriyajamshidi)
- GitHub Sponsors: [![Sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/pouriyajamshidi)

## Licentie

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
