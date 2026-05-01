package adapter

import (
	"context"
	"errors"
	"time"
)

// ErrNotConfigured is returned by the angel_live stub until Phase 11 wires SmartAPI.
var ErrNotConfigured = errors.New("md: angel_live adapter not configured (Phase 11)")

// Kind identifies which BrokerAdapter implementation is active.
type Kind string

const (
	KindNSEReplay Kind = "nse_replay"
	KindAngelLive Kind = "angel_live"
)

// BrokerAdapter is the only broker/replay-specific surface for the md service.
// Normalization, persistence, and fan-out sit outside this interface.
type BrokerAdapter interface {
	Kind() Kind

	// Run blocks until ctx is cancelled or the adapter hits a fatal error.
	// Hooks may be nil; implementations must not dereference without checking.
	Run(ctx context.Context, hooks *RunHooks) error
}

// RunHooks carries callbacks wired in later phases (normalizer → bus → persist).
type RunHooks struct {
	OnTick func(ctx context.Context, tick DraftTick) error
}

// DraftTick is adapter output before normalization (Phase 1.6).
type DraftTick struct {
	InstrumentID string
	Ts           time.Time
	LTP          float64
	BidPx        float64
	AskPx        float64
	Volume       int64
	Source       string // REPLAY | LIVE
}
