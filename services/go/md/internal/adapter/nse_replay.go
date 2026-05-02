package adapter

import (
	"context"

	"github.com/ganesh/papertrading/services/go/md/internal/replay"
)

// NSEReplayAdapter replays staged NSE equity bars (Phase 1.3).
type NSEReplayAdapter struct {
	Coord *replay.Coordinator
}

func (NSEReplayAdapter) Kind() Kind { return KindNSEReplay }

func (a NSEReplayAdapter) Run(ctx context.Context, hooks *RunHooks) error {
	if a.Coord == nil {
		<-ctx.Done()
		return ctx.Err()
	}
	var sink replay.Sink = replaySinkBridge{hooks: hooks}
	return a.Coord.RunIdle(ctx, sink)
}

type replaySinkBridge struct {
	hooks *RunHooks
}

func (s replaySinkBridge) OnReplayTick(ctx context.Context, t replay.Tick) error {
	if s.hooks == nil || s.hooks.OnTick == nil {
		return nil
	}
	return s.hooks.OnTick(ctx, DraftTick{
		InstrumentID: t.InstrumentID,
		Ts:           t.Ts,
		LTP:          t.LTP,
		BidPx:        t.BidPx,
		AskPx:        t.AskPx,
		Volume:       t.Volume,
		Source:       t.Source,
	})
}
