# 0020 — Asset-class module boundary

- Status: Accepted
- Date: 2026-04-30
- Deciders: Ganesh

## Context

The system must support multiple asset classes/segments (starting with **Equity cash**, then **NSE F&O**), while keeping core services stable:

- OMS should not encode contract-specific rules in its state machine.
- Portfolio should not be rewritten when adding MTM/expiry behavior.
- Risk should support multiple margin models (VAR+ELM for cash, SPAN+Exposure for derivatives) without branching logic sprawling across the codebase.

Without an explicit boundary, “asset logic” tends to leak into every service via `if segment == ...` conditionals, making future additions (and even refactors) expensive and brittle.

## Decision

Adopt an explicit **asset-class module plug-in boundary**:

- Core services (`gateway`, `oms`, `risk`, `portfolio`, `reports`) are **asset-agnostic** and route requests/events by `instrument_id` → `InstrumentSpec`.
- All domain variability lives behind **asset modules**, selected by `(assetClass, segment)` derived from `InstrumentSpec`.

The module boundary is defined by stable contracts:

- **OrderSemantics**: request validation + market-session constraints.
- **RiskModel**: margin checks + margin block/release semantics.
- **PositionModel**: trade → position/holding aggregation + MTM/expiry rules.
- **SettlementModel**: EOD jobs + corporate actions hooks.

v1 ships with:

- **Equity module** first (end-to-end vertical slice).
- **NFO module** later (futures, options, SPAN), added without refactoring core services.

## Consequences

Positive:

- Equity can be shipped first without pre-committing to F&O details everywhere.
- Adding NSE F&O becomes implementing module interfaces, not rewriting OMS/portfolio.
- Testing improves: module logic is deterministic and can be golden-tested per asset class.

Negative / trade-offs:

- Slight up-front design cost: common data contracts (`InstrumentSpec`) must exist early.
- Module boundary must be policed (review discipline) to prevent leakage.

## Alternatives considered

1. **Branching in core services** (`if instrumentType == FUT ...`)
   Rejected: grows unbounded and becomes hard to reason about; changes ripple across services.

2. **Separate services per asset class** (duplicate OMS/portfolio/risk)
   Rejected: duplicates infra/ops and makes “single user” scope heavier; harder to maintain consistent APIs.

3. **Single “domain service” with everything inside**
   Rejected: makes scaling and ownership unclear; obscures hot-path vs. business logic separation.

## References

- [docs/02-architecture.md](../02-architecture.md)
- ADR-0002 (event-sourced OMS)
- ADR-0004 (Go for hot path)
