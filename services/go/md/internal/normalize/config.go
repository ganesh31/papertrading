package normalize

import "time"

// Config controls staleness and instrument cache TTL.
type Config struct {
	LiveStaleness time.Duration // wall clock vs tick ts for LIVE only; default 60s
	InstrumentTTL time.Duration // Redis + in-process cache; default 24h
	// AdapterKind is the active md broker adapter (Prometheus "adapter" label); empty → "unknown".
	AdapterKind string
}

func (c Config) resolved() Config {
	out := c
	if out.LiveStaleness <= 0 {
		out.LiveStaleness = 60 * time.Second
	}
	if out.InstrumentTTL <= 0 {
		out.InstrumentTTL = 24 * time.Hour
	}
	if out.AdapterKind == "" {
		out.AdapterKind = "unknown"
	}
	return out
}

// DefaultConfig returns Phase 1.6 defaults.
func DefaultConfig() Config {
	return Config{
		LiveStaleness: 60 * time.Second,
		InstrumentTTL: 24 * time.Hour,
	}
}
