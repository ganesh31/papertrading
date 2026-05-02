# 00 — Overview

## North star

Build a paper-trading system that is **structurally indistinguishable from a real Indian broker** (Zerodha / Angel One / Upstox) on the inside:

- Own Central Limit Order Book (CLOB) matching engine with price-time priority.
- Event-sourced Order Management System (OMS).
- Pre-trade risk with VAR + ELM for equity cash, and SPAN + Exposure for derivatives.
- Daily MTM, T+1 settlement, corporate actions.
- Kite-clone UI + programmatic strategy SDK with live/backtest parity.

Use Angel One SmartAPI + NSE historical data behind an adapter. **No order is ever sent to a real exchange**; every order is matched in-house against synthetic market-maker liquidity seeded from real prices.

## Core principle: equity-first, asset-class plug-ins

Build a complete **Equity cash** broker slice first (orders → fills → positions/holdings → P&L → ledger → UI). Everything beyond equity is implemented as an **asset-class module plug-in** (first: NSE F&O) without refactoring core services.

Terminology used throughout the docs:

- **Asset class**: top-level product family. v1 starts with `EQUITY` and later adds `DERIVATIVES` as a plug-in (NSE F&O).
- **Segment**: exchange segment / market. Examples: `NSE_EQ`, `NFO`.
- **Instrument type**: contract kind. Examples: `EQ`, `FUT`, `OPT`.
- **Instrument**: a concrete tradable contract identified by an `instrument_id` (e.g. `INFY` equity, `NIFTY24MAYFUT`, `NIFTY24MAY24500CE`).

## Why this project

Two goals, both served by the same build:

1. **Learning / interview prep** for Senior Architect roles in Indian wealth management & broker tech. Every module ends with an ADR + a talking-points doc you can use verbatim.
2. **Actual usable paper-trading tool** you and others can trade on — hosted on a tiny VPS.

## Success criteria (v1 "done")

Must-have:

- **Equity cash works end-to-end**: place/cancel/modify, fills, positions + holdings (T+1), realised & unrealised P&L, double-entry ledger, and the UI can trade it.
- **Asset-class plug-in architecture is real**: NSE F&O is added as a module (not a refactor), and futures + options trade end-to-end with correct lifecycle semantics.
- LIMIT, MARKET, SL, SL-M, IOC, FOK order types match correctly (golden-file tests pass).
- Pre-trade margin blocks over-leveraged orders:
  - Equity cash: VAR + ELM (with sensible defaults if a daily file is missing).
  - NSE F&O: SPAN + Exposure.
- Kite-clone UI: watchlist, L5 depth, TradingView chart, order pad, option chain, positions, holdings, funds.
- Strategy SDK with one sample strategy runnable in **both live and replay** modes, deterministic in replay.
- Observability: p50/p95/p99 order-ack latency histogram, OTR, reject-reason counters, event loop lag, GC pause — visible on Grafana.
- Reproducible load test: 10k orders/min sustained on localhost, p99 order-ack < 50 ms.
- Demo deployed at a public URL with a 5-min loom.
- ≥ 15 ADRs, 1 talking-points doc per phase.

Could-have (post-v1): multi-tenant auth, currency F&O, GTT/BO/CO, basket orders, mobile UI, Kafka migration, k8s deploy.

Won't-have (v1): commodity derivatives, MF/IPO/Corp FD (your day-job domain — not the learning target), real money integration.

## Single-user scope — explicit trade-offs

### Pros

- ~30% less boilerplate (no auth, RBAC, tenancy, quotas).
- Schemas + APIs stay readable.
- Faster iteration inside a 3-month window.
- Focus budget on the hard parts: matching, SPAN, F&O lifecycle.

### Cons (real interview gaps you must acknowledge)

- Broker systems are inherently multi-tenant. Noisy neighbour, fair queuing, per-user rate limits, data isolation, cross-tenant reports are **core Architect interview topics** you won't have shipped.
- Harder to sell as a SaaS portfolio piece.
- "Scale it to 1M users" answers stay theoretical unless you've at least *designed* the multi-tenant version.

### Mitigation (do these from day 1)

- Every table has `user_id` even if always `= 1`. No schema refactor later.
- Every request carries `X-User-Id` via middleware; stub today, JWT tomorrow.
- Rate limits keyed on `user_id`, not IP.
- Matching engine partitions orders by `(symbol, user_id)` even if the second dimension is trivial.
- Ship `docs/adrs/0003-single-tenant-v1.md` explicitly describing the exact diff to multi-tenant v2 — that ADR **is an interview artifact**.

## Timeline

Budget: ~3 hours/day, target ~14–16 weeks. Phases overlap slightly; one demoable milestone per phase.


| Week  | Phase               | Milestone                                                               |
| ----- | ------------------- | ----------------------------------------------------------------------- |
| 1     | 0 — Foundation      | `docker-compose up` → Grafana + Postgres + Redis + Timescale green      |
| 2     | 1 — Market Data     | `nse_replay` default adapter; seeded data replays at 100× speed¹        |
| 3–4   | 2 — Matching Engine | 5 symbols tradeable against synthetic MMs; golden-file tests pass       |
| 5     | 3 — OMS + Risk      | OMS is event-sourced; Equity module validates + risk-checks cash orders |
| 6     | 4 — Positions/P&L   | Equity positions/holdings/P&L/ledger are correct end-to-end             |
| 7–8   | 5 — Frontend v1     | **Equity broker slice demoable**: trade from UI, see fills + P&L        |
| 9     | 6 — Futures (plug-in) | NFO futures module: contracts + daily MTM + expiry                      |
| 10–11 | 7 — Options (plug-in) | NFO options module: chain + greeks + expiry settlement                  |
| 12–13 | 8 — SPAN (plug-in)  | NFO risk module: SPAN + Exposure matches NSE daily files within tolerance |
| 14    | 9 — Strategy SDK    | SMA crossover runs in live and replay, same outcomes                    |
| 15    | 10 — Settlement     | EOD jobs, T+1, contract notes; extended for NFO lifecycle               |
| 16+   | 11 — Hardening      | Load + chaos tests, deployment, demo                                    |

¹ **Replay-first**: all of Phases 1–10 run on historical data via the `nse_replay` adapter — deterministic, offline, and fast-forwardable. The `angel_live` adapter is stubbed in Phase 1 and fully implemented in Phase 11 for **milestone / ops** reasons, **not** because Angel SmartAPI market data requires a datacenter static IP (IP whitelist is mainly an **order API** concern; live MD can work from a registered **current** IP). See [phases/phase-01-market-data.md](./phases/phase-01-market-data.md#angel-smartapi-orders-vs-market-data-and-ip).

## Non-goals (call these out)

- Not building a real broker — no SEBI registration, no actual trading, no KYC/AML beyond a stub.
- Not latency-competitive with actual NSE (they run at microsecond latencies on co-located servers). Target is **plausibly realistic**, not state-of-the-art.
- Not a trading-strategy research product. Strategy SDK is a proof of interface, not a Lean/QuantConnect replacement.

## Reading order for a reviewer / interviewer

1. This doc.
2. [05-nse-domain-primer.md](./05-nse-domain-primer.md).
3. [02-architecture.md](./02-architecture.md).
4. Any one phase doc they probe on.
5. Corresponding ADR under [adrs/](./adrs/).
6. Corresponding talking-points doc.

If they only read one thing, point them at [phases/phase-02-matching-engine.md](./phases/phase-02-matching-engine.md) and [phases/phase-08-span-margin.md](./phases/phase-08-span-margin.md) — these two carry most of the architectural differentiation.