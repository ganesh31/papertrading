// Package ticksynth turns one 1-minute OHLCV bar into a deterministic stream of synthetic ticks (Phase 1.4).
package ticksynth

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"time"
)

// BarInput is one minute of OHLCV for a single instrument.
type BarInput struct {
	InstrumentID string
	BarStart     time.Time // start of the 1m bucket (any timezone; tick times are derived from it)
	Open         float64
	High         float64
	Low          float64
	Close        float64
	Volume       int64
}

// SyntheticTick is one replay tick before full normalization (Phase 1.6).
type SyntheticTick struct {
	Ts     time.Time
	LTP    float64
	BidPx  float64
	AskPx  float64
	Volume int64
}

// Config controls synthesis; zero values pick defaults.
type Config struct {
	TicksPerBar int     // default 10; minimum 4 for the brownian-bridge path
	TickSize    float64 // default 0.05 — used for bid/ask spread around LTP
	SpreadTicks float64 // default 1 — bid/ask = LTP ± TickSize*SpreadTicks
}

const (
	defaultTicksPerBar = 10
	defaultTickSize    = 0.05
	defaultSpreadTicks = 1
)

func (c *Config) resolved() (Config, error) {
	out := *c
	switch {
	case out.TicksPerBar == 0:
		out.TicksPerBar = defaultTicksPerBar
	case out.TicksPerBar < 4:
		return Config{}, fmt.Errorf("ticksynth: ticks_per_bar must be >= 4 (got %d)", out.TicksPerBar)
	}
	if out.TickSize <= 0 {
		out.TickSize = defaultTickSize
	}
	if out.SpreadTicks <= 0 {
		out.SpreadTicks = defaultSpreadTicks
	}
	return out, nil
}

// Synthesize builds ticksPerBar ticks spanning [bar.BarStart, bar.BarStart+60s).
// Deterministic for the same (sessionID, instrumentID, bar.BarStart, OHLCV, cfg).
func Synthesize(sessionID, instrumentID string, bar BarInput, cfg Config) ([]SyntheticTick, error) {
	if bar.InstrumentID == "" {
		return nil, errors.New("ticksynth: empty instrument_id")
	}
	if bar.Volume < 0 {
		return nil, errors.New("ticksynth: negative volume")
	}
	var err error
	cfg, err = cfg.resolved()
	if err != nil {
		return nil, err
	}
	n := cfg.TicksPerBar
	if err := validateBar(bar); err != nil {
		return nil, err
	}

	rng := newSeededRNG(sessionID, instrumentID, bar.BarStart, bar.Open, bar.High, bar.Low, bar.Close, bar.Volume)

	// Distinct interior indices for forced high / low touches (phase doc: [1, N-2] on an (N)-tick grid).
	iHi, iLo := pickDistinctInterior(rng, n)

	prices := make([]float64, n)
	prices[0] = bar.Open
	prices[n-1] = bar.Close

	sigma := (bar.High - bar.Low) / 4.0
	if sigma <= 0 || math.IsNaN(sigma) {
		sigma = 1e-9
	}

	// Baseline open→close; interior noise (Brownian bridge variance, phase §1.4).
	for i := 1; i <= n-2; i++ {
		t := float64(i) / float64(n) // match phase pseudocode (i/N)
		base := bar.Open + (bar.Close-bar.Open)*float64(i)/float64(n)
		variance := (sigma * sigma) * t * (1.0 - t)
		std := math.Sqrt(variance)
		noise := rng.NormFloat64() * std
		prices[i] = clamp(base+noise, bar.Low, bar.High)
	}

	prices[iHi] = bar.High
	prices[iLo] = bar.Low

	volumes := uniformVolumes(bar.Volume, n)
	half := cfg.TickSize * cfg.SpreadTicks

	out := make([]SyntheticTick, n)
	for i := 0; i < n; i++ {
		ts := bar.BarStart.Add(time.Duration(float64(time.Minute) * float64(i) / float64(n-1)))
		ltp := prices[i]
		out[i] = SyntheticTick{
			Ts:     ts,
			LTP:    ltp,
			BidPx:  ltp - half,
			AskPx:  ltp + half,
			Volume: volumes[i],
		}
	}
	return out, nil
}

func validateBar(bar BarInput) error {
	if math.IsNaN(bar.Open) || math.IsNaN(bar.High) || math.IsNaN(bar.Low) || math.IsNaN(bar.Close) {
		return errors.New("ticksynth: NaN in OHLC")
	}
	if bar.High < bar.Low {
		return errors.New("ticksynth: high < low")
	}
	if bar.High < bar.Open || bar.High < bar.Close {
		return errors.New("ticksynth: high below open/close")
	}
	if bar.Low > bar.Open || bar.Low > bar.Close {
		return errors.New("ticksynth: low above open/close")
	}
	return nil
}

func newSeededRNG(sessionID, instrumentID string, barStart time.Time, o, h, l, c float64, vol int64) *rand.Rand {
	hh := sha256.New()
	_, _ = fmt.Fprintf(hh, "%s|%s|%d|%g|%g|%g|%g|%d", sessionID, instrumentID, barStart.UnixNano(), o, h, l, c, vol)
	sum := hh.Sum(nil)
	s0 := binary.LittleEndian.Uint64(sum[0:8])
	s1 := binary.LittleEndian.Uint64(sum[8:16])
	return rand.New(rand.NewPCG(s0, s1))
}

func pickDistinctInterior(rng *rand.Rand, n int) (iHi, iLo int) {
	// Interior indices: 1 .. n-2 inclusive.
	lo := 1
	hi := n - 2
	span := hi - lo + 1
	iHi = lo + rng.IntN(span)
	for {
		iLo = lo + rng.IntN(span)
		if iLo != iHi {
			break
		}
	}
	return iHi, iLo
}

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func uniformVolumes(total int64, n int) []int64 {
	if n <= 0 {
		return nil
	}
	out := make([]int64, n)
	if total == 0 {
		return out
	}
	base := total / int64(n)
	rem := total % int64(n)
	for i := 0; i < n; i++ {
		out[i] = base
		if int64(i) < rem {
			out[i]++
		}
	}
	return out
}
