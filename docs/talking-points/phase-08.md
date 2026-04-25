# Talking Points — Phase 8: SPAN Margin Engine

The other big one. Interviewers love probing here because most candidates wave hands at margin.

## 1. What is SPAN, in your own words?

**90-second answer**

SPAN — Standard Portfolio Analysis of Risk — is a scenario-based margin methodology invented by CME in 1988. NSE licenses and uses it for F&O. The core idea: given a portfolio of derivative positions, what's the worst-case single-day loss across a fixed set of market move scenarios? That worst-case loss is the initial margin the client must post.

It's **scenario-based** rather than parametric VaR or historical VaR. You define a grid of standard moves — price scans in multiples of a standard deviation, volatility scans, extreme moves. For each scenario you revalue every contract in the portfolio — linear for futures, Black-Scholes for options — sum P&L, and pick the worst scenario. That's your Scanning Risk.

On top of Scanning Risk, SPAN adds Spread Charges (calendar and inter-commodity), Short Option Minimum (a floor to guard against cheap short options), and then NSE layers an Exposure Margin on top — a flat percentage of contract value, basically a regulatory cushion.

Total upfront margin collected from the client = SPAN + Exposure.

**Trade-off I'd defend**

Scenario-based is **conservative** compared to parametric VaR. That's intentional — regulators prefer conservatism and explainability over statistical tightness. SPAN is easy to audit: "show me the scenario that hit your worst loss" is a question with a concrete answer. A portfolio VaR would require explaining covariance matrices to the regulator.

**At scale**

SPAN is embarrassingly parallel — each portfolio is independent. NSE Clearing runs it on thousands of portfolios in batches; the same gRPC service can scale horizontally. The optimization I didn't do: precompute the per-contract P&L vector across scenarios once and combine vectors per portfolio — my current code re-evaluates contracts. That's a 10× speedup left on the table.

---

## 2. Walk me through computing SPAN margin for a NIFTY straddle

**90-second answer**

Long NIFTY 25000 CE + Long NIFTY 25000 PE, current expiry, spot at 25000.

First, fetch NSE's SPAN parameters for NIFTY for today: price scan range (say 6% — so ±1500 points), vol scan (±25% of current IV), extreme move (2× price range = ±12%, weighted 32%).

Build 16 scenarios: 7 price moves × 2 vol moves + 2 extreme moves.

For each scenario, reprice both legs with Black-Scholes using the shifted spot and shifted IV. Compute P&L per leg. Sum.

A long straddle is long convexity — it gains on big moves in either direction and loses on time decay and vol drops. So the worst scenario is typically the "no move + vol down" one. Let's say worst-case loss is ₹18,000 per straddle. That's Scanning Risk.

Short Option Minimum: doesn't apply (we're long options, not short).

Spread charge: none — single underlying, no calendar spread.

Exposure: ~3% of contract value = 25000 × 75 × 0.03 = ₹56,250 — wait, that's way higher than Scanning Risk. Actually for long options, the premium paid is often more than the scanning risk, and NSE caps total margin at premium for long options (you can only lose what you paid). So margin for a long straddle is effectively premium paid.

Total: `min(SPAN + Exposure, premium paid)`.

**Trade-off I'd defend**

The premium-cap rule is something I had to add explicitly — CME's SPAN doesn't have it; NSE's does. Mention this as an Indian-specific deviation.

**At scale**

Cache the per-position P&L across scenarios; combine for incremental changes.

---

## 3. What's "incremental margin" and why does your risk service do it?

**90-second answer**

Incremental margin = margin required for adding a new order to an existing portfolio. It's `CalculateMargin(portfolio ∪ new_order) - CalculateMargin(portfolio)`.

Why? Because a new position in a hedged portfolio might **decrease** total margin (e.g., a short call against a long future reduces directional exposure). If I charged the new order's standalone margin, I'd over-block cash — bad UX.

So pre-trade: compute margin of current portfolio (cached 100 ms), compute with the hypothetical addition, compare to user's available cash. Reject if the delta exceeds available cash.

SEBI mandates upfront margin collection, so this check has to be **pre-trade and synchronous** — non-negotiable latency budget.

**Trade-off I'd defend**

Caching the baseline margin for 100 ms bounds stale data risk: if prices move meaningfully in 100 ms, the margin might be stale — but by less than the scenario range, so it's safe. ADR captures this.

**At scale**

Precompute and stream baseline margin as positions change — the service maintains a per-user margin cache updated on every trade. Order-path calls are then pure delta compute.

---

## 4. Why does a calendar spread get a margin credit?

**90-second answer**

A calendar spread is long one expiry, short another, same strike. Scenario-wise, a price move hurts one leg and helps the other roughly equally — but not exactly, because the two contracts have different time-to-expiry and therefore different greeks, especially vega and theta.

SPAN's scanning would give you near-full netting on the price dimension, but basis risk remains — the short-dated leg is more sensitive to near-term vol, the long-dated leg to longer-term vol. That basis risk is what the **Intra-Commodity Spread Charge** prices in. It's explicitly a small charge *added back* to the highly-netted scanning risk, acknowledging "yes, you're hedged, but not perfectly".

The alternative — computing scanning on each leg standalone — would ignore the hedge and charge huge margin, which wouldn't reflect real risk.

**Trade-off I'd defend**

Spread charge magnitudes are set by the clearing corp based on historical basis volatility. There's no universally correct number; it's a regulatory calibration.

**At scale**

Detect spreads automatically by scanning the portfolio for offsetting positions on the same underlying — O(N) per portfolio. Easy.

---

## 5. How do you reconcile your SPAN against NSE's?

**90-second answer**

NSE publishes a daily "margin file" with per-contract margin values for standard single-contract portfolios. I fetch that file and compute my own SPAN for the same portfolios. Acceptable drift: 1–2%.

Sources of drift: (a) NSE uses their own futures price scan methodology — they have a separate "futures scan range" distinct from the spot scan, I approximate with spot; (b) BS model differences — my r, q assumptions vs theirs; (c) rounding.

For multi-leg portfolios, NSE doesn't publish — so reconciliation is against a few hand-computed reference cases.

CI runs nightly; alerts on > 5% drift. When it drifts, usually NSE updated a parameter or methodology, and I need to catch up.

**Trade-off I'd defend**

Perfect parity with NSE would require licensing SPAN and using their exact binary — not available to individuals. 1–2% drift is the best a paper implementation can achieve.

**At scale**

License CME SPAN library or become a clearing member — both require registration and fees, out of scope for this project.

---

## 6. What's the Short Option Minimum and why does it exist?

**90-second answer**

A deep OTM short option has near-zero premium and near-zero Scanning Risk under standard scenarios — the option is so far from strike that no scenario brings it in-the-money meaningfully. But it's still a naked short option; tail risk is unbounded.

SOM is a floor on per-position margin that kicks in when Scanning Risk alone is too low. Say SOM is ₹5000 per short option lot. If Scanning Risk computed ₹500, we use ₹5000 instead.

It's a regulatory safety net against black-swan events that SPAN's standard scenarios don't cover.

**Trade-off I'd defend**

SOM is crude — it's the same number regardless of moneyness. A better model would scale with delta or distance from strike. But SPAN values simplicity and explicability here.

**At scale**

Dynamic SOM could be computed via tail-scenario revaluation, but NSE keeps it flat for operational simplicity.

---

## 7. What are the limitations of SPAN?

**90-second answer**

Three big ones:

1. **Linear netting assumption within spreads** — real basis risk isn't linear, and extreme moves can decorrelate legs more than the spread charge anticipates.
2. **Scenario grid is fixed** — a regime shift (COVID-level vol jumps) isn't in the daily parameter set until NSE updates it, which typically lags by days.
3. **No correlation across underlyings in standard form** — inter-commodity credit is manual, not a covariance matrix. A portfolio long 10 different stocks gets very little netting benefit even though they're correlated.

Alternatives: Portfolio VaR (historical or Monte Carlo) addresses these but is harder to audit and can be pro-cyclical (bigger margin calls in stressed markets, feeding the stress). That's why regulators don't love it.

CME has announced SPAN 2 (a more VaR-ish model) and is migrating; NSE hasn't.

**Trade-off I'd defend**

SPAN's conservatism is a feature when you're a clearing corp holding systemic risk. For an individual trader-centric view, VaR would give tighter margins but volatile requirements.

**At scale**

This is where you'd invest in a VaR engine for internal risk management while still collecting SPAN+Exposure from clients (regulatory floor).

---

## 8. How did you handle options expiry in margin?

**90-second answer**

As expiry approaches, time-to-expiry T approaches zero, and options reprice sharply. An ATM option at T=0 has vega near zero but delta near 0.5 — gamma risk spikes.

SPAN handles this naturally because its scenarios revalue with BS using the actual T. But the standard scenarios may underestimate gamma risk on expiry day. NSE historically has had special "near-expiry" vol shifts to compensate.

In my implementation, I read NSE's daily parameters which include those adjustments. On expiry day, the scanning loss for a short ATM straddle is larger than a normal day — the scenarios naturally catch it.

Post-expiry, the position either settles (ITM → intrinsic cash, margin releases) or vanishes (OTM → zero). My margin service decommissions positions as expiry processes, so the margin drop reflects at EOD.

**Trade-off I'd defend**

"Near-expiry blow-up" is a real trader-kill event; SPAN captures it imperfectly. Some brokers layer in *additional* expiry-day margin (a private add-on). I don't — ADR notes this as a known gap.

**At scale**

Monitor expiry-day margin coverage ratios as a KPI; alert ops if a client's margin dips below 120% of SPAN on expiry day.
