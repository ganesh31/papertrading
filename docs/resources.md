# Resources

One consolidated list. Starred items are high-leverage for your goal.

## Books

- ⭐ **Larry Harris — *Trading and Exchanges: Market Microstructure for Practitioners*.** The one book. Chapters 4, 5, 11 in particular.
- ⭐ **John Hull — *Options, Futures, and Other Derivatives*.** Chapters 10, 13, 15, 19 cover pricing, greeks, SPAN-style margining intuition.
- **Euan Sinclair — *Volatility Trading*.** Greeks and vol surface intuition.
- **Martin Kleppmann — *Designing Data-Intensive Applications*.** Your event-sourcing, replication, consistency vocabulary.
- **Vaughn Vernon — *Implementing Domain-Driven Design*.** Bounded contexts, aggregates — maps cleanly to services here.
- Ernie Chan — *Algorithmic Trading*. Light read; helps you think about what a strategy SDK must expose.

## Papers / white papers

- ⭐ **CME Group — *SPAN Methodology* PDF**. The canonical spec. The NSE version follows CME closely. [https://www.cmegroup.com/clearing/risk-management/span-methodology.html](https://www.cmegroup.com/clearing/risk-management/span-methodology.html)
- **LMAX Disruptor** — Martin Thompson et al. Single-writer principle.
- **Chronicle Queue** architecture notes.
- **NSE Clearing — SPAN Risk Parameter File specification** (PDF on NSE site).

## Talks (YouTube)

- ⭐ Martin Thompson — *LMAX: How to do 100k TPS at less than 1ms latency*.
- Jane Street tech talks on trading systems (OCaml specifics transferable).
- QCon talks tagged "exchange" / "matching engine".
- Kailash Nadh (Zerodha CTO) interviews on The Ken / YourStory podcasts — design-philosophy gold.

## Indian exchange references (free & official)

- NSE Circulars: [https://www.nseindia.com/resources/exchange-communication-circulars](https://www.nseindia.com/resources/exchange-communication-circulars)
- NSE Archives (bhavcopy, SPAN file, F&O files): [https://archives.nseindia.com](https://archives.nseindia.com)
- NSE Option Chain: [https://www.nseindia.com/option-chain](https://www.nseindia.com/option-chain)
- NSE Trading Holidays & Timings: [https://www.nseindia.com/resources/exchange-communication-holidays](https://www.nseindia.com/resources/exchange-communication-holidays)
- NSE Clearing (NCL): [https://www.nseclearing.com](https://www.nseclearing.com)
- SEBI Circulars: [https://www.sebi.gov.in/sebiweb/home/HomeAction.do?doListing=yes&sid=1&ssid=6](https://www.sebi.gov.in/sebiweb/home/HomeAction.do?doListing=yes&sid=1&ssid=6)
- SEBI Master Circular for Stock Brokers (latest year).
- ⭐ **Zerodha Varsity** (free, excellent): [https://zerodha.com/varsity/](https://zerodha.com/varsity/) — Modules 1, 4 (Futures), 5 (Options), 6 (Option Strategies), 7 (Markets & Taxation).

## Broker APIs (read as documentation, not as clients)

- ⭐ **Kite Connect v3 docs** — model your APIs after these: [https://kite.trade/docs/connect/v3/](https://kite.trade/docs/connect/v3/)
- Angel One SmartAPI docs + WebSocket 2.0: [https://smartapi.angelbroking.com/docs](https://smartapi.angelbroking.com/docs)
- Upstox API v2: [https://upstox.com/developer/api-documentation/](https://upstox.com/developer/api-documentation/)
- Interactive Brokers TWS API — for a global comparison.

## Libraries / OSS to read (not fork)

- `https://github.com/i25959341/orderbook` — Go limit order book, small + readable.
- `https://github.com/enewhuis/liquibook` — C++ order book, clean header file worth reading.
- `https://github.com/zerodha/kiteconnect-python` — API client ergonomics.
- `https://github.com/opensourceframeworks/openalgo` — OSS algo gateway for Indian brokers.
- `https://github.com/QuantConnect/Lean` — concepts for event-driven backtesting; don't try to clone.
- `https://github.com/timescale/timescaledb` — read the continuous aggregates internals.

## Indian fintech engineering blogs

- Zerodha Tech blog: [https://zerodha.tech](https://zerodha.tech)
- Razorpay engineering blog (payments, but distributed-systems gold): [https://engineering.razorpay.com](https://engineering.razorpay.com)
- Upstox engineering blog.

## Podcasts

- *The Seen and the Unseen* — several episodes on Indian markets and regulation.
- *Paisa Vaisa*.
- *20 Minute VC* (rare episodes on fintech infra).

## Data sources for replay (free)

- **yfinance** (primary for 1-min intraday bars, last 7 days rolling): <https://github.com/ranaroussi/yfinance>
- **NSE Equity Bhavcopy** (EOD, free, historical): <https://archives.nseindia.com/content/historical/EQUITIES/>
- **NSE F&O Bhavcopy** (EOD settlement prices + OI): <https://archives.nseindia.com/content/historical/DERIVATIVES/>
- **Angel scrip master** (static JSON, no auth): <https://margincalculator.angelbroking.com/OpenAPI_File/files/OpenAPIScripMaster.json>
- **NSE chart-databyindex API** (unofficial, cookies needed): `https://www.nseindia.com/api/chart-databyindex`

## Tooling docs you'll hit often

- Fastify: [https://fastify.dev](https://fastify.dev)
- Zod: [https://zod.dev](https://zod.dev)
- Redis Streams: [https://redis.io/docs/data-types/streams/](https://redis.io/docs/data-types/streams/)
- TimescaleDB continuous aggregates: [https://docs.timescale.com/use-timescale/latest/continuous-aggregates/](https://docs.timescale.com/use-timescale/latest/continuous-aggregates/)
- OpenTelemetry JS & Go quickstarts.
- k6: [https://k6.io/docs/](https://k6.io/docs/)
- Testcontainers: [https://testcontainers.com](https://testcontainers.com)
- TradingView Charting Library: [https://www.tradingview.com/charting-library-docs/](https://www.tradingview.com/charting-library-docs/)

## Interview prep supplements

- Alex Xu — *System Design Interview* vol I & II (superficial for senior levels, but handy warm-up).
- **Grokking the System Design Interview** (Educative) — *trading system* chapter.
- **Pramp / Interviewing.io** — mock senior-level system design.

## Paid (optional; only if free tier blocks you)

- TrueData / GFDL historical tick data — ~₹500–2000/month. Phase 2 or 7 if Angel WS rate limits bite.
- TradingView advanced charting library (free for individual use; commercial license needed if you host it publicly for others).

## What to skip

- HFT ultra-low-latency papers (kernel bypass, FPGA matching) — interesting, not your goal.
- MQL / MetaTrader tutorials — wrong market / wrong paradigm.
- Most "build a stock trading app" YouTube tutorials — they skip every hard part.

