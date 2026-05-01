-- migrate:up
create extension if not exists timescaledb;

create table if not exists ref.instruments (
  instrument_id             text primary key,
  tradingsymbol             text not null,
  exchange                  text not null,
  segment                   text not null,
  asset_class               text not null,
  instrument_type           text not null,
  underlying_instrument_id  text references ref.instruments (instrument_id),
  expiry                    date,
  strike                    numeric(18,4),
  option_type               text,
  lot_size                  integer not null default 1,
  tick_size                 numeric(18,4) not null default 0.05,
  freeze_qty                integer,
  isin                      text,
  listing_date              date,
  status                    text not null default 'ACTIVE',
  metadata                  jsonb default '{}'::jsonb,
  created_at                timestamptz default now()
);

create unique index if not exists ref_instruments_exchange_segment_tradingsymbol_idx
  on ref.instruments (exchange, segment, tradingsymbol);

create index if not exists ref_instruments_underlying_expiry_idx
  on ref.instruments (underlying_instrument_id, expiry);

create table if not exists md.ticks (
  instrument_id  text not null,
  ts             timestamptz not null,
  ltp            numeric(18,4) not null,
  bid_px         numeric(18,4),
  bid_qty        integer,
  ask_px         numeric(18,4),
  ask_qty        integer,
  volume         bigint,
  oi             bigint,
  source         text not null,
  primary key (instrument_id, ts)
);

select create_hypertable('md.ticks', 'ts', chunk_time_interval => interval '1 day', if_not_exists => true);

alter table md.ticks set (
  timescaledb.compress,
  timescaledb.compress_segmentby = 'instrument_id',
  timescaledb.compress_orderby = 'ts desc'
);

select add_compression_policy('md.ticks', interval '7 days', if_not_exists => true);

create table if not exists md.bars_1m (
  instrument_id text not null references ref.instruments (instrument_id),
  ts            timestamptz not null,
  open          numeric(18,4) not null,
  high          numeric(18,4) not null,
  low           numeric(18,4) not null,
  close         numeric(18,4) not null,
  volume        bigint not null,
  source        text not null default 'yfinance',
  primary key (instrument_id, ts)
);

create index if not exists md_bars_1m_ts_idx on md.bars_1m (ts);

create table if not exists md.bhav_eq (
  instrument_id text not null references ref.instruments (instrument_id),
  trade_date    date not null,
  open          numeric(18,4),
  high          numeric(18,4),
  low           numeric(18,4),
  close         numeric(18,4),
  last          numeric(18,4),
  prev_close    numeric(18,4),
  volume        bigint,
  turnover      numeric(18,4),
  primary key (instrument_id, trade_date)
);

create index if not exists md_bhav_eq_trade_date_idx on md.bhav_eq (trade_date);

-- migrate:down
drop table if exists md.bhav_eq;
drop table if exists md.bars_1m;

select remove_compression_policy('md.ticks', if_exists => true);
drop table if exists md.ticks;

drop table if exists ref.instruments;
