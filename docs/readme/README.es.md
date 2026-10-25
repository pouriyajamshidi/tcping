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

Un programa ping multiplataforma que usa `TCP`, `UDP` o `HTTP(S)` en lugar de `ICMP`, inspirado en la utilidad ping de Linux.

> [!TIP]
> Este documento también está disponible en [English](../../README.md), [中文](README.cn.md), [日本語](README.ja.md), [한국어](README.ko.md), [فارسی](README.fa.md), [Deutsch](README.de.md), [Français](README.fr.md) y [Nederlands](README.nl.md).

- Una alternativa a `ping` donde `ICMP` está bloqueado, que sondea por `TCP`, `HTTP(S)` o `UDP`.
- Informa de la pérdida de paquetes y del mínimo, la media, el máximo y la desviación estándar de la latencia, igual que `ping`, además del tiempo más largo en línea y fuera de línea y de cuándo ocurrieron.
- Muestra las estadísticas en cualquier momento al pulsar la tecla `Enter`, sin detener el programa.
- Produce la salida **en color**, **en texto plano**, en **JSON**, **CSV** o **sqlite3**, o envía cada sondeo como métricas por **OTLP** (Grafana Alloy, OpenTelemetry Collector y otros) o a **InfluxDB**.
- Muestra el código de estado, la versión de TLS y el cifrado, la caducidad del certificado y los tiempos de conexión, de negociación TLS y del primer byte de cada sondeo `HTTP(S)`.
- Vuelve a resolver el nombre de host del destino tras un número de fallos (`-r`) o antes de cada sondeo (`--resolve-every-probe`), e indica cuánto tardó cada resolución. Útil para probar tu balanceo de carga por `DNS` o tu balanceador de carga global `(GSLB)`.
- Te deja elegir la **interfaz de origen**, el **tiempo de espera**, el **intervalo** y el **servidor DNS**, y forzar `IPv4` o `IPv6`.
- Numera por separado los sondeos _correctos_ y los _fallidos_, para ver los totales de un vistazo.

Echa un vistazo a las [demostraciones](#demostraciones) para hacerte una idea de cómo es **tcping**.

---

## Índice

- [Demostraciones](#demostraciones)
- [Descarga e instalación](#descarga-e-instalación)
- [Uso](#uso)
- [Opciones](#opciones)
- [Contribuir](#contribuir)
- [Ayuda al proyecto](#ayuda-al-proyecto)
- [Licencia](#licencia)

---

## Demostraciones

### Uso básico

![tcping](../Images/gifs/tcping.gif)

<details>
<summary><b>Haz clic para ver más demostraciones</b></summary>

### Sondeos con marca de tiempo por IPv4 (opciones `-D -4`)

![tcping timestamp example](../Images/gifs/tcping_timestamp.gif)

---

### Salida en texto plano, con estadísticas solo al pulsar Enter (opciones `--no-color --no-stats`)

![tcping plain output example](../Images/gifs/tcping_plain.gif)

---

### Volver a resolver el nombre de host (opción `-r`)

![tcping resolve example](../Images/gifs/tcping_resolve.gif)

---

### Salida JSON (opción `-j --pretty`)

![tcping json example](../Images/gifs/tcping_json_pretty.gif)

---

### Enviar la salida JSON a un colector (opciones `--json-url --json-server`)

![tcping JSON streaming example](../Images/gifs/tcping_json_stream.gif)

---

### Tiempo de resolución del nombre de host (opción `--resolve-every-probe`)

![tcping resolution timing example](../Images/gifs/tcping_dns_timing.gif)

---

### Interfaz de origen (opción `-I`)

![tcping interface example](../Images/gifs/tcping_interface.gif)

---

### Sondeos HTTP(S)

![tcping HTTP example](../Images/gifs/tcping_http.gif)

---

### Detalles de los sondeos HTTP(S) (opción `-v`)

![tcping HTTP verbose example](../Images/gifs/tcping_http_verbose.gif)

---

### Omitir la verificación del certificado (opción `--insecure`)

![tcping skip TLS example](../Images/gifs/tcping_skip_tls.gif)

---

### Sondeos a través de un proxy (opción `--proxy`)

![tcping proxy example](../Images/gifs/tcping_proxy.gif)

---

### Sondeos UDP contra un servidor UDP (opción `--udp-server`)

![tcping UDP example](../Images/gifs/tcping_udp.gif)

---

### Solo los sondeos fallidos (opción `--failures-only`)

![tcping failures only example](../Images/gifs/tcping_failures_only.gif)

</details>

---

## Descarga e instalación

| Plataforma                                                | Instalar con                                                                                                               |
| --------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| [Windows](#windows)                                       | `winget install pj.tcping`                                                                                                 |
| [macOS](#macos)                                           | `brew install pouriyajamshidi/tap/tcping`                                                                                  |
| [Debian, Ubuntu](#linux---repositorios-de-paquetes)       | `sudo apt install tcping`, después de añadir el repositorio                                                                |
| [Fedora, RHEL, CentOS](#linux---repositorios-de-paquetes) | `sudo dnf install tcping`, después de añadir el repositorio                                                                |
| [Alpine](#linux---repositorios-de-paquetes)               | `sudo apk add tcping`, después de añadir el repositorio                                                                    |
| [Arch, Manjaro](#linux---repositorios-de-paquetes)        | `yay -S pj.tcping`, aún no disponible                                                                                      |
| [Cualquier Linux](#linux---instalación-en-una-línea)      | una [instalación en una línea](#linux---instalación-en-una-línea) o un [archivo de paquete](#linux---archivos-de-paquetes) |
| [BSD](#bsd-y-linux---instalación-manual)                  | el [binario precompilado](#bsd-y-linux---instalación-manual)                                                               |
| [Nix](#nix)                                               | `nix profile install github:pouriyajamshidi/tcping`                                                                        |
| [Docker](#otras-formas)                                   | `docker run -it pouriyajamshidi/tcping example.com 443`                                                                    |
| [Go](#otras-formas)                                       | `go install github.com/pouriyajamshidi/tcping/v3@latest`                                                                   |

Los binarios son estáticos, es decir, llevan todo lo que necesitan dentro de un solo archivo
y no dependen de que haya ninguna biblioteca instalada en tu máquina. Están en la
[página de versiones](https://github.com/pouriyajamshidi/tcping/releases/latest/), para
_amd64_ y _arm64_.

Cuando termines la instalación, pasa a la sección de [uso](#uso).

### Windows

La mejor forma de instalar **tcping** en Windows es con el _Administrador de paquetes de Windows_ usando [WinGet](https://learn.microsoft.com/es-es/windows/package-manager/winget/?ref=github.com%2Fpouriyajamshidi%2Ftcping), que viene por defecto en prácticamente todas las máquinas con Windows _10_ y _11_ desde septiembre de 2020:

```powershell
winget install pj.tcping
```

Para instalarlo a mano, extrae el archivo zip descargado y copia `tcping.exe` al [PATH](https://www.howtogeek.com/118594/how-to-edit-your-system-path-for-easy-command-line-access/) de tu sistema, por ejemplo `C:\Windows\System32`.

> [!CAUTION]
> Windows Defender o algún antimalware puede marcar TCPING como amenaza por error. Es algo común en los programas escritos en Go. Consulta la declaración oficial del equipo de Go [aquí](https://go.dev/doc/faq#virus).

> [!WARNING]
> El formato de salida `--sqlite` (sqlite3) ya no está disponible en los binarios de Windows. Todas las demás opciones funcionan como se espera.

### macOS

```bash
brew install pouriyajamshidi/tap/tcping
```

### Linux - Instalación en una línea

Pega lo siguiente en tu terminal para descargar el último binario para tu arquitectura e instalarlo:

```bash
cd /tmp &&
ARCH=$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/') &&
curl -LO "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
tar -xf "tcping-linux-$ARCH.tar.gz" &&
sudo install tcping /usr/local/bin/ &&
tcping --version
```

Consulta [Autocompletado del shell](#autocompletado-del-shell) para configurar el autocompletado con la tecla Tab.

Si no tienes `curl`, cambia su línea por `wget`:

```bash
wget "https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-$ARCH.tar.gz" &&
```

### Linux - Repositorios de paquetes

Si añades el repositorio una vez, `apt`, `dnf` o `apk` se encargan de las
instalaciones y actualizaciones a partir de entonces, como con cualquier otro paquete.

**Debian**, **Ubuntu** y sus derivados:

```bash
sudo install -d /etc/apt/keyrings &&
  sudo curl -fsSL https://packages.thegraynode.io/keys/pouriyajamshidi.gpg -o /etc/apt/keyrings/pouriyajamshidi.gpg &&
  echo "deb [signed-by=/etc/apt/keyrings/pouriyajamshidi.gpg] https://packages.thegraynode.io/deb ./" | sudo tee /etc/apt/sources.list.d/pouriyajamshidi.list &&
  sudo apt update &&
  sudo apt install tcping
```

**Fedora**, **RHEL**, **CentOS** y sus derivados:

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

**Arch**, **Manjaro**, **EndeavourOS** y sus derivados lo tienen en el
[AUR](https://aur.archlinux.org/packages/pj.tcping), así que usa tu asistente favorito:

```bash
yay -S pj.tcping
```

> [!NOTE]
> El paquete del AUR aún no está disponible, porque el registro en el AUR está cerrado por ahora. Mientras tanto, usa el [archivo de paquete de Arch](#linux---archivos-de-paquetes) o la [instalación en una línea](#linux---instalación-en-una-línea).

### Linux - Archivos de paquetes

Sin añadir ningún repositorio, descarga el paquete de tu distribución desde la
[página de versiones](https://github.com/pouriyajamshidi/tcping/releases/latest/) e
instálalo:

| Distribución         | Paquete                    | Instalar                                            |
| -------------------- | -------------------------- | --------------------------------------------------- |
| Debian, Ubuntu       | `tcping-amd64.deb`         | `sudo apt install -y ./tcping-amd64.deb`            |
| Fedora, RHEL, CentOS | `tcping-amd64.rpm`         | `sudo dnf install -y ./tcping-amd64.rpm`            |
| Arch, Manjaro        | `tcping-amd64.pkg.tar.zst` | `sudo pacman -U ./tcping-amd64.pkg.tar.zst`         |
| Alpine               | `tcping-amd64.apk`         | `sudo apk add --allow-untrusted ./tcping-amd64.apk` |

Cambia `amd64` por `arm64` en máquinas ARM. Los cuatro instalan el binario y el
[autocompletado del shell](#autocompletado-del-shell) por ti. En las máquinas sin
`dnf` se puede usar `yum`, y Alpine necesita `--allow-untrusted` porque nuestros
paquetes no están firmados con una clave de Alpine.

### BSD y Linux - Instalación manual

Descarga el archivo para tu sistema operativo y tu arquitectura:

```bash
wget https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-freebsd-amd64.tar.gz
# O para máquinas Linux ARM64 y con cURL
curl -LO https://github.com/pouriyajamshidi/tcping/releases/latest/download/tcping-linux-arm64.tar.gz
```

Extráelo y copia el ejecutable al `PATH` de tu sistema, por ejemplo `/usr/local/bin/`:

```bash
tar -xvf tcping-freebsd-amd64.tar.gz &&
  chmod +x tcping &&
  sudo cp tcping /usr/local/bin/
```

Consulta [Autocompletado del shell](#autocompletado-del-shell) para configurar el autocompletado con la tecla Tab.

### Nix

**tcping** incluye un flake, así que en cualquier máquina que tenga [Nix](https://nixos.org/download/)
puedes ejecutarlo sin instalar nada:

```bash
nix run github:pouriyajamshidi/tcping -- example.com 80
```

O instalarlo de forma permanente:

```bash
nix profile install github:pouriyajamshidi/tcping
```

Las dos formas te dan el binario y el [autocompletado del shell](#autocompletado-del-shell).
El flake también tiene un `devShells.default` con todo lo necesario para trabajar en tcping,
que obtienes con `nix develop`.

### Otras formas

- Imágenes de `Docker`:

  ```bash
  docker pull pouriyajamshidi/tcping:latest
  # O
  docker pull ghcr.io/pouriyajamshidi/tcping:latest
  ```

- Con `go install`:

  > Necesita como mínimo la versión `1.27.3` de go

  ```bash
  go install github.com/pouriyajamshidi/tcping/v3@latest
  ```

- [x tcping](https://x-cmd.com/pkg/tcping):

  **Directamente, sin instalarlo**, en [x-cmd](https://www.x-cmd.com).

  ```bash
  x tcping example.com 80
  ```

  O instala `tcping` localmente con x-cmd, sin necesitar permisos de root ni tocar tu configuración global.

  ```bash
  x env use tcping
  tcping example.com 80
  ```

- Por último, puedes compilar el código tú mismo con el comando `make`:

  ```bash
  make build
  ```

  Esto deja el ejecutable en la carpeta `build`.

### Autocompletado del shell

**tcping** lleva scripts de autocompletado para `bash`, `zsh`, `fish` y `PowerShell`
y los muestra con `--completions <shell>`. Completan las opciones, los nombres de
interfaz para `-I` y los nombres de archivo para `--csv` y `--sqlite`.

Los paquetes de Linux, Homebrew, el paquete del AUR, Nix y la
[instalación en una línea](#linux---instalación-en-una-línea) los colocan por ti.
Si no, guarda el de tu shell:

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

- `PowerShell`, cargándolo desde tu perfil:

  ```powershell
  New-Item -ItemType Directory -Force (Split-Path $PROFILE) | Out-Null
  Add-Content $PROFILE "tcping --completions powershell | Out-String | Invoke-Expression"
  ```

  Las opciones de un solo guion como `-i` necesitan PowerShell 7, ya que Windows
  PowerShell 5.1 solo completa las de `--`. PowerShell no completa los nombres de
  interfaz para `-I`.

Después, abre un shell nuevo para que se carguen.

---

## Uso

Dale a **tcping** un destino y un puerto, y empieza a sondear:

```bash
tcping www.example.com 443
tcping www.example.com:443          # host:puerto también funciona
tcping 192.168.1.1:80
tcping '[2001:db8::1]:443'          # entre comillas, para que el shell no lo toque
tcping https://www.example.com      # sondeado por HTTP(S)
tcping udp://127.0.0.1 53           # sondeado por UDP
tcping tcp://www.example.com 443    # tcp:// también funciona
```

Algunas de las cosas que puedes pedirle a una ejecución:

```bash
tcping www.example.com 443 -c 5            # parar tras 5 sondeos
tcping www.example.com 443 -i 2 -t 5       # 2 segundos entre sondeos, 5 segundos de tiempo de espera
tcping www.example.com 443 -I eth2         # enviar los sondeos desde una interfaz concreta
tcping www.example.com 443 -4              # o -6, para forzar una familia de direcciones
tcping www.example.com 443 -D              # mostrar una marca de tiempo en cada sondeo
tcping www.example.com 443 -r 5            # volver a resolver el nombre de host tras 5 fallos
```

Y los formatos de salida que puede producir, en lugar del de color por defecto:

```bash
tcping www.example.com 443 -j              # JSON, añade --pretty para darle formato
tcping www.example.com 443 --no-color      # texto plano, sin colores ANSI
tcping www.example.com 443 --csv out.csv   # archivo CSV
tcping www.example.com 443 --sqlite out.db # base de datos sqlite3
```

La imagen de Docker acepta los mismos destinos y opciones:

```bash
docker run -it pouriyajamshidi/tcping:latest example.com 443
# O, desde el registro de contenedores de GitHub
docker run -it ghcr.io/pouriyajamshidi/tcping:latest example.com:443
```

> [!TIP]
> Pulsa la tecla `Enter` mientras el programa está en marcha para ver el resumen de todos los sondeos sin detenerlo, como se ve en la sección de [demostraciones](#demostraciones).

> [!NOTE]
> Consulta las **opciones disponibles** [aquí](#opciones) para un uso más avanzado.

### Sondear por HTTP(S)

Dale una URL en lugar de un host y un puerto, y tcping la sondea por HTTP(S),
informando del código de estado y de cuánto tardó el servidor en responder. El
cuerpo no se descarga, así que una página grande no hace más lento el sondeo:

```bash
tcping https://www.example.com/health
```

El puerto sale del esquema, `80` o `443`, salvo que la URL lleve el suyo propio
o le pases uno detrás:

```bash
tcping http://www.example.com:8080/health
# Lo mismo
tcping http://www.example.com/health 8080
```

`-v` muestra todo lo que recogió un sondeo: la versión de HTTP, la versión de TLS
y el cifrado, cuántos días le quedan al certificado y los tiempos de conexión, de
negociación TLS y del primer byte. `--insecure` omite la comprobación del
certificado, que es lo que quieres con uno autofirmado o caducado:

```bash
tcping https://www.example.com/health -v
tcping https://self-signed.example.com --insecure
```

### Sondear a través de un proxy

`--proxy` envía los sondeos TCP y HTTP(S) a través de un proxy. Funcionan los
proxies `http://`, `https://` y `socks5://`, y el usuario y la contraseña van en la
URL, por ejemplo `socks5://user:password@127.0.0.1:1080`. tcping no lee las
variables de entorno `HTTPS_PROXY`, `HTTP_PROXY` y `NO_PROXY`, así que un sondeo
solo pasa por un proxy cuando se lo pides:

```bash
tcping https://www.example.com --proxy http://proxy.example.com:3128
tcping db.example.com 5432 --proxy socks5://127.0.0.1:1080
```

Un sondeo TCP pide al proxy que se conecte al destino, y tiene éxito cuando el
proxy dice que lo hizo. Su tiempo incluye el camino hasta el proxy y el camino
del propio proxy hasta el destino. Muchos proxies HTTP solo permiten conexiones
al puerto `443`, así que un sondeo a otro puerto puede fallar aunque el destino
esté en marcha. Para tener un proxy SOCKS5 hacia una red a la que puedes entrar
por SSH, `ssh -D 1080 bastion.example.com` abre uno en `127.0.0.1:1080`.

El proxy se conecta por su cuenta al nombre de host del destino, así que la
dirección de origen, y el tiempo de conexión de un sondeo HTTP(S), son los de la
conexión al proxy, y la primera línea de la ejecución lo indica. tcping no
resuelve el nombre de host él mismo, así que la salida no muestra ninguna
dirección para él, y `-r` y `--resolve-every-probe` no funcionan con un proxy.
`-4`, `-6` y `--dns-server` se aplican en cambio al nombre de host del propio
proxy. Los sondeos UDP no pasan por un proxy. Un proxy definido en un archivo PAC
hay que indicarlo a mano, ya que tcping no lee archivos PAC.

### Sondear por UDP

```bash
tcping udp://127.0.0.1 53
```

UDP no tiene negociación, así que un sondeo solo tiene éxito cuando el otro
extremo responde. Para obtener una respuesta, ejecuta tcping como servidor UDP en
el otro extremo, que devuelve cada datagrama a quien lo envió:

```bash
# en la máquina que se sondea:
tcping --udp-server 127.0.0.1 9999
# en la máquina que la sondea:
tcping udp://127.0.0.1 9999
```

> [!NOTE]
> Un sondeo UDP rechazado informa de `port unreachable`, lo que significa que
> algo nos está bloqueando. Un sondeo que no recibe ninguna respuesta también se
> cuenta como fallo, pero no puede distinguir un puerto abierto que no contesta
> de un paquete que se perdió por el camino.
>
> En Windows, un sondeo rechazado siempre aparece como sin respuesta. Go
> desactiva los errores "port unreachable" en los sockets UDP de Windows, así que
> tcping nunca los ve.

Cada sondeo UDP envía su propio número como contenido, y el servidor lo devuelve.
Con `-v` se muestra ese número en cada línea, así que se puede saber qué sondeo se
perdió:

```text
Reply from 127.0.0.1 on port 9999 UDP_conn=4 time=1.276 ms
    reply echoed back probe 4
```

### Enviar la salida JSON a un servidor

La salida `JSON` no tiene por qué imprimirse. Dale una dirección a `--json-url` y
cada evento se envía allí con `POST` en cuanto ocurre, un evento por petición, así
que una máquina que sondea puede informar a otra que recoge los resultados:

```bash
tcping www.example.com 443 --json-url http://localhost:8000/tcping
```

Los eventos y sus campos son exactamente los que imprime `-j`, así que cualquier
cosa que ya lea la salida por tubería también lee estos. Cada evento enviado
indica además la máquina que lo mandó en su campo `source`, que por defecto es el
nombre de host de esa máquina, así que un colector que recibe eventos de varias
puede saber de quién es cada ejecución. Usa `--source-label` para ponerles nombre
tú mismo:

```bash
tcping www.example.com 443 --json-url http://collector:8000/tcping --source-label brussels
```

tcping también puede ser el lado que recibe, igual que `--udp-server` responde a
los sondeos UDP. `--json-server` no sondea: escucha en el host y el puerto
indicados e imprime cada evento que le envían, en cualquier ruta, así que la
máquina que recoge tampoco necesita nada más instalado:

```bash
# en la máquina que recoge:
tcping --json-server 0.0.0.0 8000

# en la máquina que sondea:
tcping www.example.com 443 --json-url http://collector:8000/tcping
```

Los eventos van a su salida estándar y todo lo demás a la salida de error, así
que `tcping --json-server 0.0.0.0 8000 > events.jsonl` deja un archivo con solo
eventos, y todo lo que no sea `JSON` se rechaza en lugar de imprimirse.

Cualquier servidor que acepte un `POST` con un cuerpo `JSON` sirve igual de bien, por ejemplo:

```python
from fastapi import FastAPI

app = FastAPI()


@app.post("/tcping")
async def tcping(event: dict) -> None:
    print(event)
```

Todos los eventos de una ejecución van por una sola conexión, que se mantiene
abierta entre sondeos en lugar de abrirse de nuevo para cada uno. Un servidor
caído o que da errores no detiene los sondeos: tcping lo avisa una vez y sigue,
descartando los eventos que no puede entregar.

### Enviar los resultados por OTLP o a InfluxDB

En lugar de imprimir cada sondeo, tcping puede enviarlo como métrica, lo que
convierte una ejecución en una gráfica y permite que varias máquinas vigilen el
mismo destino. Por OTLP, a cualquier cosa que lo acepte, como
[Prometheus](https://prometheus.io/) arrancado con `--web.enable-otlp-receiver`:

```bash
tcping www.example.com 443 --otlp http://localhost:9090/api/v1/otlp
```

O a [Grafana Alloy](https://grafana.com/docs/alloy/latest/) y al OpenTelemetry
Collector, que pueden reenviarlo a varios sitios:

```bash
tcping www.example.com 443 --otlp http://localhost:4318
```

Los servicios alojados piden un token, cada uno en su propia cabecera. Pasa la
cabecera completa con `--otlp-header`, o en la variable de entorno `OTLP_HEADER`
para que no quede en el historial de tu shell:

```bash
export OTLP_HEADER="Authorization: Bearer your-api-token" && \
  tcping www.example.com 443 --otlp https://otlp.example.com
```

O directamente a un servidor [InfluxDB](https://www.influxdata.com/) v2 o v3 como
line protocol:

```bash
export INFLUXDB_TOKEN=your-api-token && \
  tcping www.example.com 443 --influxdb http://localhost:8086 --influxdb-org home --influxdb-bucket tcping
```

Cada sondeo lleva su tiempo de ida y vuelta, si tuvo éxito, la dirección a la que
se resolvió el destino y los contadores, junto con los detalles HTTP o UDP cuando
el destino es de uno de esos tipos. Además, el bloque completo de estadísticas que
normalmente verías al salir se envía cada 10 segundos, así que una ejecución que
nadie está mirando también lo informa. Usa `--stats-interval` para cambiar ese
intervalo.

Cada máquina se identifica en las métricas con la etiqueta `source`, que por
defecto es su nombre de host, así que varias máquinas que sondean el mismo
destino quedan en sus propias series en lugar de pisarse. Usa `--source-label`
para ponerles nombre tú mismo:

```bash
tcping www.example.com 443 --otlp http://localhost:4318 --source-label brussels
```

Todo lo que se envía, cómo consultarlo, y entornos listos con Prometheus, Alloy,
OpenTelemetry Collector e InfluxDB, junto con Grafana y tcping, que puedes
arrancar con un solo comando para probar esto sin montar antes un servidor, están
en [docs/observability](../observability/README.md).

---

## Opciones

Las opciones pueden ir antes o después del destino, y con uno o dos guiones, así
que `-c 5` y `--c 5` son la misma opción.

### General

| Opción      | Por defecto | Descripción                              |
| ----------- | ----------- | ---------------------------------------- |
| `-h`        |             | Mostrar las opciones disponibles y salir |
| `--version` |             | Mostrar la versión y salir               |
| `-u`        |             | Buscar actualizaciones y salir           |

### Sondeo

| Opción          | Por defecto | Descripción                                                                                                                                                                                |
| --------------- | ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `-c <n>`        | sin límite  | Parar tras `<n>` sondeos, sea cual sea el resultado                                                                                                                                        |
| `-i <seconds>`  | `1`         | Intervalo entre sondeos. Se admiten números reales, por ejemplo `-i 0.5`                                                                                                                   |
| `-t <seconds>`  | `1`         | Tiempo de espera de una respuesta, en segundos. Se admiten números reales. `0` significa sin límite de espera. Los destinos HTTP(S) esperan `3` segundos si no se indica `-t`              |
| `-4`            |             | Usar solo direcciones IPv4                                                                                                                                                                 |
| `-6`            |             | Usar solo direcciones IPv6                                                                                                                                                                 |
| `-I <name\|IP>` |             | Nombre de interfaz o dirección IP desde la que enviar los sondeos y las consultas DNS                                                                                                      |
| `--proxy <URL>` |             | Enviar los sondeos TCP y `HTTP(S)` a través de un proxy, por ejemplo `--proxy http://proxy.example.com:3128`. No se usan las variables de entorno `HTTPS_PROXY`, `HTTP_PROXY` y `NO_PROXY` |

> [!TIP]
> Sin las opciones `-4` y `-6`, tcping elige al azar una dirección IP de entre las que devuelve la consulta DNS.

### Resolución de nombres

| Opción                    | Por defecto    | Descripción                                                                                                                                                               |
| ------------------------- | -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-r <n>`                  | nunca          | Volver a resolver el nombre de host del destino tras `<n>` sondeos fallidos, por ejemplo `-r 10`                                                                          |
| `--resolve-every-probe`   |                | Resolver el nombre de host del destino antes de cada sondeo en lugar de solo al arrancar. Tiene prioridad sobre `-r` y no hace nada cuando el destino es una dirección IP |
| `--dns-server <IP>`       | el del sistema | Servidor DNS a usar. Se admite una combinación de IP y puerto, por ejemplo `--dns-server 1.1.1.1:53`                                                                      |
| `--dns-timeout <seconds>` | `2`            | Tiempo de espera de una respuesta DNS, en segundos. Se admiten números reales. `0` significa sin límite de espera                                                         |

### Salida en el terminal

| Opción                  | Por defecto | Descripción                                                                                                                                                                                                                                                                                                                                        |
| ----------------------- | ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-D`                    |             | Mostrar una marca de tiempo en cada sondeo                                                                                                                                                                                                                                                                                                         |
| `--no-color`            |             | No colorear la salida                                                                                                                                                                                                                                                                                                                              |
| `--show-source-address` |             | Mostrar la dirección IP y el puerto de origen usados en los sondeos                                                                                                                                                                                                                                                                                |
| `--failures-only`       |             | Mostrar solo los sondeos fallidos. Los correctos se siguen contando                                                                                                                                                                                                                                                                                |
| `--no-stats`            |             | No mostrar las estadísticas al salir del programa. Pulsar la tecla **Enter** las sigue mostrando. No hace nada cuando la salida no va al terminal                                                                                                                                                                                                  |
| `-v`                    |             | Mostrar todo lo que recogió un sondeo HTTP(S): la versión de HTTP, la versión de TLS y el cifrado, la caducidad del certificado y los tiempos de conexión, de TLS y del primer byte. Con un destino UDP, muestra el número del sondeo y si la respuesta lo devolvió, para distinguir un sondeo perdido de los demás. No hace nada con destinos TCP |

### Salida a archivo y para máquinas

| Opción             | Por defecto | Descripción                                                                                                                                                                                                                                  |
| ------------------ | ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-j`               |             | Salida en formato `JSON`                                                                                                                                                                                                                     |
| `--pretty`         |             | Dar formato legible a la salida `JSON`. No hace nada sin `-j`                                                                                                                                                                                |
| `--json-url <URL>` |             | Enviar la salida `JSON` a un servidor HTTP en lugar de imprimirla, un `POST` por evento, por ejemplo `--json-url http://localhost:8000/tcping`. Activa la salida `JSON` por sí sola, así que no hace falta `-j`                              |
| `--json-server`    |             | No sondear. Escuchar en el host y el puerto indicados e imprimir cada evento `JSON` que le envíen, para que un tcping que usa `--json-url` en otra máquina tenga adónde enviar su ejecución, por ejemplo `tcping --json-server 0.0.0.0 8000` |
| `--csv <file>`     |             | Guardar la salida en un archivo `CSV`. Las estadísticas van a un archivo con el mismo nombre y el sufijo `_stats`                                                                                                                            |
| `--csv-fixed-name` |             | Usar el nombre de archivo de `--csv` tal cual, sin sufijo de fecha y hora, así que las ejecuciones repetidas sobrescriben el mismo archivo                                                                                                   |
| `--sqlite <file>`  |             | Guardar la salida en una base de datos sqlite3, por ejemplo `--sqlite /tmp/tcping.db`. No disponible en Windows                                                                                                                              |

### Métricas

| Opción                       | Por defecto    | Descripción                                                                                                                                                                                                                                                                                                           |
| ---------------------------- | -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--otlp <URL>`               |                | Enviar los resultados a un endpoint HTTP de OTLP, como [Grafana Alloy](https://grafana.com/docs/alloy/latest/) o el OpenTelemetry Collector, como métricas en lugar de imprimirlos, por ejemplo `--otlp http://localhost:4318`                                                                                        |
| `--otlp-header <header>`     |                | Cabecera HTTP extra que enviar al endpoint de `--otlp`, como `"Name: value"`, por ejemplo `"Authorization: Bearer <token>"`. También se puede pasar en la variable de entorno `OTLP_HEADER`, para que no quede en el historial de tu shell                                                                            |
| `--influxdb <URL>`           |                | Escribir los resultados en un servidor [InfluxDB](https://www.influxdata.com/) v2 o v3 como line protocol en lugar de imprimirlos, por ejemplo `--influxdb http://localhost:8086`                                                                                                                                     |
| `--influxdb-org <org>`       |                | Organización de InfluxDB en la que escribir. Obligatoria con `--influxdb`                                                                                                                                                                                                                                             |
| `--influxdb-bucket <bucket>` |                | Bucket de InfluxDB en el que escribir. Obligatorio con `--influxdb`                                                                                                                                                                                                                                                   |
| `--influxdb-token <token>`   |                | Token de la API de InfluxDB. Obligatorio con `--influxdb`. También se puede pasar en la variable de entorno `INFLUXDB_TOKEN`, para que no quede en el historial de tu shell                                                                                                                                           |
| `--stats-interval <seconds>` | `10`           | Cada cuánto enviar las estadísticas al endpoint de OTLP o a InfluxDB. No hace nada sin `--otlp` o `--influxdb`                                                                                                                                                                                                        |
| `--source-label <name>`      | nombre de host | Nombre de esta máquina en los eventos `JSON` y en las métricas enviadas por OTLP o a InfluxDB, para poder distinguir varias máquinas que sondean el mismo destino. Los resultados que se envían a otro sitio llevan el nombre de host si no se indica esta opción; la salida que se queda en la máquina no lleva nada |

### HTTP(S) y UDP

| Opción         | Por defecto | Descripción                                                                                                                                                           |
| -------------- | ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--insecure`   |             | No verificar el certificado del servidor al sondear un destino `https://`. Útil con certificados autofirmados o caducados                                             |
| `--udp-server` |             | No sondear. Escuchar en el host y el puerto indicados y devolver cada datagrama UDP a quien lo envió, para que un sondeo UDP dirigido a esta máquina reciba respuesta |

---

## Contribuir

Los pull requests son bienvenidos para corregir errores, añadir funciones nuevas y
ayudar con los issues abiertos, que puedes encontrar [aquí](https://github.com/pouriyajamshidi/tcping/issues).
Número actual de issues abiertos: ![GitHub issues](https://img.shields.io/github/issues/pouriyajamshidi/tcping.svg).

1. Elige el issue con el que te sientas cómodo.
1. Haz un fork del repositorio.
1. Crea una rama.
1. Haz commit de tu trabajo.
1. Añade tests.
1. Ejecuta los tests con `make test` y asegúrate de que pasan.
1. Crea un pull request

Asegúrate de que tu pull request **solo cubre un issue o una función concreta** y no trata dos o más issues a la vez. Así nos resulta más fácil revisarlo y el historial de git se mantiene limpio.

Salvo que estés arreglando algo realmente pequeño, comenta primero tu intención en
un **issue** antes de empezar a trabajar.

Para probar tus cambios en una red mala, `tools/netcond.sh` puede añadir latencia,
descartar una parte de los paquetes o bloquear por completo un destino, y deshacer
después todo lo que añadió:

```bash
sudo ./tools/netcond.sh delay 1.1.1.1 800ms 200ms 30
sudo ./tools/netcond.sh loss 1.1.1.1 40
sudo ./tools/netcond.sh clear
```

## Ayuda al proyecto

Si tcping te resulta útil, compártelo con tu red para que llegue más lejos y otras personas también se beneficien de él.

Además, puedes apoyar el proyecto con los enlaces de abajo:

- Invítame a un café: [![Buy Me A Coffee](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/pouriyajamshidi)
- GitHub Sponsors: [![Sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/pouriyajamshidi)

## Licencia

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
