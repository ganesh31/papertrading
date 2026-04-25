# Phase 0 — Foundation

**Week 1 · ~20 hrs**

Goal: a one-command dev environment, CI green on an empty app, observability wired into a hello-world service in each language. Everything after depends on this being boring and solid.

## Prerequisites

- macOS / Linux with Docker, pnpm 9, Node 20, Go 1.22.
- Accounts: GitHub (for Actions + GHCR), Angel One (API access enabled), free-tier Grafana Cloud optional (not used in v1).

## Deliverables (Definition of Done)

- Monorepo scaffolded with pnpm workspaces + Turborepo + Go workspaces.
- `docker-compose up` brings up Postgres+Timescale, Redis, Grafana, Prometheus, Loki, Tempo — all healthy.
- One Node service (`gateway`) and one Go service (`md`) stand up with `/healthz`, emit OTel traces visible in Tempo and a metric visible in Prometheus.
- `pnpm turbo run lint test build` is green locally and in CI.
- `.env.example` + secrets hygiene in place.
- ADR-0001 (monorepo boundaries), ADR-0003 (single-tenant v1), ADR-0004 (Go for hot path) written.
- Talking-points doc for Phase 0 written.

## Tasks

### 0.1 Repo scaffold

- `git init`, `pnpm init`, `pnpm-workspace.yaml`, `turbo.json`.
- `go work init`, create `services/go/md`, `services/go/matching` as skeletons.
- `.editorconfig`, `.gitignore`, `.env.example`, `LICENSE` (MIT).
- `README.md` (top-level) linking to `docs/`.
- Commit hygiene: `commitlint` + `husky` or `lefthook`.

### 0.2 Docker compose

- `infra/docker-compose.yml` with services:
  - `postgres` (image `timescale/timescaledb:latest-pg16`), healthcheck, volume.
  - `redis` 7.2-alpine, healthcheck.
  - `prometheus` (config in `infra/prometheus/prometheus.yml`).
  - `grafana` (provision dashboards + datasources).
  - `loki`, `tempo`, `promtail` for logs.
- All on a shared network, named `papertrading`.
- `.env` vars: `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, etc.

### 0.3 Migrations

- `dbmate` (Go binary, language-agnostic).
- Initial migrations:
  - `001_create_schemas.sql` (`oms`, `portfolio`, `md`, `reports`, `ref`).
  - `002_ref_users.sql` + seed the single user.
- `just migrate` wired.

### 0.4 Skeleton services

- `services/gateway`: Fastify, `/healthz`, `/metrics` (Prometheus), OTel instrumentation via `@opentelemetry/auto-instrumentations-node`.
- `services/go/md`: Gin/Echo (or stdlib), `/healthz`, `/metrics`, OTel via `go.opentelemetry.io/otel`.
- Both emit one business metric: `hello_requests_total`.

### 0.5 CI

- `.github/workflows/ci.yml`:
  - `setup-node`, `setup-go`, `pnpm install --frozen-lockfile`.
  - `turbo run lint test build`.
  - `go test ./...`.
  - Build container images on `main` merges; push to GHCR.
- Required status check on PRs.

### 0.6 Observability verification

- Spin everything up, hit `/healthz` on both services 100 times.
- Grafana dashboard `Trading Overview` shows the metric.
- Tempo shows a trace with a span from each service.
- Loki shows JSON logs.
- Screenshot into `docs/phases/phase-00-foundation.md` (optional).

### 0.7 ADRs written

- `0001-monorepo-boundaries.md`
- `0003-single-tenant-v1.md`
- `0004-go-for-hot-path.md`
(The numbering is intentional — 0002 is "event-sourced OMS", belongs to Phase 3.)

### 0.8 Runbook template

- `infra/runbooks/README.md` with template headings: Symptoms / Checks / Remediation / Escalation.

## Performance targets (this phase)

- `docker compose up` cold-start to all-healthy: < 60 s on laptop.
- CI: full pipeline < 5 min.
- Service boot: < 2 s each.

## Common pitfalls

- Skipping OTel until "later" — it's 10× harder to bolt on. Do it now, even for hello-world.
- Committing `.env` — add to `.gitignore` day 1.
- Turbo cache misses because of absolute paths in outputs — ensure `dist/`** etc. are relative.
- Node + Go both wanting port 4000 — use the port table in [02-architecture.md](../02-architecture.md).
- Grafana persistence lost on `down -v` — that's expected; provisioning should regenerate.

## Interview talking points (see `talking-points/phase-00.md`)

- Why monorepo for a project this small — shared schemas, single PR spanning FE/BE.
- Why OTel day 1 — non-optional for SRE-minded architecture.
- Why Turbo + Go workspaces vs. polyrepo — cost/benefit.
- Single-tenant as an *explicit* decision with a multi-tenant ADR, not an oversight.

## Resources

- Turborepo: [https://turbo.build/repo/docs](https://turbo.build/repo/docs)
- pnpm workspaces: [https://pnpm.io/workspaces](https://pnpm.io/workspaces)
- Go workspaces: [https://go.dev/ref/mod#workspaces](https://go.dev/ref/mod#workspaces)
- Docker Compose v2 cheatsheet.
- OpenTelemetry quickstarts:
  - Node: [https://opentelemetry.io/docs/languages/js/getting-started/nodejs/](https://opentelemetry.io/docs/languages/js/getting-started/nodejs/)
  - Go: [https://opentelemetry.io/docs/languages/go/getting-started/](https://opentelemetry.io/docs/languages/go/getting-started/)
- `dbmate`: [https://github.com/amacneil/dbmate](https://github.com/amacneil/dbmate)
- Grafana OSS: [https://grafana.com/oss/grafana/](https://grafana.com/oss/grafana/)
- lefthook (git hooks, language-agnostic): [https://github.com/evilmartians/lefthook](https://github.com/evilmartians/lefthook)

## What you should NOT do in this phase

- Don't start on matching engine — it will tempt you and derail CI hygiene.
- Don't optimize anything.
- Don't add auth.
- Don't add Kafka. Redis Streams comes in Phase 1 and that's soon enough.

## Exit checklist before starting Phase 1

- Everything in "Deliverables" above is checked.
- You can show a stranger `just up` and it works.
- You've written 2 ADRs and committed them.
- You've taken a screenshot of Grafana and pinned it somewhere.