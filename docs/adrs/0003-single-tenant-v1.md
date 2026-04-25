# 0003 — Single-tenant v1, multi-tenant v2

- Status: Accepted
- Date: 2026-04-20
- Deciders: Project owner

## Context

A real broker is inherently multi-tenant: many clients share one broker, one exchange connection, one clearing relationship. Multi-tenancy affects:

- Auth, session, RBAC.
- Per-user rate limiting, OTR, position limits.
- Fair-queuing on the order path.
- Data isolation at the DB layer.
- Noisy neighbour management.

Building all of that in v1 eats 30% of the calendar for little learning payoff, because the hard, architecturally-interesting parts are matching, risk/margin, and F&O lifecycle — not auth plumbing.

This project has a 2–3 month budget and two goals: deep learning and a personal paper-trading tool. Multi-tenancy serves neither directly.

## Decision

**v1 is single-tenant** with one seeded user. But the data model and APIs are structured such that multi-tenant v2 requires no schema refactor:

- Every table has `user_id` (always `= '01USER'` in v1).
- Every API call carries `X-User-Id` via middleware (stub today; JWT tomorrow).
- Rate limits keyed on `user_id`, not IP.
- Matching engine partitions orders by `(symbol, user_id)` — trivial today, load-bearing tomorrow.
- Per-user config (rate limit, max lots) lives on `ref.users`.
- No hardcoded "single user" assumptions in business logic — queries filter by `user_id` parameter unconditionally.

## Consequences

**Positive**

- Saves ~1 week vs. building auth + RBAC + session + user management.
- Schemas remain readable.
- Velocity on the interesting parts.

**Negative**

- The "how would you scale to 1M users" interview answer stays theoretical in the codebase.
- Cannot publicly share the demo as a product (one user = no isolation).
- v2 migration is *easy* but not *zero* — still need to add auth, tenancy middleware, quota enforcement.

**Neutral**

- Actually one-user operation aligns with the secondary goal (personal paper trading tool).

## Alternatives considered

- **Multi-tenant from day 1** — rejected: too much non-core work, not enough ROI for the learning goal.
- **Stub auth with hardcoded assumptions scattered in code** — rejected: accumulates debt; discipline of `user_id`-everywhere is low cost today, high value tomorrow.

## The v2 migration (what it would take)

1. Add `auth` middleware (Fastify plugin) that verifies JWT and injects `X-User-Id`.
2. Add `POST /users/register` and OAuth flow.
3. Add `ref.user_settings` with per-user rate limit, max lots, features.
4. Add tenant-aware rate limiter already keyed on `user_id`.
5. Add per-user order queue in OMS (fair-queue consumer).
6. Add per-user surveillance dashboards in Grafana.
7. No schema change. No business logic change.

Estimated: 1–2 weeks.

## References

- AWS multi-tenant SaaS patterns.
- [Shopify] on shard-per-tenant strategies.
- Linear's engineering blog on workspace isolation.
- This project's `docs/00-overview.md`.
