# 03 — Data Model

Same logical DB for v1, split by **schema** per bounded context. Every row carries `user_id` even in single-tenant mode.

- Schemas: `oms`, `portfolio`, `md`, `reports`, `ref` (reference / shared).
- Conventions: `snake_case`, `created_at`/`updated_at` timestamptz default `now()`, monetary values in `numeric(18,4)`, IDs are ULIDs (`text`) not auto-ints.

## `ref` — reference data (shared read-only)

### `ref.users`

```sql
create table ref.users (
  user_id        text primary key,
  email          text unique not null,
  display_name   text not null,
  created_at     timestamptz default now()
);
```

v1 seeded with a single row.

### `ref.instruments` (contract master)

Re-ingested daily from Angel's `OpenAPIScripMaster.json` and NSE's F&O contract file.

```sql
create table ref.instruments (
  instrument_id   text primary key,                 -- ULID
  tradingsymbol   text not null,                    -- e.g. "NIFTY24DEC25000CE"
  exchange        text not null,                    -- NSE | BSE | NFO | BFO
  segment         text not null,                    -- EQ | FUT | OPT | INDEX
  instrument_type text,                             -- CE | PE | FUT | EQ
  underlying      text,                             -- NIFTY, INFY...
  expiry          date,
  strike          numeric(18,4),
  lot_size        integer not null default 1,
  tick_size       numeric(18,4) not null default 0.05,
  freeze_qty      integer,
  isin            text,
  listing_date    date,
  status          text not null default 'ACTIVE',   -- ACTIVE | SUSPENDED | EXPIRED
  metadata        jsonb default '{}'::jsonb,
  created_at      timestamptz default now()
);
create unique index on ref.instruments (exchange, tradingsymbol);
create index on ref.instruments (underlying, expiry);
```

## `oms` — orders (event-sourced)

### `oms.order_events` (source of truth, append-only)

```sql
create table oms.order_events (
  event_id    text primary key,            -- ULID
  order_id    text not null,               -- ULID
  user_id     text not null,
  seq         bigint not null,             -- monotonic per order_id
  event_type  text not null,               -- PLACED | VALIDATED | ACCEPTED | REJECTED | OPEN | PARTIAL | FILLED | CANCELLED | MODIFIED | EXPIRED
  payload     jsonb not null,
  occurred_at timestamptz not null default now()
);
create unique index on oms.order_events (order_id, seq);
create index on oms.order_events (user_id, occurred_at desc);
```

No `UPDATE`, no `DELETE` — ever.

### `oms.orders` (projection, rebuildable)

```sql
create table oms.orders (
  order_id          text primary key,
  user_id           text not null,
  instrument_id     text not null references ref.instruments,
  side              text not null,             -- BUY | SELL
  product           text not null,             -- MIS | CNC | NRML
  order_type        text not null,             -- LIMIT | MARKET | SL | SL_M | IOC | FOK
  validity          text not null,             -- DAY | IOC | GTT
  quantity          integer not null,
  disclosed_qty     integer,
  price             numeric(18,4),
  trigger_price     numeric(18,4),
  iceberg_legs      integer,                   -- NULL unless ICEBERG
  status            text not null,
  filled_qty        integer not null default 0,
  avg_fill_price    numeric(18,4),
  reject_reason     text,
  placed_at         timestamptz not null,
  updated_at        timestamptz not null,
  client_order_id   text,                      -- for idempotency
  parent_order_id   text                       -- for BO/CO later
);
create index on oms.orders (user_id, status, placed_at desc);
create index on oms.orders (instrument_id, status);
```

### `oms.trades` (immutable, authoritative)

```sql
create table oms.trades (
  trade_id      text primary key,
  order_id      text not null references oms.orders,
  user_id       text not null,
  instrument_id text not null,
  side          text not null,
  quantity      integer not null,
  price         numeric(18,4) not null,
  traded_at     timestamptz not null,
  counter_id    text,                          -- synthetic MM id or another order_id
  taker         boolean not null               -- aggressor flag
);
create index on oms.trades (user_id, traded_at desc);
create index on oms.trades (instrument_id, traded_at desc);
```

## `portfolio` — derived state

### `portfolio.positions` (projection from trades; keyed by product)

```sql
create table portfolio.positions (
  position_id   text primary key,
  user_id       text not null,
  instrument_id text not null,
  product       text not null,                 -- MIS | NRML (CNC rolls into holdings)
  net_qty       integer not null,
  avg_price     numeric(18,4) not null,
  realised_pnl  numeric(18,4) not null default 0,
  updated_at    timestamptz not null default now(),
  unique (user_id, instrument_id, product)
);
```

`unrealised_pnl` is computed on read from `net_qty * (ltp - avg_price)`; not stored.

### `portfolio.holdings` (post T+1 CNC positions)

```sql
create table portfolio.holdings (
  holding_id    text primary key,
  user_id       text not null,
  instrument_id text not null,
  quantity      integer not null,
  avg_price     numeric(18,4) not null,
  acquired_on   date not null,
  updated_at    timestamptz not null default now(),
  unique (user_id, instrument_id, acquired_on)   -- lot-wise
);
```

Lot-wise rows support FIFO cost basis & LTCG/STCG split.

### `portfolio.ledger` (double-entry)

```sql
create table portfolio.ledger_entries (
  entry_id     text primary key,
  user_id      text not null,
  account      text not null,               -- CASH | MARGIN_BLOCKED | MTM_REALISED | MTM_UNREALISED | CHARGES
  debit        numeric(18,4) not null default 0,
  credit       numeric(18,4) not null default 0,
  ref_type     text not null,               -- TRADE | MTM | FUND_IN | FUND_OUT | CHARGE | CORP_ACTION
  ref_id       text,
  narration    text,
  occurred_at  timestamptz not null default now(),
  check (debit = 0 or credit = 0)
);
create index on portfolio.ledger_entries (user_id, account, occurred_at desc);
```

Every business event produces a pair (e.g., buy trade: debit `MARGIN_BLOCKED`, credit `CASH` → on fill: reverse + adjust).

## `md` — market data

### `md.ticks` (hypertable)

```sql
create table md.ticks (
  instrument_id  text not null,
  ts             timestamptz not null,
  ltp            numeric(18,4) not null,
  bid_px         numeric(18,4),
  bid_qty        integer,
  ask_px         numeric(18,4),
  ask_qty        integer,
  volume         bigint,
  oi             bigint,
  source         text not null,              -- LIVE | REPLAY
  primary key (instrument_id, ts)
);
select create_hypertable('md.ticks', 'ts', chunk_time_interval => interval '1 day');
alter table md.ticks set (timescaledb.compress, timescaledb.compress_segmentby='instrument_id');
select add_compression_policy('md.ticks', interval '7 days');
```

### `md.candles_1m` (continuous aggregate)

```sql
create materialized view md.candles_1m
with (timescaledb.continuous) as
select
  instrument_id,
  time_bucket('1 minute', ts) as bucket,
  first(ltp, ts) as open,
  max(ltp) as high,
  min(ltp) as low,
  last(ltp, ts) as close,
  sum(coalesce(volume, 0)) as volume
from md.ticks
group by instrument_id, bucket;
select add_continuous_aggregate_policy('md.candles_1m',
  start_offset => interval '1 day',
  end_offset   => interval '1 minute',
  schedule_interval => interval '1 minute');
```

Repeat for 5m, 15m, 60m, 1d.

## `reports`

### `reports.contract_notes`

```sql
create table reports.contract_notes (
  note_id      text primary key,
  user_id      text not null,
  trade_date   date not null,
  pdf_path     text not null,
  totals       jsonb not null,              -- turnover, stt, sebi, gst, stamp, total
  created_at   timestamptz default now()
);
```

## Event catalog (bus)

All events are Protobuf-encoded on `packages/protos/events.proto`. Topics are Redis Streams.


| Event                              | Producer     | Consumers                | Purpose                          |
| ---------------------------------- | ------------ | ------------------------ | -------------------------------- |
| `Tick`                             | MD           | MM, FE via GW, Strategy  | Live prices                      |
| `OrderPlaced`                      | OMS          | Risk, audit              | Lifecycle                        |
| `OrderAccepted`                    | OMS          | —                        | Audit                            |
| `OrderRejected`                    | OMS / Risk   | FE (via GW), audit       | Feedback                         |
| `OrderOpen`                        | ME           | OMS                      | Book entry confirmed             |
| `OrderUpdate`                      | ME           | OMS                      | PARTIAL/FILLED/CANCELLED/EXPIRED |
| `Trade`                            | ME           | OMS, Portfolio, Strategy | Fills                            |
| `PositionUpdated`                  | Portfolio    | FE (via GW), Strategy    | Position changes                 |
| `Candle1m`                         | MD           | FE (via GW), Strategy    | Chart updates                    |
| `DayClosed`                        | Scheduler    | All                      | EOD trigger                      |
| `CorpActionApplied`                | Scheduler    | Portfolio                | Splits/bonus/dividend            |
| `MarginBlocked` / `MarginReleased` | Risk         | Portfolio (ledger)       | Double-entry bookkeeping         |
| `KillSwitchArmed`                  | Surveillance | OMS                      | Halt trading                     |


## Idempotency

- Every client-facing write accepts `Idempotency-Key` header (OMS stores `client_order_id` in `oms.orders`; Redis SETNX TTL 60s caches the response).
- Every bus consumer is idempotent: deduplication via `(event_id, consumer_group)` Redis set, TTL 24h.
- Projections never `INSERT` without `ON CONFLICT DO UPDATE` keyed on deterministic ULIDs derived from the source event.

## Migrations

- Tool: **sqitch** or **dbmate** (language-agnostic). Avoid ORM-coupled migrations.
- `infra/migrations/<schema>/<seq>_<name>.sql` files.
- CI runs migrations against a throwaway Postgres before tests.

## Seed data

- `ref.users`: single user.
- `ref.instruments`: NIFTY, BANKNIFTY, INFY, RELIANCE, TCS, HDFCBANK for v1 → expand later.
- F&O contracts: current-month + next-month + current-quarter for NIFTY and BANKNIFTY; ATM ± 10 strikes.

## Data retention


| Table                    | Retention                                 | Justification                  |
| ------------------------ | ----------------------------------------- | ------------------------------ |
| `md.ticks`               | 90 days raw, then compressed indefinitely | Storage cost manageable        |
| `oms.order_events`       | Forever                                   | Audit / replay source of truth |
| `oms.trades`             | Forever                                   | Same                           |
| `portfolio.`*            | Forever (projections are cheap)           | Replayable                     |
| `reports.contract_notes` | 7 years                                   | SEBI record-keeping norm       |


## Projection rebuild

`pt admin rebuild positions --user=<id>` walks `oms.trades` chronologically and recomputes `portfolio.positions`. Sanity-checks against the live projection; prints diffs. Run in CI against a fixture dataset.