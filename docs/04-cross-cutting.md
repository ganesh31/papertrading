# 04 — Cross-Cutting Concerns

Everything that cuts across phases. Set up in Phase 0 and kept healthy throughout.

## Observability

### Signals

- **Traces** (Tempo): every HTTP request, every gRPC call, every bus publish/consume. Context propagated via W3C Trace Context headers; Protobuf envelope carries trace IDs between services.
- **Metrics** (Prometheus): RED (Rate, Errors, Duration) for every endpoint; USE (Utilization, Saturation, Errors) for queues and DB.
- **Logs** (Loki): structured JSON; `user_id`, `order_id`, `trace_id`, `span_id` as standard labels.

### Critical metrics (name → type)


| Metric                                | Type      | Why                                      |
| ------------------------------------- | --------- | ---------------------------------------- |
| `order_ack_latency_ms{service}`       | histogram | Client perception; your interview number |
| `order_reject_total{reason}`          | counter   | Surveillance + UX                        |
| `order_to_trade_ratio{user}`          | gauge     | SEBI OTR compliance narrative            |
| `match_engine_queue_depth{symbol}`    | gauge     | Hotspot detection                        |
| `match_engine_event_loop_lag_ms`      | histogram | Single-writer health                     |
| `bus_consumer_lag_ms{consumer,topic}` | gauge     | Stream health                            |
| `md_tick_staleness_sec{symbol}`       | gauge     | Feed health                              |
| `db_query_duration_ms{query}`         | histogram | Slow-query detection                     |
| `gc_pause_ms{service}`                | histogram | Go GC story                              |
| `node_event_loop_lag_ms{service}`     | histogram | Node health                              |
| `span_calc_duration_ms`               | histogram | Risk path latency                        |
| `positions_reconcile_diff_total`      | counter   | Projection drift alarm                   |


### Dashboards (Grafana, checked into `infra/grafana/`)

1. **Trading Overview** — orders placed, accepted, rejected, filled; OTR; latency p50/p95/p99.
2. **Matching Engine** — per-symbol queue depth, matches/sec, book size, snapshot lag.
3. **Market Data** — tick rate, staleness, reconnect count per broker adapter.
4. **Risk** — SPAN calc latency, incremental margin calls/sec, margin-blocked totals.
5. **Infra** — DB, Redis, bus lag, disk, memory.

### SLOs (v1, on dev laptop)

- API availability: 99.5% during business hours (weekly window).
- p99 order-ack < 50 ms.
- p99 SPAN calc < 100 ms for a 10-position portfolio.
- MD tick staleness < 2 s for subscribed symbols during market hours.

SLO burn alerts are cheap wins — set them up.

## Testing strategy

### Pyramid

- **Unit (60%)**: matching engine price-time logic; BS pricer; SPAN scenario math; reject-reason mapping; idempotency.
- **Integration (30%)**: OMS↔Risk↔ME round-trip with Testcontainers Postgres + Redis; MD replay → candle aggregation; projection rebuild.
- **E2E (5%)**: Playwright — place order, see fill, cancel order, positions update.
- **Load (5%)**: k6 — sustained 10k orders/min, spike to 20k.

### Golden files (matching engine)

- Fixture order streams in `services/go/matching/testdata/scenarios/*.orders.jsonl`.
- Expected trade streams in `*.trades.jsonl`.
- Test: feed orders → assert exact trade stream (order + price + qty + side + taker).

### Property tests

- Matching engine invariants: book is always sorted; total filled qty ≤ order qty; sum of trade qty == fill qty; self-trade prevention never produces a crossing trade for the same account.
- Ledger invariant: sum(debit) == sum(credit) per user per day.
- SPAN monotonicity: adding a protective option leg never *increases* total margin.

### Chaos tests (simple, effective)

- Kill matching engine mid-fill → boot → assert book reconstructs from event log.
- Drop bus for 5 s → assert nothing double-counted.
- Postgres restart → assert OMS returns 503 cleanly, recovers.

## Security & secrets

- No secrets in code or committed env files. `.env.example` checked in; `.env` git-ignored.
- Broker API keys in `.env`; loaded via `@fastify/env` / `os.Getenv`.
- Planned path for production: SOPS + age-encrypted files in repo, or Doppler free tier.
- JWT secret rotated via env variable.
- Rate limit: 50 orders/min/user (Redis token bucket); 500 reads/min/user.
- Outbound: only whitelisted hosts (Angel One, NSE). Enforced at egress if deployed.

## Release & versioning

- **Monorepo**; services released together in v1. Tag `vYYYY.MM.DD-<seq>`.
- Each commit must pass: lint + test + build in CI.
- Container images tagged with `git sha` + `semver`.
- DB migrations forward-only; rollbacks via compensating migrations.

## ADR process

- `docs/adrs/NNNN-kebab-title.md` — Nygard template (Context / Decision / Consequences / Alternatives).
- One ADR per non-obvious decision. If a teammate (or future you) would ask "why?", write it.
- ADRs are immutable after merge; supersede via new ADR referencing the old (`Supersedes: 0003`).

## Talking-points process

- Every phase ends with `docs/talking-points/<phase>.md`.
- Format per topic: **Question** → **Answer (90 sec)** → **Trade-off I'd defend** → **What I'd do differently at scale**.
- These become your interview warm-up pack.

## Coding conventions

### TypeScript

- `strict: true`, `noUncheckedIndexedAccess: true`.
- No `any` in committed code; `unknown` + narrowing.
- Biome or ESLint + Prettier.
- No default exports.
- Zod for every external boundary; never trust inbound JSON.

### Go

- `gofmt`, `golangci-lint` with `errcheck`, `gosec`, `revive`, `staticcheck`.
- Return errors, don't panic on business logic.
- Context plumbed through every request.
- No global mutable state.

### SQL

- Migrations are idempotent-safe or explicitly not (stated at top).
- Named constraints.
- Explain-analyze every query that hits a table > 100k rows.

## Git hygiene

- Conventional commits (`feat:`, `fix:`, `chore:`, `refactor:`, `docs:`).
- Feature branches; squash-merge.
- PR template: summary, screenshots/tracebacks, ADR link if applicable, test plan.
- Pre-commit: lint + unit tests for changed packages (via Turbo).

## Documentation hygiene

- Every service has a `README.md` with: purpose, ports, envs, `make dev` instructions, on-call runbook link.
- Every phase doc is updated when reality diverges from plan.
- The repo `README.md` links here.

## Compliance framing (talk about, don't implement)

Keep a `docs/compliance.md` (v2) stub noting:

- SEBI record retention (7 years).
- Audit trail (event log covers it).
- OTR reporting (would be submitted daily in a real broker).
- Peak margin reporting (4 snapshots/day).
- Upfront margin collection norm.

Interviewers probe these. Knowing they exist (and where they'd plug into your system) matters more than implementing them.

## Runbooks (write these the first time something breaks)

- `infra/runbooks/matching-engine-crash.md`
- `infra/runbooks/md-feed-stale.md`
- `infra/runbooks/kill-switch.md`
- `infra/runbooks/projection-drift.md`

One page each, command-level detail, updated after every incident.