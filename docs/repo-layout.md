# Repo Layout

## Top-level

```text
papertrading/
  apps/
    web/                          # React + Vite frontend
  services/
    gateway/                      # Node, Fastify, WS fan-out
    oms/                          # Node, Fastify
    risk/                         # Node, Fastify (orchestrator)
    portfolio/                    # Node, Fastify
    reports/                      # Node, Fastify + Puppeteer
    strategy-runner/              # Node, process supervisor
    go/
      matching/                   # Go matching engine
      md/                         # Go market-data gateway
      span/                       # Go SPAN calculator
      mm/                         # Go synthetic market makers
  packages/
    protos/                       # .proto files + generated Go/TS
    sdk/                          # TS client SDK (published or workspace)
    quant/                        # TS BS pricer, greeks, IV solver
    ui/                           # shared React components
    contracts/                    # Zod schemas + TS types
    bus/                          # Redis Streams wrapper (TS); Go equiv in services/go/bus
    config/                       # Shared env-loading, logging, tracing helpers
  infra/
    docker-compose.yml
    grafana/ prometheus/ loki/ tempo/   # dashboards, rules, configs
    seed/                         # bhavcopy, SPAN files, seed scripts
    migrations/                   # sqitch / dbmate SQL
    runbooks/                     # ops docs
  docs/                           # you are here
    adrs/
    phases/
    talking-points/
  .github/workflows/
  .env.example
  go.work
  package.json
  pnpm-workspace.yaml
  turbo.json
  justfile                        # or Makefile
```

## pnpm-workspace.yaml

```yaml
packages:
  - 'apps/*'
  - 'services/*'
  - 'packages/*'
```

## Turbo pipeline (abbreviated)

```json
{
  "pipeline": {
    "build": { "dependsOn": ["^build"], "outputs": ["dist/**"] },
    "test":  { "dependsOn": ["build"] },
    "lint":  {},
    "dev":   { "cache": false, "persistent": true }
  }
}
```

## Go workspace (`go.work`)

```go
go 1.22

use (
  ./services/go/matching
  ./services/go/md
  ./services/go/span
  ./services/go/mm
  ./services/go/bus
  ./packages/protos/go
)
```

## Protobuf layout

```text
packages/protos/
  events/
    tick.proto
    order.proto
    trade.proto
    margin.proto
  services/
    matching.proto     # gRPC service
    span.proto         # gRPC service
    md.proto           # gRPC + WS
  build.sh             # invokes protoc for both Go and TS
  go/                  # generated Go
  ts/                  # generated TS
```

## Conventions

### Naming

- Service folders: lowercase, single word.
- TS packages: `@papertrading/<name>` scoped.
- Go modules: `github.com/<you>/papertrading/services/go/<name>`.

### Imports

- No cross-service imports. Ever.
- Services import from `packages/*` only.
- `packages/*` never import from `services/*` or `apps/*`.

### Config

- Every service reads from env + validates via Zod (TS) or `envconfig` (Go).
- Shared env keys live in `packages/config`.
- `.env.example` is canonical; CI fails if a new `getEnv("X")` has no entry in the example.

### Logging

- Pino (TS) + slog (Go).
- Fields: `service`, `trace_id`, `span_id`, `user_id`, `order_id` when applicable.

### Error handling

- TS: Result-ish pattern — return `{ ok: true, value } | { ok: false, error }` for domain errors; throw only on programmer errors. Centralize reject-reason codes.
- Go: wrap errors with `fmt.Errorf("...: %w", err)`; sentinel errors for public API boundaries.

### Tests

- Co-located: `foo.ts` + `foo.test.ts`.
- Integration tests in `*.integration.test.ts`; run with `TEST_MODE=integration`.
- Go: `_test.go` in the same package.

## Scripts (`justfile` excerpt)

```makefile
up:        docker compose -f infra/docker-compose.yml up -d
down:      docker compose -f infra/docker-compose.yml down
seed:      pnpm -F seed run all
dev:       pnpm turbo run dev --parallel
test:      pnpm turbo run test && go test ./...
lint:      pnpm turbo run lint && golangci-lint run ./...
proto:     bash packages/protos/build.sh
migrate:   dbmate up
loadtest:  k6 run infra/k6/orders.js
```

## What each service must export as its `README.md`

- Purpose (one sentence).
- Env vars (ref to `.env.example`).
- Ports.
- `make dev` instructions.
- gRPC / HTTP endpoints.
- On-call runbook link.
- Known limitations.

## Growth rules

- New service? ADR required.
- New top-level folder? ADR required.
- New database? Probably no — keep the schema split inside Postgres.
- New language? Hard no in v1.

