package replay

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ganesh/papertrading/services/go/md/internal/mdmetrics"
	"github.com/ganesh/papertrading/services/go/md/internal/ticksynth"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Status is a snapshot for GET /replay/status.
type Status struct {
	Running      bool      `json:"running"`
	VirtualTime  time.Time `json:"-"`
	Speed        float64   `json:"speed"`
	TicksEmitted uint64    `json:"ticksEmitted"`
	SessionID    string    `json:"sessionId"`
	Error        string    `json:"error"`
}

// Coordinator loads bars, synthesizes ticks, and drives virtual clock + sink.
type Coordinator struct {
	appCtx context.Context
	pool   *pgxpool.Pool

	mu   sync.RWMutex
	sink Sink

	runMu     sync.Mutex
	runCancel context.CancelFunc

	statusMu sync.RWMutex
	status   Status
}

func NewCoordinator(appCtx context.Context, pool *pgxpool.Pool) *Coordinator {
	return &Coordinator{
		appCtx: appCtx,
		pool:   pool,
		sink:   noopSink{},
	}
}

func (c *Coordinator) SetSink(s Sink) {
	if s == nil {
		s = noopSink{}
	}
	c.mu.Lock()
	c.sink = s
	c.mu.Unlock()
}

func (c *Coordinator) Snapshot() Status {
	c.statusMu.RLock()
	defer c.statusMu.RUnlock()
	return c.status
}

func (c *Coordinator) Stop() {
	c.runMu.Lock()
	cancel := c.runCancel
	c.runCancel = nil
	c.runMu.Unlock()
	if cancel != nil {
		cancel()
	}
	c.statusMu.Lock()
	st := c.status
	st.Running = false
	c.status = st
	c.statusMu.Unlock()
}

// RunIdle optionally starts env-driven replay, then blocks until ctx cancelled.
func (c *Coordinator) RunIdle(ctx context.Context, sink Sink) error {
	c.SetSink(sink)

	cfg, err := EnvConfig()
	if err != nil {
		return err
	}
	if cfg != nil {
		rctx, cancel := context.WithCancel(ctx)
		c.runMu.Lock()
		c.runCancel = cancel
		c.runMu.Unlock()
		go func() {
			defer cancel()
			err := c.runReplay(rctx, *cfg)
			if err != nil && !errors.Is(err, context.Canceled) {
				c.setStatusErr(err.Error())
			}
		}()
	}

	<-ctx.Done()
	c.Stop()
	return ctx.Err()
}

func (c *Coordinator) Start(req Request) error {
	if c.pool == nil {
		return fmt.Errorf("replay: DATABASE_URL / pool not configured")
	}
	if err := mergeDefaults(&req); err != nil {
		return err
	}

	c.Stop()

	rctx, cancel := context.WithCancel(c.appCtx)
	c.runMu.Lock()
	c.runCancel = cancel
	c.runMu.Unlock()

	go func() {
		defer cancel()
		err := c.runReplay(rctx, req)
		if err != nil && !errors.Is(err, context.Canceled) {
			c.setStatusErr(err.Error())
		}
	}()
	return nil
}

func (c *Coordinator) setStatusErr(msg string) {
	c.statusMu.Lock()
	st := c.status
	st.Error = msg
	st.Running = false
	c.status = st
	c.statusMu.Unlock()
}

func (c *Coordinator) runReplay(ctx context.Context, req Request) error {
	c.statusMu.Lock()
	c.status = Status{
		Running:     true,
		Speed:       req.Speed,
		SessionID:   req.SessionID,
		Error:       "",
		TicksEmitted: 0,
	}
	c.statusMu.Unlock()

	defer func() {
		c.statusMu.Lock()
		st := c.status
		st.Running = false
		c.status = st
		c.statusMu.Unlock()
	}()

	dayStart, dayEnd, err := parseDateIST(req.Date)
	if err != nil {
		return err
	}

	idBySym, err := c.resolveInstrumentIDs(ctx, req.Symbols)
	if err != nil {
		return err
	}
	if len(idBySym) == 0 {
		return fmt.Errorf("replay: no instruments matched symbols %v", req.Symbols)
	}

	var ids []string
	for _, sym := range req.Symbols {
		if id, ok := idBySym[sym]; ok {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return fmt.Errorf("replay: symbols not in ref.instruments (sync scrip master?)")
	}

	bars, err := c.loadBars(ctx, ids, dayStart, dayEnd)
	if err != nil {
		return err
	}
	if len(bars) == 0 {
		return fmt.Errorf("replay: no bars in md.bars_1m for date %s", req.Date)
	}

	tsCfg := req.ticksynthConfig()
	type queued struct {
		iid string
		ts  time.Time
		st  ticksynth.SyntheticTick
	}
	var q []queued
	for _, b := range bars {
		synth, err := ticksynth.Synthesize(req.SessionID, b.InstrumentID, b, tsCfg)
		if err != nil {
			return err
		}
		for _, t := range synth {
			q = append(q, queued{iid: b.InstrumentID, ts: t.Ts, st: t})
		}
	}

	sort.Slice(q, func(i, j int) bool {
		if !q[i].ts.Equal(q[j].ts) {
			return q[i].ts.Before(q[j].ts)
		}
		return q[i].iid < q[j].iid
	})

	totalTicks := len(q)
	defer func() {
		mdmetrics.ReplayRunning.Set(0)
		mdmetrics.ReplaySpeedMultiplier.Set(0)
		mdmetrics.ReplayPendingTicks.Set(0)
		mdmetrics.ReplayVirtualTimestampSeconds.Set(0)
		mdmetrics.ReplaySessionBarRows.Set(0)
	}()
	mdmetrics.ReplayRunning.Set(1)
	mdmetrics.ReplaySpeedMultiplier.Set(req.Speed)
	mdmetrics.ReplaySessionBarRows.Set(float64(len(bars)))
	mdmetrics.ReplayPendingTicks.Set(float64(totalTicks))

	realStart := time.Now()
	virtualAnchor := q[0].ts

	c.mu.RLock()
	sink := c.sink
	c.mu.RUnlock()
	if sink == nil {
		sink = noopSink{}
	}

	var n uint64
	for _, item := range q {
		if err := ctx.Err(); err != nil {
			return err
		}
		waitVirtual(ctx, realStart, virtualAnchor, item.ts, req.Speed)

		tick := Tick{
			InstrumentID: item.iid,
			Ts:           item.st.Ts,
			LTP:          item.st.LTP,
			BidPx:        item.st.BidPx,
			AskPx:        item.st.AskPx,
			Volume:       item.st.Volume,
			Source:       "REPLAY",
		}
		if err := sink.OnReplayTick(ctx, tick); err != nil {
			return err
		}
		n++
		mdmetrics.ReplayVirtualTimestampSeconds.Set(float64(item.ts.UnixNano()) / 1e9)
		mdmetrics.ReplayPendingTicks.Set(float64(totalTicks - int(n)))
		c.statusMu.Lock()
		st := c.status
		st.VirtualTime = item.ts
		st.TicksEmitted = n
		c.status = st
		c.statusMu.Unlock()
	}

	return nil
}

func waitVirtual(ctx context.Context, realStart, virtualAnchor, target time.Time, speed float64) {
	if speed <= 0 {
		speed = defaultSpeed
	}
	for {
		if err := ctx.Err(); err != nil {
			return
		}
		elapsed := time.Since(realStart)
		virtElapsed := time.Duration(float64(elapsed) * speed)
		virtualNow := virtualAnchor.Add(virtElapsed)
		if !virtualNow.Before(target) {
			return
		}
		gap := target.Sub(virtualNow)
		sleep := time.Duration(float64(gap) / speed)
		if sleep < time.Millisecond {
			sleep = time.Millisecond
		}
		t := time.NewTimer(sleep)
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
		}
		t.Stop()
	}
}

func (c *Coordinator) resolveInstrumentIDs(ctx context.Context, tradingsymbols []string) (map[string]string, error) {
	const q = `
SELECT tradingsymbol, instrument_id
FROM ref.instruments
WHERE exchange = 'NSE' AND segment = 'NSE_EQ' AND tradingsymbol = ANY($1::text[])
`
	rows, err := c.pool.Query(ctx, q, tradingsymbols)
	if err != nil {
		return nil, fmt.Errorf("replay: resolve instruments: %w", err)
	}
	defer rows.Close()

	out := make(map[string]string)
	for rows.Next() {
		var sym, id string
		if err := rows.Scan(&sym, &id); err != nil {
			return nil, err
		}
		out[sym] = id
	}
	return out, rows.Err()
}

func (c *Coordinator) loadBars(ctx context.Context, instrumentIDs []string, dayStart, dayEnd time.Time) ([]ticksynth.BarInput, error) {
	const q = `
SELECT instrument_id, ts,
       open::float8, high::float8, low::float8, close::float8, volume
FROM md.bars_1m
WHERE instrument_id = ANY($1::text[])
  AND ts >= $2 AND ts < $3
ORDER BY ts ASC, instrument_id ASC
`
	rows, err := c.pool.Query(ctx, q, instrumentIDs, dayStart, dayEnd)
	if err != nil {
		return nil, fmt.Errorf("replay: load bars: %w", err)
	}
	defer rows.Close()

	var bars []ticksynth.BarInput
	for rows.Next() {
		var b ticksynth.BarInput
		if err := rows.Scan(&b.InstrumentID, &b.BarStart, &b.Open, &b.High, &b.Low, &b.Close, &b.Volume); err != nil {
			return nil, err
		}
		bars = append(bars, b)
	}
	return bars, rows.Err()
}
