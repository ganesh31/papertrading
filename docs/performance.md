# Performance (published numbers)

This doc is the single source of truth for performance numbers under load.

## Environment (always record)

- **Date**: YYYY-MM-DD
- **Git SHA**: `________`
- **Machine**: CPU, cores/threads, RAM
- **OS**: Linux/macOS + version
- **Runtime**: Docker Desktop / bare metal; resource limits (CPU/memory)
- **Go**: version, `GOGC`, `GOMEMLIMIT`
- **Node**: version
- **Postgres / Redis**: versions + config notes (pool sizes, persistence)

## Workloads

### 1) Matching engine — orders (steady)

- **Goal**: sustained throughput with stable p99.
- **Load**: `infra/k6/orders_steady.js`
- **Mix**: place/cancel mix (record actual %), LIMIT/MARKET/IOC/FOK mix
- **Symbols**: count, distribution (single hot symbol vs many)

**Results**

- **Throughput**:
  - Orders accepted/sec: ___
  - Acks/sec: ___
  - Trades/sec: ___
- **Latency (NewOrder → Ack)**:
  - p50: ___ ms
  - p95: ___ ms
  - p99: ___ ms
- **Errors**:
  - Error rate: ___ %
  - Top errors: ___
- **Resource**:
  - ME CPU: ___ %
  - ME RSS: ___ MB
  - Go GC pause p99 (`gc_pause_ms{service="matching"}`): ___ ms
  - Per-symbol queue p99 (`match_engine_queue_depth`): ___

**Notes / bottlenecks found**

- ___

### 2) Matching engine — orders (burst)

- **Load**: `infra/k6/orders_burst.js`
- **Burst profile**: ramp + peak duration

**Results**

- Peak orders/sec: ___
- p99 order-ack at peak: ___ ms
- Queue depth at peak: ___
- Recovery time to steady-state (queue back to baseline): ___ s

### 3) Market data gateway — tick firehose

- **Load**: `infra/k6/md_firehose.js` (or equivalent generator)
- **Goal**: sustain tick ingest + persist + WS fanout SLOs.

**Results**

- Ticks in/sec: ___
- Persist p99 (ingest → durable): ___ ms
- WS push p99 (ingest → client): ___ ms
- CPU/RSS, GC pause p99: ___

### 4) SPAN hot path — pre-trade margin checks

- **Load**: add a k6 scenario that drives the order path with portfolios sized:
  - Small (5 positions), medium (25), large (100)

**Results**

- SPAN calls/sec: ___
- p99 `span_calc_duration_ms`: ___ ms (for each portfolio size)
- Reject rate due to margin: ___ %

### 5) Synthetic market makers (MM) under load

- **Goal**: MM should not be the throughput limiter; it should degrade gracefully.

**Results**

- Quotes/sec: ___
- Cancel/sec: ___
- ME impact (delta in p99 order-ack): ___ ms

## Dashboards & artifacts

- Grafana dashboard JSON committed: `infra/grafana/dashboards/load-test.json` (or actual path)
- Raw k6 outputs saved: `infra/k6/results/<date>-<sha>/`

## Targets (v1)

These are the targets we try to hit on a dev laptop with Docker networking:

- **Sustained**: 10k orders/min
- **Burst**: 20k orders/min
- **p99 order-ack (E2E)**: < 50 ms
- **MD tick-to-FE p99**: < 300 ms
- **Go GC pause p99 (hot services)**: < 10 ms

