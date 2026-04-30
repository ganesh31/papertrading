# 0022 — Risk model interface (VAR+ELM vs SPAN)

- Status: Accepted
- Date: 2026-04-30
- Deciders: Ganesh

## Context

Pre-trade risk must block orders that would exceed available funds/margin. The rules differ by asset class:

- **Equity cash**: cash availability + (VAR + ELM) style margin for intraday (MIS) and delivery constraints for CNC sells.
- **NSE F&O**: SPAN + Exposure + spread credits, computed from a portfolio view and a hypothetical order.

If risk logic is embedded directly into `services/risk` as branching rules, adding SPAN later tends to force large refactors (new portfolio snapshot shape, new “blocked margin” semantics, new error reporting).

## Decision

Define a stable **RiskModel** interface called by `services/risk`, implemented per asset module:

- Input:
  - `orderIntent` (validated order request + resolved `InstrumentSpec`)
  - `portfolioSnapshot` (cash, holdings, open positions, currently blocked margin)
  - `marketState` (session open/close, symbol staleness flags)
- Output:
  - `result`: `OK` | `REJECT`
  - `rejectCode` (stable taxonomy)
  - `marginBlockedDelta` (how much additional margin must be blocked if accepted)
  - optional `breakdown` (for UI “margin preview” and debugging)

Semantics:

- Risk runs **synchronously** on the order path.
- If the order is accepted, OMS records the `marginBlockedDelta` so portfolio can reflect the new blocked amount.
- On terminal outcomes (cancel/reject/expiry/fill), the module defines when and how blocked margin is released.

Implementations:

- **Equity RiskModel**: cash + holdings checks; VAR+ELM (or conservative defaults) for MIS.
- **NFO RiskModel**: calls `go/span` to compute SPAN+Exposure incrementally; returns a breakdown for the order pad.

## Consequences

Positive:

- Equity-first delivery is straightforward, and SPAN integration later is additive.
- A single “margin preview” endpoint can work across modules (returns `breakdown` when available).
- Margin blocking is explicit, enabling ledger invariants and clean release logic.

Negative / trade-offs:

- Requires defining (and keeping consistent) the `portfolioSnapshot` shape early.
- Risk must remain deterministic (given same snapshot + order + params), which constrains “live broker” quirks.

## Alternatives considered

1. Hardcode equity now, refactor for SPAN later.
   Rejected: tends to create breaking changes in portfolio and OMS event semantics.

2. Always compute “% of notional” for everything.
   Rejected: fails the realism goal and blocks learning value; also mismatches broker UX.

## References

- [docs/phases/phase-03-oms-risk.md](../phases/phase-03-oms-risk.md)
- [docs/phases/phase-08-span-margin.md](../phases/phase-08-span-margin.md)
- ADR-0020 (asset-class module boundary)
