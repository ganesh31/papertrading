# Resource URLs by NotebookLM notebook

Paste **URLs** into NotebookLM as *website* sources, or **download / print to PDF** when you need stable grounding.  
Details on **which notebook is which**: [notebooklm-study-structure.md](./notebooklm-study-structure.md).

---

## How to use this file

| Short name | Full name in NotebookLM | Answers questions like… |
|------------|-------------------------|---------------------------|
| **Notebook A** | **Markets & interview prep** | What is T+1? What is SPAN *as a market/regulatory concept*? What do NSE circulars say? |
| **Notebook B** | **Delivery & phases** | How do I fetch bhavcopy? What tools run in Phase 0/11? Seed data, CLIs, vendors. |
| **Notebook C** | **Architecture & system design** | Why Go for hot path? How do services talk? APIs to mimic? Stack docs, patterns, reference code. |
| **Notebook D** | **External theory (optional)** | Textbook-only answers without mixing your repo design. |

**Local sources (not URLs):** **B** always includes PDF exports of `docs/README.md`, `00-overview`, and `docs/phases/*.md`. **C** always includes PDF exports of `02-architecture`, `01-tech-stack`, `03-data-model`, `04-cross-cutting`, `repo-layout`, and `docs/adrs/*.md` — see study-structure doc for the full list.

**NotebookLM tip legend**

| Tag | Meaning |
|-----|---------|
| **Website** | Paste URL; snapshot may go stale—re-pin after site redesigns. |
| **PDF** | Save official PDF from browser; upload file. |
| **Export PDF** | Print page → Save as PDF. |
| **JSON** | Download; summarised subset for NotebookLM if file is huge. |
| **Books** | Do not upload full copyrighted books—notes or fair-use excerpts only. |

---

## Notebook A — Markets & interview prep

*Regulators, exchanges, product mechanics, retail education, SPAN *concept*, interview vocabulary.  
**Do not load** full `02-architecture.md` or ADRs here — use **Notebook C** so answers don’t mix code layout with market rules.*

### A.1 India — regulators & exchange official

| Resource | Purpose | URL | NotebookLM tip |
|----------|---------|-----|----------------|
| SEBI — home | Master circulars, search | https://www.sebi.gov.in | **Website** — search “master circular stock brokers”, peak margin, RMS. |
| SEBI — legacy circulars listing *(path may change)* | Older listing UI | https://www.sebi.gov.in/sebiweb/other/OtherAction.do?doPaging=yes&subsystemId=54&foreignInst=Yes | **Website** — if 404, use SEBI home + search. |
| NSE — exchange circulars | Sessions, order types, MWPL | https://www.nseindia.com/resources/exchange-communication-circulars | **Website** or **Export PDF** per circular you care about. |
| NSE Clearing (NCL) — home | Settlement, SPAN landing | https://www.nseclearing.com | **Website** |
| NSE Clearing — SPAN / risk parameters *(derivatives)* | India parameter files, PDFs | https://www.nseclearing.com/products/content/der/risk_management/span_risk_parameter_files.htm | **Website** + attach small **PDF**/README only (zips are huge). |
| NSE — trading holidays & timings | Calendar for schedulers *(domain understanding)* | https://www.nseindia.com/resources/exchange-communication-holidays | **Export PDF** if holiday list PDF is linked. |
| NSE — F&O FAQs | Futures/options mechanics | https://www.nseindia.com/products-services/equity-derivatives-faqs | **Export PDF** |
| NSE Equity — cash market FAQs | Settlement, day trading | Search NSE site **“equity market trading faqs”** | **Export PDF** |
| NSE — option chain *(reference UI)* | Field names, layout language | https://www.nseindia.com/option-chain | **Website** |
| NSE — home *(thin use)* | General reference | https://www.nseindia.com | **Website** |
| BSE — corporate actions / corp tools | Cross-check corp actions | https://www.bseindia.com/corporates/ind_corporate.html | **Website** — path changes; search “BSE corporate actions” if broken. |

### A.2 SPAN & margin *(concept + global spec)*

| Resource | Purpose | URL | NotebookLM tip |
|----------|---------|-----|----------------|
| CME — SPAN methodology | Canonical scenario narrative | https://www.cmegroup.com/clearing/risk-management/span-methodology.html | **PDF** from page if offered. |
| CME — clearing / risk hub | Extra SPAN PDFs | https://www.cmegroup.com/clearing/risk-management | **PDF** picks from hub |

### A.3 Retail education & charges

| Resource | Purpose | URL | NotebookLM tip |
|----------|---------|-----|----------------|
| Zerodha Varsity — home | Modules 1, 4, 5, 6, 7 | https://zerodha.com/varsity/ | **Export PDF** per module/chapter. |
| Zerodha — brokerage calculator | Fee structure reference | Search web **“Zerodha brokerage calculator”** | **Export PDF** one snapshot page. |

---

## Notebook B — Delivery & phases

*Implementation order, data seeding, infra CLIs, replay pipeline, load tests, paid data fallbacks. Pair with **PDF exports** of phase docs from `docs/phases/`.*

### B.1 Repo docs *(local — export to PDF)*

| Resource | Purpose | NotebookLM tip |
|----------|---------|----------------|
| Entire `docs/` tree | Authoritative plan | **Local** — export `README`, `00-overview`, `phases/phase-*.md` as PDFs; refresh after edits. |
| `docs/notebooklm-study-structure.md` | Notebook naming & workflow | Same |

### B.2 Market data — fetch & archive *(build replay)*

| Resource | Purpose | URL | NotebookLM tip |
|----------|---------|-----|----------------|
| NSE — archives index | Navigate bhavcopy trees | https://archives.nseindia.com | **Website** |
| NSE — equity bhavcopy *(historical)* | EOD cash OHLCV zip pattern | https://archives.nseindia.com/content/historical/EQUITIES/ | **Website** — implement fetch locally; paste **one** CSV header row into a Doc for LM if needed. |
| NSE — derivatives bhavcopy *(historical)* | F&O EOD settle, OI | https://archives.nseindia.com/content/historical/DERIVATIVES/ | Same |
| yfinance — repo | 1m bars in Phase 1 importer | https://github.com/ranaroussi/yfinance | **Website** — README |

### B.3 Contract master & seed files

| Resource | Purpose | URL | NotebookLM tip |
|----------|---------|-----|----------------|
| Angel — public scrip master *(no auth)* | `instruments` sync | https://margincalculator.angelbroking.com/OpenAPI_File/files/OpenAPIScripMaster.json | **JSON** — too large raw; paste **field description** + tiny sample in Google Doc. |

### B.4 Phase 0 / Phase 11 tooling

| Resource | Purpose | URL | NotebookLM tip |
|----------|---------|-----|----------------|
| Docker Compose | Compose v2 reference | https://docs.docker.com/compose/ | **Website** |
| dbmate | SQL migrations | https://github.com/amacneil/dbmate | **Website** |
| k6 | Load tests | https://k6.io/docs/ | **Website** |
| Testcontainers | Integration tests | https://testcontainers.com | **Website** |

### B.5 Paid data vendors *(only if free path blocks you)*

| Resource | Purpose | URL | NotebookLM tip |
|----------|---------|-----|----------------|
| TrueData India *(verify domain)* | Tick / historical | https://www.truedata.in | **Export PDF** pricing/features page. |
| GFDL / others | Alternative vendors | Search **“GFDL India market data”** | Commercial site **Export PDF** only. |

---

## Notebook C — Architecture & system design

*Service boundaries, stack, integration patterns, broker API shapes, reference implementations, distributed-systems reading. Pair with **PDF exports** of `02-architecture`, `01-tech-stack`, `03-data-model`, `04-cross-cutting`, `repo-layout`, `docs/adrs/*`, and architecture-heavy phase PDFs (1, 2, 3, 8, 9).*

### C.1 Broker & public API documentation *(design your gateway/OMS/SDK)*

| Resource | Purpose | URL | NotebookLM tip |
|----------|---------|-----|----------------|
| Kite Connect v3 | REST/WebSocket shape reference | https://kite.trade/docs/connect/v3/ | **Website** or **Export PDF** Orders + WebSocket sections. |
| Angel One SmartAPI | Live adapter (Phase 11) | https://smartapi.angelbroking.com/docs | **Website** |
| Upstox API v2 | Alternative patterns | https://upstox.com/developer/api-documentation/ | **Website** |

### C.2 Stack — runtime, data, observability

| Resource | Purpose | URL | NotebookLM tip |
|----------|---------|-----|----------------|
| Fastify | HTTP gateway / services | https://fastify.dev | **Website** |
| Zod | Validation | https://zod.dev | **Website** |
| TimescaleDB — continuous aggregates | Ticks & candles | https://docs.timescale.com/use-timescale/latest/continuous-aggregates/ | **Export PDF** key page. |
| Redis — Streams | Event bus | https://redis.io/docs/latest/develop/data-types/streams/ | **Website** |
| OpenTelemetry — Go | Instrumentation | https://opentelemetry.io/docs/languages/go/getting-started/ | **Website** |
| OpenTelemetry — JS | Node instrumentation | https://opentelemetry.io/docs/languages/js/getting-started/ | **Website** |
| PostgreSQL — docs | SQL, constraints | https://www.postgresql.org/docs/ | **Website** |
| Protocol Buffers | IDL between Go/TS | https://protobuf.dev | **Website** |
| Grafana | Dashboards | https://grafana.com/docs/ | **Website** |
| Prometheus | Metrics | https://prometheus.io/docs/ | **Website** |
| Turborepo | Monorepo tasks | https://turbo.build/repo/docs | **Website** |
| pnpm — workspaces | Workspaces | https://pnpm.io/workspaces | **Website** |
| Go — workspaces | `go.work` | https://go.dev/ref/mod#workspaces | **Website** |
| TradingView — Charting Library | Phase 5 UI | https://www.tradingview.com/charting-library-docs/ | **Website** |

### C.3 Patterns & distributed systems *(excerpts, not whole books)*

| Resource | Purpose | URL | NotebookLM tip |
|----------|---------|-----|----------------|
| LMAX Disruptor | Single-writer / ring buffer *idea* | https://lmax-exchange.github.io/disruptor/disruptor.html | **Website** |
| Martin Kleppmann — *DDIA* *(book site)* | Logs, replication, stream processing | https://dataintensive.net | **Books** — your **notes PDF** only. |

### C.4 Indian fintech engineering *(architecture culture)*

| Resource | Purpose | URL | NotebookLM tip |
|----------|---------|-----|----------------|
| Zerodha Tech | Broker-scale engineering posts | https://zerodha.tech | **Website** — **Export PDF** favourite posts. |
| Razorpay Engineering | Payments / ledger-scale patterns | https://engineering.razorpay.com | **Website** |

### C.5 Reference OSS *(read order books & clients — don’t paste whole repos)*

| Resource | Purpose | URL | NotebookLM tip |
|----------|---------|-----|----------------|
| Go orderbook | CLOB reference | https://github.com/i25959341/orderbook | **Website** — paste README + one `.go` excerpt into Doc. |
| Liquibook (C++) | Order book patterns | https://github.com/enewhuis/liquibook | Same |
| pykiteconnect | Client ergonomics | https://github.com/zerodha/pykiteconnect | **README** in Doc |
| TimescaleDB — GitHub | Project home *(prefer docs §C.2)* | https://github.com/timescale/timescaledb | **Website** |

### C.6 Talk discovery *(paste your own transcripts or notes PDFs)*

| Idea | Query / action |
|------|----------------|
| Martin Thompson — LMAX 100 TPS narrative | Search YouTube: `Martin Thompson LMAX 100k TPS`; export transcript to Doc if needed |
| Broker / exchange craft | Search `Kailash Nadh Zerodha`, `Jane Street exchange talks` |

---

## Notebook D — External theory *(optional)*

*Long-form textbooks and fair-use excerpts only — keeps “pure theory” chats separate from your system design.*

| Book | Focus | Acquire | NotebookLM tip |
|------|-------|---------|----------------|
| Larry Harris — *Trading and Exchanges* | Microstructure | Library / purchase | **Your notes PDF** — do not upload full book. |
| John Hull — *Options, Futures, and Other Derivatives* | Derivatives pricing | Same | Fair-use **chapter excerpts** PDF only. |
| Euan Sinclair — *Volatility Trading* | Vol / Greeks | Same | Notes PDF |
| Martin Kleppmann — *Designing Data-Intensive Applications* | Same as Notebook C cite — or keep long notes **only here** if C has short excerpts | Same | Notes PDF |

---

## Cross-notebook cheatsheet *(where one URL overlaps)*

| Topic | Primary notebook | Also useful in |
|-------|-------------------|----------------|
| CME SPAN methodology PDF | **A** *(market/regulatory framing)* | **C** *(when designing margin service behaviour)* |
| NSE Clearing SPAN file page | **A** | **B** *(when parsing files in code)* |
| Kite Connect docs | **C** | **B** *(when copying endpoint paths during coding)* |
| yfinance README | **B** | — |
| SEBI/NSE circulars | **A** | — |

---

## Maintenance & suggested first upload

| Notebook | First external URLs to add |
|-----------|----------------------------|
| **A** | Varsity exported PDF + NSE F&O FAQ **Export PDF** + CME SPAN **PDF** |
| **B** | `docs/` phase PDFs + yfinance README + bhav archive index |
| **C** | `02-architecture` PDF + Kite **Export PDF** + Fastify + Timescale aggregates page **Export PDF** |
| **D** | One notes PDF |

- **Quarterly:** Re-verify SEBI / NSE / BSE links (sites move).


---

## See also

- Narrative reading list — [resources.md](./resources.md)  
- Notebook naming & workflows — [notebooklm-study-structure.md](./notebooklm-study-structure.md)

