package normalize

import (
	"context"
	"strings"
	"time"

	"github.com/ganesh/papertrading/services/go/md/internal/adapter"
	"github.com/ganesh/papertrading/services/go/md/internal/mdmetrics"
)

// Normalize maps a draft tick to a canonical Tick, optionally dropping stale LIVE ticks.
// drop is true when a LIVE tick is older than LiveStaleness vs now (REPLAY is never dropped for staleness).
func (n *Normalizer) Normalize(ctx context.Context, d adapter.DraftTick, now time.Time) (adapter.Tick, bool, error) {
	cfg := DefaultConfig().resolved()
	if n != nil {
		cfg = n.cfg.resolved()
	}
	src := canonicalSource(d.Source)
	if staleLive(src, d.Ts, now, cfg.LiveStaleness) {
		if src == "LIVE" {
			mdmetrics.TicksStaleDropped.WithLabelValues(cfg.AdapterKind).Inc()
		}
		return adapter.Tick{}, true, nil
	}
	if n == nil || n.pool == nil {
		if src == "LIVE" {
			mdmetrics.TickStalenessSeconds.WithLabelValues(cfg.AdapterKind).Observe(now.Sub(d.Ts).Seconds())
		}
		return buildTick(d, src), false, nil
	}
	if _, err := n.loadInstrument(ctx, d.InstrumentID); err != nil {
		return adapter.Tick{}, false, err
	}
	if src == "LIVE" {
		mdmetrics.TickStalenessSeconds.WithLabelValues(cfg.AdapterKind).Observe(now.Sub(d.Ts).Seconds())
	}
	return buildTick(d, src), false, nil
}

// WrapWithNormalizer returns RunHooks whose OnTick runs normalization then inner.OnNormalizedTick.
func WrapWithNormalizer(inner *adapter.RunHooks, n *Normalizer) *adapter.RunHooks {
	if inner == nil {
		inner = &adapter.RunHooks{}
	}
	return &adapter.RunHooks{
		OnTick: func(ctx context.Context, d adapter.DraftTick) error {
			t, drop, err := n.Normalize(ctx, d, time.Now())
			if drop {
				return nil
			}
			if err != nil {
				return err
			}
			if inner.OnNormalizedTick != nil {
				return inner.OnNormalizedTick(ctx, t)
			}
			return nil
		},
	}
}

func canonicalSource(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "LIVE":
		return "LIVE"
	default:
		return "REPLAY"
	}
}

func staleLive(source string, tickTs, now time.Time, maxAge time.Duration) bool {
	if source != "LIVE" {
		return false
	}
	return now.Sub(tickTs) > maxAge
}

func buildTick(d adapter.DraftTick, src string) adapter.Tick {
	bp := d.BidPx
	ap := d.AskPx
	return adapter.Tick{
		InstrumentID: d.InstrumentID,
		Ts:           d.Ts,
		LTP:          d.LTP,
		BidPx:        &bp,
		BidQty:       nil,
		AskPx:        &ap,
		AskQty:       nil,
		Volume:       d.Volume,
		OI:           0,
		Source:       src,
	}
}
