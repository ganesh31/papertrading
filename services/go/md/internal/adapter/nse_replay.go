package adapter

import (
	"context"
)

// NSEReplayAdapter replays staged NSE equity bars (Phase 1.3). Phase 1-T02: Run waits on ctx only.
type NSEReplayAdapter struct{}

func (NSEReplayAdapter) Kind() Kind { return KindNSEReplay }

func (NSEReplayAdapter) Run(ctx context.Context, _ *RunHooks) error {
	<-ctx.Done()
	return ctx.Err()
}
