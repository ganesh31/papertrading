package normalize

import (
	"testing"
	"time"

	"github.com/ganesh/papertrading/services/go/md/internal/adapter"
)

func TestCanonicalSource(t *testing.T) {
	if g, w := canonicalSource(" live "), "LIVE"; g != w {
		t.Fatalf("got %q want %q", g, w)
	}
	if g, w := canonicalSource("replay"), "REPLAY"; g != w {
		t.Fatalf("got %q want %q", g, w)
	}
	if g, w := canonicalSource(""), "REPLAY"; g != w {
		t.Fatalf("got %q want %q", g, w)
	}
}

func TestStaleLive(t *testing.T) {
	now := time.Unix(1700000000, 0)
	ts := now.Add(-61 * time.Second)
	if !staleLive("LIVE", ts, now, 60*time.Second) {
		t.Fatal("expected stale")
	}
	if staleLive("LIVE", now.Add(-30*time.Second), now, 60*time.Second) {
		t.Fatal("expected fresh")
	}
	if staleLive("REPLAY", ts, now, 60*time.Second) {
		t.Fatal("replay must not use wall staleness")
	}
}

func TestBuildTick(t *testing.T) {
	d := adapter.DraftTick{
		InstrumentID: "abc",
		Ts:           time.Unix(1, 0).UTC(),
		LTP:          100,
		BidPx:        99.95,
		AskPx:        100.05,
		Volume:       10,
		Source:       "replay",
	}
	got := buildTick(d, "REPLAY")
	if got.Source != "REPLAY" || got.LTP != 100 || got.Volume != 10 || got.OI != 0 {
		t.Fatalf("unexpected tick: %+v", got)
	}
	if got.BidPx == nil || *got.BidPx != 99.95 || got.AskPx == nil || *got.AskPx != 100.05 {
		t.Fatalf("bid/ask: %+v", got)
	}
}
