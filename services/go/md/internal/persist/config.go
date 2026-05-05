package persist

import "time"

// Config controls tick batch flush behavior (Phase 1.7).
type Config struct {
	MaxRows    int           // flush when buffer reaches this count (default 500)
	FlushEvery time.Duration // periodic flush when buffer non-empty (default 100ms)
}

func (c Config) resolved() Config {
	out := c
	if out.MaxRows <= 0 {
		out.MaxRows = 500
	}
	if out.FlushEvery <= 0 {
		out.FlushEvery = 100 * time.Millisecond
	}
	return out
}

// DefaultConfig matches docs/phases/phase-01-market-data.md §1.7.
func DefaultConfig() Config {
	return Config{
		MaxRows:    500,
		FlushEvery: 100 * time.Millisecond,
	}
}
