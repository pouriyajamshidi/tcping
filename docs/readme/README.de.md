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

> [!CAUTION]
> Dies ist ein Branch in Entwicklung, nicht der Hauptbranch.

Ein plattformübergreifendes Ping-Programm, das `TCP`, `UDP` oder `HTTP(S)` statt `ICMP` verwendet und vom Ping-Werkzeug von Linux inspiriert ist.

> [!TIP]
> Dieses Dokument gibt es auch auf [English](../../README.md), [中文](README.cn.md), [日本語](README.ja.md), [한국어](README.ko.md), [فارسی](README.fa.md), [Français](README.fr.md) und [Nederlands](README.nl.md).

- Eine Alternative zu `ping`, wo `ICMP` blockiert ist: Es prüft die Erreichbarkeit über `TCP`, `HTTP(S)` oder `UDP`.
- Meldet den Paketverlust sowie Minimum, Durchschnitt, Maximum und Standardabweichung der Latenz, genau wie `ping`, dazu die längste Betriebs- und Ausfallzeit und wann sie aufgetreten sind.
- Gibt die Statistik jederzeit aus, wenn man die Taste `Enter` drückt, ohne das Programm zu beenden.
- Gibt in den Formaten **farbig**, **einfach**, **JSON**, **CSV** oder **sqlite3** aus oder sendet jede Probe als Metrik per **OTLP** (Grafana Alloy, OpenTelemetry Collector und andere) oder an **InfluxDB**.
- Zeigt für jede `HTTP(S)`-Probe den Statuscode, die TLS-Version und die Cipher Suite, den Ablauf des Zertifikats sowie die Zeiten für Verbindungsaufbau, TLS-Handshake und das erste Byte.
- Löst den Hostnamen des Ziels nach einer Anzahl von Fehlschlägen (`-r`) oder vor jeder Probe (`--resolve-every-probe`) erneut auf und meldet, wie lange jede Abfrage gedauert hat. Geeignet, um Ihr `DNS`-Load-Balancing oder Ihren Global Server Load Balancer `(GSLB)` zu testen.
- Lässt Sie die **Quellschnittstelle**, das **Timeout**, das **Intervall** und den **DNS-Server** wählen und `IPv4` oder `IPv6` erzwingen.
- Zählt _erfolgreiche_ und _fehlgeschlagene_ Probes getrennt, sodass die Summen auf einen Blick sichtbar sind.

In den [Demos](#demos) bekommen Sie einen Eindruck vom Aussehen und der Bedienung von **tcping**.

---

## Inhaltsverzeichnis

- [Demos](#demos)
- [Download und Installation](#download-und-installation)
- [Verwendung](#verwendung)
- [Flags](#flags)
- [Mitwirken](#mitwirken)
- [Das Projekt unterstützen](#das-projekt-unterstützen)
- [Lizenz](#lizenz)

---

## Demos

<details>
<summary>Zum Aufklappen klicken</summary>

### Grundlegende Verwendung

![tcping](../Images/gifs/tcping.gif)

---

### Probes mit Zeitstempel über IPv4 (`-D -4`) Flags

![tcping timestamp example](../Images/gifs/tcping_timestamp.gif)

---

### Einfache Ausgabe, Statistik nur bei Druck auf Enter (`--no-color --no-stats`) Flags

![tcping plain output example](../Images/gifs/tcping_plain.gif)

---

### Hostnamen-Abfrage wiederholen (`-r`) Flag

![tcping resolve example](../Images/gifs/tcping_resolve.gif)

---

### JSON-Ausgabe (`-j --pretty`) Flag

![tcping json example](../Images/gifs/tcping_json_pretty.gif)

---

### Die JSON-Ausgabe an einen Collector streamen (`--json-url --json-server`) Flags

![tcping JSON streaming example](../Images/gifs/tcping_json_stream.gif)

---

### Dauer der Hostnamen-Auflösung (`--resolve-every-probe`) Flag

![tcping resolution timing example](../Images/gifs/tcping_dns_timing.gif)

---

### Quellschnittstelle (`-I`) Flag

![tcping interface example](../Images/gifs/tcping_interface.gif)

---

### HTTP(S)-Probes

![tcping HTTP example](../Images/gifs/tcping_http.gif)

---

### Details der HTTP(S)-Probes (`-v`) Flag

![tcping HTTP verbose example](../Images/gifs/tcping_http_verbose.gif)

---

### Zertifikatsprüfung überspringen (`--insecure`) Flag

![tcping skip TLS example](../Images/gifs/tcping_skip_tls.gif)

---

### UDP-Probes gegen einen UDP-Server (`--udp-server`) Flag

![tcping UDP example](../Images/gifs/tcping_udp.gif)

---

### Nur die fehlgeschlagenen Probes (`--failures-only`) Flag

![tcping failures only example](../Images/gifs/tcping_failures_only.gif)

</details>

---

## Download und Installation

| Plattform                                          | Installation mit                                                                           |
| -------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| [Windows](#windows)                                | `winget install pj.tcping`                                                                 |
| [macOS](#macos)                                    | `brew install pouriyajamshidi/tap/tcping`                                                  |
| [Debian, Ubuntu](#linux---paketquellen)            | `sudo apt install tcping`, nachdem die Paketquelle hinzugefügt wurde                       |
| [Fedora, RHEL, CentOS](#linux---paketquellen)      | `sudo dnf install tcping`, nachdem die Paketquelle hinzugefügt wurde                       |
| [Alpine](#linux---paketquellen)                    | `sudo apk add tcping`, nachdem die Paketquelle hinzugefügt wurde                           |
| [Arch, Manjaro](#linux---paketquellen)             | `yay -S tcping-bin`                                                                        |
| [Jedes Linux](#linux---ein-zeilen-installation)    | einer [Ein-Zeilen-Installation](#linux---ein-zeilen-installation) oder einer [Paketdatei](#linux---paketdateien) |
| [BSD](#bsd-und-linux---manuell)                    | der [vorkompilierten Binärdatei](#bsd-und-linux---manuell)                                 |
| [Nix](#nix)                                        | `nix profile install github:pouriyajamshidi/tcping`                                        |
| [Docker](#weitere-möglichkeiten)                   | `docker run -it pouriyajamshidi/tcping example.com 443`                                    |
| [Go](#weitere-möglichkeiten)                       | `go install github.com/pouriyajamshidi/tcping/v3@latest`                                   |

Die Binärdateien sind statisch. Das heißt, sie bringen alles Nötige in einer einzigen Datei mit
und hängen nicht davon ab, dass eine Bibliothek auf Ihrem Rechner vorhanden ist. Sie liegen auf der
[Release-Seite](https://github.com/pouriyajamshidi/tcping/releases/latest/), für
_amd64_ und _arm64_.

Wenn die Installation abgeschlossen ist, gehen Sie zum Abschnitt [Verwendung](#verwendung).

### Windows

Der beste Weg, **tcping** unter Windows zu installieren, ist der _Windows-Paket-Manager_ mit [WinGet](https://learn.microsoft.com/de-de/windows/package-manager/winget/?ref=github.com%2Fpouriyajamshidi%2Ftcping), der seit September 2020 auf praktisch allen Windows-_10_- und -_11_-Rechnern standardmäßig vorhanden ist:

```powershell
winget install pj.tcping
```

Um es von Hand zu installieren, entpacken Sie die heruntergeladene Zip-Datei und kopieren `tcping.exe` in Ihren System-[PATH](https://www.howtogeek.com/118594/how-to-edit-your-system-path-for-easy-command-line-access/), zum Beispiel nach `C:\Windows\System32`.

> [!CAUTION]
> TCPING wird von Windows Defender oder anderer Anti-Malware-Software möglicherweise fälschlich gemeldet. Das ist bei Go-Programmen üblich. Die offizielle Stellungnahme des Go-Teams finden Sie [hier](https://go.dev/doc/faq#virus).

> [!WARNING]
> Das Ausgabeformat `--sqlite` (sqlite3) ist in den Windows-Binärdateien nicht mehr verfügbar. Alle anderen Flags funktionieren wie erwartet.

### macOS

```bash
brew install pouriyajamshidi/tap/tcping
```

### Linux - Ein-Zeilen-Installation

Fügen Sie Folgendes in Ihr Terminal ein, um die neueste Binärdatei für Ihre Architektur zu laden und zu installieren:

```bash
cd /tmp &&
ARCH=$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/') &&
curl -LO "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
tar -xf "tcping-linux-$ARCH.tar.gz" &&
sudo install tcping /usr/local/bin/ &&
tcping --version
```

Wie Sie die Tab-Vervollständigung einrichten, steht unter [Shell-Vervollständigung](#shell-vervollständigung).

Wenn Sie kein `curl` haben, ersetzen Sie dessen Zeile durch `wget`:

```bash
wget "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
```

### Linux - Paketquellen

Wenn Sie die Paketquelle einmal hinzufügen, übernehmen `apt`, `dnf` oder `apk` ab dann
Installation und Updates wie bei jedem anderen Paket.

**Debian**, **Ubuntu** und ihre Ableger:

```bash
sudo install -d /etc/apt/keyrings &&
  sudo curl -fsSL https://pouriyajamshidi.github.io/packages/keys/pouriyajamshidi.gpg -o /etc/apt/keyrings/pouriyajamshidi.gpg &&
  echo "deb [signed-by=/etc/apt/keyrings/pouriyajamshidi.gpg] https://pouriyajamshidi.github.io/packages/deb ./" | sudo tee /etc/apt/sources.list.d/pouriyajamshidi.list &&
  sudo apt update &&
  sudo apt install tcping
```

**Fedora**, **RHEL**, **CentOS** und ihre Ableger:

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

**Arch**, **Manjaro**, **EndeavourOS** und ihre Ableger finden es im
[AUR](https://aur.archlinux.org/packages/tcping-bin). Nutzen Sie dafür Ihren bevorzugten Helfer:

```bash
yay -S tcping-bin
```

### Linux - Paketdateien

Ohne eine Paketquelle hinzuzufügen, laden Sie das Paket für Ihre Distribution von der
[Release-Seite](https://github.com/pouriyajamshidi/tcping/releases/latest/) herunter und
installieren es:

| Distribution         | Paket                      | Installation                                        |
| -------------------- | -------------------------- | --------------------------------------------------- |
| Debian, Ubuntu       | `tcping-amd64.deb`         | `sudo apt install -y ./tcping-amd64.deb`            |
| Fedora, RHEL, CentOS | `tcping-amd64.rpm`         | `sudo dnf install -y ./tcping-amd64.rpm`            |
| Arch, Manjaro        | `tcping-amd64.pkg.tar.zst` | `sudo pacman -U ./tcping-amd64.pkg.tar.zst`         |
| Alpine               | `tcping-amd64.apk`         | `sudo apk add --allow-untrusted ./tcping-amd64.apk` |

Ersetzen Sie `amd64` auf ARM-Rechnern durch `arm64`. Alle vier installieren die Binärdatei und die
[Shell-Vervollständigung](#shell-vervollständigung) für Sie. Rechner ohne `dnf` können stattdessen
`yum` benutzen, und Alpine braucht `--allow-untrusted`, weil unsere Pakete nicht
mit einem Alpine-Schlüssel signiert sind.

### BSD und Linux - Manuell

Laden Sie die Datei für Ihr jeweiliges Betriebssystem und Ihre Architektur herunter:

```bash
wget https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-freebsd-amd64.tar.gz
# Oder für Linux-ARM64-Rechner und mit cURL
curl -LO https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-arm64.tar.gz
```

Entpacken Sie sie und kopieren Sie die ausführbare Datei in Ihren System-`PATH`, zum Beispiel nach `/usr/local/bin/`:

```bash
tar -xvf tcping-freebsd-amd64.tar.gz &&
  chmod +x tcping &&
  sudo cp tcping /usr/local/bin/
```

Wie Sie die Tab-Vervollständigung einrichten, steht unter [Shell-Vervollständigung](#shell-vervollständigung).

### Nix

**tcping** liefert ein Flake mit. Auf jedem Rechner mit [Nix](https://nixos.org/download/)
können Sie es also ausführen, ohne etwas zu installieren:

```bash
nix run github:pouriyajamshidi/tcping -- example.com 80
```

Oder dauerhaft installieren:

```bash
nix profile install github:pouriyajamshidi/tcping
```

Beides liefert die Binärdatei und die [Shell-Vervollständigung](#shell-vervollständigung). Das
Flake hat außerdem eine `devShells.default` mit allem, was man für die Arbeit an tcping braucht,
die Sie mit `nix develop` erhalten.

### Weitere Möglichkeiten

- `Docker`-Images:

  ```bash
  docker pull pouriyajamshidi/tcping:latest
  # Oder
  docker pull ghcr.io/pouriyajamshidi/tcping:latest
  ```

- Mit `go install`:

  > Dazu wird mindestens Go in Version `1.26.8` benötigt

  ```bash
  go install github.com/pouriyajamshidi/tcping/v3@latest
  ```

- [x tcping](https://x-cmd.com/pkg/tcping):

  **Direkt ohne Installation** in [x-cmd](https://www.x-cmd.com).

  ```bash
  x tcping example.com 80
  ```

  Oder installieren Sie `tcping` mit x-cmd lokal, ohne Root-Rechte und ohne Ihre globale Umgebung zu verändern.

  ```bash
  x env use tcping
  tcping example.com 80
  ```

- Schließlich können Sie den Code selbst kompilieren, indem Sie den Befehl `make` ausführen:

  ```bash
  make build
  ```

  Dadurch landet die ausführbare Datei im Ordner `target`.

### Shell-Vervollständigung

**tcping** bringt Vervollständigungsskripte für `bash`, `zsh`, `fish` und `PowerShell` mit
und gibt sie mit `--completions <shell>` aus. Sie vervollständigen die Flags, die
Schnittstellennamen für `-I` und die Dateinamen für `--csv` und `--sqlite`.

Die Linux-Pakete, Homebrew, das AUR-Paket, Nix und die
[Ein-Zeilen-Installation](#linux---ein-zeilen-installation) legen sie für Sie an die richtige Stelle.
Andernfalls speichern Sie das Skript für Ihre Shell:

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

- `PowerShell`, indem Sie es aus Ihrem Profil laden:

  ```powershell
  Add-Content $PROFILE "tcping --completions powershell | Out-String | Invoke-Expression"
  ```

Starten Sie danach eine neue Shell, damit sie wirksam werden.

---

## Verwendung

Geben Sie **tcping** ein Ziel und einen Port, dann beginnt es mit den Probes:

```bash
tcping www.example.com 443
tcping www.example.com:443          # host:port funktioniert auch
tcping 192.168.1.1:80
tcping '[2001:db8::1]:443'          # in Anführungszeichen, damit die Shell es in Ruhe lässt
tcping https://www.example.com      # über HTTP(S) geprüft
tcping udp://127.0.0.1 53           # über UDP geprüft
```

Einiges, was Sie von einem Lauf verlangen können:

```bash
tcping www.example.com 443 -c 5            # nach 5 Probes anhalten
tcping www.example.com 443 -i 2 -t 5       # 2 Sekunden zwischen den Probes, 5 Sekunden Timeout
tcping www.example.com 443 -I eth2         # die Probes von einer bestimmten Schnittstelle senden
tcping www.example.com 443 -4              # oder -6, um eine Adressfamilie zu erzwingen
tcping www.example.com 443 -D              # für jede Probe einen Zeitstempel zeigen
tcping www.example.com 443 -r 5            # den Hostnamen nach 5 Fehlschlägen erneut auflösen
```

Und die Ausgabeformate, die es statt der farbigen Standardausgabe erzeugen kann:

```bash
tcping www.example.com 443 -j              # JSON, mit --pretty besser lesbar
tcping www.example.com 443 --no-color      # einfach, ohne ANSI-Farben
tcping www.example.com 443 --csv out.csv   # CSV-Datei
tcping www.example.com 443 --sqlite out.db # sqlite3-Datenbank
```

Das Docker-Image nimmt dieselben Ziele und Flags:

```bash
docker run -it pouriyajamshidi/tcping:latest example.com 443
# Oder aus der GitHub Container Registry
docker run -it ghcr.io/pouriyajamshidi/tcping:latest example.com:443
```

> [!TIP]
> Drücken Sie die Taste `Enter`, während das Programm läuft, um die Zusammenfassung aller Probes zu sehen, ohne das Programm zu beenden, wie im Abschnitt [Demos](#demos) gezeigt.

> [!NOTE]
> Die **verfügbaren Flags** für eine fortgeschrittenere Verwendung finden Sie [hier](#flags).

### Probes über HTTP(S)

Geben Sie statt Host und Port eine URL an, dann prüft tcping sie über HTTP(S)
und meldet den Statuscode und wie lange der Server für die Antwort brauchte. Der Body
wird nicht heruntergeladen, eine große Seite bremst die Probe also nicht aus:

```bash
tcping https://www.example.com/health
```

Der Port ergibt sich aus dem Schema, `80` oder `443`, es sei denn, die URL enthält einen eigenen
oder Sie geben danach einen an:

```bash
tcping http://www.example.com:8080/health
# Dasselbe
tcping http://www.example.com/health 8080
```

`-v` zeigt alles, was eine Probe gesammelt hat: die HTTP-Version, die TLS-Version und
Cipher Suite, wie viele Tage das Zertifikat noch gültig ist sowie die Zeiten für Verbindungsaufbau, TLS-Handshake
und erstes Byte. `--insecure` überspringt die Zertifikatsprüfung,
was Sie bei einem selbstsignierten oder abgelaufenen Zertifikat wollen:

```bash
tcping https://www.example.com/health -v
tcping https://self-signed.example.com --insecure
```

### Probes über UDP

```bash
tcping udp://127.0.0.1 53
```

UDP hat keinen Handshake, eine Probe ist also nur erfolgreich, wenn die Gegenseite antwortet.
Um eine Antwort zu bekommen, starten Sie tcping auf der Gegenseite als UDP-Server, der
jedes Datagramm an den Absender zurückschickt:

```bash
# auf dem Rechner, der geprüft wird:
tcping --udp-server 127.0.0.1 9999
# auf dem Rechner, der prüft:
tcping udp://127.0.0.1 9999
```

> [!NOTE]
> Eine abgelehnte UDP-Probe meldet `port unreachable`, was bedeutet, dass
> etwas uns blockiert. Eine Probe ohne jede Antwort wird ebenfalls als
> Fehlschlag gemeldet, kann aber einen offenen Port, der still bleibt, nicht von einem
> unterwegs verworfenen Paket unterscheiden.
>
> Unter Windows erscheint eine abgelehnte Probe immer als keine Antwort. Go schaltet die
> Fehler "port unreachable" bei UDP-Sockets unter Windows ab, tcping sieht sie also nie.

Jede UDP-Probe sendet ihre eigene Nummer als Nutzlast, die der Server zurückschickt.
Mit `-v` erscheint diese Nummer in jeder Zeile, sodass eine verlorene Probe benannt werden kann:

```text
Reply from 127.0.0.1 on port 9999 UDP_conn=4 time=1.276 ms
    reply echoed back probe 4
```

### Die JSON-Ausgabe an einen Server streamen

Die `JSON`-Ausgabe muss nicht ausgegeben werden. Geben Sie `--json-url` eine Adresse, dann wird
jedes Ereignis sofort dorthin per `POST` geschickt, ein Ereignis pro Anfrage, sodass ein Rechner,
der prüft, an einen Rechner berichten kann, der sammelt:

```bash
tcping www.example.com 443 --json-url http://localhost:8000/tcping
```

Die Ereignisse und ihre Felder sind genau die, die `-j` ausgibt, alles, was die
weitergeleitete Ausgabe schon liest, liest also auch diese. Jedes gestreamte Ereignis nennt außerdem
den Rechner, der es gesendet hat, im Feld `source`, das standardmäßig der Hostname
dieses Rechners ist. So kann ein Collector, der Ereignisse von mehreren Rechnern annimmt, erkennen,
wessen Lauf er gerade sieht. Mit `--source-label` benennen Sie sie selbst:

```bash
tcping www.example.com 443 --json-url http://collector:8000/tcping --source-label brussels
```

tcping kann auch die empfangende Seite sein, so wie `--udp-server` UDP-Probes
beantwortet. `--json-server` prüft nicht: Es lauscht auf dem angegebenen Host und
Port und gibt jedes dorthin gesendete Ereignis aus, auf jedem Pfad, der sammelnde Rechner
braucht also ebenfalls nichts zu installieren:

```bash
# auf dem sammelnden Rechner:
tcping --json-server 0.0.0.0 8000

# auf dem prüfenden Rechner:
tcping www.example.com 443 --json-url http://collector:8000/tcping
```

Die Ereignisse gehen auf die Standardausgabe und alles andere auf die Standardfehlerausgabe, sodass
`tcping --json-server 0.0.0.0 8000 > events.jsonl` eine Datei hinterlässt, die nur aus
Ereignissen besteht, und alles, was kein `JSON` ist, abgelehnt statt ausgegeben wird.

Jeder Server, der einen `POST` mit `JSON`-Body annimmt, funktioniert genauso gut, z. B.:

```python
from fastapi import FastAPI

app = FastAPI()


@app.post("/tcping")
async def tcping(event: dict) -> None:
    print(event)
```

Alle Ereignisse eines Laufs gehen über eine Verbindung, die zwischen den
Probes offen gehalten und nicht für jeden neu aufgebaut wird. Ein Server, der ausgefallen oder gestört ist,
hält die Probes nicht an: tcping meldet es einmal und macht weiter, wobei
Ereignisse, die es nicht zustellen kann, verworfen werden.

### Die Ergebnisse per OTLP oder an InfluxDB senden

Statt jede Probe auszugeben, kann tcping sie als Metrik senden, wodurch aus einem
Lauf ein Diagramm wird und mehrere Rechner dasselbe Ziel beobachten können. Per OTLP,
an alles, was es annimmt, etwa [Prometheus](https://prometheus.io/),
gestartet mit `--web.enable-otlp-receiver`:

```bash
tcping www.example.com 443 --otlp http://localhost:9090/api/v1/otlp
```

Oder [Grafana Alloy](https://grafana.com/docs/alloy/latest/) und den
OpenTelemetry Collector, die es an mehrere Stellen weiterleiten können:

```bash
tcping www.example.com 443 --otlp http://localhost:4318
```

Gehostete Backends wollen ein Token, jedes in einem eigenen Header. Geben Sie den ganzen
Header mit `--otlp-header` an, oder in der Umgebungsvariable `OTLP_HEADER`, damit er
nicht in Ihrer Shell-Historie landet:

```bash
export OTLP_HEADER="Authorization: Bearer your-api-token" && \
  tcping www.example.com 443 --otlp https://otlp.example.com
```

Oder direkt an einen [InfluxDB](https://www.influxdata.com/)-v2- oder -v3-Server im
Line Protocol:

```bash
export INFLUXDB_TOKEN=your-api-token && \
  tcping www.example.com 443 --influxdb http://localhost:8086 --influxdb-org home --influxdb-bucket tcping
```

Jede Probe enthält ihre Round-Trip-Zeit, ob sie erfolgreich war, die Adresse, in die das
Ziel aufgelöst wurde, und die Zähler, dazu die HTTP- oder UDP-Details, wenn das Ziel
eines von beiden ist. Der gesamte Statistikblock, den Sie normalerweise beim Beenden sehen,
wird zusätzlich alle 10 Sekunden gesendet, sodass auch ein Lauf, den niemand
beobachtet, ihn meldet. Mit `--stats-interval` ändern Sie dieses Intervall.

Jeder Rechner benennt sich in den Metriken mit dem Label `source`, das
standardmäßig sein Hostname ist, sodass mehrere Rechner, die dasselbe Ziel prüfen, in
ihren eigenen Reihen landen statt übereinander. Mit `--source-label` benennen Sie
sie selbst:

```bash
tcping www.example.com 443 --otlp http://localhost:4318 --source-label brussels
```

Alles, was gesendet wird, wie man es abfragt, und fertige Prometheus-, Alloy-,
OpenTelemetry-Collector- und InfluxDB-Stacks mit Grafana und tcping, die Sie
mit einem Befehl starten können, um das auszuprobieren, ohne zuerst einen Server aufzusetzen, stehen in
[docs/observability](../observability/README.md).

---

## Flags

Flags können vor oder nach dem Ziel stehen, mit einem oder zwei
Bindestrichen, `-c 5` und `--c 5` sind also dasselbe Flag.

### Allgemein

| Flag        | Standard | Beschreibung                          |
| ----------- | -------- | ------------------------------------- |
| `-h`        |          | Zeigt die verfügbaren Flags und beendet |
| `--version` |          | Zeigt die Version und beendet         |
| `-u`        |          | Sucht nach Updates und beendet        |

### Probes

| Flag            | Standard    | Beschreibung                                                                                        |
| --------------- | ----------- | --------------------------------------------------------------------------------------------------- |
| `-c <n>`        | unbegrenzt  | Nach `<n>` Probes anhalten, unabhängig vom Ergebnis                                                 |
| `-i <seconds>`  | `1`         | Intervall zwischen den Probes. Dezimalzahl erlaubt, z. B. `-i 0.5`                                  |
| `-t <seconds>`  | `1`         | Wartezeit auf eine Antwort in Sekunden. Dezimalzahl erlaubt. `0` bedeutet unendliches Timeout       |
| `-4`            |             | Nur IPv4-Adressen verwenden                                                                         |
| `-6`            |             | Nur IPv6-Adressen verwenden                                                                         |
| `-I <name\|IP>` |             | Schnittstellenname oder IP-Adresse, von der die Probes und die DNS-Abfragen gesendet werden        |

> [!TIP]
> Ohne die Flags `-4` und `-6` wählt tcping anhand der DNS-Abfragen zufällig eine IP-Adresse aus.

### Namensauflösung

| Flag                      | Standard    | Beschreibung                                                                                                                                                              |
| ------------------------- | ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-r <n>`                  | nie         | Den Hostnamen des Ziels nach `<n>` fehlgeschlagenen Probes erneut auflösen, z. B. `-r 10`                                                                                 |
| `--resolve-every-probe`   |             | Den Hostnamen des Ziels vor jeder einzelnen Probe statt nur beim Start auflösen. Hat Vorrang vor `-r` und wirkt nicht, wenn das Ziel eine IP-Adresse ist                   |
| `--dns-server <IP>`       | systemweit  | Zu verwendender eigener DNS-Server. Eine Kombination aus IP und Port ist erlaubt, z. B. `--dns-server 1.1.1.1:53`                                                         |
| `--dns-timeout <seconds>` | `2`         | Wartezeit auf eine DNS-Antwort in Sekunden. Dezimalzahl erlaubt. `0` bedeutet unendliches Timeout                                                                         |

### Terminalausgabe

| Flag                    | Standard | Beschreibung                                                                                                                                                                                                                                                                                                                  |
| ----------------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-D`                    |          | Für jede Probe einen Zeitstempel zeigen                                                                                                                                                                                                                                                                                      |
| `--no-color`            |          | Die Ausgabe nicht einfärben                                                                                                                                                                                                                                                                                                   |
| `--show-source-address` |          | Die für die Probes verwendete Quell-IP-Adresse und den Port zeigen                                                                                                                                                                                                                                                            |
| `--failures-only`       |          | Nur die fehlgeschlagenen Probes zeigen. Die erfolgreichen werden weiterhin gezählt                                                                                                                                                                                                                                            |
| `--no-stats`            |          | Die Statistik beim Beenden des Programms nicht zeigen. Ein Druck auf **Enter** zeigt sie weiterhin. Wirkt nicht, wenn die Ausgabe nicht ins Terminal geht                                                                                                                                                                    |
| `-v`                    |          | Alles zeigen, was eine HTTP(S)-Probe gesammelt hat: die HTTP-Version, die TLS-Version und Cipher Suite, den Ablauf des Zertifikats und die Zeiten für Verbindung, TLS und erstes Byte. Bei einem UDP-Ziel die Nummer der Probe und ob die Antwort sie zurückgebracht hat, sodass eine verlorene Probe erkennbar ist. Wirkt nicht bei TCP-Zielen |

### Datei- und maschinenlesbare Ausgabe

| Flag               | Standard | Beschreibung                                                                                                                                                                                                              |
| ------------------ | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-j`               |          | Im `JSON`-Format ausgeben                                                                                                                                                                                                 |
| `--pretty`         |          | Die `JSON`-Ausgabe lesbarer formatieren. Wirkt nicht ohne `-j`                                                                                                                                                            |
| `--json-url <URL>` |          | Die `JSON`-Ausgabe an einen HTTP-Server senden statt auszugeben, ein `POST` pro Ereignis, z. B. `--json-url http://localhost:8000/tcping`. Schaltet die `JSON`-Ausgabe von selbst ein, `-j` ist also nicht nötig           |
| `--json-server`    |          | Nicht prüfen. Auf dem angegebenen Host und Port lauschen und jedes dorthin gesendete `JSON`-Ereignis ausgeben, damit ein tcping mit `--json-url` anderswo ein Ziel für seinen Lauf hat, z. B. `tcping --json-server 0.0.0.0 8000` |
| `--csv <file>`     |          | Die Ausgabe in einer `CSV`-Datei speichern. Die Statistik geht in denselben Namen mit der Endung `_stats`                                                                                                                 |
| `--csv-fixed-name` |          | Den `--csv`-Dateinamen unverändert ohne Datums-/Zeitsuffix verwenden, sodass wiederholte Läufe dieselbe Datei überschreiben                                                                                                |
| `--sqlite <file>`  |          | Die Ausgabe in einer sqlite3-Datenbank speichern, z. B. `--sqlite /tmp/tcping.db`. Unter Windows nicht verfügbar                                                                                                          |

### Metriken

| Flag                         | Standard | Beschreibung                                                                                                                                                                                                                                                                               |
| ---------------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--otlp <URL>`               |          | Die Ergebnisse als Metriken an einen OTLP-HTTP-Endpunkt senden, etwa [Grafana Alloy](https://grafana.com/docs/alloy/latest/) oder den OpenTelemetry Collector, statt sie auszugeben, z. B. `--otlp http://localhost:4318`                                                                    |
| `--otlp-header <header>`     |          | Zusätzlicher HTTP-Header für den `--otlp`-Endpunkt als `"Name: value"`, z. B. `"Authorization: Bearer <token>"`. Kann auch in der Umgebungsvariable `OTLP_HEADER` angegeben werden, damit er nicht in Ihrer Shell-Historie landet                                                          |
| `--influxdb <URL>`           |          | Die Ergebnisse als Line Protocol auf einen [InfluxDB](https://www.influxdata.com/)-v2- oder -v3-Server schreiben statt sie auszugeben, z. B. `--influxdb http://localhost:8086`                                                                                                            |
| `--influxdb-org <org>`       |          | InfluxDB-Organisation, in die geschrieben wird. Erforderlich mit `--influxdb`                                                                                                                                                                                                              |
| `--influxdb-bucket <bucket>` |          | InfluxDB-Bucket, in den geschrieben wird. Erforderlich mit `--influxdb`                                                                                                                                                                                                                    |
| `--influxdb-token <token>`   |          | InfluxDB-API-Token. Erforderlich mit `--influxdb`. Kann auch in der Umgebungsvariable `INFLUXDB_TOKEN` angegeben werden, damit es nicht in Ihrer Shell-Historie landet                                                                                                                      |
| `--stats-interval <seconds>` | `10`     | Wie oft die Statistik an den OTLP-Endpunkt oder InfluxDB gesendet wird. Wirkt nicht ohne `--otlp` oder `--influxdb`                                                                                                                                                                        |
| `--source-label <name>`      | Hostname | Diesen Rechner in den `JSON`-Ereignissen und in den per OTLP oder an InfluxDB gesendeten Metriken benennen, damit mehrere Rechner, die dasselbe Ziel prüfen, unterscheidbar sind. Ergebnisse, die anderswohin gesendet werden, tragen den Hostnamen, wenn dies nicht angegeben ist; Ausgaben, die auf dem Rechner bleiben, tragen nichts |

### HTTP(S) und UDP

| Flag           | Standard | Beschreibung                                                                                                                                                           |
| -------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--insecure`   |          | Das Serverzertifikat beim Prüfen eines `https://`-Ziels nicht verifizieren. Nützlich bei selbstsignierten oder abgelaufenen Zertifikaten                                |
| `--udp-server` |          | Nicht prüfen. Auf dem angegebenen Host und Port lauschen und jedes UDP-Datagramm an den Absender zurückschicken, sodass eine auf diesen Rechner gerichtete UDP-Probe eine Antwort bekommt |

---

## Mitwirken

Pull Requests sind willkommen, um Fehler zu beheben, neue Funktionen hinzuzufügen und bei den
offenen Issues zu helfen, die Sie [hier](https://github.com/pouriyajamshidi/tcping/issues) finden.
Aktuelle Anzahl offener Issues: ![GitHub issues](https://img.shields.io/github/issues/pouriyajamshidi/tcping.svg).

1. Wählen Sie ein Issue, mit dem Sie sich wohlfühlen.
1. Forken Sie das Repository.
1. Erstellen Sie einen Branch.
1. Committen Sie Ihre Arbeit.
1. Fügen Sie Tests hinzu.
1. Führen Sie die Tests mit `make test` aus und stellen Sie sicher, dass sie erfolgreich sind.
1. Erstellen Sie einen Pull Request.

Bitte stellen Sie sicher, dass Ihr Pull Request **nur ein bestimmtes Issue bzw. Feature abdeckt** und nicht zwei oder mehr Issues behandelt. Das macht es uns leichter, Ihren Pull Request zu prüfen, und hilft, die Git-Historie sauber zu halten.

Sofern Sie kein wirklich winziges Problem beheben, teilen Sie Ihre Absicht bitte zuerst in einem **Issue** mit, bevor Sie mit der Arbeit beginnen.

Um Ihre Änderungen in einem schlechten Netzwerk auszuprobieren, kann `tools/netcond.sh` Latenz hinzufügen,
einen Teil der Pakete verwerfen oder ein Ziel komplett blockieren und anschließend
alles wieder rückgängig machen, was es hinzugefügt hat:

```bash
sudo ./tools/netcond.sh delay 1.1.1.1 800ms 200ms 30
sudo ./tools/netcond.sh loss 1.1.1.1 40
sudo ./tools/netcond.sh clear
```

## Das Projekt unterstützen

Wenn tcping für Sie nützlich ist, erwägen Sie, es in Ihrem Netzwerk weiterzuempfehlen, um seine Reichweite zu vergrößern und auch anderen Menschen zu helfen, davon zu profitieren.

Außerdem können Sie das Projekt über die folgenden Links unterstützen:

- Einen Kaffee spendieren: [![Buy Me A Coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/pouriyajamshidi)
- GitHub Sponsors: [![Sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/pouriyajamshidi)

## Lizenz

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
