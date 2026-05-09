package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ganesh/papertrading/services/go/md/internal/marketstatus"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var instrumentIDRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// Register mounts Phase 1 REST handlers that require Postgres (pool non-nil).
func Register(mux *http.ServeMux, pool *pgxpool.Pool) {
	mux.HandleFunc("GET /instruments", func(w http.ResponseWriter, r *http.Request) {
		handleInstruments(w, r, pool)
	})
	mux.HandleFunc("GET /candles", func(w http.ResponseWriter, r *http.Request) {
		handleCandles(w, r, pool)
	})
	mux.HandleFunc("GET /market/status", handleMarketStatus)
}

func handleMarketStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	now := time.Now()
	loc := marketstatus.ISTLocation()
	local := now.In(loc)
	session := marketstatus.NSEEQSession(now)
	asOf := local.Format(time.RFC3339Nano)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"segment": "NSE_EQ",
		"session": session,
		"asOf":    asOf,
		"weekday": local.Weekday().String(),
	})
}

func handleInstruments(w http.ResponseWriter, r *http.Request, pool *pgxpool.Pool) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if pool == nil {
		jsonErr(w, http.StatusServiceUnavailable, "instruments: DATABASE_URL not configured")
		return
	}
	ex := strings.TrimSpace(strings.ToUpper(r.URL.Query().Get("exchange")))
	if ex == "" {
		ex = "NSE"
	}
	q := strings.TrimSpace(r.URL.Query().Get("query"))

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	const base = `
SELECT instrument_id, tradingsymbol, exchange, segment, asset_class, instrument_type,
       lot_size, tick_size::float8, status
FROM ref.instruments
WHERE exchange = $1
`
	var rows pgx.Rows
	var err error
	if q != "" {
		pattern := "%" + q + "%"
		rows, err = pool.Query(ctx, base+` AND tradingsymbol ILIKE $2 ORDER BY tradingsymbol LIMIT 200`, ex, pattern)
	} else {
		rows, err = pool.Query(ctx, base+` ORDER BY tradingsymbol LIMIT 200`, ex)
	}
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type row struct {
		InstrumentID     string  `json:"instrumentId"`
		Tradingsymbol    string  `json:"tradingsymbol"`
		Exchange         string  `json:"exchange"`
		Segment          string  `json:"segment"`
		AssetClass       string  `json:"assetClass"`
		InstrumentType   string  `json:"instrumentType"`
		LotSize          int32   `json:"lotSize"`
		TickSize         float64 `json:"tickSize"`
		Status           string  `json:"status"`
	}
	out := make([]row, 0)
	for rows.Next() {
		var x row
		if err := rows.Scan(&x.InstrumentID, &x.Tradingsymbol, &x.Exchange, &x.Segment,
			&x.AssetClass, &x.InstrumentType, &x.LotSize, &x.TickSize, &x.Status); err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		jsonErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"instruments": out})
}

func caggRel(interval string) (string, error) {
	switch interval {
	case "1m":
		return "md.cagg_ticks_1m", nil
	case "5m":
		return "md.cagg_ticks_5m", nil
	case "15m":
		return "md.cagg_ticks_15m", nil
	case "1h":
		return "md.cagg_ticks_1h", nil
	case "1d":
		return "md.cagg_ticks_1d", nil
	default:
		return "", errors.New("invalid interval (use 1m, 5m, 15m, 1h, 1d)")
	}
}

func handleCandles(w http.ResponseWriter, r *http.Request, pool *pgxpool.Pool) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if pool == nil {
		jsonErr(w, http.StatusServiceUnavailable, "candles: DATABASE_URL not configured")
		return
	}
	q := r.URL.Query()
	inst := strings.TrimSpace(q.Get("instrument_id"))
	if inst == "" {
		jsonErr(w, http.StatusBadRequest, "missing instrument_id")
		return
	}
	if !instrumentIDRe.MatchString(inst) {
		jsonErr(w, http.StatusBadRequest, "invalid instrument_id")
		return
	}
	interval := strings.TrimSpace(q.Get("interval"))
	if interval == "" {
		interval = "1m"
	}
	rel, err := caggRel(interval)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, err.Error())
		return
	}
	fromS := strings.TrimSpace(q.Get("from"))
	toS := strings.TrimSpace(q.Get("to"))
	if fromS == "" || toS == "" {
		jsonErr(w, http.StatusBadRequest, "missing from or to (RFC3339)")
		return
	}
	fromT, err := time.Parse(time.RFC3339, fromS)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid from")
		return
	}
	toT, err := time.Parse(time.RFC3339, toS)
	if err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid to")
		return
	}
	if !toT.After(fromT) {
		jsonErr(w, http.StatusBadRequest, "to must be after from")
		return
	}
	limit := 5000
	if v := strings.TrimSpace(q.Get("limit")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 20000 {
			jsonErr(w, http.StatusBadRequest, "invalid limit (1–20000)")
			return
		}
		limit = n
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// rel is whitelisted above — safe to interpolate relation name only here.
	sql := `
SELECT bucket, open::float8, high::float8, low::float8, close::float8, volume::bigint
FROM ` + rel + `
WHERE instrument_id = $1 AND bucket >= $2 AND bucket < $3
ORDER BY bucket ASC
LIMIT $4
`
	rows, err := pool.Query(ctx, sql, inst, fromT.UTC(), toT.UTC(), limit)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type candle struct {
		Ts     string  `json:"ts"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume int64   `json:"volume"`
	}
	candles := make([]candle, 0)
	for rows.Next() {
		var b time.Time
		var o, h, l, c float64
		var vol int64
		if err := rows.Scan(&b, &o, &h, &l, &c, &vol); err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		candles = append(candles, candle{
			Ts:     b.UTC().Format(time.RFC3339Nano),
			Open:   o,
			High:   h,
			Low:    l,
			Close:  c,
			Volume: vol,
		})
	}
	if err := rows.Err(); err != nil {
		jsonErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"instrumentId": inst,
		"interval":     interval,
		"candles":      candles,
	})
}

func jsonErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
