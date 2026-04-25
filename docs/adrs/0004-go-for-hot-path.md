# 0004 — Go for the hot path (matching, MD, SPAN, MMs)

- Status: Accepted
- Date: 2026-04-20
- Deciders: Project owner

## Context

The system has a heterogeneous workload:

- **Hot path**: matching engine, market data gateway, SPAN calculator, synthetic MMs. Throughput and tail-latency matter. Long-running stateful processes with specific concurrency patterns (single-writer per symbol).
- **Warm path**: OMS, Risk orchestrator, Portfolio, Reports. Business logic changes often; developer velocity matters; throughput is well within what Node can handle.
- **Cold path**: Frontend, SDK, tooling. TS is the natural choice.

The developer is fluent in JavaScript / TypeScript and comfortable-enough in Go. Writing everything in TS loses latency predictability (V8 GC, event-loop lag); writing everything in Go loses iteration speed in the OMS/risk/portfolio layer where business logic churns.

## Decision

Split by path:

- **Go 1.22+** for: matching engine, MD gateway, SPAN calculator, synthetic market makers.
- **Node 20 + TypeScript + Fastify** for: gateway, OMS, risk, portfolio, reports, strategy-runner.
- **React + TS** for frontend.

Inter-service contracts are Protobuf (Go↔Node) and OpenAPI (FE↔BE).

## Consequences

**Positive**

- Latency-critical components get Go's predictable GC and concurrency primitives (goroutines + channels map cleanly to single-writer-per-symbol).
- Business logic stays in TS where iteration cost is lowest.
- Matches the realistic stack of Indian brokers (Zerodha runs C++/Go for core, Python/Node for other pieces — publicly known).
- Interview-credible: "Go for hot path, TS for everything else" is a defensible pattern.

**Negative**

- Two languages mean two build toolchains, two CI pipelines, two dependency managers.
- Some contracts are duplicated (TS types from Zod, Go types from Protobuf). Protobuf + codegen mitigates most.
- Context switching tax for the solo developer.

**Neutral**

- Protobuf boundary is healthy architectural discipline regardless of language.

## Alternatives considered

- **All Go**: rejected. Frontend is a non-starter in Go; and the business-logic iteration velocity loss in OMS/risk would slow the project noticeably for this developer.
- **All TS (Node + Bun)**: rejected for matching engine — GC pauses, single-threaded event loop, and arbitrary allocation under load would miss the p99 target. Bun is intriguing but too young for a learning project's core.
- **Rust for hot path**: rejected for v1 — steeper learning curve, slower iteration; possible v2 swap for the matching engine.
- **Java / Kotlin**: rejected — the realistic Indian broker stack is Go/C++/Python; Java is fine but off the learning goal.

## References

- LMAX Disruptor paper (single-writer principle; language-agnostic).
- Zerodha Tech blog posts on their stack.
- Go runtime GC papers (Rick Hudson et al).
- Node event loop + GC analysis posts.

## Revisit triggers

- If p99 order-ack stays above 50 ms after profiling — consider Rust for matching.
- If OMS becomes CPU-bound in v2 — consider Go port for OMS.
