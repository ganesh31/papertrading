# Redis Stream `ticks.v1` — consumer groups

The `md` service **XADD**s each normalized tick to stream **`ticks.v1`** (override with `MD_TICKS_STREAM`).
Payload field: **`payload`** — JSON object aligned with WebSocket tick wire (`instrumentId`, `ts`, `ltp`, bid/ask, volume, `source`, …).

Approximate **MINID** trim keeps roughly **one hour** of history by stream IDs (see `MD_TICKS_STREAM_RETENTION_SEC`).
**Postgres `md.ticks` remains the durable store**; this stream is a firehose for subscribers.

## Consumer group labels (Phase 1 contract)

| Group name      | Intended subscriber                          |
|-----------------|----------------------------------------------|
| `mm`            | Market-making / liquidity simulation (later) |
| `strategy`      | Strategy runner / signals                    |
| `surveillance`  | Limits, alerts, abuse detection              |

Create groups once per Redis (idempotent: ignore `BUSYGROUP`).

```bash
# Replace host/port/db index as needed; stream must exist (md emits first) or add MKSTREAM.
redis-cli -u "$REDIS_URL" XGROUP CREATE ticks.v1 mm        $ MKSTREAM
redis-cli -u "$REDIS_URL" XGROUP CREATE ticks.v1 strategy $ MKSTREAM
redis-cli -u "$REDIS_URL" XGROUP CREATE ticks.v1 surveillance $ MKSTREAM
```

Use `$` so each group only sees **new** entries after creation; use `0` if you need backlog from stream start.

## Observability

Prometheus (on `md` `/metrics`):

- `md_ticks_stream_published_total`
- `md_ticks_stream_publish_errors_total`
- `md_ticks_ingested_total`, `md_ws_*`, `replay_*`, etc. — see **Market Data (md)** Grafana dashboard (`infra/grafana/dashboards/market-data-md.json`).

## Env (`md`)

| Variable                        | Default      | Purpose                              |
|---------------------------------|-------------|--------------------------------------|
| `MD_TICKS_STREAM`               | `ticks.v1`| Stream key                           |
| `MD_TICKS_STREAM_RETENTION_SEC`| `3600`      | Trim window (approx MINID)           |
| `MD_TICKS_PUBLISH_TIMEOUT_MS`  | `75`        | Per-tick XADD timeout                |
