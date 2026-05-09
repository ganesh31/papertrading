package stream

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/ganesh/papertrading/services/go/md/internal/adapter"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type HubConfig struct {
	PerClientBuffer int
	WriteTimeout    time.Duration
}

func DefaultConfig() HubConfig {
	return HubConfig{
		PerClientBuffer: 1024,
		WriteTimeout:    3 * time.Second,
	}
}

type Hub struct {
	pool *pgxpool.Pool
	cfg  HubConfig

	mu      sync.RWMutex
	clients map[*client]struct{}

	wsClients prometheus.Gauge
	drops     *prometheus.CounterVec
}

type client struct {
	conn *websocket.Conn
	ch   chan adapter.Tick

	mu   sync.RWMutex
	subs map[string]struct{} // instrument_id
}

func New(pool *pgxpool.Pool, cfg HubConfig) *Hub {
	wsClients := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "md_ws_clients_gauge",
		Help: "Connected md /stream websocket clients",
	})
	drops := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "md_ws_dropped_ticks_total",
		Help: "Ticks dropped due to per-client backpressure buffer saturation",
	}, []string{"reason"})

	prometheus.MustRegister(wsClients, drops)

	return &Hub{
		pool:      pool,
		cfg:       cfg,
		clients:   make(map[*client]struct{}),
		wsClients: wsClients,
		drops:     drops,
	}
}

type subscribeMsg struct {
	Subscribe []string `json:"subscribe"`
}

// HandleStream upgrades an HTTP connection and serves bidirectional stream:
// - client sends {"subscribe":["INFY","RELIANCE"]} (tradingsymbol or instrument_id)
// - server sends {"type":"tick","tick":{...}} messages
func (h *Hub) HandleStream(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		return
	}
	cl := &client{
		conn: c,
		ch:   make(chan adapter.Tick, h.cfg.PerClientBuffer),
		subs: make(map[string]struct{}),
	}

	h.mu.Lock()
	h.clients[cl] = struct{}{}
	h.mu.Unlock()
	h.wsClients.Inc()

	ctx := r.Context()
	go h.readLoop(ctx, cl)
	h.writeLoop(ctx, cl)

	h.mu.Lock()
	delete(h.clients, cl)
	h.mu.Unlock()
	h.wsClients.Dec()
	_ = c.Close(websocket.StatusNormalClosure, "bye")
}

func (h *Hub) readLoop(ctx context.Context, cl *client) {
	defer func() {
		_ = cl.conn.Close(websocket.StatusNormalClosure, "readLoop exit")
	}()
	for {
		_, b, err := cl.conn.Read(ctx)
		if err != nil {
			return
		}
		var msg subscribeMsg
		if err := json.Unmarshal(b, &msg); err != nil {
			continue
		}
		if len(msg.Subscribe) == 0 {
			continue
		}
		ids := h.resolveIDs(ctx, msg.Subscribe)
		if len(ids) == 0 {
			continue
		}
		cl.mu.Lock()
		for _, id := range ids {
			cl.subs[id] = struct{}{}
		}
		cl.mu.Unlock()
	}
}

func (h *Hub) writeLoop(ctx context.Context, cl *client) {
	type tickWire struct {
		InstrumentID string   `json:"instrumentId"`
		Ts           string   `json:"ts"`
		LTP          float64  `json:"ltp"`
		BidPx        *float64 `json:"bidPx,omitempty"`
		BidQty       *int     `json:"bidQty,omitempty"`
		AskPx        *float64 `json:"askPx,omitempty"`
		AskQty       *int     `json:"askQty,omitempty"`
		Volume       int64    `json:"volume"`
		OI           int64    `json:"oi"`
		Source       string   `json:"source"`
	}
	type env struct {
		Type string   `json:"type"`
		Tick tickWire `json:"tick"`
	}
	for {
		select {
		case <-ctx.Done():
			return
		case t, ok := <-cl.ch:
			if !ok {
				return
			}
			wire := env{
				Type: "tick",
				Tick: tickWire{
					InstrumentID: t.InstrumentID,
					Ts:           t.Ts.UTC().Format(time.RFC3339Nano),
					LTP:          t.LTP,
					BidPx:        t.BidPx,
					BidQty:       t.BidQty,
					AskPx:        t.AskPx,
					AskQty:       t.AskQty,
					Volume:       t.Volume,
					OI:           t.OI,
					Source:       t.Source,
				},
			}
			writeCtx, cancel := context.WithTimeout(ctx, h.cfg.WriteTimeout)
			_ = cl.conn.Write(writeCtx, websocket.MessageText, mustJSON(wire))
			cancel()
		}
	}
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

// Publish fans out ticks to subscribed clients with drop-oldest backpressure.
func (h *Hub) Publish(t adapter.Tick) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for cl := range h.clients {
		if !cl.isSubscribed(t.InstrumentID) {
			continue
		}
		// drop-oldest when full
		select {
		case cl.ch <- t:
		default:
			select {
			case <-cl.ch:
				h.drops.WithLabelValues("buffer_full").Inc()
			default:
			}
			select {
			case cl.ch <- t:
			default:
				h.drops.WithLabelValues("buffer_full").Inc()
			}
		}
	}
}

func (cl *client) isSubscribed(instrumentID string) bool {
	cl.mu.RLock()
	_, ok := cl.subs[instrumentID]
	cl.mu.RUnlock()
	return ok
}

func (h *Hub) resolveIDs(ctx context.Context, symbolsOrIDs []string) []string {
	out := make([]string, 0, len(symbolsOrIDs))
	seen := make(map[string]struct{}, len(symbolsOrIDs))
	for _, s := range symbolsOrIDs {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		// If it looks like an instrument_id, accept as-is.
		if strings.Contains(s, "|") || strings.Contains(s, ":") || len(s) > 12 {
			if _, ok := seen[s]; !ok {
				seen[s] = struct{}{}
				out = append(out, s)
			}
			continue
		}
		id, err := h.lookupInstrumentID(ctx, s)
		if err != nil || id == "" {
			continue
		}
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

func (h *Hub) lookupInstrumentID(ctx context.Context, sym string) (string, error) {
	if h.pool == nil {
		return "", errors.New("md stream: DATABASE_URL not configured")
	}
	// Support INFY / INFY-EQ / RELIANCE.
	candidates := []string{sym}
	if !strings.Contains(sym, "-") {
		candidates = append(candidates, sym+"-EQ")
	}
	const q = `
SELECT instrument_id
FROM ref.instruments
WHERE tradingsymbol = ANY($1)
ORDER BY
  CASE
    WHEN tradingsymbol LIKE '%-EQ' THEN 0
    ELSE 1
  END
LIMIT 1
`
	var id string
	err := h.pool.QueryRow(ctx, q, candidates).Scan(&id)
	return id, err
}

