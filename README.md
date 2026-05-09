# papertrading

Equity-first paper trading system for Indian markets, designed to add new asset classes as plug-in modules (NSE F&O next).

## Docs

Start at:

- [docs/README.md](./docs/README.md)

## Dev (Phase 0 target)

Quickstart:

- `make up`
- `make migrate`
- `make check`

### Manual smoke test

- `curl -sf http://localhost:4000/healthz`
- `curl -sf http://localhost:6011/healthz` (expect `broker_adapter`, optional `replay_db`, `redis_instrument_cache` when DB/Redis are wired)
- `curl -sf http://localhost:4000/metrics | rg hello_requests_total`
- `curl -sf http://localhost:6011/metrics | rg md_ticks_ingested_total` (idle **`0`** until replay/live emits ticks)

### Observability

- **Grafana**: `http://localhost:3000` (default login `admin` / `admin`)
- **Prometheus**: `http://localhost:9090` (Targets API: `http://localhost:9090/api/v1/targets`)
- **Loki** queries (Grafana Explore → Loki):
  - `{container=~".*gateway.*"}`
  - `{container=~".*md.*"}`
  - `{container=~"/papertrading-.*"}`
- **Tempo** traces: use Grafana Explore → Tempo after hitting `/healthz` a few times.

### DB viewer (Beekeeper / DataGrip / DBeaver)

Postgres (from `infra/docker-compose.yml` defaults):

- **Host**: `localhost`
- **Port**: `5432`
- **Database**: `papertrading`
- **Username**: `papertrading`
- **Password**: `papertrading`
- **URL**: `postgres://papertrading:papertrading@localhost:5432/papertrading?sslmode=disable`
