# Phase 0 — Talking points

Short prompts you can rehearse for **foundation / repo hygiene / observability**.

---

## 1) Why a monorepo for something this small?

**Question:**  
Why not separate repos for gateway, Go services, and docs?

**90-second answer:**  
The hard parts here aren’t “many apps”, they’re **cross-cutting contracts**: protos, shared types, migrations, infra, and docs that change together. A monorepo keeps **one PR** for “schema + service + migration + ADR”, which matches how brokers evolve (everything ships together). Turborepo caches builds so we don’t pay much for colocation.

**Trade-off:**  
Tooling noise upfront vs **integration friction forever** in polyrepo.

**At scale:**  
Still monorepo-first for shared contracts; isolate deploy units per service with clear boundaries (schemas per bounded context).

---

## 2) Why OTel + Prometheus + Loki + Tempo on day one?

**Question:**  
Isn’t this overkill before real features?

**90-second answer:**  
The order path will get subtle fast (idempotency, bus lag, GC pauses). If we bolt observability later, we ship blind during the hardest phases. Day-one baseline means every service exposes **health**, **metrics**, and **traces** so regressions show up as data, not anecdotes.

**Trade-off:**  
Compose footprint vs **debuggability** when things break at night.

**At scale:**  
Same signals; exporters become vendor-managed (managed Prometheus / Grafana Cloud / Tempo).

---

## 3) Why Go on the hot path and TS elsewhere?

**Question:**  
Could you do everything in one language?

**90-second answer:**  
Matching and market-data paths care about **latency variance** and predictable runtime behavior; Go’s simplicity + goroutines fit a single-writer style book. Business rules that churn (OMS state machine, reporting) fit Node/TS iteration speed. We document the boundary so we don’t blur it casually.

**Trade-off:**  
Two runtimes vs **wrong language on the wrong tier**.

**At scale:**  
Still split by tier; maybe Rust for matching if microsecond tail latency matters — but not this project’s goal.

---

## 4) Single-tenant v1 — aren’t brokers multi-tenant?

**Question:**  
Doesn’t skipping tenancy make this toy?

**90-second answer:**  
It’s an explicit scope cut to ship core mechanics in weeks, **not an oversight**. We keep `user_id` everywhere and middleware hooks so the mechanical multi-tenant upgrade is straightforward; ADR-0003 captures the exact delta to v2.

**Trade-off:**  
Velocity vs **enterprise realism**.

**At scale:**  
Add tenancy, quotas, noisy-neighbor isolation, and auth — but keep domain boundaries stable.

---

## 5) Developer UX: Make vs Just vs Compose

**Question:**  
Why both Make and Just?

**90-second answer:**  
**Make** is the universal entrypoint for strangers cloning a public repo; **Just** delegates to Make so personal ergonomics doesn’t fork scripts. Compose stays the source of truth for infra.

**Trade-off:**  
Two runners vs **drifting duplicated commands**.
