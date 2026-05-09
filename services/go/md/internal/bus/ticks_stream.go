package bus

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ganesh/papertrading/services/go/md/internal/adapter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

const (
	defaultStreamName        = "ticks.v1"
	defaultRetention         = time.Hour
	defaultPublishTimeout    = 75 * time.Millisecond
	envTicksStream           = "MD_TICKS_STREAM"
	envTicksRetentionSeconds = "MD_TICKS_STREAM_RETENTION_SEC"
	envTicksPublishTimeoutMs = "MD_TICKS_PUBLISH_TIMEOUT_MS"
)

// TicksPublisherConfig configures Redis Streams publishing for normalized ticks.
type TicksPublisherConfig struct {
	StreamName      string
	Retention       time.Duration
	PublishTimeout  time.Duration
	PayloadFieldKey string // Redis stream entry field name (single JSON blob)
}

// DefaultTicksPublisherConfig returns defaults (override via env in ConfigFromEnv).
func DefaultTicksPublisherConfig() TicksPublisherConfig {
	return TicksPublisherConfig{
		StreamName:      defaultStreamName,
		Retention:       defaultRetention,
		PublishTimeout:  defaultPublishTimeout,
		PayloadFieldKey: "payload",
	}
}

// ConfigFromEnv overlays MD_* env vars onto defaults.
func ConfigFromEnv(base TicksPublisherConfig) TicksPublisherConfig {
	if v := os.Getenv(envTicksStream); v != "" {
		base.StreamName = v
	}
	if v := os.Getenv(envTicksRetentionSeconds); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			base.Retention = time.Duration(n) * time.Second
		}
	}
	if v := os.Getenv(envTicksPublishTimeoutMs); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			base.PublishTimeout = time.Duration(n) * time.Millisecond
		}
	}
	return base
}

// TicksPublisher pushes each canonical tick to a Redis Stream (firehose).
// Postgres remains durable storage; the stream is for cross-service fan-out.
type TicksPublisher struct {
	rdb *redis.Client
	cfg TicksPublisherConfig

	pubOK prometheus.Counter
	pubErr prometheus.Counter
}

// NewTicksPublisher registers Prometheus metrics and returns a publisher.
// rdb must be non-nil.
func NewTicksPublisher(rdb *redis.Client, cfg TicksPublisherConfig) *TicksPublisher {
	pubOK := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "md_ticks_stream_published_total",
		Help: "Ticks successfully XADDed to Redis Stream",
	})
	pubErr := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "md_ticks_stream_publish_errors_total",
		Help: "Failed Redis XADD attempts for ticks stream",
	})
	prometheus.MustRegister(pubOK, pubErr)

	return &TicksPublisher{
		rdb:    rdb,
		cfg:    cfg,
		pubOK:  pubOK,
		pubErr: pubErr,
	}
}

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

func wireFromTick(t adapter.Tick) tickWire {
	return tickWire{
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
	}
}

// Publish runs XADD with approximate MINID trim so entries older than Retention
// are dropped (Redis >= 6.2). Conflicts with MAXLEN in Redis; we only use MINID.
func (p *TicksPublisher) Publish(parent context.Context, t adapter.Tick) error {
	if p == nil || p.rdb == nil {
		return nil
	}
	payload, err := json.Marshal(wireFromTick(t))
	if err != nil {
		p.pubErr.Inc()
		return err
	}

	minMs := time.Now().Add(-p.cfg.Retention).UnixMilli()
	if minMs < 0 {
		minMs = 0
	}
	minID := fmt.Sprintf("%d-0", minMs)

	pubCtx, cancel := context.WithTimeout(parent, p.cfg.PublishTimeout)
	defer cancel()

	err = p.rdb.XAdd(pubCtx, &redis.XAddArgs{
		Stream: p.cfg.StreamName,
		MinID:  minID,
		Approx: true,
		Values: map[string]interface{}{
			p.cfg.PayloadFieldKey: string(payload),
		},
	}).Err()
	if err != nil {
		p.pubErr.Inc()
		return err
	}
	p.pubOK.Inc()
	return nil
}

// StreamName returns the configured Redis stream key (for health/doc introspection).
func (p *TicksPublisher) StreamName() string {
	if p == nil {
		return ""
	}
	return p.cfg.StreamName
}
