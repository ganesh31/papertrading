# Paper Trading System — Documentation

A learning-first build of an NSE/BSE-style paper trading platform: CLOB matching engine, event-sourced OMS, broker-grade risk + portfolio/ledger, Kite-clone UI.

The plan is **Equity-first**, with additional asset classes (starting with **NSE F&O**) implemented as **plug-in asset modules** rather than refactors of OMS/portfolio core.

## How to read these docs

1. Start with [00-overview.md](./00-overview.md) — north star, success criteria, single-tenant pros/cons.
2. Then [05-nse-domain-primer.md](./05-nse-domain-primer.md) — the exchange concepts the rest of the docs assume.
3. Then [02-architecture.md](./02-architecture.md) — services, data flow, diagrams.
4. Then walk the phases in order: [phases/](./phases/).
5. Each architectural decision lives in [adrs/](./adrs/); each phase ships with [talking-points/](./talking-points/) you can reuse verbatim in interviews.

## Document index

### Foundations

- [00-overview.md](./00-overview.md) — goal, success criteria, single-tenant trade-offs, timeline.
- [01-tech-stack.md](./01-tech-stack.md) — stack choices with justification.
- [02-architecture.md](./02-architecture.md) — services, bounded contexts, data flow, diagrams.
- [03-data-model.md](./03-data-model.md) — schemas, event sourcing, ledger, projections.
- [04-cross-cutting.md](./04-cross-cutting.md) — observability, testing, security, release, ADR process.
- [05-nse-domain-primer.md](./05-nse-domain-primer.md) — NSE/BSE sessions, order types, circuit filters, freeze qty, T+1, peak margin, MWPL.
- [resources.md](./resources.md) — curated books, papers, OSS, official NSE/SEBI links.
- [repo-layout.md](./repo-layout.md) — monorepo structure, package conventions.
- [notebooklm-study-structure.md](./notebooklm-study-structure.md) — NotebookLM notebooks, sources, prompts, weekly workflow.
- [notebooklm-resource-urls.md](./notebooklm-resource-urls.md) — **External URLs by NotebookLM notebook** (Markets / Delivery / Architecture / External theory) + paste/download tips.

### Phases

- [phases/phase-00-foundation.md](./phases/phase-00-foundation.md) — Week 1 — monorepo, docker-compose, o11y skeleton, CI.
- [phases/phase-01-market-data.md](./phases/phase-01-market-data.md) — Week 2 — MD gateway, broker adapter, virtual clock, tick store.
- [phases/phase-02-matching-engine.md](./phases/phase-02-matching-engine.md) — Weeks 3–4 — CLOB, order types, synthetic MMs, snapshotting.
- [phases/phase-03-oms-risk.md](./phases/phase-03-oms-risk.md) — Week 5 — event-sourced OMS, pre-trade risk, reject taxonomy, drop-copy.
- [phases/phase-04-positions-pnl.md](./phases/phase-04-positions-pnl.md) — Week 6 — positions, P&L, ledger, MIS/CNC/NRML, position conversion.
- [phases/phase-05-frontend.md](./phases/phase-05-frontend.md) — Weeks 7–8 — Kite-clone UI, TradingView charts, order pad.
- [phases/phase-06-futures.md](./phases/phase-06-futures.md) — Week 9 — **NSE F&O module**: futures contracts, daily MTM, expiry, rollover.
- [phases/phase-07-options.md](./phases/phase-07-options.md) — Weeks 10–11 — **NSE F&O module**: option chain, greeks, IV solver, expiry settlement.
- [phases/phase-08-span-margin.md](./phases/phase-08-span-margin.md) — Weeks 12–13 — **NSE F&O module**: SPAN scenario engine, spread charges, exposure margin.
- [phases/phase-09-strategy-sdk.md](./phases/phase-09-strategy-sdk.md) — Week 14 — strategy runtime, deterministic backtest/live parity.
- [phases/phase-10-clearing-settlement.md](./phases/phase-10-clearing-settlement.md) — Week 15 — EOD scheduler, T+1, corporate actions, contract notes.
- [phases/phase-11-hardening.md](./phases/phase-11-hardening.md) — Week 16+ — load/chaos, kill switch, deploy, demo.

### Architectural decisions

- [adrs/README.md](./adrs/README.md) — ADR index + template.
- Seed ADRs: 0001 monorepo boundaries, 0002 event-sourced OMS, 0003 single-tenant v1, 0004 Go for hot path.

### Interview prep

- [talking-points/README.md](./talking-points/README.md) — how to use these in interviews.
- One doc per phase under [talking-points/](./talking-points/).

## Conventions used across docs

- **Bolded terms** are domain concepts defined in the [primer](./05-nse-domain-primer.md).
- Code references use `package/file.ext` relative to repo root.
- "Must / Should / Could / Won't (v1)" flags every requirement (MoSCoW).
- SLOs are stated as `p50 / p95 / p99` on your local dev box (M-class Mac or 4 vCPU / 8 GB).
- Currency is INR; time is Asia/Kolkata unless noted.

## Status

Plan authored: v1. No code yet. Start at [phases/phase-00-foundation.md](./phases/phase-00-foundation.md).