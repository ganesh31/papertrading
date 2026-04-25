# 0002 — Event-sourced OMS

- Status: Accepted
- Date: 2026-04-20
- Deciders: Project owner

## Context

An OMS is heavily audited in the real world. SEBI requires 7-year retention of audit trails; every state transition must be reconstructable. Additionally, the matching engine is authoritative for some state transitions (OPEN/PARTIAL/FILLED) while OMS itself owns others (NEW/VALIDATED/RISK_CHECKED/REJECTED). Conflicting sources of truth easily cause drift.

Paper-trading users won't notice, but interviewers will probe exactly this area: "how do you audit order lineage?", "what happens if projection drifts from reality?".

## Decision

The OMS uses **event sourcing** as the canonical model:

- Append-only `oms.order_events` table; one row per lifecycle transition.
- `oms.orders` is a **projection** of the event stream, kept in sync via a projector.
- Projections can always be rebuilt via `pt admin rebuild orders`.
- Trades (`oms.trades`) are immutable; authoritative from the matching engine's `Trade` event.
- Every service that mutates state (OMS, ME) publishes events; the projection layer only consumes.

Constraints:

- No `UPDATE` or `DELETE` on `oms.order_events`.
- Projection writes are idempotent (keyed on `event_id`).
- Schema of the event payload is versioned (`payload.version`).

## Consequences

**Positive**

- Full audit trail for free.
- Replayability for bug hunts.
- Clean story for interviewers.
- New projections (e.g., a per-symbol analytics cube) can be added later without touching writers.

**Negative**

- Slightly more code than state-based persistence.
- Schema migrations on event payloads need care (accept old versions forever).
- Care required to keep projections fast enough (batching, indexes).

**Neutral**

- Event store is in the same Postgres instance — easy to manage; a "real" system might use Kafka + Debezium, but we don't need that here.

## Alternatives considered

- **Plain state-based CRUD** — rejected: no audit trail without bolting one on; projections from trades for positions are already event-sourced, so consistency across aggregate types is easier with a unified model.
- **Dedicated event store (EventStoreDB)** — rejected: one more DB to operate. Postgres-as-event-log is sufficient at our scale.
- **CQRS with separate read DB** — partially adopted (projections *are* a separate table), but same DB instance.

## References

- Martin Fowler — Event Sourcing: <https://martinfowler.com/eaaDev/EventSourcing.html>
- Greg Young talks on CQRS + ES.
- Kleppmann, *DDIA*, ch. 11.
- SEBI Master Circular — record retention norms.
- Stripe API idempotency blog.
