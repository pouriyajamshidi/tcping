# Observing tcping

tcping can send its probes to [Grafana Alloy](https://grafana.com/docs/alloy/latest/)
or to [InfluxDB](https://www.influxdata.com/) instead of printing them, which
turns a run into a graph and lets several machines watch the same target.

This directory is a complete stack you can start in one command to try that
out, or to copy pieces of into your own setup.

## What is in here

| File | What it does |
| --- | --- |
| `compose.yml` | Alloy, Prometheus, InfluxDB and Grafana, wired together |
| `config.alloy` | Alloy taking OTLP from tcping and pushing it to Prometheus |
| `prometheus.yml` | Prometheus with nothing to scrape, it only receives |
| `grafana/provisioning/` | Both data sources, so nothing has to be clicked |
| `grafana/dashboards/tcping.json` | The dashboard |

> [!WARNING]
> Every password and token in here is a throwaway one, written in plain text
> so the stack starts without any setup. Do not reuse them anywhere, and do
> not put this stack on a network you do not control.

## Start it

```bash
cd docs/observability
docker compose up -d
```

That gives you:

- Grafana on <http://localhost:3000>, no login needed
- Alloy's UI on <http://localhost:12346>, to check it is receiving anything
- Prometheus on <http://localhost:9090>
- InfluxDB on <http://localhost:8086>, user `tcping`, password `tcping-dev-password`

Then point tcping at it. Through Alloy:

```bash
tcping --alloy http://localhost:4318 example.com 443
```

Or straight to InfluxDB:

```bash
export INFLUXDB_TOKEN=tcping-dev-token
tcping --influxdb http://localhost:8086 --influxdb-org home --influxdb-bucket tcping example.com 443
```

Both work at the same time. Open the `tcping` dashboard in Grafana and the
probes start showing up.

When you are done:

```bash
docker compose down -v
```

The `-v` throws the stored metrics away too. Leave it off to keep them. The
stack only holds 6 hours of data either way.

## Several machines probing the same target

Every probe carries a `source` label, which defaults to the hostname of the
machine that sent it. Two machines probing the same target therefore land in
their own series instead of on top of each other, and you can see the target
from both at once.

Use `--source-label` to name a machine yourself, which is also how to fake a
second machine on one laptop:

```bash
tcping --alloy http://localhost:4318 --source-label paris example.com 443
tcping --alloy http://localhost:4318 --source-label tokyo example.com 443
```

The **Source** dropdown at the top of the dashboard picks which ones to show.
It is filled from InfluxDB, so a machine that only sends to Alloy will not be
in the list, but **All** leaves every panel unfiltered and shows it anyway.

## The dashboard

The panels are split by where the data came from, because the two paths do not
store the same thing:

- **Round trip time**, **Probe result**, **Packet loss**, **Latency min /
  average / max / mdev** and **HTTP timings** read from InfluxDB.
- **Round trip time, through Alloy** and **Resolved address, through Alloy**
  read from Prometheus, and show the same run arriving the other way.

`allowUiUpdates` is on, so you can edit panels in Grafana and try things. The
edits live in Grafana's volume, not in the JSON file here, and `docker compose
down -v` throws them away. To keep one, export the dashboard JSON and write it
over `grafana/dashboards/tcping.json`.

## Using this outside the playground

The only part of this that is really tcping-specific is `config.alloy`: an
OTLP receiver on 4318, an exporter that turns the metrics into Prometheus
ones, and a remote write to wherever your Prometheus lives. Point the URL at
your own Prometheus and it works the same.

> [!NOTE]
> Prometheus needs `--web.enable-remote-write-receiver` for Alloy to be able
> to push to it.

For InfluxDB there is no middle piece at all, tcping writes line protocol
directly, so an existing v2 or v3 server only needs an org, a bucket and a
token.

The metrics, the fields and how to query them are described in the
[main README](../../README.md#usage), under the Alloy and InfluxDB examples.

## Nothing is showing up

- Alloy's UI on <http://localhost:12346> shows the health of each component.
  If the receiver is healthy but the remote write is not, Prometheus is the
  problem.
- A rejected write does not stop the run. tcping prints the error to stderr
  and says the metrics are being dropped, then keeps probing, so watch stderr
  rather than the probe output. A wrong InfluxDB token shows up this way.
- The statistics panels only fill in after the first statistics push, which is
  every 10 seconds by default. `--alloy-stats-interval` and
  `--influxdb-stats-interval` change that.
