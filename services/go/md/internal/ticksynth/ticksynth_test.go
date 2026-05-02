package ticksynth

import (
	"math"
	"testing"
	"time"
)

func barFixture() BarInput {
	ts := time.Date(2026, 4, 17, 9, 15, 0, 0, time.UTC)
	return BarInput{
		InstrumentID: "inst_1",
		BarStart:     ts,
		Open:         100,
		High:         105,
		Low:          99,
		Close:        103,
		Volume:       1000,
	}
}

func TestSynthesize_OpenCloseAndHighLow(t *testing.T) {
	b := barFixture()
	cfg := Config{TicksPerBar: 10, TickSize: 0.05, SpreadTicks: 1}
	ticks, err := Synthesize("sess-a", b.InstrumentID, b, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(ticks) != 10 {
		t.Fatalf("len=%d", len(ticks))
	}
	const eps = 1e-9
	if math.Abs(ticks[0].LTP-b.Open) > eps {
		t.Fatalf("open: got %v want %v", ticks[0].LTP, b.Open)
	}
	if math.Abs(ticks[len(ticks)-1].LTP-b.Close) > eps {
		t.Fatalf("close: got %v want %v", ticks[len(ticks)-1].LTP, b.Close)
	}
	var sawHi, sawLo bool
	for _, tk := range ticks {
		if math.Abs(tk.LTP-b.High) < 1e-6 {
			sawHi = true
		}
		if math.Abs(tk.LTP-b.Low) < 1e-6 {
			sawLo = true
		}
	}
	if !sawHi || !sawLo {
		t.Fatalf("high/low visit: sawHi=%v sawLo=%v", sawHi, sawLo)
	}
}

func TestSynthesize_VolumeConserved(t *testing.T) {
	b := barFixture()
	b.Volume = 10_007
	cfg := Config{TicksPerBar: 10}
	ticks, err := Synthesize("sess-v", b.InstrumentID, b, cfg)
	if err != nil {
		t.Fatal(err)
	}
	var sum int64
	for _, tk := range ticks {
		sum += tk.Volume
	}
	if sum != b.Volume {
		t.Fatalf("volume sum %d != %d", sum, b.Volume)
	}
}

func TestSynthesize_TimeMonotonic(t *testing.T) {
	b := barFixture()
	ticks, err := Synthesize("sess-t", b.InstrumentID, b, Config{TicksPerBar: 10})
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(ticks); i++ {
		if !ticks[i].Ts.After(ticks[i-1].Ts) {
			t.Fatalf("ts not strictly increasing at %d: %v %v", i, ticks[i-1].Ts, ticks[i].Ts)
		}
	}
	last := b.BarStart.Add(time.Minute)
	if !ticks[len(ticks)-1].Ts.Before(last) && !ticks[len(ticks)-1].Ts.Equal(last) {
		t.Fatalf("last ts %v not <= bar end %v", ticks[len(ticks)-1].Ts, last)
	}
}

func TestSynthesize_Determinism(t *testing.T) {
	b := barFixture()
	cfg := Config{TicksPerBar: 10, TickSize: 0.05, SpreadTicks: 1}
	a, err := Synthesize("replay-1", b.InstrumentID, b, cfg)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := Synthesize("replay-1", b.InstrumentID, b, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != len(b2) {
		t.Fatal("length mismatch")
	}
	for i := range a {
		if a[i].Ts != b2[i].Ts || a[i].LTP != b2[i].LTP || a[i].Volume != b2[i].Volume {
			t.Fatalf("diff at %d: %+v vs %+v", i, a[i], b2[i])
		}
	}
}

func TestSynthesize_DifferentSessionChangesStream(t *testing.T) {
	b := barFixture()
	cfg := Config{TicksPerBar: 10}
	x, _ := Synthesize("sess-1", b.InstrumentID, b, cfg)
	y, _ := Synthesize("sess-2", b.InstrumentID, b, cfg)
	same := true
	for i := range x {
		if x[i].LTP != y[i].LTP {
			same = false
			break
		}
	}
	if same {
		t.Fatal("expected different LTP stream for different session id")
	}
}

func TestSynthesize_BidAskSpread(t *testing.T) {
	b := barFixture()
	cfg := Config{TicksPerBar: 10, TickSize: 0.25, SpreadTicks: 2}
	ticks, err := Synthesize("sess-b", b.InstrumentID, b, cfg)
	if err != nil {
		t.Fatal(err)
	}
	half := 0.25 * 2
	for _, tk := range ticks {
		if math.Abs((tk.AskPx-tk.BidPx)-2*half) > 1e-9 {
			t.Fatalf("spread mismatch %+v", tk)
		}
		if math.Abs(tk.LTP-(tk.BidPx+half)) > 1e-9 {
			t.Fatalf("ltp not mid %+v", tk)
		}
	}
}

func TestSynthesize_FlatBar(t *testing.T) {
	ts := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	b := BarInput{
		InstrumentID: "x",
		BarStart:     ts,
		Open:         50,
		High:         50,
		Low:          50,
		Close:        50,
		Volume:       100,
	}
	ticks, err := Synthesize("s", b.InstrumentID, b, Config{TicksPerBar: 10})
	if err != nil {
		t.Fatal(err)
	}
	for _, tk := range ticks {
		if tk.LTP != 50 {
			t.Fatalf("flat bar ltp %v", tk.LTP)
		}
	}
}

func TestSynthesize_InvalidBar(t *testing.T) {
	b := barFixture()
	b.High = 90
	_, err := Synthesize("s", b.InstrumentID, b, Config{TicksPerBar: 10})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSynthesize_TicksPerBarTooSmall(t *testing.T) {
	b := barFixture()
	_, err := Synthesize("s", b.InstrumentID, b, Config{TicksPerBar: 3})
	if err == nil {
		t.Fatal("expected error for ticks_per_bar < 4")
	}
}
