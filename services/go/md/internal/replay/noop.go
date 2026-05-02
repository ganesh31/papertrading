package replay

import "context"

type noopSink struct{}

func (noopSink) OnReplayTick(context.Context, Tick) error { return nil }
