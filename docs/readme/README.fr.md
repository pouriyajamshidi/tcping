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

Un programme ping multiplateforme qui utilise `TCP`, `UDP` ou `HTTP(S)` à la place d'`ICMP`, inspiré de l'utilitaire ping de Linux.

> [!TIP]
> Ce document est aussi disponible en [English](../../README.md), [中文](README.cn.md), [日本語](README.ja.md), [한국어](README.ko.md), [فارسی](README.fa.md), [Deutsch](README.de.md) et [Nederlands](README.nl.md).

- Une alternative à `ping` là où `ICMP` est bloqué, qui sonde en `TCP`, `HTTP(S)` ou `UDP`.
- Indique la perte de paquets ainsi que le minimum, la moyenne, le maximum et l'écart type de la latence, comme le fait `ping`, plus la plus longue durée de disponibilité et d'indisponibilité et le moment où elles ont eu lieu.
- Affiche les statistiques à tout moment en appuyant sur la touche `Entrée`, sans arrêter le programme.
- Produit une sortie **en couleur**, **brute**, **JSON**, **CSV** ou **sqlite3**, ou envoie chaque sonde sous forme de métriques par **OTLP** (Grafana Alloy, OpenTelemetry Collector et autres) ou vers **InfluxDB**.
- Montre le code d'état, la version de TLS et la suite de chiffrement, l'expiration du certificat ainsi que les temps de connexion, de négociation TLS et d'arrivée du premier octet de chaque sonde `HTTP(S)`.
- Résout à nouveau le nom d'hôte de la cible après un certain nombre d'échecs (`-r`) ou avant chaque sonde (`--resolve-every-probe`), et indique la durée de chaque résolution. Pratique pour tester votre répartition de charge `DNS` ou votre répartiteur de charge global `(GSLB)`.
- Permet de choisir l'**interface source**, le **délai d'attente**, l'**intervalle** et le **serveur DNS**, et d'imposer `IPv4` ou `IPv6`.
- Numérote séparément les sondes _réussies_ et _échouées_, pour que les totaux soient visibles d'un coup d'œil.

Regardez les [démos](#démos) pour vous faire une idée de l'aspect et de la prise en main de **tcping**.

---

## Table des matières

- [Démos](#démos)
- [Téléchargement et installation](#téléchargement-et-installation)
- [Utilisation](#utilisation)
- [Options](#options)
- [Contribuer](#contribuer)
- [Aider le projet](#aider-le-projet)
- [Licence](#licence)

---

## Démos

<details>
<summary>Cliquez pour déplier</summary>

### Utilisation de base

![tcping](../Images/gifs/tcping.gif)

---

### Sondes horodatées en IPv4 (options `-D -4`)

![tcping timestamp example](../Images/gifs/tcping_timestamp.gif)

---

### Sortie brute, statistiques seulement quand on appuie sur Entrée (options `--no-color --no-stats`)

![tcping plain output example](../Images/gifs/tcping_plain.gif)

---

### Nouvelle tentative de résolution du nom d'hôte (option `-r`)

![tcping resolve example](../Images/gifs/tcping_resolve.gif)

---

### Sortie JSON (options `-j --pretty`)

![tcping json example](../Images/gifs/tcping_json_pretty.gif)

---

### Diffuser la sortie JSON vers un collecteur (options `--json-url --json-server`)

![tcping JSON streaming example](../Images/gifs/tcping_json_stream.gif)

---

### Durée de la résolution du nom d'hôte (option `--resolve-every-probe`)

![tcping resolution timing example](../Images/gifs/tcping_dns_timing.gif)

---

### Interface source (option `-I`)

![tcping interface example](../Images/gifs/tcping_interface.gif)

---

### Sondes HTTP(S)

![tcping HTTP example](../Images/gifs/tcping_http.gif)

---

### Détails des sondes HTTP(S) (option `-v`)

![tcping HTTP verbose example](../Images/gifs/tcping_http_verbose.gif)

---

### Ignorer la vérification du certificat (option `--insecure`)

![tcping skip TLS example](../Images/gifs/tcping_skip_tls.gif)

---

### Sondes à travers un proxy (option `--proxy`)

![tcping proxy example](../Images/gifs/tcping_proxy.gif)

---

### Sondes UDP vers un serveur UDP (option `--udp-server`)

![tcping UDP example](../Images/gifs/tcping_udp.gif)

---

### Uniquement les sondes échouées (option `--failures-only`)

![tcping failures only example](../Images/gifs/tcping_failures_only.gif)

</details>

---

## Téléchargement et installation

| Plateforme                                                 | Installer avec                                                                                                              |
| ---------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| [Windows](#windows)                                        | `winget install pj.tcping`                                                                                                  |
| [macOS](#macos)                                            | `brew install pouriyajamshidi/tap/tcping`                                                                                   |
| [Debian, Ubuntu](#linux---dépôts-de-paquets)               | `sudo apt install tcping`, après avoir ajouté le dépôt                                                                      |
| [Fedora, RHEL, CentOS](#linux---dépôts-de-paquets)         | `sudo dnf install tcping`, après avoir ajouté le dépôt                                                                      |
| [Alpine](#linux---dépôts-de-paquets)                       | `sudo apk add tcping`, après avoir ajouté le dépôt                                                                          |
| [Arch, Manjaro](#linux---dépôts-de-paquets)                | `yay -S tcping-bin`                                                                                                         |
| [N'importe quel Linux](#linux---installation-en-une-ligne) | une [installation en une ligne](#linux---installation-en-une-ligne) ou un [fichier de paquet](#linux---fichiers-de-paquets) |
| [BSD](#bsd-et-linux---installation-manuelle)               | le [binaire précompilé](#bsd-et-linux---installation-manuelle)                                                              |
| [Nix](#nix)                                                | `nix profile install github:pouriyajamshidi/tcping`                                                                         |
| [Docker](#autres-méthodes)                                 | `docker run -it pouriyajamshidi/tcping example.com 443`                                                                     |
| [Go](#autres-méthodes)                                     | `go install github.com/pouriyajamshidi/tcping/v3@latest`                                                                    |

Les binaires sont statiques : ils contiennent tout ce dont ils ont besoin dans un seul fichier
et ne dépendent d'aucune bibliothèque présente sur votre machine. Ils se trouvent sur la
[page des versions](https://github.com/pouriyajamshidi/tcping/releases/latest/), pour
_amd64_ et _arm64_.

Une fois l'installation terminée, passez à la section [utilisation](#utilisation).

### Windows

La meilleure façon d'installer **tcping** sous Windows est le _Gestionnaire de package Windows_ avec [WinGet](https://learn.microsoft.com/fr-fr/windows/package-manager/winget/?ref=github.com%2Fpouriyajamshidi%2Ftcping), disponible par défaut sur pratiquement toutes les machines Windows _10_ et _11_ depuis septembre 2020 :

```powershell
winget install pj.tcping
```

Pour l'installer à la main, extrayez le fichier zip téléchargé et copiez `tcping.exe` dans le [PATH](https://www.howtogeek.com/118594/how-to-edit-your-system-path-for-easy-command-line-access/) de votre système, par exemple `C:\Windows\System32`.

> [!CAUTION]
> TCPING peut être signalé à tort par Windows Defender ou par certains logiciels anti-malware. C'est courant avec les programmes Go. Consultez la déclaration officielle de l'équipe Go [ici](https://go.dev/doc/faq#virus).

> [!WARNING]
> Le format de sortie `--sqlite` (sqlite3) n'est plus disponible dans les binaires Windows. Toutes les autres options fonctionnent comme prévu.

### macOS

```bash
brew install pouriyajamshidi/tap/tcping
```

### Linux - Installation en une ligne

Collez ce qui suit dans votre terminal pour récupérer le dernier binaire pour votre architecture et l'installer :

```bash
cd /tmp &&
ARCH=$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/') &&
curl -LO "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
tar -xf "tcping-linux-$ARCH.tar.gz" &&
sudo install tcping /usr/local/bin/ &&
tcping --version
```

Voir [Complétions du shell](#complétions-du-shell) pour configurer la complétion avec la touche Tab.

Si vous n'avez pas `curl`, remplacez sa ligne par `wget` :

```bash
wget "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
```

### Linux - Dépôts de paquets

Une fois le dépôt ajouté, `apt`, `dnf` ou `apk` s'occupe des installations et des
mises à jour comme pour n'importe quel autre paquet.

**Debian**, **Ubuntu** et leurs dérivées :

```bash
sudo install -d /etc/apt/keyrings &&
  sudo curl -fsSL https://pouriyajamshidi.github.io/packages/keys/pouriyajamshidi.gpg -o /etc/apt/keyrings/pouriyajamshidi.gpg &&
  echo "deb [signed-by=/etc/apt/keyrings/pouriyajamshidi.gpg] https://pouriyajamshidi.github.io/packages/deb ./" | sudo tee /etc/apt/sources.list.d/pouriyajamshidi.list &&
  sudo apt update &&
  sudo apt install tcping
```

**Fedora**, **RHEL**, **CentOS** et leurs dérivées :

```bash
sudo curl -fsSL https://pouriyajamshidi.github.io/packages/rpm/pouriyajamshidi.repo -o /etc/yum.repos.d/pouriyajamshidi.repo &&
  sudo dnf install tcping
```

**Alpine** :

```bash
sudo curl -fsSL https://pouriyajamshidi.github.io/packages/keys/pouriyajamshidi.rsa.pub -o /etc/apk/keys/pouriyajamshidi.rsa.pub &&
  echo "https://pouriyajamshidi.github.io/packages/apk" | sudo tee -a /etc/apk/repositories &&
  sudo apk update &&
  sudo apk add tcping
```

**Arch**, **Manjaro**, **EndeavourOS** et leurs dérivées le trouvent dans l'
[AUR](https://aur.archlinux.org/packages/tcping-bin), avec votre assistant préféré :

```bash
yay -S tcping-bin
```

### Linux - Fichiers de paquets

Sans ajouter de dépôt, téléchargez le paquet de votre distribution depuis la
[page des versions](https://github.com/pouriyajamshidi/tcping/releases/latest/) et
installez-le :

| Distribution         | Paquet                     | Installation                                        |
| -------------------- | -------------------------- | --------------------------------------------------- |
| Debian, Ubuntu       | `tcping-amd64.deb`         | `sudo apt install -y ./tcping-amd64.deb`            |
| Fedora, RHEL, CentOS | `tcping-amd64.rpm`         | `sudo dnf install -y ./tcping-amd64.rpm`            |
| Arch, Manjaro        | `tcping-amd64.pkg.tar.zst` | `sudo pacman -U ./tcping-amd64.pkg.tar.zst`         |
| Alpine               | `tcping-amd64.apk`         | `sudo apk add --allow-untrusted ./tcping-amd64.apk` |

Remplacez `amd64` par `arm64` sur les machines ARM. Les quatre installent le binaire et les
[complétions du shell](#complétions-du-shell) pour vous. Les machines sans `dnf` peuvent utiliser
`yum` à la place, et Alpine a besoin de `--allow-untrusted` car nos paquets ne sont pas
signés avec une clé Alpine.

### BSD et Linux - Installation manuelle

Téléchargez le fichier correspondant à votre système et à votre architecture :

```bash
wget https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-freebsd-amd64.tar.gz
# Ou, pour les machines Linux ARM64 et avec cURL
curl -LO https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-arm64.tar.gz
```

Extrayez-le et copiez l'exécutable dans le `PATH` de votre système, par exemple `/usr/local/bin/` :

```bash
tar -xvf tcping-freebsd-amd64.tar.gz &&
  chmod +x tcping &&
  sudo cp tcping /usr/local/bin/
```

Voir [Complétions du shell](#complétions-du-shell) pour configurer la complétion avec la touche Tab.

### Nix

**tcping** fournit un flake : sur toute machine qui a [Nix](https://nixos.org/download/),
vous pouvez donc l'exécuter sans rien installer :

```bash
nix run github:pouriyajamshidi/tcping -- example.com 80
```

Ou l'installer pour de bon :

```bash
nix profile install github:pouriyajamshidi/tcping
```

Les deux vous donnent le binaire et les [complétions du shell](#complétions-du-shell). Le
flake propose aussi un `devShells.default` avec tout ce qu'il faut pour travailler sur tcping,
que vous obtenez avec `nix develop`.

### Autres méthodes

- Images `Docker` :

  ```bash
  docker pull pouriyajamshidi/tcping:latest
  # Ou
  docker pull ghcr.io/pouriyajamshidi/tcping:latest
  ```

- Avec `go install` :

  > Cela demande au moins la version `1.26.8` de go

  ```bash
  go install github.com/pouriyajamshidi/tcping/v3@latest
  ```

- [x tcping](https://x-cmd.com/pkg/tcping) :

  **Directement sans installation** dans [x-cmd](https://www.x-cmd.com).

  ```bash
  x tcping example.com 80
  ```

  Ou installez `tcping` en local avec x-cmd, sans avoir besoin des droits root et sans toucher à votre configuration globale.

  ```bash
  x env use tcping
  tcping example.com 80
  ```

- Enfin, vous pouvez compiler le code vous-même en lançant la commande `make` :

  ```bash
  make build
  ```

  Cela place l'exécutable dans le dossier `target`.

### Complétions du shell

**tcping** embarque des scripts de complétion pour `bash`, `zsh`, `fish` et `PowerShell`
et les affiche avec `--completions <shell>`. Ils complètent les options, les
noms d'interface pour `-I` et les noms de fichier pour `--csv` et `--sqlite`.

Les paquets Linux, Homebrew, le paquet AUR, Nix et l'
[installation en une ligne](#linux---installation-en-une-ligne) les mettent en place pour vous.
Sinon, enregistrez celui de votre shell :

- `bash` :

  ```bash
  tcping --completions bash | sudo tee /usr/share/bash-completion/completions/tcping >/dev/null
  ```

- `zsh` :

  ```bash
  tcping --completions zsh | sudo tee /usr/share/zsh/site-functions/_tcping >/dev/null
  ```

- `fish` :

  ```bash
  mkdir -p ~/.config/fish/completions &&
    tcping --completions fish > ~/.config/fish/completions/tcping.fish
  ```

- `PowerShell`, en le chargeant depuis votre profil :

  ```powershell
  New-Item -ItemType Directory -Force (Split-Path $PROFILE) | Out-Null
  Add-Content $PROFILE "tcping --completions powershell | Out-String | Invoke-Expression"
  ```

Ouvrez ensuite un nouveau shell pour qu'elles soient prises en compte.

---

## Utilisation

Donnez à **tcping** une cible et un port, et il commence à sonder :

```bash
tcping www.example.com 443
tcping www.example.com:443          # hôte:port fonctionne aussi
tcping 192.168.1.1:80
tcping '[2001:db8::1]:443'          # entre guillemets, pour que le shell n'y touche pas
tcping https://www.example.com      # sondé en HTTP(S)
tcping udp://127.0.0.1 53           # sondé en UDP
tcping tcp://www.example.com 443    # tcp:// fonctionne aussi
```

Quelques exemples de ce que vous pouvez demander à une exécution :

```bash
tcping www.example.com 443 -c 5            # s'arrêter après 5 sondes
tcping www.example.com 443 -i 2 -t 5       # 2 secondes entre les sondes, 5 secondes de délai d'attente
tcping www.example.com 443 -I eth2         # envoyer les sondes depuis une interface donnée
tcping www.example.com 443 -4              # ou -6, pour imposer une famille d'adresses
tcping www.example.com 443 -D              # afficher un horodatage pour chaque sonde
tcping www.example.com 443 -r 5            # résoudre à nouveau le nom d'hôte après 5 échecs
```

Et les formats de sortie possibles, à la place de la sortie en couleur par défaut :

```bash
tcping www.example.com 443 -j              # JSON, ajoutez --pretty pour l'embellir
tcping www.example.com 443 --no-color      # brut, sans couleurs ANSI
tcping www.example.com 443 --csv out.csv   # fichier CSV
tcping www.example.com 443 --sqlite out.db # base de données sqlite3
```

L'image Docker accepte les mêmes cibles et les mêmes options :

```bash
docker run -it pouriyajamshidi/tcping:latest example.com 443
# Ou, depuis le registre de conteneurs GitHub
docker run -it ghcr.io/pouriyajamshidi/tcping:latest example.com:443
```

> [!TIP]
> Appuyez sur la touche `Entrée` pendant que le programme tourne pour voir le résumé de toutes les sondes sans l'arrêter, comme le montre la section [démos](#démos).

> [!NOTE]
> Consultez les **options disponibles** [ici](#options) pour une utilisation plus avancée.

### Sonder en HTTP(S)

Donnez une URL à la place d'un hôte et d'un port, et tcping la sonde en HTTP(S),
en indiquant le code d'état et le temps que le serveur a mis à répondre. Le corps
n'est pas téléchargé, donc une grosse page ne ralentit pas la sonde :

```bash
tcping https://www.example.com/health
```

Le port vient du schéma, `80` ou `443`, sauf si l'URL porte le sien
ou si vous en donnez un après elle :

```bash
tcping http://www.example.com:8080/health
# Même chose
tcping http://www.example.com/health 8080
```

`-v` montre tout ce qu'une sonde a recueilli : la version de HTTP, la version de TLS et la suite
de chiffrement, le nombre de jours restants sur le certificat et les temps de connexion, de négociation
TLS et d'arrivée du premier octet. `--insecure` ignore la vérification du certificat,
ce qu'il faut face à un certificat autosigné ou expiré :

```bash
tcping https://www.example.com/health -v
tcping https://self-signed.example.com --insecure
```

### Sonder à travers un proxy

`--proxy` fait passer les sondes TCP et HTTP(S) par un proxy. Les proxys
`http://`, `https://` et `socks5://` fonctionnent, et un identifiant se met dans
l'URL, par ex. `socks5://user:password@127.0.0.1:1080`. tcping ne lit pas les
variables d'environnement `HTTPS_PROXY`, `HTTP_PROXY` et `NO_PROXY`, une sonde
ne passe donc par un proxy que si vous le demandez :

```bash
tcping https://www.example.com --proxy http://proxy.example.com:3128
tcping db.example.com 5432 --proxy socks5://127.0.0.1:1080
```

Une sonde TCP demande au proxy de se connecter à la cible, et réussit quand le
proxy dit l'avoir fait. Son temps couvre le trajet jusqu'au proxy et celui du
proxy jusqu'à la cible. Beaucoup de proxys HTTP n'autorisent que les connexions
vers le port `443` : une sonde vers un autre port peut donc échouer alors que la
cible répond. Pour un proxy SOCKS5 vers un réseau où vous avez un accès SSH,
`ssh -D 1080 bastion.example.com` en ouvre un sur `127.0.0.1:1080`.

Le proxy se connecte lui-même au nom d'hôte de la cible. L'adresse source, et le
temps de connexion d'une sonde HTTP(S), sont donc ceux de la connexion au proxy,
et la première ligne de l'exécution le nomme. tcping ne résout pas le nom d'hôte
lui-même : la sortie n'affiche donc aucune adresse pour lui, et `-r` et
`--resolve-every-probe` ne fonctionnent pas avec un proxy. `-4`, `-6` et
`--dns-server` s'appliquent à la place au nom d'hôte du proxy lui-même. Les
sondes UDP ne passent pas par un proxy. Un proxy défini dans un fichier PAC doit
être donné à la main, car tcping ne lit pas les fichiers PAC.

### Sonder en UDP

```bash
tcping udp://127.0.0.1 53
```

UDP n'a pas de négociation, donc une sonde ne réussit que si l'autre bout répond.
Pour obtenir une réponse, lancez tcping comme serveur UDP de l'autre côté : il renvoie
chaque datagramme à son expéditeur :

```bash
# sur la machine sondée :
tcping --udp-server 127.0.0.1 9999
# sur la machine qui sonde :
tcping udp://127.0.0.1 9999
```

> [!NOTE]
> Une sonde UDP refusée signale `port unreachable` (port inaccessible), ce qui signifie
> que quelque chose nous bloque. Une sonde qui ne reçoit aucune réponse est aussi signalée
> comme un échec, mais elle ne peut pas distinguer un port ouvert qui reste silencieux d'un
> paquet perdu en route.
>
> Sous Windows, une sonde refusée apparaît toujours comme sans réponse. Go désactive les
> erreurs « port unreachable » sur les sockets UDP de Windows, donc tcping ne les voit jamais.

Chaque sonde UDP envoie son propre numéro comme charge utile, que le serveur renvoie.
Avec `-v`, ce numéro apparaît sur chaque ligne, ce qui permet de nommer une sonde perdue :

```text
Reply from 127.0.0.1 on port 9999 UDP_conn=4 time=1.276 ms
    reply echoed back probe 4
```

### Diffuser la sortie JSON vers un serveur

La sortie `JSON` n'a pas à être affichée. Donnez une adresse à `--json-url` et
chaque événement y est envoyé en `POST` au moment où il se produit, un événement par requête, pour qu'une machine
qui sonde puisse rendre compte à une machine qui collecte :

```bash
tcping www.example.com 443 --json-url http://localhost:8000/tcping
```

Les événements et leurs champs sont exactement ceux que `-j` affiche, donc tout ce qui
lit déjà la sortie redirigée lit aussi ceux-ci. Chaque événement diffusé indique aussi
la machine qui l'a envoyé dans son champ `source`, qui vaut par défaut le nom d'hôte de
cette machine, de sorte qu'un collecteur qui reçoit des événements de plusieurs machines peut savoir
de laquelle vient l'exécution qu'il regarde. Utilisez `--source-label` pour les nommer vous-même :

```bash
tcping www.example.com 443 --json-url http://collector:8000/tcping --source-label brussels
```

tcping peut aussi être le côté qui reçoit, comme `--udp-server` répond aux sondes
UDP. `--json-server` ne sonde pas : il écoute sur l'hôte et le port donnés
et affiche chaque événement qui lui est envoyé, sur n'importe quel chemin, donc la machine
qui collecte n'a rien à installer non plus :

```bash
# sur la machine qui collecte :
tcping --json-server 0.0.0.0 8000

# sur la machine qui sonde :
tcping www.example.com 443 --json-url http://collector:8000/tcping
```

Les événements vont sur sa sortie standard et tout le reste sur l'erreur standard, donc
`tcping --json-server 0.0.0.0 8000 > events.jsonl` laisse un fichier qui ne contient que
des événements, et tout ce qui n'est pas du `JSON` est refusé au lieu d'être affiché.

Tout serveur qui accepte un `POST` avec un corps `JSON` convient aussi, par exemple :

```python
from fastapi import FastAPI

app = FastAPI()


@app.post("/tcping")
async def tcping(event: dict) -> None:
    print(event)
```

Tous les événements d'une exécution passent par une seule connexion, qui reste ouverte entre
les sondes au lieu d'être recomposée pour chacune. Un serveur arrêté ou en difficulté
n'interrompt pas les sondes : tcping le signale une fois et continue, en abandonnant les
événements qu'il ne peut pas remettre.

### Envoyer les résultats par OTLP ou vers InfluxDB

Au lieu d'afficher chaque sonde, tcping peut l'envoyer comme métrique, ce qui transforme une
exécution en graphique et permet à plusieurs machines de surveiller la même cible. Par OTLP,
vers tout ce qui l'accepte, comme [Prometheus](https://prometheus.io/)
lancé avec `--web.enable-otlp-receiver` :

```bash
tcping www.example.com 443 --otlp http://localhost:9090/api/v1/otlp
```

Ou [Grafana Alloy](https://grafana.com/docs/alloy/latest/) et
OpenTelemetry Collector, qui peuvent le transmettre à plusieurs endroits :

```bash
tcping www.example.com 443 --otlp http://localhost:4318
```

Les backends hébergés veulent un jeton, chacun dans un en-tête qui lui est propre. Donnez l'en-tête
complet avec `--otlp-header`, ou dans la variable d'environnement `OTLP_HEADER` pour
le garder hors de l'historique de votre shell :

```bash
export OTLP_HEADER="Authorization: Bearer your-api-token" && \
  tcping www.example.com 443 --otlp https://otlp.example.com
```

Ou directement vers un serveur [InfluxDB](https://www.influxdata.com/) v2 ou v3 au
format line protocol :

```bash
export INFLUXDB_TOKEN=your-api-token && \
  tcping www.example.com 443 --influxdb http://localhost:8086 --influxdb-org home --influxdb-bucket tcping
```

Chaque sonde porte son temps d'aller-retour, son succès ou son échec, l'adresse en laquelle la
cible a été résolue et les compteurs, ainsi que les détails HTTP ou UDP quand la cible
en est une. Le bloc complet de statistiques que vous verriez normalement à la sortie
est envoyé toutes les 10 secondes en plus, pour qu'une exécution que personne ne
surveille le signale quand même. Utilisez `--stats-interval` pour changer cet intervalle.

Chaque machine se nomme dans les métriques avec l'étiquette `source`, qui vaut par défaut
son nom d'hôte, de sorte que plusieurs machines qui sondent la même cible tombent dans
leurs propres séries au lieu de se superposer. Utilisez `--source-label` pour les nommer
vous-même :

```bash
tcping www.example.com 443 --otlp http://localhost:4318 --source-label brussels
```

Tout ce qui est envoyé, comment l'interroger, et des environnements Prometheus, Alloy,
OpenTelemetry Collector et InfluxDB prêts à l'emploi, avec Grafana et tcping, que vous
pouvez lancer en une commande pour essayer sans installer de serveur d'abord, se trouvent dans
[docs/observability](../observability/README.md).

---

## Options

Les options peuvent être données avant ou après la cible, avec un ou deux
tirets, donc `-c 5` et `--c 5` sont la même option.

### Général

| Option      | Par défaut | Description                                 |
| ----------- | ---------- | ------------------------------------------- |
| `-h`        |            | Afficher les options disponibles et quitter |
| `--version` |            | Afficher la version et quitter              |
| `-u`        |            | Vérifier les mises à jour et quitter        |

### Sondes

| Option          | Par défaut | Description                                                                                           |
| --------------- | ---------- | ----------------------------------------------------------------------------------------------------- |
| `-c <n>`        | illimité   | S'arrêter après `<n>` sondes, quel que soit le résultat                                               |
| `-i <seconds>`  | `1`        | Intervalle entre les sondes. Nombre réel autorisé, par ex. `-i 0.5`                                   |
| `-t <seconds>`  | `1`        | Temps d'attente d'une réponse, en secondes. Nombre réel autorisé. `0` signifie pas de délai d'attente |
| `-4`            |            | N'utiliser que des adresses IPv4                                                                      |
| `-6`            |            | N'utiliser que des adresses IPv6                                                                      |
| `-I <name\|IP>` |            | Nom d'interface ou adresse IP depuis laquelle envoyer les sondes et les requêtes DNS                  |
| `--proxy <URL>` |            | Envoyer les sondes TCP et `HTTP(S)` à travers un proxy, par ex. `--proxy http://proxy.example.com:3128`. Les variables d'environnement `HTTPS_PROXY`, `HTTP_PROXY` et `NO_PROXY` ne sont pas utilisées |

> [!TIP]
> Sans les options `-4` et `-6`, tcping choisit une adresse IP au hasard parmi les résultats des requêtes DNS.

### Résolution de noms

| Option                    | Par défaut | Description                                                                                                                                                       |
| ------------------------- | ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-r <n>`                  | jamais     | Résoudre à nouveau le nom d'hôte de la cible après `<n>` sondes échouées, par ex. `-r 10`                                                                         |
| `--resolve-every-probe`   |            | Résoudre le nom d'hôte de la cible avant chaque sonde au lieu de seulement au démarrage. A priorité sur `-r` et n'a aucun effet quand la cible est une adresse IP |
| `--dns-server <IP>`       | du système | Serveur DNS personnalisé à utiliser. Une combinaison IP et port est autorisée, par ex. `--dns-server 1.1.1.1:53`                                                  |
| `--dns-timeout <seconds>` | `2`        | Temps d'attente d'une réponse DNS, en secondes. Nombre réel autorisé. `0` signifie pas de délai d'attente                                                         |

### Sortie du terminal

| Option                  | Par défaut | Description                                                                                                                                                                                                                                                                                                                                                        |
| ----------------------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `-D`                    |            | Afficher un horodatage pour chaque sonde                                                                                                                                                                                                                                                                                                                           |
| `--no-color`            |            | Ne pas colorer la sortie                                                                                                                                                                                                                                                                                                                                           |
| `--show-source-address` |            | Afficher l'adresse IP source et le port utilisés pour les sondes                                                                                                                                                                                                                                                                                                   |
| `--failures-only`       |            | N'afficher que les sondes échouées. Les réussies sont toujours comptées                                                                                                                                                                                                                                                                                            |
| `--no-stats`            |            | Ne pas afficher les statistiques à la fermeture du programme. Appuyer sur la touche **Entrée** les affiche toujours. Sans effet quand la sortie ne va pas vers le terminal                                                                                                                                                                                         |
| `-v`                    |            | Afficher tout ce qu'une sonde HTTP(S) a recueilli : la version de HTTP, la version de TLS et la suite de chiffrement, l'expiration du certificat et les temps de connexion, de TLS et du premier octet. Pour une cible UDP, affiche le numéro de la sonde et si la réponse l'a renvoyé, pour distinguer une sonde perdue des autres. Sans effet sur les cibles TCP |

### Sortie fichier et lisible par machine

| Option             | Par défaut | Description                                                                                                                                                                                                                                          |
| ------------------ | ---------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-j`               |            | Sortie au format `JSON`                                                                                                                                                                                                                              |
| `--pretty`         |            | Embellir la sortie `JSON`. Sans effet sans `-j`                                                                                                                                                                                                      |
| `--json-url <URL>` |            | Envoyer la sortie `JSON` à un serveur HTTP au lieu de l'afficher, un `POST` par événement, par ex. `--json-url http://localhost:8000/tcping`. Active la sortie `JSON` d'elle-même, donc `-j` n'est pas nécessaire                                    |
| `--json-server`    |            | Ne pas sonder. Écouter sur l'hôte et le port donnés et afficher chaque événement `JSON` qui lui est envoyé, pour qu'un tcping qui utilise `--json-url` ailleurs ait un endroit où envoyer son exécution, par ex. `tcping --json-server 0.0.0.0 8000` |
| `--csv <file>`     |            | Enregistrer la sortie dans un fichier `CSV`. Les statistiques vont dans le même nom avec le suffixe `_stats`                                                                                                                                         |
| `--csv-fixed-name` |            | Utiliser le nom de fichier de `--csv` tel quel, sans suffixe de date/heure, de sorte que des exécutions répétées écrasent le même fichier                                                                                                            |
| `--sqlite <file>`  |            | Enregistrer la sortie dans une base de données sqlite3, par ex. `--sqlite /tmp/tcping.db`. Indisponible sous Windows                                                                                                                                 |

### Métriques

| Option                       | Par défaut | Description                                                                                                                                                                                                                                                                                                |
| ---------------------------- | ---------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--otlp <URL>`               |            | Envoyer les résultats comme métriques vers un point de terminaison OTLP HTTP, comme [Grafana Alloy](https://grafana.com/docs/alloy/latest/) ou OpenTelemetry Collector, au lieu de les afficher, par ex. `--otlp http://localhost:4318`                                                                    |
| `--otlp-header <header>`     |            | En-tête HTTP supplémentaire à envoyer au point de terminaison `--otlp`, sous la forme `"Name: value"`, par ex. `"Authorization: Bearer <token>"`. Peut aussi être donné dans la variable d'environnement `OTLP_HEADER`, ce qui le garde hors de l'historique de votre shell                                |
| `--influxdb <URL>`           |            | Écrire les résultats vers un serveur [InfluxDB](https://www.influxdata.com/) v2 ou v3 au format line protocol au lieu de les afficher, par ex. `--influxdb http://localhost:8086`                                                                                                                          |
| `--influxdb-org <org>`       |            | Organisation InfluxDB dans laquelle écrire. Obligatoire avec `--influxdb`                                                                                                                                                                                                                                  |
| `--influxdb-bucket <bucket>` |            | Bucket InfluxDB dans lequel écrire. Obligatoire avec `--influxdb`                                                                                                                                                                                                                                          |
| `--influxdb-token <token>`   |            | Jeton d'API InfluxDB. Obligatoire avec `--influxdb`. Peut aussi être donné dans la variable d'environnement `INFLUXDB_TOKEN`, ce qui le garde hors de l'historique de votre shell                                                                                                                          |
| `--stats-interval <seconds>` | `10`       | Fréquence d'envoi des statistiques au point de terminaison OTLP ou à InfluxDB. Sans effet sans `--otlp` ou `--influxdb`                                                                                                                                                                                    |
| `--source-label <name>`      | nom d'hôte | Nommer cette machine dans les événements `JSON` et dans les métriques envoyées par OTLP ou vers InfluxDB, pour distinguer plusieurs machines qui sondent la même cible. Les résultats envoyés ailleurs portent le nom d'hôte quand ce n'est pas indiqué ; la sortie qui reste sur la machine ne porte rien |

### HTTP(S) et UDP

| Option          | Par défaut | Description                                                                                                                                                                                                 |
| --------------- | ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--insecure`    |            | Ne pas vérifier le certificat du serveur en sondant une cible `https://`. Utile pour les certificats autosignés ou expirés                                                                                  |
| `--udp-server`  |            | Ne pas sonder. Écouter sur l'hôte et le port donnés et renvoyer chaque datagramme UDP à son expéditeur, pour qu'une sonde UDP dirigée vers cette machine reçoive une réponse                                |

---

## Contribuer

Les pull requests sont les bienvenues pour corriger des bogues, ajouter de nouvelles fonctionnalités et aider
sur les tickets ouverts que l'on trouve [ici](https://github.com/pouriyajamshidi/tcping/issues).
Nombre actuel de tickets ouverts : ![GitHub issues](https://img.shields.io/github/issues/pouriyajamshidi/tcping.svg).

1. Choisissez un ticket avec lequel vous êtes à l'aise.
1. Faites un fork du dépôt.
1. Créez une branche.
1. Validez votre travail avec un commit.
1. Ajoutez des tests.
1. Lancez les tests avec `make test` et assurez-vous qu'ils réussissent.
1. Créez une pull request.

Veillez à ce que votre pull request **ne couvre qu'un seul ticket ou une seule fonctionnalité** et ne traite pas deux sujets ou plus. Cela nous simplifie la relecture de votre pull request et aide à garder un historique git propre.

Sauf si vous corrigez un tout petit problème, indiquez d'abord votre intention dans un **ticket** avant de commencer à travailler.

Pour essayer vos changements avec un mauvais réseau, `tools/netcond.sh` peut ajouter de la latence,
perdre une partie des paquets ou bloquer complètement une destination, puis défaire
tout ce qu'il a ajouté :

```bash
sudo ./tools/netcond.sh delay 1.1.1.1 800ms 200ms 30
sudo ./tools/netcond.sh loss 1.1.1.1 40
sudo ./tools/netcond.sh clear
```

## Aider le projet

Si tcping vous est utile, pensez à le partager avec votre réseau pour étendre sa portée et permettre à d'autres personnes d'en profiter aussi.

Vous pouvez en outre soutenir le projet avec les liens ci-dessous :

- Offrez-moi un café : [![Buy Me A Coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/pouriyajamshidi)
- GitHub Sponsors : [![Sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/pouriyajamshidi)

## Licence

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
