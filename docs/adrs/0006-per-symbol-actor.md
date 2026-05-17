# 0006 — Per-symbol actor for the matching engine

- Status: Accepted
- Date: 2026-05-15
- Deciders: Project owner

## Context

The matching engine must maintain a limit order book per symbol with **price-time priority**, partial fills, stop triggers on LTP, and deterministic recovery from an event log. Concurrent access to one book from many goroutines (gRPC handlers, tick subscribers, snapshot writers) would require fine-grained locking; lock ordering bugs and priority inversions are hard to test and violate the mental model interviewers expect for exchange cores.

Phase 2 targets **p99 order-ack &lt; 50 ms** at ~10k orders/min on a laptop, with **deterministic replay** after crash. The gRPC surface is already defined (`MatchingService` in `packages/protos/papertrading/matching/v1/matching.proto`); handlers must not run matching logic on the RPC thread.

Forces:

- **Ordering**: For a given symbol, every command (new/cancel/modify) and every LTP update must be applied in a single total order so replay reproduces the same book.
- **No re-entrancy**: An LTP tick must not synchronously call into matching in a nested stack that might publish another tick (stop → match → …).
- **Isolation**: Symbols are independent; cross-symbol work should not block hot symbols.
- **Scale shape**: NSE cash has on the order of thousands of active symbols — one goroutine per active symbol is acceptable for v1.

## Decision

Use a **single-writer actor per symbol**:

1. **Registry** — `map[symbol]*SymbolActor` (or equivalent) owned by the matching process. First touch for a symbol starts the actor; idle actors may be stopped later (out of scope for v1).

2. **Inbox** — Each actor reads from one Go channel carrying a sealed union of work items, for example:
   - `OrderCommand` — NewOrder / CancelOrder / ModifyOrder (already validated for tick/lot/band where applicable).
   - `LTPUpdate` — last traded price (+ optional timestamp) from `ticks.v1`.
   - `MarketHalt` / `Resume` — from `market.flags` (Phase 2 §2.9).
   - `Shutdown` — drain and persist (graceful stop).

3. **Processing** — The actor loop handles **one message at a time** (no concurrent handlers for the same symbol):
   - Run matcher / book mutations / stop evaluation on the actor goroutine only.
   - Append to `match.events` (and publish bus messages) before or as part of acknowledging the RPC — exact ack vs log ordering is fixed in the persistence ADR path (issue #14); the actor is still the sole mutator of in-memory book state.

4. **gRPC layer** — Thin: resolve symbol → enqueue → wait for result (channel or future) → return `*Response`. No book locks in handlers.

5. **No locks inside the book** — `Book` data structures are owned exclusively by the actor; other goroutines never hold pointers into the book.

6. **Fairness** — Go’s channel + scheduler multiplexes symbols; we do not implement strict cross-symbol fairness in v1. Per-symbol FIFO is guaranteed.

7. **Snapshots** — Redis L5 and Postgres full-book snapshots read **copies** or levels exported by the actor on a timer/event count (not concurrent unsynchronized reads of live nodes).

```text
  gRPC / bus          SymbolActor (one goroutine)
       │                      │
       ├─ enqueue ───────────►│ select inbox
       │                      │   → validate (if not done upstream)
       │                      │   → matcher + book
       │                      │   → emit events / trades
       │                      │   → signal ack to waiter
       └◄── response ─────────┘
```

**Pitfall (explicit):** Do not run matching inside a dedicated “LTP callback” that can re-enter the actor. LTP and orders share the **same** inbox and the same loop (Phase 2 common pitfalls).

## Consequences

**Positive**

- Trivial reasoning: total order per symbol = channel delivery order.
- Deterministic replay: feed the same event stream into a fresh actor → identical book.
- Aligns with ADR-0004 (Go hot path) and LMAX-style single-writer guidance.
- No mutex contention on the book hot path.

**Negative**

- One goroutine per active symbol — memory and scheduler overhead; need caps or eviction for pathological symbol counts.
- A single ultra-hot symbol can saturate one core (acceptable for v1; sharding by symbol hash is the scale-out story).
- Enqueue + wait adds latency vs in-handler matching — bounded by channel ops; still within Phase 2 SLO if validation stays thin.

**Neutral**

- Cross-symbol operations (market-wide halt) broadcast to all actors via a supervisor, not a global book lock.

## Alternatives considered

- **Global book + `sync.RWMutex`** — rejected: harder to prove ordering; writer locks under load; poor interview story for an exchange core.
- **Sharded worker pool (N goroutines, hash symbol → worker)** — viable at scale; v1 uses 1:1 symbol:goroutine for clarity; pool is a drop-in once registry maps symbol → worker id.
- **Lock-free book structures** — rejected for v1: single-writer makes them unnecessary; higher implementation risk.
- **Thread-per-core Disruptor (Java LMAX)** — pattern reference only; Go channels + single-writer per symbol capture the same idea with less machinery.

## References

- Phase 2: `docs/phases/phase-02-matching-engine.md` (§Core design 1, architecture diagram, common pitfalls).
- ADR-0004 — Go for the hot path.
- ADR-0007 — order book data structure (owned by the actor).
- ADR-0002 — event-sourced OMS (downstream consumer of trades/order updates).
- Larry Harris, *Trading and Exchanges*, ch. 5 (order-driven markets).
- LMAX Disruptor: [https://lmax-exchange.github.io/disruptor/disruptor.html](https://lmax-exchange.github.io/disruptor/disruptor.html)

## Revisit triggers

- Process RSS or goroutine count exceeds budget with realistic symbol fan-out → worker pool + symbol→worker mapping.
- p99 ack dominated by enqueue contention → bounded MPMC inbox per worker or coalesce LTP updates.
