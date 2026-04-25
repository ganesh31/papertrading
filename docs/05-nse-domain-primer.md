# 05 — NSE / BSE Domain Primer

The exchange concepts the rest of the docs assume. Read this before Phase 2.

## Trading sessions (NSE equity)


| Session                    | Timing (IST)                | Purpose                                                                |
| -------------------------- | --------------------------- | ---------------------------------------------------------------------- |
| Pre-open                   | 09:00 – 09:08 (order entry) | Call auction: orders collected, equilibrium price discovered at 09:08. |
| Pre-open matching + buffer | 09:08 – 09:15               | Matching then transition.                                              |
| Continuous / Normal        | 09:15 – 15:30               | Continuous matching, price-time priority.                              |
| Closing session            | 15:40 – 16:00               | Post-close orders at close price only (limited use).                   |
| AMO entry                  | 16:00 – 08:59 next day      | After-market orders queued for pre-open.                               |


**F&O (NFO)**: Continuous 09:15 – 15:30 (no pre-open in F&O). Currency: 09:00 – 17:00.

## Order types (must support in v1)


| Type                          | Semantics                                                                                                          |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| **LIMIT**                     | Execute at specified price or better; rest of qty on book.                                                         |
| **MARKET**                    | Execute immediately at best available; walks the book. NSE allows MARKET only with price protection bands.         |
| **SL (Stop-Loss Limit)**      | Trigger when LTP crosses `trigger_price`; enters book as LIMIT at `price`.                                         |
| **SL-M (Stop-Loss Market)**   | Trigger when LTP crosses trigger; enters as MARKET. Disabled by NSE on some segments (2020+) — simulate that rule. |
| **IOC (Immediate-or-Cancel)** | Fill whatever you can; cancel the rest instantly.                                                                  |
| **FOK (Fill-or-Kill)**        | Fill fully immediately, or cancel. Rarely used; include for completeness.                                          |


### Could-have (Phase 11 stretch)

- **Iceberg** — total qty split into N visible legs.
- **GTT (Good-Till-Triggered)** — client-side trigger order, not a real exchange order type; Zerodha's invention.
- **Cover Order (CO)** — entry + compulsory stop-loss, reduced margin.
- **Bracket Order (BO)** — entry + SL + target (discontinued by Zerodha post-peak-margin; worth referencing in ADR).
- **AMO** — After Market Order; queued in OMS until pre-open.
- **Basket** — submit N orders atomically; half-success is OK.

## Validity

- `DAY` — cancel at EOD if unfilled.
- `IOC` — immediate.
- `GTC` / `GTD` — not available on Indian exchanges directly; Zerodha's GTT simulates it.

## Product types (the defining Indian broker concept)


| Product                              | Segment       | Meaning                             | Margin                     |
| ------------------------------------ | ------------- | ----------------------------------- | -------------------------- |
| **CNC** (Cash & Carry)               | Equity        | Delivery; holdings after T+1.       | Full (100% cash up front). |
| **MIS** (Margin Intraday Square-off) | Equity, F&O   | Intraday; auto-square-off at 15:15. | Reduced intraday margin.   |
| **NRML** (Normal)                    | F&O, currency | Carry-forward position.             | Full SPAN + Exposure.      |


**Position conversion**: a user may convert MIS → CNC (needs full margin) or MIS → NRML before square-off. Real brokers support this; v1 must too.

## Tick size & lot size

- Equity tick size: usually ₹0.05; some scrips ₹0.01 or ₹0.10. Pulled from contract master.
- F&O lot size: per-symbol (NIFTY 75, BANKNIFTY 30 as of 2024 revisions — refresh from NSE daily). Lot sizes change periodically; never hard-code.

## Freeze quantity

- Max order qty per single order on F&O (e.g., NIFTY options: 1800 qty = 24 lots at lot 75). Orders above freeze qty are rejected by the exchange.
- Enforce in OMS validation. NSE publishes a daily **freeze quantity file**.

## Circuit filters & price bands

- **Daily price band** per scrip: 2%, 5%, 10%, 20% or "no band" (for F&O underlyings). Prevents orders outside the band.
- **Market-wide circuit breaker (MWCB)**: 10%, 15%, 20% moves in NIFTY 50 → market halt for 45 min / 1h 15 / rest of day depending on trigger time. (Rare; simulate as a feature flag.)
- **Upper / Lower circuit** for individual scrips; when hit, matching continues only in one direction.

## Peak margin & upfront margin (SEBI, 2020–2021)

- Brokers must collect upfront margin from clients **before** order placement (pre-trade).
- Clearing Corp snaps each broker's margin utilization **4 times a day at random**; any shortfall → penalty.
- Consequence for your design: margin check is **pre-trade, synchronous, non-negotiable**. This is why Risk is on the order path.

## Settlement

- **Equity**: T+1 rolling settlement (live since Jan 2023). Some scrips moved to **T+0** in 2024 (optional). v1: T+1.
- **F&O**: daily MTM for futures; option premiums settled intraday (cash). Final settlement on expiry.
- **Currency**: T+2.

## Corporate actions

- **Bonus** (e.g., 1:1): `holdings.qty *= 2`, `avg_price /= 2`.
- **Split** (e.g., face value 10 → 1): same math as bonus for ratio 1:10.
- **Dividend**: cash credit to ledger; adjust F&O contracts by dividend amount if ex-date between now and expiry.
- **Rights issue**, **buyback**: out of v1 scope.

## Surveillance concepts (mention, lightly implement)

- **Order-to-Trade Ratio (OTR)**: orders placed / trades executed. Exchanges penalize excessive OTR in F&O. Track per user per symbol per day.
- **Market-wide Position Limit (MWPL)**: per-underlying cap on open interest. NSE publishes daily. Reject new positions when breached.
- **Client-level position limit**: per client per underlying.
- **Kill switch**: broker can halt all client orders. Manual trigger in v1.

## STT / charges (for contract notes)

Approximate v1 values (update from latest gov notifications):


| Charge       | Equity Delivery          | Equity Intraday | Equity Futures | Equity Options                                         |
| ------------ | ------------------------ | --------------- | -------------- | ------------------------------------------------------ |
| STT (buy)    | 0.1%                     | —               | —              | —                                                      |
| STT (sell)   | 0.1%                     | 0.025%          | 0.02%          | 0.1% on premium (sell), 0.125% on intrinsic (exercise) |
| Exchange txn | ~0.00345%                | ~0.00345%       | ~0.002%        | ~0.053% on premium                                     |
| SEBI         | ₹10 / crore              | same            | same           | same                                                   |
| Stamp        | 0.015% (buy)             | 0.003% (buy)    | 0.002% (buy)   | 0.003% (buy)                                           |
| GST          | 18% on (brokerage + txn) | same            | same           | same                                                   |


Brokerage (paper trader): 0 for delivery, ₹20/order for intraday & F&O (Zerodha-style).

## Key identifiers

- **ISIN**: 12-char instrument identifier (`INE009A01021` = Infosys).
- **Symbol**: exchange-local code (`INFY`, `NIFTY24DEC25000CE`).
- **Token**: numeric ID used by Angel/Kite in WebSocket frames.
- **Segment codes** (Angel): `1` NSE, `2` NFO, `3` BSE, `5` MCX, `7` NCDEX, `13` CDS.

## Useful references to bookmark

- [NSE — Trading FAQ](https://www.nseindia.com/products-services/equity-market-trading-faqs)
- [NSE — Circulars](https://www.nseindia.com/resources/exchange-communication-circulars)
- [NSE — Archives (bhavcopy, SPAN files)](https://archives.nseindia.com)
- [SEBI — Master Circular for Stock Brokers](https://www.sebi.gov.in)
- [Angel One SmartAPI Docs](https://smartapi.angelbroking.com/docs)
- [Kite Connect API Docs](https://kite.trade/docs/connect/v3/)
- [Zerodha Varsity — Modules 1, 4, 5, 6, 7](https://zerodha.com/varsity/)
- [NSE Option Chain](https://www.nseindia.com/option-chain)

## Cheat-sheet: which concepts live where in this system


| Concept               | Primary module          | Docs section |
| --------------------- | ----------------------- | ------------ |
| Session timings       | MD gateway + Scheduler  | Phase 1, 10  |
| Order types           | OMS + Matching engine   | Phase 2, 3   |
| Tick / lot / freeze   | OMS validation          | Phase 3      |
| Circuit filter        | OMS validation + ME     | Phase 2, 3   |
| MWCB                  | Scheduler → kill switch | Phase 11     |
| Peak / upfront margin | Risk + SPAN             | Phase 3, 8   |
| T+1 settlement        | Scheduler (Portfolio)   | Phase 10     |
| OTR                   | Surveillance consumer   | Phase 11     |
| Corp actions          | Scheduler (Portfolio)   | Phase 10     |
| Contract notes        | Reports                 | Phase 10     |


