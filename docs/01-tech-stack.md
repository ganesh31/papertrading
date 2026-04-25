# 01 — Tech Stack

Every choice justified; every choice has an "escape hatch" so Phase 11 can evolve without a rewrite.

## Languages

### Go (1.22+) — hot path

Used for: **matching engine**, **market data gateway**, **SPAN calculator**, **synthetic market makers**.

Why:

- Goroutines + channels map cleanly onto the "one actor per symbol" design for the CLOB.
- Predictable GC pauses under sustained allocation — critical for p99 latency targets.
- Compiles to a static binary → tiny containers → fast deploys.
- Strong `encoding/binary` and `sync/atomic` stories for the binary tick protocol Angel One uses.
- Standard broker tech stack in India (Zerodha's C++/Go mix is public knowledge; most new services in the space are Go).

Escape hatch: if you want to go further, Rust or C++ for the matching engine; the `protos/` interface doesn't care.

### TypeScript / Node 20 — everything else

Used for: **API gateway**, **OMS**, **risk orchestrator**, **portfolio**, **reports**, **strategy runtime**, **frontend**, **SDK**.

Why:

- Your fluency → velocity. You finish the project.
- Single language front-to-back-to-SDK → share Zod schemas in `packages/contracts`.
- Fastify is fast enough for OMS throughput you'll hit on your laptop.
- Massive ecosystem for the boring stuff (PDF gen, cron, OAuth-when-you-need-it).

Escape hatch: OMS can be rewritten in Go later without protocol change, since all inter-service messages are Protobuf.

## Web framework: Fastify (not Express, not NestJS)

- Fastify: 2–3× Express throughput, first-class schema validation, built-in lifecycle hooks, great WebSocket support via `@fastify/websocket`.
- NestJS is idiomatic but heavier — dependency injection buys you little here, and decorators slow you down in a learning project.

## Frontend: React 18 + Vite + TS

- Vite dev server is instant; HMR matters when iterating on a depth widget.
- State: **TanStack Query** for server state, **Zustand** for local state. No Redux.
- UI: **shadcn/ui** + **Tailwind** for composition; **Radix** primitives underneath.
- Charting: **TradingView `charting_library`** (free after signing the agreement) — the only way to match Kite's chart UX without reinventing it.
- Forms: **react-hook-form** + **zod**. Same zod schemas from `packages/contracts`.
- Routing: **TanStack Router** (type-safe) or React Router 6 if you're faster there.

## Data stores

### Postgres 16 + TimescaleDB extension

- Postgres: orders, trades, positions projection, contract master, users, ledger.
- Timescale (as a Postgres extension): ticks + 1m/5m/15m/60m/1d candle continuous aggregates.
- One DB instance = one less moving part in docker-compose.

### Redis 7

- Order book snapshots (ZSET for price levels gives O(log N) best-bid/ask).
- Session/rate-limit storage.
- Idempotency keys (TTL 60s on client order IDs).
- **Redis Streams** as event bus (v1).

### Event bus: Redis Streams → Kafka (deferred)

- Start with Redis Streams. Consumer groups give you at-least-once; XADD + XREADGROUP cover 95% of Kafka ergonomics.
- Abstract behind a `packages/contracts/bus.ts` `Bus` interface with methods `publish`, `subscribe`, `ack`.
- Phase 11 can swap to **Redpanda** (Kafka-compatible, single binary, lighter than Kafka) — interface unchanged.

## Inter-service contracts

- **Protobuf** for Go↔Node service-to-service (matching engine, MD, SPAN). Versioned under `packages/protos/`. `protoc-gen-go` + `ts-proto`.
- **OpenAPI 3.1** for FE↔BE. Generated via `fastify-swagger` from Zod schemas; client codegen via `openapi-typescript`.
- **Zod** schemas as the single source of truth for TS-side validation.

## Observability (OTel → OSS stack)

- **OpenTelemetry SDK** in every service. Traces + metrics + logs.
- **Tempo** for traces, **Prometheus** for metrics, **Loki** for logs, **Grafana** to visualize.
- **k6** for load tests; metrics pushed into Prometheus during runs.
- No paid vendor in v1 (Datadog/New Relic: post-v1 if hosted).

## Testing

- **Vitest** for TS unit + integration.
- **Go `testing`** + **testify** assertions.
- **Testcontainers** for Postgres/Redis-backed integration tests in both languages.
- **Playwright** for two critical E2E flows (place order → see in positions; cancel order).
- **k6** for load.
- **Golden-file** tests for matching engine outputs (feed fixture order stream → assert trade stream byte-for-byte).

## Build, monorepo, CI

- **pnpm** workspaces + **Turborepo** for TS.
- **Go workspaces** (`go.work`) for the Go services.
- **GitHub Actions**: lint, test, build, container images pushed to GHCR.
- **Changesets** for versioning `packages/sdk` if/when you publish it.

## Runtime, dev environment

- **docker-compose** for v1 — everything on one laptop.
- `.env` + `.env.example`; never commit secrets. Use **direnv** + **mise** (or asdf) for tool-version pinning.
- Makefile or `just` at repo root with targets: `up`, `down`, `seed`, `test`, `lint`, `proto`, `e2e`, `loadtest`.

## Deployment (post-v1)

- Single VPS (Hetzner CX22 or Contabo) running docker-compose behind **Caddy** (free HTTPS).
- Or: single-node **k3s** if you want k8s on the resume.
- Cost target: ≤ ₹800/month.

## What you're explicitly NOT using (and why)


| Rejected      | Reason                                                                                                                                                 |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Nest.js       | DI + decorators buy little here; Fastify + Zod is cleaner.                                                                                             |
| Prisma        | Great tool, but the event-sourcing pattern is cleaner with raw SQL + a thin layer (`postgres.js` or `slonik`). Projections are easier to own manually. |
| MongoDB       | Your data is relational (double-entry ledger, FK-heavy).                                                                                               |
| ClickHouse    | Overkill for v1 scale; Timescale hypertables suffice. Reconsider if you ever replay 10+ years of ticks.                                                |
| Kafka day 1   | Operationally heavy; Redis Streams covers v1.                                                                                                          |
| gRPC for FE   | Browsers still need gRPC-Web proxy; REST + WS is simpler. gRPC stays between Go ↔ Node services.                                                       |
| GraphQL       | Fixed, well-understood resource shape. REST is fine.                                                                                                   |
| Next.js       | Trading UI is a dashboard, not a content site. SPA (Vite) is the right call.                                                                           |
| Redux         | Replaced by Zustand + TanStack Query for this app's shape.                                                                                             |
| Auth0 / Clerk | Single-tenant v1. A stubbed `user_id=1` middleware is sufficient.                                                                                      |


## Version floor

Fix minimum versions in `package.json` `engines` and Go `go.mod`. Example:

- Node: `>=20.11.0`
- pnpm: `>=9`
- Go: `1.22`
- Postgres: `16`
- Redis: `7.2`
- Docker Compose: v2

