package persist

import (
	"context"
	"testing"
	"time"

	"github.com/ganesh/papertrading/services/go/md/internal/adapter"
)

func TestBatcher_FlushAtMaxRows(t *testing.T) {
	var flushCount int
	var lastN int
	b := &Batcher{
		cfg: Config{MaxRows: 3, FlushEvery: time.Hour}.resolved(),
		flushFn: func(_ context.Context, rows []adapter.Tick) error {
			flushCount++
			lastN = len(rows)
			return nil
		},
	}
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := b.Enqueue(ctx, adapter.Tick{InstrumentID: "a", Ts: time.Unix(int64(i), 0).UTC(), LTP: 1, Source: "REPLAY"}); err != nil {
			t.Fatal(err)
		}
	}
	if flushCount != 1 || lastN != 3 {
		t.Fatalf("want flush after 3 rows, got flushCount=%d lastN=%d", flushCount, lastN)
	}
	if err := b.Enqueue(ctx, adapter.Tick{InstrumentID: "b", Ts: time.Unix(10, 0).UTC(), LTP: 2, Source: "REPLAY"}); err != nil {
		t.Fatal(err)
	}
	if flushCount != 1 {
		t.Fatalf("want no second flush yet, got %d", flushCount)
	}
	if err := b.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if flushCount != 2 || lastN != 1 {
		t.Fatalf("want final flush 1 row, got flushCount=%d lastN=%d", flushCount, lastN)
	}
}

func TestTickArgs_NilPointers(t *testing.T) {
	tick := adapter.Tick{
		InstrumentID: "x",
		Ts:           time.Unix(1, 0).UTC(),
		LTP:          100,
		Volume:       5,
		OI:           0,
		Source:       "REPLAY",
	}
	args, err := tickArgs(&tick)
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 10 {
		t.Fatalf("len %d", len(args))
	}
	if args[3] != nil || args[4] != nil || args[5] != nil || args[6] != nil {
		t.Fatalf("expected nil bid/ask fields, got %#v", args[3:7])
	}
}
