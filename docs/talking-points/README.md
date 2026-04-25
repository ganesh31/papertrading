# Talking Points — Interview Prep Pack

One doc per phase. Each doc contains a handful of prompts you're likely to get in a Senior Architect interview, with:

1. **Question** — how it'll actually be asked.
2. **90-second answer** — what you say, phrased as you'd say it out loud.
3. **Trade-off you'd defend** — the nuance that signals seniority.
4. **At scale** — what you'd do differently for a real broker running at 1M users.

Use these to rehearse, not to memorize. The goal is fluency, not scripts.

## Recommended rehearsal pattern

1. **Solo rehearsal** — record yourself answering one prompt in 90 seconds. Watch back.
2. **Peer rehearsal** — have someone ask you prompts from a random phase, cold.
3. **Whiteboard rehearsal** — pick the matching-engine or SPAN doc. Diagram from memory.

## Index

- `phase-00.md` — monorepo, boundaries, observability.
- `phase-01.md` — broker adapters, virtual clock, tick firehoses.
- `phase-02.md` — matching engine, CLOB, single-writer, order types. **The one they'll probe most.**
- `phase-03.md` — event-sourcing, idempotency, reject taxonomy.
- `phase-04.md` — projections, ledger, FIFO cost basis.
- `phase-05.md` — frontend state split, WS multiplexing, chart.
- `phase-06.md` — futures MTM flow, settlement.
- `phase-07.md` — options pricing, greeks, expiry edge cases.
- `phase-08.md` — SPAN internals. **The other big one.**
- `phase-09.md` — backtest/live parity, determinism, isolation.
- `phase-10.md` — schedulers, DAGs, corporate actions.
- `phase-11.md` — SLOs, load/chaos, how I'd scale this.

## The "one-pager elevator pitch"

If you have 60 seconds at a coffee chat:

> I built a paper-trading platform modeled on NSE/BSE end-to-end. Own CLOB matching engine in Go, event-sourced OMS, real SPAN margin calculator with scenario revaluation, Kite-clone UI, and a strategy SDK with live/backtest parity. Synthetic market makers seeded from real Angel One prices so every paper order gets matched against realistic liquidity. Deployable on one VPS. p99 order-ack under 50 ms at 10k orders/min. Every architectural decision is an ADR; every phase has a talking-points doc.

## What interviewers are usually testing

- **Can you explain trade-offs?** Every answer should name at least one alternative you considered.
- **Do you know where the complexity hides?** Matching priorities, margin scenarios, T+1 timing, idempotency edges.
- **Can you scale it?** Not just "add a load balancer" — per-symbol partitioning, event-log throughput, pre-trade risk latency budget.
- **Do you know the domain?** Peak margin, circuit filters, MWPL, T+1, STT-on-exercise. This project teaches you all of them.
- **Ops mindset?** Runbooks, kill switch, reconciliation, projection rebuild.

## Red flags to avoid

- Saying "I'd use Kafka" for every event. Justify it.
- Saying "microservices" without saying what each owns.
- Saying "we use event sourcing" without explaining when you wouldn't.
- Not having a number for p99 anything.
- Dodging "how would you scale to 1M users?".
- Not having a clear reason for each language choice.

## Sample prompts across phases (sanity drill)

Cold-ask yourself these. If any feel shaky, re-read the phase doc.

1. Walk me through what happens when a user places a BUY LIMIT order in your system.
2. A trader places 5000 orders in 10 seconds. How does your system respond?
3. Your matching engine process crashes while actively matching. What happens?
4. How do you enforce "upfront margin collection" per SEBI's peak margin norms?
5. Explain how SPAN arrives at the initial margin for a long call + short put.
6. A user's position projection doesn't match their trade log. What's your debug flow?
7. How does your strategy backtest ensure the same result run-to-run?
8. How would you shard this to 100 brokers × 100k clients each?
9. A stock split on INFY tomorrow — walk me through what runs tonight.
10. STT on option exercise — why does it surprise retail traders?
