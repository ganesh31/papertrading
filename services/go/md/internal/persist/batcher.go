package persist

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ganesh/papertrading/services/go/md/internal/adapter"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const insertTickSQL = `INSERT INTO md.ticks (instrument_id, ts, ltp, bid_px, bid_qty, ask_px, ask_qty, volume, oi, source)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
ON CONFLICT (instrument_id, ts) DO NOTHING`

// Batcher buffers normalized ticks and flushes to md.ticks (500 rows or FlushEvery, whichever first).
type Batcher struct {
	pool    *pgxpool.Pool
	cfg     Config
	flushFn func(context.Context, []adapter.Tick) error

	mu  sync.Mutex
	buf []adapter.Tick
}

// NewBatcher creates a batcher that inserts via pool. cfg is resolved with defaults.
func NewBatcher(pool *pgxpool.Pool, cfg Config) *Batcher {
	b := &Batcher{pool: pool, cfg: cfg.resolved()}
	b.flushFn = b.insertPostgres
	return b
}

// Run drives periodic flush until ctx is cancelled, then performs a final flush.
func (b *Batcher) Run(ctx context.Context) {
	if b == nil || b.flushFn == nil {
		return
	}
	cfg := b.cfg.resolved()
	ticker := time.NewTicker(cfg.FlushEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			if err := b.Flush(context.Background()); err != nil {
				log.Printf("md persist: final flush: %v", err)
			}
			return
		case <-ticker.C:
			if err := b.Flush(ctx); err != nil {
				log.Printf("md persist: periodic flush: %v", err)
			}
		}
	}
}

// Enqueue adds a tick; may flush synchronously when the batch reaches MaxRows.
func (b *Batcher) Enqueue(ctx context.Context, t adapter.Tick) error {
	if b == nil || b.flushFn == nil {
		return nil
	}
	cfg := b.cfg.resolved()
	b.mu.Lock()
	b.buf = append(b.buf, t)
	need := len(b.buf) >= cfg.MaxRows
	b.mu.Unlock()
	if need {
		return b.Flush(ctx)
	}
	return nil
}

// Flush writes all buffered ticks (if any) to Postgres.
func (b *Batcher) Flush(ctx context.Context) error {
	if b == nil || b.flushFn == nil {
		return nil
	}
	b.mu.Lock()
	if len(b.buf) == 0 {
		b.mu.Unlock()
		return nil
	}
	chunk := b.buf
	b.buf = make([]adapter.Tick, 0, b.cfg.resolved().MaxRows)
	b.mu.Unlock()
	return b.flushFn(ctx, chunk)
}

func (b *Batcher) insertPostgres(ctx context.Context, rows []adapter.Tick) error {
	if b.pool == nil {
		return fmt.Errorf("persist: insert without database pool")
	}
	if len(rows) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for i := range rows {
		args, err := tickArgs(&rows[i])
		if err != nil {
			return err
		}
		batch.Queue(insertTickSQL, args...)
	}
	br := b.pool.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("md persist: batch exec row %d: %w", i, err)
		}
	}
	return br.Close()
}

func tickArgs(t *adapter.Tick) ([]any, error) {
	if t == nil {
		return nil, fmt.Errorf("persist: nil tick")
	}
	var bidPx, askPx any
	if t.BidPx != nil {
		bidPx = *t.BidPx
	}
	if t.AskPx != nil {
		askPx = *t.AskPx
	}
	var bidQty, askQty any
	if t.BidQty != nil {
		bidQty = *t.BidQty
	}
	if t.AskQty != nil {
		askQty = *t.AskQty
	}
	return []any{
		t.InstrumentID,
		t.Ts,
		t.LTP,
		bidPx,
		bidQty,
		askPx,
		askQty,
		t.Volume,
		t.OI,
		t.Source,
	}, nil
}
