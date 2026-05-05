-- migrate:up
-- Continuous aggregates from md.ticks (hypertable). Policies: fine buckets refresh often; 1d refreshes hourly.

create materialized view if not exists md.cagg_ticks_1m
with (timescaledb.continuous) as
select
  instrument_id,
  time_bucket(interval '1 minute', ts) as bucket,
  first(ltp, ts) as open,
  max(ltp) as high,
  min(ltp) as low,
  last(ltp, ts) as close,
  sum(coalesce(volume, 0::bigint)) as volume
from md.ticks
group by instrument_id, time_bucket(interval '1 minute', ts)
with no data;

create materialized view if not exists md.cagg_ticks_5m
with (timescaledb.continuous) as
select
  instrument_id,
  time_bucket(interval '5 minutes', ts) as bucket,
  first(ltp, ts) as open,
  max(ltp) as high,
  min(ltp) as low,
  last(ltp, ts) as close,
  sum(coalesce(volume, 0::bigint)) as volume
from md.ticks
group by instrument_id, time_bucket(interval '5 minutes', ts)
with no data;

create materialized view if not exists md.cagg_ticks_15m
with (timescaledb.continuous) as
select
  instrument_id,
  time_bucket(interval '15 minutes', ts) as bucket,
  first(ltp, ts) as open,
  max(ltp) as high,
  min(ltp) as low,
  last(ltp, ts) as close,
  sum(coalesce(volume, 0::bigint)) as volume
from md.ticks
group by instrument_id, time_bucket(interval '15 minutes', ts)
with no data;

create materialized view if not exists md.cagg_ticks_1h
with (timescaledb.continuous) as
select
  instrument_id,
  time_bucket(interval '1 hour', ts) as bucket,
  first(ltp, ts) as open,
  max(ltp) as high,
  min(ltp) as low,
  last(ltp, ts) as close,
  sum(coalesce(volume, 0::bigint)) as volume
from md.ticks
group by instrument_id, time_bucket(interval '1 hour', ts)
with no data;

create materialized view if not exists md.cagg_ticks_1d
with (timescaledb.continuous) as
select
  instrument_id,
  time_bucket(interval '1 day', ts) as bucket,
  first(ltp, ts) as open,
  max(ltp) as high,
  min(ltp) as low,
  last(ltp, ts) as close,
  sum(coalesce(volume, 0::bigint)) as volume
from md.ticks
group by instrument_id, time_bucket(interval '1 day', ts)
with no data;

-- 1m / 5m / 15m: refresh every minute (Phase 1.7 “minute” cadence for intraday).
select add_continuous_aggregate_policy(
  continuous_aggregate => 'md.cagg_ticks_1m',
  start_offset => interval '3 days',
  end_offset => interval '1 minute',
  schedule_interval => interval '1 minute',
  if_not_exists => true
);

select add_continuous_aggregate_policy(
  continuous_aggregate => 'md.cagg_ticks_5m',
  start_offset => interval '7 days',
  end_offset => interval '5 minutes',
  schedule_interval => interval '1 minute',
  if_not_exists => true
);

select add_continuous_aggregate_policy(
  continuous_aggregate => 'md.cagg_ticks_15m',
  start_offset => interval '14 days',
  end_offset => interval '15 minutes',
  schedule_interval => interval '1 minute',
  if_not_exists => true
);

-- 1h: hourly refresh window.
select add_continuous_aggregate_policy(
  continuous_aggregate => 'md.cagg_ticks_1h',
  start_offset => interval '30 days',
  end_offset => interval '1 hour',
  schedule_interval => interval '1 hour',
  if_not_exists => true
);

-- 1d: daily bucket, hourly policy (Phase 1.7 “hour” cadence for coarser views).
select add_continuous_aggregate_policy(
  continuous_aggregate => 'md.cagg_ticks_1d',
  start_offset => interval '400 days',
  end_offset => interval '1 day',
  schedule_interval => interval '1 hour',
  if_not_exists => true
);

-- migrate:down
select remove_continuous_aggregate_policy('md.cagg_ticks_1m', if_exists => true);
select remove_continuous_aggregate_policy('md.cagg_ticks_5m', if_exists => true);
select remove_continuous_aggregate_policy('md.cagg_ticks_15m', if_exists => true);
select remove_continuous_aggregate_policy('md.cagg_ticks_1h', if_exists => true);
select remove_continuous_aggregate_policy('md.cagg_ticks_1d', if_exists => true);

drop materialized view if exists md.cagg_ticks_1d;
drop materialized view if exists md.cagg_ticks_1h;
drop materialized view if exists md.cagg_ticks_15m;
drop materialized view if exists md.cagg_ticks_5m;
drop materialized view if exists md.cagg_ticks_1m;
