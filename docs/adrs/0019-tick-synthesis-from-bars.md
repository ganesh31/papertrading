# 0019 — Tick synthesis from 1-minute bars

- Status: Accepted
- Date: 2026-05-02
- Deciders: Project owner

## Context

Phase 1 replay uses **free** Yahoo / NSE sources that only provide **1-minute OHLCV**. Downstream matching, risk, and strategies still expect a **tick-shaped** stream (timestamped LTP, volume slices, bid/ask placeholders) for pacing, WS fan-out, and persistence tests. True exchange tick data is paid or restricted.

We need **deterministic**, **offline-friendly** ticks that respect each bar’s **OHLCV** envelope without pretending to reconstruct real microstructure.

## Decision

Implement a **tick synthesizer** in Go (`services/go/md/internal/ticksynth`) that, for each 1m bar and configured `ticks_per_bar` (default **10**):

1. Places the **first** tick at **open** and the **last** at **close**, evenly spaced across the 60s window.
2. Builds interior prices with a **Brownian-bridge-style** Gaussian noise path (variance ∝ `t(1−t)`), **clamped** to `[low, high]`, then **forces** one interior index to **high** and another (distinct) to **low**.
3. Splits **volume** **uniformly** across ticks with remainder on the first ticks (integer-safe).
4. Sets **bid/ask** as `ltp ± tick_size × spread_ticks` (defaults `0.05` and `1`) as a **placeholder** spread until Phase 2 synthetic market makers supply realistic quotes.

**Determinism:** RNG is `math/rand/v2` **PCG** seeded from **SHA-256** of `(session_id, instrument_id, bar_start_unix, OHLCV, volume)` so the same replay session + inputs yields the **same tick sequence** (Phase 9 backtest parity).

**Minimum `ticks_per_bar`:** **4** so there is room for two distinct forced interior high/low touches.

## Consequences

**Preserves (approximately)**

- Bar **open** and **close** on boundary ticks, **high** and **low** visited at least once.
- **Total volume** per bar.
- **Deterministic** streams for tests and replay.

**Does not preserve**

- True **micro-price path**, **order-book** dynamics, **trade aggressor** side, or **intra-bar** volume microstructure.

Those gaps are acceptable for Phases 1–6 and replay-first testing; **Phase 2** synthetic MMs are the planned place for richer microstructure. Paid tick feeds can later replace the adapter behind the same boundary.

## Alternatives considered

- **Linear interpolation only (OHLC order)**: rejected — too rigid; Brownian bridge adds plausible path diversity without leaving the bar range (after clamp).
- **Random walk without forced high/low**: rejected — could miss the bar’s high/low after clamping noise.
- **Emit only OHLC four “ticks”**: rejected — too few events for WS / persistence stress tests; configurable N is more useful.

## References

- [Phase 1 — §1.4 Tick synthesizer](../phases/phase-01-market-data.md)
- ADR-0005 (broker adapter)
- Brownian bridge (stochastic processes texts / Wikipedia)

## Revisit triggers

- Subscribing to **paid tick** feeds or exchange tick drops — swap adapter implementation, keep normalizer contract.
- **Regulatory** or **backtest parity** review requiring documented synthesis error bounds vs. real ticks.
