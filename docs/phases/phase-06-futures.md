# Phase 6 — Equity Futures

**Week 9 · ~20 hrs**

Goal: trade index and stock futures end-to-end. Daily MTM, expiry settlement, rollover helper. Interim margin stays on the placeholder VAR model until Phase 8 replaces it with SPAN.

## Prerequisites

- Phases 0–5 complete.
- Equity cash is boring and working.

## Deliverables

- [ ] Contract master loads NIFTY, BANKNIFTY, FINNIFTY + stock futures for current + next month + current quarter.
- [ ] Matching, OMS, positions all work for `segment='FUT'`.
- [ ] Daily MTM job: settles P&L at the daily settlement price.
- [ ] Expiry settlement:
  - Index futures: cash-settled at final settlement price (index close on expiry day).
  - Stock futures: cash-settled at final settlement price.
- [ ] Rollover helper: "Rollover to next expiry" action closes current + opens next at ask.
- [ ] Product `NRML` (carry-forward) + `MIS` (intraday) both supported.
- [ ] Margin: placeholder `initial_margin% × contract_value` (configurable per underlying; e.g., 15% NIFTY, 20% stock); Phase 8 replaces with SPAN.
- [ ] Option chain widget still works; futures chain shown as a simple row per expiry.
- [ ] ADR-0011 (MTM-as-ledger-entry vs. settlement-as-trade).
- [ ] Talking-points doc.

## Why futures are (slightly) tricky

Futures don't have cash exchanging hands at trade time — you post margin, and you mark-to-market daily. Your ledger pattern needs to reflect that:

- **At open**: debit `MARGIN_BLOCKED`, credit `CASH` (initial margin).
- **Each EOD**: debit/credit `CASH` by MTM (settlement price − prev settlement price) × qty × lot_size × side sign.
- **On close**: release initial margin back; realise remaining MTM.

This is categorically different from equity cash where cash moved at trade.

## Data model additions

No schema changes — the existing structures handle it:

- `ref.instruments` already has `expiry`, `lot_size`, `underlying`, `instrument_type='FUT'`.
- `portfolio.positions` keyed `(user, instrument, product)` naturally separates MIS from NRML.
- Ledger account `MTM_REALISED` grows per-day.

New reference table:

```sql
create table ref.daily_settlement (
  instrument_id     text not null references ref.instruments,
  settlement_date   date not null,
  settlement_price  numeric(18,4) not null,
  final_settlement  boolean not null default false,
  source            text not null,           -- NSE | COMPUTED
  primary key (instrument_id, settlement_date)
);
```

Populated daily from NSE's bhavcopy.

## Tasks

### 6.1 Contract master expansion

- Extend `pt instruments sync` to pull NFO instruments.
- Freeze qty per contract (from NSE's daily freeze file).
- Expiry dates for index: weekly (Thursday/Tuesday depending on index) + monthly; for stock: monthly only.

### 6.2 Validation additions

- Reject orders on expired contracts.
- Reject modify on contract whose last trading day is past.

### 6.3 Interim margin logic (replaces by Phase 8)

- For NRML FUT: `initial_margin = contract_value × im_pct` where `im_pct` comes from a config table `ref.margin_pct` seeded with rough values.
- For MIS: `im_pct × intraday_multiplier` (e.g., 0.5) — Zerodha gives ~2× leverage on intraday futures.
- ADR notes the simplification and the Phase 8 plan.

### 6.4 Daily MTM job

- Scheduled at EOD (23:00 IST; scheduler wired in Phase 10 but can be a CLI for now: `pt admin mtm --date=YYYY-MM-DD`).
- Steps:
  - Fetch NSE bhavcopy for the day → populate `daily_settlement`.
  - For each open FUT position across all users:
    - `pnl = (today_settle - prior_settle) × net_qty × lot_size × side_sign`
    - Write ledger: debit/credit `CASH` vs. `MTM_REALISED`.
    - Emit `MTMApplied` event.
    - Update `avg_price = today_settle` for next-day delta (this is the "carry forward" trick that makes daily MTM work).

### 6.5 Expiry settlement

- On expiry day EOD (after `DayClosed`):
  - Mark contract as `EXPIRED` in `ref.instruments`.
  - Final MTM at final settlement price (same math as 6.4 but last time).
  - Close position: release initial margin.
  - If stock future: final settlement price = NSE's weighted average of last 30 min in cash segment.

### 6.6 Rollover helper

- UI button on a NIFTY-current-month position: "Roll to next month".
- Server: in a single *logical* transaction, place MARKET SELL on current + MARKET BUY same qty on next. Not atomic at the exchange — real brokers do it leg-by-leg; mimic that, including the calendar spread margin benefit (Phase 8).

### 6.7 FE additions

- Watchlist row for futures shows `LTP (spot) / basis`.
- Positions show `MTM_today` separately from `realised` and `unrealised` since last MTM.
- Expiry countdown on future position rows.

### 6.8 Intraday square-off extension

- 15:25 IST for F&O (different from equity 15:15). Parameterize by segment.

## Metrics

- `mtm_applied_total{underlying}`
- `expiry_settlements_total`
- `rollover_events_total`
- Gauge: `futures_open_positions{underlying,expiry}`

## Performance targets

- Daily MTM job for 10k positions < 30 s.
- Expiry settlement for a week of contracts < 60 s.

## Testing

- Unit: MTM math (long, short, multi-day carry).
- Integration: seed fixture with 3-day trajectory of NIFTY future → compare ledger & position state to expected.
- Property: ∑ MTM over life of contract == final P&L (modulo rounding).
- Edge: position opened Friday, MTM Monday (holiday in between) — correct prev_settle used.

## Common pitfalls

- Treating MTM as just another trade — it's a *settlement* event; don't add trades to `oms.trades` for it.
- Not advancing `avg_price` post-MTM → P&L double-counts.
- Forgetting weekly vs. monthly expiry for indices.
- Futures position on expiry day not getting auto-closed → position lingers forever.
- Calculating MTM at LTP instead of settlement price → off by close/settle gap.
- Not propagating corporate actions (stock split) to F&O contract adjustment.

## Interview talking points

- MTM as a daily settlement flow; why it exists (credit risk reduction by clearing corp).
- Cash-settled vs. physically-settled futures (stock F&O in India moved to physical delivery in 2019 for certain stocks — mention as a v2).
- Calendar spread: two legs that net into a lower margin.
- Basis (fut − spot) and roll yield.
- Why avg_price resets to settlement price each day.

## Resources

- NSE F&O FAQ: <https://www.nseindia.com/products-services/equity-derivatives-faqs>
- NSE Bhavcopy for F&O: under `archives.nseindia.com/content/fo/fo_mktlots.csv` and `fo_bhav.csv`.
- Zerodha Varsity Module 4 — Futures Trading (free).
- Hull, ch. 2–5 (mechanics, hedging, arbitrage).
- NSE's physical settlement circular (for interview context; not implemented).

## Exit checklist

- [ ] You can buy NIFTY FUT via UI, see MTM apply next day, see ledger balance.
- [ ] On expiry day, position auto-closes at final settlement price.
- [ ] Rollover button works; two trades appear; positions update correctly.
- [ ] ADR-0011 merged.
