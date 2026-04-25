# Phase 10 — Clearing, Settlement, Scheduler, Reports

**Week 15 · ~20 hrs**

Goal: the "EOD machine". Every operation a real broker runs between 15:30 and 00:00, implemented as scheduled jobs: MIS square-off, daily MTM, T+1 holdings transition, contract notes, corporate actions, day open reset.

## Prerequisites

- Phases 0–9 complete (SPAN ideal, not mandatory).

## Deliverables

- [ ] Scheduler service (Temporal or simple cron runner) executing named jobs in dependency order.
- [ ] Trading calendar (holidays, weekends) drives job schedule.
- [ ] Jobs: `intraday_squareoff`, `close_session`, `daily_mtm`, `holdings_rollover_t1`, `contract_notes`, `corp_actions_apply`, `rotate_kill_switch`, `backup_event_log`, `day_open_reset`.
- [ ] Contract note PDFs generated per user per day (Puppeteer + HTML template).
- [ ] Corporate actions ingested from NSE corporate actions file; applied on ex-date.
- [ ] Day-open reset: clears intraday caches, rotates circuit flags, refreshes contract master.
- [ ] Job observability: run log, duration, success/failure metrics.
- [ ] Retry + idempotency for every job.
- [ ] ADR-0016 (scheduler choice), ADR-0017 (settlement & corp-action model).
- [ ] Talking-points doc.

## Schedule (IST)

| Time | Job | Notes |
|------|-----|-------|
| 15:15 | `intraday_squareoff_equity` | MIS equity flatten at market. |
| 15:25 | `intraday_squareoff_fo` | MIS F&O flatten. |
| 15:30 | `close_session` | Stop accepting orders; snapshot books; mark day close. |
| 16:30 | `daily_mtm` | Run MTM for futures using bhavcopy. |
| 17:00 | `expiry_settlement` | If expiry day, settle expired contracts. |
| 18:00 | `contract_notes` | Generate per-user PDF; file in `reports.contract_notes`. |
| 20:00 | `holdings_rollover_t1` | T+1 CNC → holdings (actually: run at 00:00 for positions closed yesterday). |
| 22:00 | `corp_actions_apply` | For contracts with ex-date tomorrow. |
| 23:30 | `backup_event_log` | Dump event log to cold storage (v1: a `.tar.gz` in `/backups`). |
| 08:30 (next day) | `instruments_sync` | Refresh contract master; pull today's freeze & SPAN files. |
| 09:00 | `day_open_reset` | Reset intraday counters; arm circuit flags; clear idempotency cache. |

Job dependencies expressed as DAG; scheduler refuses to run `daily_mtm` before `close_session`.

## Job anatomy (every job)

Every scheduled job must be:

1. **Idempotent** — running twice on the same day produces the same result.
2. **Checkpointed** — long jobs (MTM, corp actions) write progress so they can resume after crash.
3. **Observable** — start/end/duration, rows processed, errors, emit `JobCompleted` event.
4. **Retryable** — failed jobs are retried with backoff; after 3 failures, alert.
5. **Restorable** — every write is compensable via reversal entry (ledger) or projection rebuild (positions).

## Scheduler choice

Options:

- **Temporal.io self-hosted** — workflow orchestration, retries, visibility. Overkill, but impressive.
- **BullMQ + cron** — lightweight, Node-native, good enough.
- **Hand-rolled + pg advisory locks** — minimal, but you own the reliability.

Recommendation for v1: **BullMQ** with Redis (you already have Redis). One queue per job; cron triggers via `BullMQ Queue#add`. Advisory lock on job name prevents double-run across replicas.

ADR notes the upgrade path to Temporal when multi-tenant/multi-region matters.

## Tasks

### 10.1 Scheduler service

- `services/scheduler` (Node).
- Loads job manifest from YAML; builds DAG; schedules via BullMQ.
- Each job is a Node module exporting `run(ctx)`; ctx has DB, bus, logger, metrics.
- `GET /jobs/status` lists recent runs.

### 10.2 Trading calendar

- `infra/seed/holidays.json` refreshed yearly from NSE.
- `getNextTradingDay(date)`, `isTradingDay(date)`, `previousTradingDay(date)` helpers.
- Scheduler skips weekends + holidays.
- Job names are `YYYY-MM-DD:<job>` for idempotency keys.

### 10.3 Intraday square-off

- Query MIS positions with non-zero qty.
- Place `MARKET` orders with `system=true, reason=MIS_AUTO_SQUAREOFF`.
- Bypass user rate limit; still go through risk (for form, not blocking).
- Track failures (e.g., no liquidity) and surface in tomorrow's funds page.

### 10.4 Close session

- Flip `market.flags` stream: `halted=true, reason=DAY_CLOSE`.
- Cancel all DAY orders still open → state `EXPIRED`.
- Snapshot all order books (already done every 1s via Phase 2; this is a final forced one).
- Emit `DayClosed` event.

### 10.5 Daily MTM (Phase 6 already specced)

- Now runs under scheduler. Picks settlement prices from ingested bhavcopy.

### 10.6 Expiry settlement (Phase 6 + 7)

- Detect contracts with expiry == today. Settle per segment (future cash-settle, option intrinsic).
- Mark instrument `EXPIRED`.

### 10.7 Contract notes

- Puppeteer template at `services/reports/templates/contract-note.html`.
- Content: trades list with charges breakdown, totals, disclaimer, PAN/UCC (from user ref).
- Render PDF; store path in `reports.contract_notes`.
- `GET /reports/contract-notes?from=&to=` returns list; `GET /reports/contract-notes/:id/pdf` streams.

### 10.8 T+1 holdings rollover

- For each CNC position with `net_qty > 0`, transferred on the night *after* trade day → so today's job processes yesterday's CNC buys.
- Create `portfolio.holdings` lot row with `acquired_on = yesterday`.
- Zero out CNC position row.
- Emit `HoldingsUpdated`.

### 10.9 Corporate actions

- Source: NSE corporate actions file (`CF_CA_<date>.csv`).
- Types handled: bonus, split, dividend (cash).
- Applied on ex-date (T-1 night for T ex-date).
- For **bonus 1:1**: `holdings.qty ×= 2; holdings.avg_price /= 2;`
- For **split 1:N**: `qty ×= N; avg_price /= N;`
- For **dividend**: credit CASH by `div_per_share × qty`; no position change for equity; for F&O: some contracts get dividend-adjusted per NSE rules — include adjustment to futures' fair price in reval (optional v1, document gap).
- Emit `CorpActionApplied` audit event.

### 10.10 Day-open reset

- Clear Redis idempotency keys (`DEL idemp:*` older than 24h).
- Re-arm MWCB flags.
- Refresh in-memory caches in services.
- Emit `DayOpen` event.

### 10.11 Backup event log

- `pg_dump` only `oms.order_events`, `oms.trades`, `md.ticks` (hot subset) → `/backups/YYYY-MM-DD.tar.gz`.
- Rotate: keep 30 days; older → delete (v1) or move to S3 (v2).

### 10.12 FE additions

- "Reports" section: contract notes download, P&L report (range), margin statement.
- "Corporate actions" panel: upcoming CAs affecting your holdings.
- Scheduler status page (admin-only).

## Metrics

- `job_runs_total{name,status}`
- `job_duration_ms{name}`
- `job_failures_total{name}`
- `contract_notes_generated_total`
- `corp_actions_applied_total{type}`

## Performance targets

- Daily MTM for 10k positions < 60 s.
- Contract notes batch for 1000 users < 5 min.
- Holdings rollover < 10 s.
- Day-open reset < 5 s (must finish before 09:15).

## Testing

- Unit: every job's pure logic (MTM math, CA adjustments, rollover).
- Integration: end-to-end day cycle on fixture data → assert all projections consistent after.
- Idempotency: run each job twice → second is no-op.
- Crash-recovery: kill job mid-run → resume → correct outcome.

## Common pitfalls

- Timezone drift: scheduler runs in UTC by default; force IST or convert.
- Job overlap: long MTM blocking contract notes — model DAG dependencies, not just cron.
- CA ex-date confusion: NSE uses "ex-date" (T from which price excludes dividend). Apply night *before* ex-date to reflect opening price fairly.
- Corporate action on a stock that doesn't have positions → skip (don't error).
- Contract note regenerated differently on re-run → idempotency miss. Use deterministic file names.
- Not rotating event-log backups → disk fills up.
- Not testing DST transitions (India doesn't DST, but if you ever support global, this bites).

## Interview talking points

- Scheduler as a DAG vs. cron — why DAG for dependencies.
- Why BullMQ over Temporal for v1 (and when you'd switch).
- Advisory locks as a poor-man's leader election.
- Idempotency in jobs — designs that survive duplicate invocations.
- Corporate actions and F&O adjustment — surprisingly intricate.
- Backups: event log is enough to reconstruct everything — projections are cheap.
- "Runbook-first" operations — every job has a runbook before it has a green tick in CI.

## Resources

- BullMQ docs: <https://docs.bullmq.io>
- Temporal concepts (for the v2 narrative): <https://docs.temporal.io/concepts>
- NSE corporate actions file location: <https://archives.nseindia.com/archives/equities/bhavcopy/pr/>
- SEBI T+1 circular (Jan 2023 rollout).
- Puppeteer PDF recipes.
- Postgres advisory locks: <https://www.postgresql.org/docs/current/explicit-locking.html#ADVISORY-LOCKS>

## Exit checklist

- [ ] End a simulated trading day; all jobs run green.
- [ ] Contract note PDF exists for the user with charges adding up correctly.
- [ ] A fake split corporate action applied to INFY doubles holdings qty & halves avg_price.
- [ ] `GET /jobs/status` shows green across the board.
- [ ] ADR-0016, ADR-0017 merged.
