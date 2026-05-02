# 0005 — Broker-adapter abstraction for market data

- Status: Accepted
- Date: 2026-05-01
- Deciders: Project owner

## Context

Market data can come from **replay** (deterministic, offline-friendly, no broker credentials) or **live broker feeds** (Angel SmartAPI WebSocket after session login, e.g. TOTP). Angel’s **IP whitelist** is primarily associated with **order-placement** APIs, not with replay and often not with **market-data** WebSocket from a registered client IP—confirm in Angel’s dashboard and docs. Downstream components (normalizer, persistence, WS fan-out, strategies) should depend on a **single contract**, not on vendor SDKs or replay-specific loops.

Phase 1 is replay-first; full **`angel_live`** implementation is deferred to Phase 11 for schedule and hardening, not because residential IP inherently blocks SmartAPI market data. We still need a stable seam so switching `MD_ADAPTER` requires **no code changes** outside the adapter registry.

## Decision

Introduce a **`BrokerAdapter`** interface in `services/go/md`, implemented by:

- **`nse_replay`** (default) — reads staged bars from Postgres and drives a virtual clock (Phase 1.3).
- **`angel_live`** — broker WebSocket and session lifecycle (Phase 11); Phase 1 ships a **stub** that returns **`ErrNotConfigured`** when `Run` is invoked.

Selection is **`MD_ADAPTER`** (`nse_replay` | `angel_live`), defaulting to `nse_replay`. Invalid values fail fast at process start.

Adapter output before full normalization is modeled as **`DraftTick`** + **`RunHooks.OnTick`**; the normalizer (Phase 1.6) refines types and enrichment.

## Consequences

**Positive**

- Deterministic replay stays the default development path; no broker API keys (or session churn) required for Phases 1–10 when using `nse_replay`.
- Vendor churn (Angel vs Kite vs paid ticks) is isolated behind adapters.
- Metrics and ops can label behavior by `adapter` dimension consistently.

**Negative**

- One more abstraction layer; new engineers must learn adapter vs normalizer boundaries.
- `DraftTick` will evolve until protobuf/canonical types settle — small refactor risk.

**Neutral**

- Stub `angel_live` deliberately fails `Run` until Phase 11 so misconfigured env is obvious without half-working connections.

## Alternatives considered

- **Single replay-only binary, fork for live later**: rejected — duplicated HTTP/WS/persist paths and harder parity testing.
- **Plugin SO/DLL loading**: rejected — unnecessary complexity for two known implementations.
- **Node adapter**: rejected — MD hot path stays Go per ADR-0004.

## References

- [Phase 1 — Market Data Gateway](../phases/phase-01-market-data.md)
- ADR-0004 (Go for hot path)
- ADR-0021 (canonical instrument spec)

## Revisit triggers

- Adding a third primary source (e.g. paid tick vendor) with different auth and rate limits.
- Needing multiple concurrent adapters per region — would push toward a router/fan-in design.
