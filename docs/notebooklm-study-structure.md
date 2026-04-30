# NotebookLM — Study structure for this project

NotebookLM (“Your sources, grounded answers”) fits this repo well because you control what it reads: your phased plan, ADRs, NSE primer, and external PDFs/book chapters.

**→ [notebooklm-resource-urls.md](./notebooklm-resource-urls.md)** — external URLs **by notebook name** (Markets / Delivery / Architecture / External theory).

---

## How NotebookLM helps (mental model)

| Feature | Best for |
|---------|----------|
| **Chat** | “Explain Phase 8 vs Phase 9,” “Quiz me on SPAN,” “What do I implement first?” |
| **Audio Overview** | Commute recap of one phase or one ADR (~10–15 min). |
| **Study guide / FAQs** (_if your region shows these_) | Condensed bullets from uploaded sources. |
| **Mind map** (_sometimes available_) | One-screen map of phases → services → data stores. |

**Constraint:** Answers are grounded **only** in sources you attached to *that notebook*. If a topic isn’t in a source, it may hallucinate — so attach the right docs (or paste key excerpts).

---

## Architecture vs domain vs delivery (three different questions)

Treat these as **three different notebooks**. If you mix them, answers blend “NSE rules,” “your microservices,” and “this week’s phase” in one paragraph — weak for both learning and interviews.

| Lens | Question | What “good” grounding is |
|------|----------|---------------------------|
| **Markets** | What are the rules and products of Indian exchanges? | Domain primer, Varsity, regulators, product mechanics. |
| **Architecture** | How is *this* system shaped — boundaries, flows, data, trade-offs? | `02-architecture`, data model, tech stack, ADRs, cross-cutting, pattern refs. |
| **Delivery** | What do I build, in what order, with what checkpoints? | `00-overview`, `README`, Phase 0–11 docs as the execution spine. |

**Architecture** here means: bounded contexts, **who owns which data**, sync vs async paths (e.g. risk on order path), persistence model (event log + projections), buses and gRPC edges, failure modes, observability, and **why** (ADRs). That is **not** the same document set as “tell me about MIS vs NRML.”

---

## Recommended: three core notebooks + optional fourth

### Notebook A — **Markets & interview prep**

**Purpose:** NSE/BSE semantics, products, charges, margin *as market concepts*, interview language that doesn’t assume your repo layout.

**Sources**

- `docs/05-nse-domain-primer.md` (core)
- `docs/phases/phase-06-futures.md`, `phase-07-options.md`
- `docs/phases/phase-08-span-margin.md` — *read for “what SPAN is for”* alongside CME/NSE PDFs
- `docs/talking-points/phase-08.md`, `phase-02.md` — market/microstructure angles
- External: Zerodha Varsity PDFs, SEBI/NSE PDFs, CME SPAN PDF, Hull excerpts

**Leave out of A:** `docs/02-architecture.md`, full `repo-layout.md`, and **ADRs** — those move to **C** so you aren’t asking “what is NRML?” and getting “the portfolio service exposes REST.”

**Typical prompts**

- Explain peak margin vs SPAN scanning risk *without* referring to my codebase.
- Quiz me on settlement and STT for options — sources only.

---

### Notebook B — **Delivery & phases**

**Purpose:** Timeline, DoD checklists, phase dependencies, “what must exist before Phase N,” replay setup steps.

**Sources**

- `docs/00-overview.md`, `docs/README.md`
- `docs/phases/phase-00-*.md` through `phase-11-*.md` (refresh PDFs when docs change)
- `docs/notebooklm-study-structure.md` — optional; same folder coherency

**Typical prompts**

- Phase 3 deliverables as a checkbox list from uploaded phase docs only.
- Minimal scope if I skip options for v1 — cite phases.

**Note:** Phase docs *contain* architecture detail. **B** is still the right place for “what week am I in?” **C** holds the canonical design docs for “why is it built this way?”

---

### Notebook C — **Architecture & system design**

**Purpose:** Interview answers of the form “walk me through your system,” trade-off discussions, and ADR-aligned reasoning.

**Sources (load in this order)**

1. `docs/02-architecture.md` — diagrams, ports, bounded contexts, scaling story, failure modes
2. `docs/01-tech-stack.md` — Go vs Node split, rationale
3. `docs/03-data-model.md` — events, projections, ledger
4. `docs/04-cross-cutting.md` — o11y, testing, security
5. `docs/repo-layout.md`
6. **All** `docs/adrs/*.md`
7. **Phase excerpts** (PDF each or one combined “arch phases” doc): `phase-01` (adapters + replay), `phase-02` (matching), `phase-03` (OMS + risk path), `phase-08` (SPAN service boundary), `phase-09` (strategy + clock)
8. Short external PDFs: LMAX Disruptor page; *DDIA* notes or 1–2 fair-use chapter excerpts (not full book)

**Typical prompts**

- List every named service, port, and what protocol it uses to talk downstream — from `02-architecture.md` only.
- Trace a place order from Gateway to Trade event: which services touch it and in what order?
- What happens if matching engine dies mid-fill, per uploaded failure-mode section?
- Summarize ADR 0002 consequences if we removed event sourcing.

**Audio Overview:** Pair `02-architecture.md` + one hot ADR (e.g. event-sourced OMS) for a compact “system design podcast.”

---

### Notebook D (optional) — **External theory only**

**Purpose:** Hull, Harris, regulators **without** your design mixed in — textbook Q&A only.

---

## At-a-glance: which notebook answers which prompt?

| Your question | Open notebook |
|---------------|----------------|
| What is T+1 peak margin? | **A** |
| Why is the matching engine single-writer per symbol? | **C** (+ phase-02 as source) |
| What do I ship in Phase 5? | **B** |
| How does `oms.order_events` relate to `oms.orders`? | **C** (`03-data-model` + ADR 0002) |
| Full SPAN maths from regulatory PDFs | **A** or **D** (not C unless your phase-08 is also loaded) |

---


NotebookLM accepts **Google Docs, PDFs, pasted text, and URLs** (capabilities vary by account/region).

| Your artifact | Practical approach |
|----------------|-------------------|
| Markdown (`.md` in repo) | **Print to PDF** from VS Code / GitHub preview, **or** paste each file into a Google Doc titled `papertrading-phase-02` and upload Doc. PDF is often fastest. |
| Long phase doc | Split into logical PDFs: e.g. `phase-08-span-part1-intro.pdf`, `part2-engine.pdf`. |
| Web pages | Use **NotebookLM “Website” source** only for stable URLs; scrape key content into PDF if the page changes often. |
| Huge PDF | Split by chapter — smaller sources = more precise citations. |

**Naming uploads** consistently: `YYYY-MM-domain-primer.pdf`, `phase-07-options.pdf` — helps you and the model cite “which doc.”

---

## Effective habits (NotebookLM hygiene)

### 1. One question = one notebook theme

Don’t merge “interview cram” with “implement MD gateway” unless you intentionally want crossover questions.

### 2. Prefer “according to my uploaded sources…”

Starts every important answer with that phrase in your prompts so answers stay anchored.

Example

> According to only the uploaded phase documents, what services exist in Phase 3 and what does each own?

### 3. Verification loop

NotebookLM compresses — **facts can drift.**

After a session, **spot-check** 2–3 claims against the actual markdown file (`docs/phases/...`). If wrong, shorten sources or paste the authoritative paragraph into its own Doc.

### 4. Incremental ingestion

Upload **incrementally** — e.g.

- Week you start Phase 0: notebooks **B** (`overview`, `readme`, `phase-00`) + **C** skeleton (`02-architecture` + ADR 0001–0004 only).
- When you start coding a phase: add that `phase-N` PDF to **B**; if the phase is architecture-heavy, also add the same PDF to **C**.
- **A (Markets)** grows steadily with Varsity/regulator PDFs — no need to upload all at once.

Updating is less cognitive load than one giant notebook with 500 pages.

### 5. Duplicate “must remember” excerpts

Paste a **tiny** “Ground truth” Google Doc bullet list inside the notebook:

- Default MD adapter is `nse_replay`.
- Event source of truth: `oms.order_events`.

Short pages get cited reliably.

---

## Prompt templates you can reuse

### Learning (markets — Notebook A)

- Explain [topic] assuming I trade Indian markets but have not built brokers before.

- Quiz me with 10 questions on [NSE topic / Varsity module]. Show answers in a second message.

- Summarize prerequisites for Phase N as a numbered list — use **Delivery** notebook (B), not A, unless the prerequisite is purely domain.

### Architecture (Notebook C)

- What are the trade-offs in uploaded ADRs for [decision X]?

- Describe the order-placement path from Gateway to Portfolio update using only `02-architecture` and `03-data-model`.

- Compare event-sourced OMS vs naive CRUD — cite ADR 0002 and data model.

### Delivery (Notebook B)

- From uploaded phase docs only, list deliverables due by end of Phase 2 — checkbox format.

- What could I ship in MVP if I cut scope — cite phase docs?

- Produce a dependency graph: Phase order as mermaid-ish text nodes.

### Interview

- Simulate a senior architect interviewer asking about matching engine consistency and recovery — answer using **Architecture** notebook sources (phase-02 + relevant ADRs).

---

## Weekly workflow suggestion

| Day / when | Notebook | Action |
|------------|----------|--------|
| **Start of week** | **B — Delivery** | Paste “This week goal: Phase N” — ask for a 5-bullet focus list from `phase-N` doc only |
| **Mid-week** | **A** or **C** | **A**: domain quiz (circuits, settlement, products). **C**: architecture quiz (services, event log, failure modes) — alternate weeks |
| **End of week** | **B — Delivery** | Retrospective: what Phase N promised vs checklist — reconcile with markdown in repo |
| **Commute** | **A** | Audio Overview on Varsity chapter PDF or domain primer |
| **Commute alternate** | **C** | Audio Overview on `02-architecture` + one ADR |

---

## Limitations — plan around them

- **No live code**: NotebookLM doesn’t compile or run Docker; use it for *planning and explanations*, Cursor/IDE for code.
- **Stale uploads**: After you edit docs, **re-export PDF** / re-upload Docs.
- **Copyright**: Avoid uploading full copyrighted books — use **short excerpts**, Varsity-exported chapters, official circulars PDFs.

---

## Minimum viable setup (~45 minutes)

1. **Notebook A (Markets)** → `05-nse-domain-primer` as PDF + Varsity one module as PDF (or domain primer only if short on time).

2. **Notebook B (Delivery)** → `README`, `00-overview`, `phase-00`, `phase-01` as PDFs.

3. **Notebook C (Architecture)** → `02-architecture`, `03-data-model`, ADRs **0001–0004** as PDFs. Add `01-tech-stack` when you have another 10 minutes.

4. Smoke-test: ask **B** “What are Phase 0 deliverables?”; ask **C** “Which services own OMS vs matching?”; ask **A** “What is MWPL?” — each answer should cite a different corpus.

---

## Tie-in to your repo folder layout

Suggested sync between disk and notebooks:

```
docs/05-nse-domain-primer.md       → Notebook A (Markets)
docs/00-overview.md + README       → Notebook B (Delivery) top
docs/phases/phase-*.md             → Notebook B (Delivery) spine; copy architecture-heavy phases also into C as needed
docs/02-architecture.md            → Notebook C (Architecture) anchor
docs/01-tech-stack.md              → C
docs/03-data-model.md              → C
docs/04-cross-cutting.md           → C
docs/repo-layout.md               → C
docs/adrs/*.md                     → C only (not A — keeps markets chat clean)
docs/talking-points/*.md           → Split: markets angles → A; system-design → C
External PDFs (Varsity, regulators) → A
External PDFs (DDIA notes, LMAX web page export) → C or D
```

URLs for stack/patterns: [notebooklm-resource-urls.md](./notebooklm-resource-urls.md) **Notebook C** tables supplement **architecture PDF exports**.

This file stays in-repo as the checklist; NotebookLM mirrors it logically, not physically.
