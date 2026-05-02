# Architectural Decision Records (ADRs)

One ADR per non-obvious decision. Immutable after merge; supersede via new ADR.

Format: **Michael Nygard**. File name: `NNNN-kebab-title.md`. Numbering is permanent.

## Template

```markdown
# NNNN — <Title>

- Status: Proposed | Accepted | Superseded by NNNN | Deprecated
- Date: YYYY-MM-DD
- Deciders: <name(s)>

## Context

The forces at play. What problem are we solving? What constraints exist?

## Decision

The specific choice made, stated clearly in prose.

## Consequences

Positive, negative, neutral consequences. What becomes easier. What becomes harder. What we're explicitly giving up.

## Alternatives considered

At least two, with why they were not chosen.

## References

Links to prior art, external docs, related ADRs.
```

## Index


| #    | Title                                                   | Status   | Phase |
| ---- | ------------------------------------------------------- | -------- | ----- |
| 0001 | Monorepo with pnpm + Turborepo + Go workspaces          | Accepted | 0     |
| 0002 | Event-sourced OMS                                       | Accepted | 3     |
| 0003 | Single-tenant v1, multi-tenant v2                       | Accepted | 0     |
| 0004 | Go for the hot path                                     | Accepted | 0     |
| 0005 | Broker-adapter abstraction                              | Accepted | 1     |
| 0006 | Per-symbol actor model for matching (planned)           | Proposed | 2     |
| 0007 | Skiplist order book; ladder alternative noted (planned) | Proposed | 2     |
| 0008 | Reject-reason taxonomy (planned)                        | Proposed | 3     |
| 0009 | Projections as derived state (planned)                  | Proposed | 4     |
| 0010 | FE state split (planned)                                | Proposed | 5     |
| 0011 | MTM as settlement event, not trade (planned)            | Proposed | 6     |
| 0012 | Greeks on server, not client (planned)                  | Proposed | 7     |
| 0013 | SPAN scenarios — implementation details (planned)       | Proposed | 8     |
| 0014 | SPAN for F&O vs. VAR+ELM for cash (planned)             | Proposed | 8     |
| 0015 | Strategy runtime isolation model (planned)              | Proposed | 9     |
| 0016 | BullMQ scheduler for v1 (planned)                       | Proposed | 10    |
| 0017 | Corporate action application model (planned)            | Proposed | 10    |
| 0018 | What I'd do differently at scale (retrospective)        | Proposed | 11    |
| 0019 | Tick synthesis from 1-minute bars                     | Accepted | 1     |
| 0020 | Asset-class module boundary                             | Accepted | 0     |
| 0021 | Canonical instrument spec + contract metadata           | Accepted | 1     |
| 0022 | Risk model interface (VAR+ELM vs SPAN)                  | Accepted | 3     |


## When to write an ADR

- Picked one non-trivial option among several.
- A future you would ask "why did we do it this way?".
- An interviewer might challenge the choice.
- Reversing the decision would cost days.

Don't ADR:

- Which linter rules to turn on.
- Specific library versions.
- One-off bug workarounds (comment in code is enough).

## Review cadence

- ADRs are written when the decision is made, not retrospectively.
- Revisit quarterly; mark stale ones Deprecated with explanation.

