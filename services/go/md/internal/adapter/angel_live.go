package adapter

import (
	"context"
)

// AngelLiveAdapter is a Phase 11 placeholder; Run fails until SmartAPI WS is implemented.
type AngelLiveAdapter struct{}

func (AngelLiveAdapter) Kind() Kind { return KindAngelLive }

func (AngelLiveAdapter) Run(ctx context.Context, _ *RunHooks) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrNotConfigured
	}
}
