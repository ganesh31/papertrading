# Talking Points — Phase 2: Matching Engine

Expect this to be the most-probed area in a Senior Architect interview at a broker or exchange.

## 1. Walk me through what happens when a LIMIT BUY arrives at your matching engine

**90-second answer**

The order comes in over gRPC to a thin server layer that does nothing but route it to the symbol's actor. Each symbol has a dedicated goroutine that owns its order book — a single-writer pattern, so there are no locks inside the book. The actor receives the order on a channel and processes it synchronously relative to other events for that symbol.

Matching: we walk the asks side starting at the best ask. If the BUY price ≥ best ask, we match — price is the resting side's price, quantity is min of both remaining qtys. We emit a Trade event, decrement both orders, and if the resting order is fully filled, pop it; if the level is empty, remove the level. Loop until we can't match. Any remainder enters the book at its limit price, appended to the FIFO queue at that price level.

After matching, we emit events to the bus — one `OrderUpdate` for the taker, one for each resting order touched, and one `Trade` per fill. The book snapshot in Redis is updated by a separate writer within 50 ms.

**Trade-off I'd defend**

Single-writer per symbol gives me determinism and no lock contention, at the cost of a hot goroutine per symbol — fine for NSE's ~6000 symbols but wouldn't scale to FX (millions of pairs). I'd shard across hosts by symbol hash once per-process scales out.

**At scale**

NUMA pinning; symbols co-located with their hottest clients via consistent hashing in the gateway. Matching engine and gateway share the same machine for latency; everything else is async.

---

## 2. How do you implement price-time priority with partial fills?

**90-second answer**

Price levels are organized in a skiplist keyed by price. At each price level we keep a doubly-linked list of orders — each order is a node with a pointer. When an order arrives at an existing level, it's appended to the tail: this is the "time" part of price-time. When matching, we take from the head. Partial fills just decrement the head's remaining qty; if it hits zero, we pop it.

On `ModifyOrder`: NSE's rule is that a qty-down keeps priority, but any price change or qty-up resets priority. So qty-down is an O(1) in-place edit; price-change is a remove-and-reinsert at the new level.

**Trade-off I'd defend**

Skiplist (O(log N) best-bid/ask) vs array-indexed ladder (O(1)). I picked skiplist because NSE price ranges are wide (some illiquid names trade across ₹1 to ₹10,000) — a ladder would waste memory and require dynamic resizing. On options, where strikes are dense, the ladder is actually better — ADR notes I'd specialize per segment in v2.

**At scale**

Lock-free variants exist (Disruptor-style ring buffer for the input channel), but lock-free inside the book isn't necessary given single-writer. The ring buffer at input is worth it for the inbox if we go to tens of millions of orders/sec.

---

## 3. How does your matching engine recover from a crash?

**90-second answer**

Every mutation — order added, order cancelled, match event — is logged to Postgres before responding. On boot, each symbol's actor reads its event log from the last snapshot onward and replays. The full book snapshot to Postgres happens every minute; Redis L5-depth snapshot every second for FE.

Because matching is deterministic given the input event stream, replay reconstructs an identical book state. No lost trades; no phantom orders. Cancelled orders don't reappear because the cancel event is in the log too.

The only thing we lose is in-flight orders that were received but not yet logged — those are dropped with a warning; the clients will retry via idempotency.

**Trade-off I'd defend**

Event-log-as-source-of-truth vs. periodic full-state persistence. I chose the log because recovery time is the same either way (replay N seconds of events), but the log gives me audit + replay testing for free. Storage overhead is negligible at NSE scale.

**At scale**

Use a distributed log (Kafka) with partitions per symbol. Replication factor 3. Commit ack before user ack. Leader election per symbol partition.

---

## 4. What's self-trade prevention and how did you implement it?

**90-second answer**

Self-trade prevention is when the same client appears on both sides of a potential match — buyer and seller are the same account. NSE requires this to be prevented (it'd be market manipulation). The three common variants: cancel-newest (CN), cancel-oldest (CO), and decrement-and-cancel (DC).

NSE uses DC: match as much as possible against *other* resting orders; the matching qty that would cross the same account is cancelled on the smaller side. I implemented DC.

When the taker is walking the asks and hits a resting ask from the same user, I skip that level (or that order) and continue — if qty remains after walking everyone else, we cancel that remainder.

**Trade-off I'd defend**

DC is a bit more complex than CN/CO but is what NSE actually uses; I prioritized realism because the interview value is higher.

**At scale**

The check is O(1) lookup per order comparison. Fine. You'd cache `user_id` on the order node to avoid any joins.

---

## 5. Why price-time and not pro-rata?

**90-second answer**

NSE uses price-time — the first order at a better price gets filled first, and within a price level, oldest first. It's simple, rewards being first, and concentrates liquidity at top of book.

Pro-rata allocates fills proportionally to quantity at the level, regardless of time. It's used by CME on some futures to reduce queue-jumping games. It creates different market microstructure incentives.

For a system mirroring NSE, price-time is the correct choice. Pro-rata would be a surprising departure that'd mislead anyone using the tool for strategy research.

**Trade-off I'd defend**

Price-time is more gameable by high-frequency players (queue jumping); pro-rata reduces that but fragments liquidity at every level. NSE's choice aligns with most equity exchanges globally.

**At scale**

Architecturally identical — only the matcher's inner loop differs. Trivial to support both behind a per-symbol config.

---

## 6. Synthetic market makers — why and how?

**90-second answer**

A paper-trading system needs counterparty liquidity. Options: (a) replay real order-book deltas from historical NSE data — ideal but licensing is paid and sourcing is hard; (b) synthetic market makers quoting around real LTP — pragmatic and realistic enough; (c) match-at-LTP — too simplistic, hides slippage.

I picked (b). The MM service subscribes to MD. For each symbol, it places 5 bid levels at ticks 1..5 below LTP and 5 ask levels above, with Poisson-distributed quantities. It refreshes quotes every 500 ms and on each significant LTP move. Configurable per-symbol: spread, skew, depth multiplier, refresh rate.

This gives realistic fills with plausible slippage on large orders without needing historical tick licenses.

**Trade-off I'd defend**

Not modelling the true order-book shape per symbol — my MMs are statistically stationary when real markets have regime shifts. For a learning project, this is fine; for a real strategy-research tool, I'd feed in fitted L5 shapes from snapshots.

**At scale**

Each MM is trivially scalable — one goroutine per symbol. The total throughput is dominated by the matching engine, not the MM.

---

## 7. Your p99 is 50 ms. Zerodha's is single-digit ms. Why?

**90-second answer**

A few reasons. Real NSE brokers co-locate in NSE's data center — network latency to the exchange is sub-millisecond. My system is entirely synthetic and local, but network between gateway/OMS/risk/ME adds milliseconds via Docker networking on a laptop.

Second, real brokers optimize at the OS level: kernel bypass (DPDK), pinned CPUs, huge pages. I run stock Linux and the Go runtime; GC is well-tuned but not single-digit-millisecond tuned.

Third, they have SPAN precomputed per user with deltas; I compute incrementally but still on the order path.

What I'm demonstrating is the *architecture* that makes those optimizations possible later — single-writer engine, no locks, event-sourced recovery. Given that, moving to sub-ms is a matter of deploying differently, not rewriting.

**Trade-off I'd defend**

I optimized for learning and observability, not raw latency. Different product, different priorities.

**At scale**

DPDK userspace networking; pin matching engine to isolated cores; FPGA for option reval; co-locate strategy runners with the engine.

---

## 8. How do stop-loss orders work in your engine?

**90-second answer**

Stop orders — SL and SL-M — sit in a separate set keyed by trigger price. They're not in the active book yet. Each tick flowing through the actor checks the stop set: for BUY stops, trigger when LTP ≥ trigger price; for SELL stops, LTP ≤ trigger.

When a stop triggers, it's promoted to a LIMIT (for SL) or MARKET (for SL-M) and enters the matcher as a fresh order. The event log captures the `StopTriggered` event separately for audit.

Trigger evaluation is O(k) where k is the number of stops crossed in this tick — typically 0, occasionally a small number. Worst case, a gap open triggers many simultaneously; we process them in the order they were placed (fairness).

**Trade-off I'd defend**

Storing stops in a separate structure rather than the book keeps matching O(log N) — mixing would complicate the inner loop. NSE does the same conceptually.

NSE actually disallowed SL-M on several segments post-2020 for risk reasons; I mirror that rule in validation.

**At scale**

Stops could move to a separate service if the volume justified it — but they don't. Keep them with their symbol's actor.