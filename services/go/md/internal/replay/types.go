package replay

import (
	"context"
	"time"
)

// Tick is one replay emit line before Phase 1.6 normalizer (same shape as adapter.DraftTick).
type Tick struct {
	InstrumentID string
	Ts           time.Time
	LTP          float64
	BidPx        float64
	AskPx        float64
	Volume       int64
	Source       string // REPLAY
}

// Sink receives synthesized replay ticks.
type Sink interface {
	OnReplayTick(ctx context.Context, t Tick) error
}
