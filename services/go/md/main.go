package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ganesh/papertrading/services/go/md/internal/adapter"
	"github.com/ganesh/papertrading/services/go/md/internal/bus"
	"github.com/ganesh/papertrading/services/go/md/internal/marketstatus"
	"github.com/ganesh/papertrading/services/go/md/internal/mdmetrics"
	"github.com/ganesh/papertrading/services/go/md/internal/normalize"
	"github.com/ganesh/papertrading/services/go/md/internal/persist"
	"github.com/ganesh/papertrading/services/go/md/internal/replay"
	"github.com/ganesh/papertrading/services/go/md/internal/rest"
	"github.com/ganesh/papertrading/services/go/md/internal/stream"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func initTracer(ctx context.Context) (shutdown func(context.Context) error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4318"
	}

	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(endpoint+"/v1/traces"),
	)
	if err != nil {
		log.Fatalf("otel exporter init: %v", err)
	}

	svcName := os.Getenv("OTEL_SERVICE_NAME")
	if svcName == "" {
		svcName = "md"
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(svcName),
		),
	)
	if err != nil {
		log.Fatalf("otel resource: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exp),
	)
	otel.SetTracerProvider(tp)

	return tp.Shutdown
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownTracer := initTracer(ctx)
	defer func() { _ = shutdownTracer(context.Background()) }()

	kind, err := adapter.KindFromEnv()
	if err != nil {
		log.Fatalf("broker adapter kind: %v", err)
	}

	adapterCtx, stopAdapter := context.WithCancel(ctx)
	defer stopAdapter()

	dsn := os.Getenv("DATABASE_URL")
	var pool *pgxpool.Pool
	if dsn != "" {
		pool, err = pgxpool.New(ctx, dsn)
		if err != nil {
			log.Printf("md: DATABASE_URL pool open failed (replay DB idle): %v", err)
			pool = nil
		} else {
			defer pool.Close()
		}
	}

	var coord *replay.Coordinator
	if kind == adapter.KindNSEReplay {
		coord = replay.NewCoordinator(adapterCtx, pool)
	}

	broker, err := adapter.NewBroker(kind, coord)
	if err != nil {
		log.Fatalf("broker adapter: %v", err)
	}
	log.Printf("md broker adapter: %s", broker.Kind())
	brokerKind := string(broker.Kind())

	var rdb *redis.Client
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		opt, parseErr := redis.ParseURL(redisURL)
		if parseErr != nil {
			log.Printf("md: REDIS_URL invalid (%v); instrument cache will use Postgres + in-process TTL only", parseErr)
		} else {
			rdb = redis.NewClient(opt)
			defer func() { _ = rdb.Close() }()
		}
	}

	holidayCal := loadHolidayCalendar()

	nrmCfg := normalize.DefaultConfig()
	nrmCfg.AdapterKind = string(kind)
	norm := normalize.New(pool, rdb, nrmCfg)

	hub := stream.New(pool, stream.DefaultConfig())

	var tickPub *bus.TicksPublisher
	if rdb != nil {
		tickPub = bus.NewTicksPublisher(rdb, bus.ConfigFromEnv(bus.DefaultTicksPublisherConfig()))
	}

	var tickBatcher *persist.Batcher
	if pool != nil {
		tickBatcher = persist.NewBatcher(pool, persist.DefaultConfig())
		go tickBatcher.Run(adapterCtx)
	}

	runHooks := normalize.WrapWithNormalizer(&adapter.RunHooks{
		OnNormalizedTick: func(ctx context.Context, t adapter.Tick) error {
			mdmetrics.TicksIngested.WithLabelValues(brokerKind).Inc()
			hub.Publish(t)
			if tickPub != nil {
				_ = tickPub.Publish(ctx, t)
			}
			if tickBatcher != nil {
				return tickBatcher.Enqueue(ctx, t)
			}
			return nil
		},
	}, norm)

	go func() {
		err := broker.Run(adapterCtx, runHooks)
		switch {
		case err == nil:
		case errors.Is(err, context.Canceled):
		case errors.Is(err, adapter.ErrNotConfigured):
			log.Printf("broker adapter: %v", err)
		default:
			log.Printf("broker adapter stopped: %v", err)
		}
	}()

	port := 6011
	if v := os.Getenv("MD_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		h := map[string]any{
			"ok":             true,
			"service":        "md",
			"broker_adapter": brokerKind,
		}
		if coord != nil && pool != nil {
			h["replay_db"] = true
		}
		if rdb != nil {
			h["redis_instrument_cache"] = true
		}
		if tickPub != nil {
			h["redis_ticks_stream"] = tickPub.StreamName()
		}
		_ = json.NewEncoder(w).Encode(h)
	})

	if coord != nil {
		coord.RegisterHTTP(mux)
	}

	rest.Register(mux, pool, holidayCal)

	mux.HandleFunc("GET /stream", hub.HandleStream)

	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           otelhttp.NewHandler(mux, "http"),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("md listening on :%d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
}

func loadHolidayCalendar() *marketstatus.Calendar {
	seen := make(map[string]struct{})
	paths := make([]string, 0, 4)
	if v := strings.TrimSpace(os.Getenv("MARKET_HOLIDAYS_PATH")); v != "" {
		seen[v] = struct{}{}
		paths = append(paths, v)
	}
	for _, p := range []string{"/etc/papertrading/holidays.json", "infra/seed/holidays.json"} {
		if _, dup := seen[p]; dup {
			continue
		}
		seen[p] = struct{}{}
		paths = append(paths, p)
	}
	for _, p := range paths {
		cal, err := marketstatus.TryLoadCalendar(p)
		if err != nil {
			log.Printf("md: holidays file %q: %v", p, err)
			continue
		}
		if cal != nil {
			log.Printf("md: NSE_EQ holidays loaded from %q (%d days)", p, cal.HolidayCount())
			return cal
		}
	}
	log.Printf("md: no holidays file found (optional MARKET_HOLIDAYS_PATH / infra/seed/holidays.json)")
	return nil
}
