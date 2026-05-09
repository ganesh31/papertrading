# Phase 1 — Talking points

Prompts for **market data gateway**: replay-first, adapters, synthetic ticks, and operational shape.

---

## 1) Why replay-first instead of wiring Angel WebSocket on day one?

**Question:**  
Couldn’t you pull “real” ticks immediately and skip the fake clock?

**90-second answer:**  
Determinism is the product feature I’m buying: same bars, same session id, same synthesized path → **byte-identical tick logs** for regression and for Phase 9 backtest parity. Replay also removes API keys, TOTP churn, and ISP/IP surprises from the critical path while I build the normalizer, persistence, bus, and WS fan-out. **`angel_live`** stays behind one **`BrokerAdapter`** implementation so production wiring is a swap, not a rewrite.

**Trade-off:**  
Synthetic intra-bar paths don’t reproduce true microstructure; that’s acceptable until real tick feeds matter.

**At scale:**  
Replay stays in CI and staging; prod runs live adapters with reconnect budgets and staleness gates—and the same **`adapter`** metric labels.

---

## 2) What does “tick synthesis” preserve and what does it lie about?

**Question:**  
If ticks are generated from 1-minute OHLCV bars, how do you trust stops or fills?

**90-second answer:**  
The synthesizer is explicit: OHLC are honored in order, volume is conserved across synthetic points, timestamps stay inside the bar, and the bridge is **documented** (ADR-0019). It’s good enough for CLOB smoke tests, risk plumbing, and P&L-shaped strategies that care about minute-level regime—not for HFT microstructure research.

**Trade-off:**  
Fidelity vs **offline velocity** and dataset cost.

**At scale:**  
Swap the adapter to paid tick vendors without changing downstream **`Tick`** contracts.

---

## 3) Why Postgres + Timescale + Redis Stream instead of “just Kafka”?

**Question:**  
Everyone says Kafka for market data—why Redis Streams here?

**90-second answer:**  
Authoritative history lives in **`md.ticks`** with Timescale continuous aggregates for candles—query path matches how charts and replay debugging think. Redis **`ticks.v1`** is a **bounded-retention fan-out** for concurrent consumers (`mm`, `strategy`, `surveillance`) with minimal ops footprint on a laptop Compose stack. Kafka is justified when we need cross-region replay, long retention, or massive partition parallelism—not v1.

**Trade-off:**  
Operational simplicity vs **Kafka’s scaling story**.

**At scale:**  
Mirror ticks from Postgres outbox or stream bridge into Kafka if consumer teams require it.

---

## 4) Virtual clock vs wall clock—where does “market open” come from?

**Question:**  
How does the UI know session state during replay?

**90-second answer:**  
Replay advances **`Coordinator`** virtual time; **`GET /market/status`** accepts optional **`virtualTime`** so gateways and UIs evaluate **`session`** against IST bands + **`holidays.json`** using replay time, not **`time.Now()`**. Prometheus gauges (**`replay_running`**, **`replay_virtual_timestamp_seconds`**, **`replay_pending_ticks`**) make it obvious in Grafana whether ticks are live fire or replay.

**Trade-off:**  
Extra clock plumbing vs **correct calendars** when speeding through history.

**At scale:**  
Same contract—live adapters stamp broker receive time; replay stamps synthesized **`Ts`**; staleness logic stays **`LIVE`**-only.
