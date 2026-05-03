package normalize

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const redisKeyPrefix = "md:inst:v1:"

type instrumentRow struct {
	InstrumentID  string  `json:"instrument_id"`
	Tradingsymbol string  `json:"tradingsymbol"`
	TickSize      float64 `json:"tick_size"`
}

func redisInstKey(id string) string {
	return redisKeyPrefix + id
}

func (n *Normalizer) loadInstrument(ctx context.Context, id string) (instrumentRow, error) {
	if id == "" {
		return instrumentRow{}, fmt.Errorf("normalize: empty instrument_id")
	}
	cfg := n.cfg.resolved()
	if row, ok := n.mem.get(id); ok {
		return row, nil
	}
	if n.rdb != nil {
		raw, err := n.rdb.Get(ctx, redisInstKey(id)).Bytes()
		if err == nil {
			var row instrumentRow
			if err := json.Unmarshal(raw, &row); err == nil && row.InstrumentID != "" {
				n.mem.set(id, row, cfg.InstrumentTTL)
				return row, nil
			}
		} else if err != redis.Nil {
			log.Printf("normalize: redis GET %s: %v", id, err)
		}
	}
	if n.pool == nil {
		return instrumentRow{}, fmt.Errorf("normalize: instrument %q not in cache and no DATABASE_URL", id)
	}
	row, err := n.loadInstrumentPG(ctx, id)
	if err != nil {
		return instrumentRow{}, err
	}
	n.mem.set(id, row, cfg.InstrumentTTL)
	if n.rdb != nil {
		if b, err := json.Marshal(row); err == nil {
			if err := n.rdb.Set(ctx, redisInstKey(id), b, cfg.InstrumentTTL).Err(); err != nil {
				log.Printf("normalize: redis SET %s: %v", id, err)
			}
		}
	}
	return row, nil
}

func (n *Normalizer) loadInstrumentPG(ctx context.Context, id string) (instrumentRow, error) {
	const q = `
SELECT instrument_id, tradingsymbol, COALESCE(tick_size::float8, 0.05)
FROM ref.instruments
WHERE instrument_id = $1
`
	var row instrumentRow
	err := n.pool.QueryRow(ctx, q, id).Scan(&row.InstrumentID, &row.Tradingsymbol, &row.TickSize)
	if err != nil {
		if err == pgx.ErrNoRows {
			return instrumentRow{}, fmt.Errorf("normalize: unknown instrument_id %q", id)
		}
		return instrumentRow{}, fmt.Errorf("normalize: load instrument: %w", err)
	}
	return row, nil
}

// Normalizer validates instruments and builds canonical ticks.
type Normalizer struct {
	pool *pgxpool.Pool
	rdb  *redis.Client
	mem  *memCache
	cfg  Config
}

// New builds a normalizer. pool may be nil (skips instrument master lookup; not recommended for production).
// rdb may be nil (instrument rows load from Postgres each miss, cached in-process only).
func New(pool *pgxpool.Pool, rdb *redis.Client, cfg Config) *Normalizer {
	return &Normalizer{
		pool: pool,
		rdb:  rdb,
		mem:  newMemCache(),
		cfg:  cfg,
	}
}
