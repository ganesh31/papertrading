# 0001 — Monorepo with pnpm + Turborepo + Go workspaces

- Status: Accepted
- Date: 2026-04-20
- Deciders: Project owner

## Context

This project spans three deliverables that share contracts:

- Node/TS services (gateway, OMS, risk, portfolio, reports, strategy-runner).
- Go services (matching engine, MD gateway, SPAN, synthetic MMs).
- React frontend.

These components share Protobuf IDLs, Zod schemas, and a TypeScript SDK. Version drift across repos would cost time disproportionate to the project size (single developer, ~3 hrs/day budget).

A polyrepo would also force early publication of internal packages, PR coordination across repos, and release sequencing — complexity tax we don't want.

## Decision

Single monorepo with:

- **pnpm workspaces** for TS packages and services.
- **Turborepo** for task orchestration (lint/test/build) with remote caching off in v1.
- **Go workspaces** (`go.work`) for the Go services and shared Go modules.
- Protobuf sources in `packages/protos/`, generated outputs in `packages/protos/go/` and `packages/protos/ts/`.
- Zod schemas in `packages/contracts/`.

Folder boundaries:

- `apps/` — runtime apps (only `web` in v1).
- `services/` — backend services (Node + `services/go/*` for Go).
- `packages/` — shared libraries.
- `infra/` — docker-compose, grafana, seed, migrations.
- `docs/` — this.

## Consequences

**Positive**

- A single PR can change a proto, the server consumer, and the client consumer; CI runs all affected packages.
- No version drift across contracts.
- Single `just up` brings the whole world up.
- Shared tooling (lint, TS config, CI).

**Negative**

- Tooling complexity at the start (Turbo pipeline, Go workspaces).
- CI cache warming takes a few iterations.
- If the project ever multi-tenant-splits, we'll need to extract packages — planned for later.

**Neutral**

- Mixed language in one repo is fine with separate task graphs.

## Alternatives considered

- **Polyrepo** — rejected: too much coordination cost for a single-dev project.
- **Nx** — rejected: heavier than Turbo for this size; the task graph we need is simple.
- **Bazel** — rejected: overkill; steep learning curve.

## References

- <https://turbo.build/repo/docs>
- <https://pnpm.io/workspaces>
- <https://go.dev/ref/mod#workspaces>
- `docs/repo-layout.md`
