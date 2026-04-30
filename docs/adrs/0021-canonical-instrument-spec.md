# 0021 — Canonical instrument spec + contract metadata

- Status: Accepted
- Date: 2026-04-30
- Deciders: Ganesh

## Context

Every part of the system references tradable things:

- Market data streams ticks/candles for instruments.
- OMS validates and routes orders by instrument.
- Matching engines partition by instrument.
- Risk and portfolio need trading constraints (tick/lot/freeze) and contract terms (expiry/strike).

If each service invents its own “instrument shape”, adding NFO introduces repeated parsing and inconsistent logic (e.g., one service treating `lot_size` differently, or misclassifying options vs futures).

## Decision

Adopt a single canonical **InstrumentSpec** model (persisted in `ref.instruments` and cached), resolved by `instrument_id` and used everywhere.

Minimum required fields for v1:

- Identity and routing
  - `instrument_id` (stable internal id)
  - `exchange`, `segment`
  - `assetClass` (`EQUITY` | `DERIVATIVES`)
  - `instrumentType` (`EQ` | `FUT` | `OPT`)
- Trading constraints
  - `tickSize`, `lotSize`, `freezeQty`
  - `priceBands` (when available)
  - `status` (`ACTIVE` | `SUSPENDED` | `EXPIRED`)
- Contract metadata (only for derivatives)
  - `underlyingInstrumentId`
  - `expiry`
  - `strike` (options)
  - `optionType` (`CE` | `PE`)

All validation, risk, and position logic must take `InstrumentSpec` as input (not re-derive terms from strings like `tradingsymbol`).

## Consequences

Positive:

- One source of truth for constraints and contract terms.
- Services become simpler: route by `instrument_id`, not by bespoke parsing.
- Adding new contract types (weekly expiries, new underlyings) is localized to the instrument sync/ingest pipeline.

Negative / trade-offs:

- The instrument ingestion/sync step becomes a critical dependency (must be reliable and versioned).
- Requires careful migration strategy if the internal `instrument_id` scheme changes (avoid if possible).

## Alternatives considered

1. Parse `tradingsymbol` everywhere.
   Rejected: fragile and inconsistent; string formats vary across brokers/exchanges.

2. Keep separate schemas per segment (equity vs derivatives).
   Rejected: makes cross-asset routing harder and forces conditional logic into core services.

## References

- [docs/phases/phase-01-market-data.md](../phases/phase-01-market-data.md) (instrument sync pipeline)
- [docs/02-architecture.md](../02-architecture.md)
- ADR-0020 (asset-class module boundary)
