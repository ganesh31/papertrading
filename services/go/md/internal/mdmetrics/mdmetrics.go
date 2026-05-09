package mdmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Central Prometheus metrics for md (Phase 1 Metrics section). Registered once via promauto.

var (
	TicksIngested = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "md_ticks_ingested_total",
		Help: "Normalized ticks entering persist/bus/WS after the normalizer",
	}, []string{"adapter"})

	TicksStaleDropped = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "md_ticks_stale_dropped_total",
		Help: "LIVE ticks dropped because tick timestamp exceeded LiveStaleness vs wall clock",
	}, []string{"adapter"})

	TickStalenessSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "md_tick_staleness_seconds",
		Help:    "Wall clock minus LIVE tick timestamp for ticks that passed the staleness gate",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 15, 30, 60, 120},
	}, []string{"adapter"})

	AdapterReconnects = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "md_adapter_reconnects_total",
		Help: "Broker-adapter reconnect attempts (meaningful for angel_live in Phase 11)",
	}, []string{"adapter"})

	PersistBatchSize = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "md_persist_batch_size",
		Help:    "Number of ticks flushed per Postgres batch insert",
		Buckets: prometheus.ExponentialBuckets(1, 2, 14),
	})

	ReplayRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "replay_running",
		Help: "1 while an nse_replay session is active, else 0",
	})

	ReplayVirtualTimestampSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "replay_virtual_timestamp_seconds",
		Help: "Unix seconds for replay virtual clock (last emitted tick); 0 when idle",
	})

	ReplaySpeedMultiplier = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "replay_speed_multiplier",
		Help: "Configured replay speed multiplier while running; 0 when idle",
	})

	ReplayPendingTicks = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "replay_pending_ticks",
		Help: "Synthetic ticks not yet emitted in the current replay run",
	})

	ReplaySessionBarRows = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "replay_session_bar_rows",
		Help: "md.bars_1m rows loaded for the active replay session (0 when idle)",
	})
)
