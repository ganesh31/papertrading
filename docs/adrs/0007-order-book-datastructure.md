# 0007 — Order book data structure (skiplist + price-level FIFO)

- Status: Accepted
- Date: 2026-05-15
- Deciders: Project owner

## Context

The matching engine implements an **order-driven** central limit order book (CLOB). NSE equity cash uses **price-time priority**: better price first; at the same price, earlier orders first. The engine must support fast best bid/ask, insert and cancel of resting orders, partial fills at the head of a level, and modification rules aligned with Indian convention (price change loses time priority; quantity decrease retains priority; quantity increase loses priority).

Prices and quantities on the wire are **integer** (`price_paise`, `quantity` in `matching.proto`) — no floating-point money in the book.

Operational constraints from Phase 2:

- Typical NSE symbols span **wide** absolute price ranges on a single tick grid (e.g. ₹1–₹10,000+ on the same instrument).
- Options (later phases) have **dense** strike ladders where an array-indexed ladder can be O(1).
- v1 targets clarity and predictable latency over squeezing last nanoseconds.

## Decision

### Layer 1 — `PriceLevel` (FIFO queue at one price)

At each distinct price on a side:

- Aggregate displayed quantity (`total_quantity`) for Redis/FE snapshots.
- **Doubly-linked list** of resting `Order` nodes for O(1) remove on cancel/modify and O(1) `PopFirst` after a full fill at the head.
- **`AddOrder`** appends to the **tail** (time priority for new arrivals at this price).
- Matching always consumes from the **head** (oldest at that price).

Partial fill: decrement head `remaining`; if zero, `PopFirst`; if level empty, remove level from the side map.

### Layer 2 — `Book` (two sides + stops)

```text
Book
  bids:  SkipList[price_paise] → *PriceLevel   // best bid = max price
  asks:  SkipList[price_paise] → *PriceLevel   // best ask = min price
  stops: structure for pending SL / SL-M (see below)
```

- **`BestBid` / `BestAsk` / `BBO`** — O(log N) via skiplist extremal key (bids: highest; asks: lowest).
- **Matching walk** — iterate asks ascending from best ask (buy taker) or bids descending from best bid (sell taker) until limit price blocks further crosses.

**Stops (SL / SL-M):** Not stored in the bid/ask skiplists until triggered. v1 keeps pending stops in a structure keyed for efficient scan on LTP (e.g. sorted by `trigger_price_paise` per side, or two trees). On trigger, promote to LIMIT or MARKET and invoke the matcher on the **actor goroutine** (ADR-0006). Exact stop index is an implementation detail; requirement is O(k) work for k triggers crossed by a tick.

### Layer 3 — v1 ordered map: **skiplist**

Use a skiplist (or Go `map` + sorted slice for very small N — prefer a dedicated ordered structure for stable O(log N)) keyed by `price_paise`:

| Operation | Complexity |
|-----------|------------|
| Insert level / add order at new price | O(log N) |
| Remove empty level | O(log N) |
| Best bid / best ask | O(log N) |
| Match walk to next level | O(log N) per level |

**N** = number of distinct price levels on a side, not number of orders.

### Aggregated vs order-level depth

- **Matching** is always **order-level** (FIFO within level).
- **Redis / `GetBook`** expose **aggregated** `PriceLevel` (price + total qty) for L5/L20 FE — sufficient at NSE tick granularity for v1.

### Modify semantics (Indian convention)

| Change | Action |
|--------|--------|
| Price change | Remove from level, insert at new price at **tail** (lose time priority) |
| Qty decrease only | In-place decrement on node (keep position) |
| Qty increase | Treat as losing priority: remove + re-append at tail (same as price change) |

### Cancel during match

When matching, iterate using the **head at match start**; if cancel arrives concurrently, it is serialized by the actor **before or after** the current event — never mid-iteration on a stale pointer without the actor model (ADR-0006).

## Consequences

**Positive**

- Standard exchange semantics; easy golden-file tests (price-time, partials, FIFO).
- Skiplist is simple to implement or import; performance predictable for wide price ranges.
- DLL at each level gives O(1) cancel for resting orders.

**Negative**

- O(log N) per level step vs O(1) for a dense ladder — acceptable for cash equities in v1.
- More allocations than a packed array ladder unless `sync.Pool` is used for order nodes (Phase 2 performance note).
- Stop book maintenance adds a second index to keep consistent.

**Neutral**

- Pro-rata matching (some CME products) is **out of scope**; NSE cash is price-time.

## Alternatives considered

| Option | Why not v1 |
|--------|------------|
| **Array-indexed ladder** `(price - floor) / tick` | O(1) ops but huge sparse arrays for wide cash ranges; excellent for dense option strikes — **planned specialization in v2** per segment. |
| **Treap / red-black tree** | Equivalent asymptotics; skiplist chosen for simplicity and readable reference implementations. |
| **Heap of levels** | Does not give O(1) FIFO within level without a secondary structure anyway. |
| **Single `map` + scan for best** | O(N) best bid/ask — rejected for hot path. |
| **Aggregated book only (no per-order nodes)** | Cannot implement partial fills + cancel/modify priority correctly. |

## References

- Phase 2: `docs/phases/phase-02-matching-engine.md` (§2.2, §2.3, common pitfalls).
- ADR-0006 — per-symbol actor (single mutator of `Book`).
- `docs/talking-points/phase-02.md` — skiplist vs ladder interview answer.
- Reference implementations: [orderbook (Go)](https://github.com/i25959341/orderbook), [liquibook (C++)](https://github.com/enewhuis/liquibook).
- Larry Harris, *Trading and Exchanges*, ch. 4–5.

## Revisit triggers

- Options module ships with dense strike grids → benchmark ladder for F&O symbols only (hybrid book factory by segment).
- Profiling shows skiplist dominates latency → consider flat intraday ladder with rebase on circuit band change.
