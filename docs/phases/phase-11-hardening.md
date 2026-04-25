# Phase 11 — Hardening, Deployment, Showcase

**Week 16+ · ~20 hrs**

Goal: turn the project into a credible portfolio artifact. Load + chaos tests, kill switch, deployment, demo, docs polish.

## Prerequisites

- Phases 0–10 complete.

## Deliverables

- [ ] Load test baseline: 10k orders/min sustained, burst 20k/min, p99 SLOs met.
- [ ] Chaos test suite covering 5 failure modes.
- [ ] Kill switch implemented + runbook written.
- [ ] Surveillance consumer: OTR alerts, position-limit alerts.
- [ ] Public demo URL with HTTPS, demo user credentials.
- [ ] Seed script that populates replay data → demo works offline.
- [ ] 5-minute demo video recorded (loom/youtube), linked in README.
- [ ] README.md polished: architecture diagram, quickstart, demo link, ADR index, SLO numbers.
- [ ] Public blog post / gist summarizing "How I built NSE in my laptop".
- [ ] LinkedIn + resume one-liner ready.
- [ ] Final ADR-0018 (what I'd do differently at scale).

## Tasks

### 11.1 Load testing

- k6 scripts in `infra/k6/`:
  - `orders_steady.js` — 10k/min for 30 min, place + cancel mix (70/30).
  - `orders_burst.js` — ramp to 20k/min.
  - `md_firehose.js` — 20k ticks/sec into MD for 10 min.
- Capture: p50/p95/p99 per endpoint, error rate, GC/event-loop metrics.
- Commit Grafana dashboard JSON `Load Test`.
- Fix bottlenecks uncovered (common: DB connection pool; WS fan-out).
- Publish numbers in `docs/performance.md`.

### 11.2 Chaos tests

- `chaos/kill-matching.sh` — kill ME mid-test; assert recovery from event log; no phantom orders; no lost trades.
- `chaos/postgres-bounce.sh` — restart Postgres; assert OMS returns 503 during, resumes cleanly.
- `chaos/redis-bounce.sh` — restart Redis; idempotency falls back, rate-limit gracefully degrades.
- `chaos/md-disconnect.sh` — block MD egress; assert staleness alerts + order rejection on stale symbols.
- `chaos/bus-lag.sh` — throttle Redis Streams; assert consumer lag alert.
- Each is a scripted test; run manually + in a weekly CI job.

### 11.3 Kill switch

- `POST /admin/kill-switch { armed: true, scope: 'ALL'|'USER:<id>'|'SEGMENT:<seg>' }`.
- OMS consumes; rejects new orders with `MARKET_HALTED` (or a new `KILL_SWITCH_ACTIVE`).
- ME halts matching; orders queue.
- Cancel outstanding orders option.
- Runbook: `infra/runbooks/kill-switch.md` with triggers and recovery.
- Metric: `kill_switch_active` gauge + audit log.

### 11.4 Surveillance (expanded)

- OTR consumer from Phase 3 now generates daily report.
- Position-limit consumer: per-client limits from config + MWPL from NSE file; alerts when hit.
- Alerts into Grafana; optional webhook to a Discord / Slack channel.

### 11.5 Deployment

- Target: single VPS (Hetzner CX22 or Contabo, ~₹500/month) running docker-compose.
- Caddy for HTTPS + reverse proxy: `app.example.com` → gateway.
- Traefik alternative.
- GitHub Actions deploy on `main`: SSH + `docker compose pull && up -d`.
- Backups: nightly `pg_dump` to Backblaze B2 (free tier) or local volume.
- Uptime monitoring: free uptime-kuma container on same VPS.

Alternative: single-node k3s for the k8s-on-resume story. `helm` charts per service. Heavier to operate; pick based on interview targets.

### 11.6 Demo data

- `pt seed demo` command: imports 3 months of replay data for 10 symbols, places some fixture trades, generates contract notes. Result: a visitor opening the demo URL sees a populated account.

### 11.7 README overhaul

Structure:

1. One-sentence elevator pitch.
2. Screenshot or GIF.
3. Live demo link + credentials.
4. Video link.
5. Architecture diagram.
6. Stack summary (badges).
7. Quickstart (`git clone; just up; open :5173`).
8. Feature matrix.
9. SLOs achieved (table).
10. ADR index table.
11. Resources.
12. License.

### 11.8 Blog post / writeup

- ~3000 words: "Building a paper-trading system that mirrors NSE".
- Sections: motivation, architecture tour, hardest problem (pick: matching engine or SPAN), what I'd do differently, resources.
- Publish on Medium + crosspost to Dev.to. Link on LinkedIn.

### 11.9 Resume one-liner + LinkedIn headline

> Built a paper-trading platform modelled on NSE/BSE end-to-end: own CLOB matching engine (Go), event-sourced OMS, SPAN + Exposure margin engine, Kite-clone UI (React/TS), strategy SDK with live/backtest parity. p99 order-ack < 50 ms at 10k orders/min. Code + live demo + writeup.

### 11.10 Interview rehearsal

- Do 3 mock system-design interviews using the project as the reference system.
- Record yourself on Loom answering: "Walk me through the order lifecycle." "How does your SPAN differ from real SPAN?" "How would you shard this to 1M users?"
- Watch back; fix filler words, tighten narrative.

## Performance targets (final SLOs to publish)

| Metric | Target | Measured |
|--------|--------|----------|
| p99 order-ack (E2E) | < 50 ms | ___ |
| p99 SPAN calc | < 50 ms | ___ |
| p99 chart load | < 1.5 s | ___ |
| Sustained orders/min | 10k | ___ |
| MD tick-to-FE p99 | < 300 ms | ___ |
| Recovery after ME crash | < 30 s | ___ |
| Cold boot time | < 60 s | ___ |

Publish the actual measured numbers in README.

## Resources

- k6 docs + "k6 patterns" blog posts.
- Chaos engineering literature (Gremlin, Netflix Chaos Monkey).
- Kailash Nadh / Nithin Kamath on running Zerodha systems — for tonal inspiration.
- Hetzner pricing; Contabo pricing.
- Caddy: <https://caddyserver.com>
- Uptime Kuma: <https://github.com/louislam/uptime-kuma>

## Exit checklist (project-complete signal)

- [ ] Load tests checked in, run green.
- [ ] Chaos scripts checked in, run green.
- [ ] Demo URL live; you can give it to a stranger.
- [ ] Video published.
- [ ] README linked from your resume, LinkedIn, GitHub profile.
- [ ] Blog post live.
- [ ] ADR count ≥ 18.
- [ ] You can talk about any phase for 10 minutes cold.

## What comes after (explicitly out of scope)

- Multi-tenant / SaaS mode (separate project).
- Kubernetes + Helm + service mesh.
- GTT/BO/CO, basket orders.
- Currency & commodity derivatives.
- Mobile app (PWA is enough).
- Tax reports (Form 3CD style) — big, distinct effort.
- Real broker integration (requires SEBI registration — not your path here).

Ship, then move on.
