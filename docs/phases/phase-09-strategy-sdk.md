# Phase 9 — Strategy Runtime & SDK

**Week 14 · ~20 hrs**

Goal: run code against your system the way you'd run code against Kite Connect. Same API in **live** and **replay** modes, deterministic in replay, supervisor-managed.

## Prerequisites

- Phases 0–8 complete (SPAN optional but ideal).
- Replay adapter (Phase 1) works end-to-end.

## Deliverables

- [ ] `@papertrading/sdk` TS package published to workspace (and privately to npm if you want).
- [ ] Strategy lifecycle: `init`, `onTick`, `onOrderUpdate`, `onTrade`, `onBar`, `shutdown`.
- [ ] `services/strategy-runner`: spawns strategy processes; supervisor restart on crash.
- [ ] Live mode: wall clock + Angel adapter (or replay adapter at 1× speed).
- [ ] Backtest mode: virtual clock + replay adapter at N× speed; deterministic outcomes.
- [ ] Strategy manifest: config file defining symbols, parameters, universe, risk caps.
- [ ] Sample strategy: SMA crossover on NIFTY future.
- [ ] CLI: `pt strategy run --name=sma --mode=live|backtest --date=...`.
- [ ] Backtest report: trades, P&L curve, drawdown, Sharpe, turnover.
- [ ] ADR-0015 (strategy runtime isolation model).
- [ ] Talking-points doc.

## Design principles

1. **Same code, two clocks.** The strategy doesn't know if it's live or replay. Only the SDK knows.
2. **Deterministic replay.** Given the same inputs + seed, a strategy produces identical trades. This is the killer property for bug hunting.
3. **One strategy, one process.** Crash isolation. CPU isolation. Easy to reason about state.
4. **Risk caps at SDK boundary.** Strategy cannot place orders above configured lot/notional — enforced in the SDK before hitting the wire.

## Architecture

```mermaid
flowchart LR
  subgraph Supervisor[services/strategy-runner]
    S[Supervisor]
  end
  subgraph Strat[Strategy Process]
    SDK[@papertrading/sdk]
    USERCODE[User Strategy]
  end
  GW[Gateway REST + WS]
  S -->|fork| Strat
  SDK --> GW
  USERCODE --> SDK
  SDK -->|onTick/onBar/onOrderUpdate| USERCODE
```

## SDK API (TS)

```ts
export interface StrategyContext {
  // subscriptions
  subscribe(symbols: string[]): Promise<void>;
  subscribeBars(symbol: string, interval: Interval): Promise<void>;

  // orders
  placeOrder(req: PlaceOrderReq): Promise<{ orderId: string }>;
  cancelOrder(orderId: string): Promise<void>;
  modifyOrder(orderId: string, patch: ModifyReq): Promise<void>;

  // queries
  getPositions(): Promise<Position[]>;
  getHoldings(): Promise<Holding[]>;
  getFunds(): Promise<Funds>;
  getMarginPreview(req: PlaceOrderReq): Promise<MarginBreakdown>;

  // clock
  now(): Date;                 // virtual in backtest
  setTimeout(ms: number, cb: () => void): number;  // virtual-aware

  // logging
  log: Logger;
  metrics: MetricsEmitter;
}

export interface Strategy {
  name: string;
  init(ctx: StrategyContext, params: Record<string, unknown>): Promise<void>;
  onTick?(ctx: StrategyContext, tick: Tick): Promise<void>;
  onBar?(ctx: StrategyContext, bar: Bar): Promise<void>;
  onOrderUpdate?(ctx: StrategyContext, update: OrderUpdate): Promise<void>;
  onTrade?(ctx: StrategyContext, trade: Trade): Promise<void>;
  shutdown?(ctx: StrategyContext): Promise<void>;
}
```

## Strategy manifest

`strategies/sma/manifest.yaml`:

```yaml
name: sma
entry: ./index.ts
mode: live
symbols: [NIFTY24DECFUT]
params:
  fast: 10
  slow: 30
  lotQty: 1
risk:
  maxLots: 5
  maxDailyLossRupees: 5000
  maxOrdersPerMinute: 10
```

## Sample strategy (SMA crossover)

`strategies/sma/index.ts`:

```ts
import { Strategy, StrategyContext } from '@papertrading/sdk';

const closes: Record<string, number[]> = {};

export const strategy: Strategy = {
  name: 'sma',
  async init(ctx, params) {
    await ctx.subscribeBars(params.symbol as string, '1m');
  },
  async onBar(ctx, bar) {
    const c = (closes[bar.symbol] ??= []);
    c.push(bar.close);
    if (c.length > 100) c.shift();
    if (c.length < 30) return;
    const fast = avg(c.slice(-10));
    const slow = avg(c.slice(-30));
    const pos = (await ctx.getPositions()).find(p => p.symbol === bar.symbol);
    const flat = !pos || pos.netQty === 0;
    if (fast > slow && flat) {
      await ctx.placeOrder({ symbol: bar.symbol, side: 'BUY', qty: 1, type: 'MARKET', product: 'NRML' });
    } else if (fast < slow && pos && pos.netQty > 0) {
      await ctx.placeOrder({ symbol: bar.symbol, side: 'SELL', qty: pos.netQty, type: 'MARKET', product: 'NRML' });
    }
  },
};

const avg = (xs: number[]) => xs.reduce((a, b) => a + b, 0) / xs.length;
```

## Tasks

### 9.1 SDK

- `packages/sdk/`: TypeScript client wrapping gateway REST + WS.
- Internal `Clock` interface; `RealClock` (Date.now) vs. `VirtualClock` (driven by replay).
- In backtest mode, `now()`, `setTimeout`, any async I/O are routed through the clock.
- Order placement sync-awaits until OMS ack (not fill).
- Re-subscribe on reconnect.

### 9.2 Supervisor

- `services/strategy-runner`: takes manifest, forks a Node worker (`node --enable-source-maps ./strategies/<name>/index.ts`).
- Restart policy: exponential backoff; max 5 retries per hour.
- IPC channel for lifecycle: STOP, STATUS.
- Memory/CPU limits via `--max-old-space-size` and `cpulimit` or ulimit (dev only).

### 9.3 Risk caps at SDK boundary

- SDK enforces `maxLots`, `maxOrdersPerMinute`, `maxDailyLossRupees`. Rejects with an error before HTTP call.
- Server-side risk remains authoritative; SDK is a polite guard.

### 9.4 Backtest mode

- Virtual clock driven by replay adapter's timestamps.
- Strategy `now()` returns virtual time.
- All WS delivery is synchronous from the runner's perspective — tick handlers run to completion before next tick.
- Trades are "filled" via the matching engine still, but synthetic MMs are quoting off the virtual LTP.

### 9.5 Backtest report

- At end of run, compute: cumulative P&L, max drawdown, Sharpe (assume r=7% annual, 252 trading days), number of trades, win rate, avg win/loss, turnover.
- Render: HTML report with equity curve (uses a tiny charting lib like `uPlot`).
- Save to `./backtests/<strategy>-<date>.html`.

### 9.6 Determinism tests

- Run the sample strategy twice on the same replay date → assert identical trade log (byte-for-byte).
- Run at 1× vs. 100× speed → assert identical trade log.
- Swap the seed of synthetic MMs → expect *different* results, but each run internally deterministic.

## Metrics

- `strategy_heartbeat{name}` — per-strategy gauge.
- `strategy_orders_placed_total{name}`
- `strategy_errors_total{name,kind}`
- `strategy_handler_duration_ms{name,handler}`

## Performance targets

- SDK place-order round-trip p99 < 100 ms (live).
- Backtest a trading day at 1000× speed (~6.5 hrs → 23 s); end-to-end under 60 s including report.
- Replay determinism: 0 diffs across 10 runs.

## Testing

- Unit: SDK clock abstractions; risk caps enforcement.
- Integration: supervisor restart on strategy crash; SDK reconnect resilience.
- Contract: SDK ↔ Gateway API types match (generated from the same schemas).
- Replay determinism: as above.

## Common pitfalls

- Mixing wall-clock sleeps with virtual clock → backtest runs for real time.
- Strategy reading `process.env` vs. SDK config → config drift.
- Non-deterministic hashing / iteration order (Map insertion vs. Object keys).
- Forgetting to persist strategy state — restart loses position tracking. SDK should auto-reconcile from `getPositions()` on init.
- Rate-limit lockouts during high-frequency strategy testing — SDK should throttle gracefully.
- Not isolating strategies → bad strategy eats shared CPU.

## Interview talking points

- Live/backtest parity as a testing *superpower*; you caught bugs in `X` during replay runs.
- Why the SDK is the only component that knows about time.
- Process-level isolation vs. VM/container isolation (fits the "v2 scale" narrative).
- Determinism in distributed systems — what it costs, what it buys.
- Rate-limit tokens as a contract.
- Risk caps at SDK vs. server — defense in depth.
- Comparing to Kite Connect / zerodha-python — architectural similarities and differences.

## Resources

- Kite Connect Python client for API shape: <https://github.com/zerodha/pykiteconnect>
- QuantConnect Lean (conceptual reference only): <https://github.com/QuantConnect/Lean>
- Backtrader (Python) design.
- *Advances in Financial Machine Learning* — López de Prado (cross-validation pitfalls in backtests).
- "Survivorship bias in backtests" literature.
- Node `worker_threads` docs.
- `execa` for subprocess management.

## Exit checklist

- [ ] `pt strategy run --name=sma --mode=backtest --date=2024-03-15` produces a report.
- [ ] Same command twice → identical report.
- [ ] `pt strategy run --name=sma --mode=live` places orders during market hours.
- [ ] Strategy crash → supervisor restarts → resubscribes.
- [ ] ADR-0015 merged.
