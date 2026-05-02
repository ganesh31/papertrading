package ptcli

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
	"github.com/spf13/cobra"
)

const upsertInstrumentSQL = `
INSERT INTO ref.instruments (
	instrument_id, tradingsymbol, exchange, segment, asset_class, instrument_type,
	underlying_instrument_id, expiry, strike, option_type, lot_size, tick_size,
	freeze_qty, isin, listing_date, status, metadata
) VALUES (
	$1, $2, $3, $4, $5, $6,
	NULL, NULL, NULL, NULL,
	$7, $8,
	NULL, NULL, NULL,
	'ACTIVE', $9::jsonb
)
ON CONFLICT (exchange, segment, tradingsymbol) DO UPDATE SET
	lot_size = EXCLUDED.lot_size,
	tick_size = EXCLUDED.tick_size,
	metadata = COALESCE(ref.instruments.metadata, '{}'::jsonb) || EXCLUDED.metadata,
	status = EXCLUDED.status
`

func runInstrumentsSync(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	filePath, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}
	dsn, err := cmd.Flags().GetString("database-url")
	if err != nil {
		return err
	}
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is not set (or pass --database-url)")
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}
	if force {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "pt: --force is accepted but not yet used (reserved for TTL / skip-fresh logic)")
	}

	var body io.ReadCloser
	if filePath != "" {
		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("open scrip file: %w", err)
		}
		body = f
	} else {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, angelMasterURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "papertrading-pt/1.0 (+https://github.com/ganesh/papertrading)")
		client := &http.Client{Timeout: 3 * time.Minute}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("download scrip master: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			return fmt.Errorf("download scrip master: HTTP %s", resp.Status)
		}
		body = resp.Body
	}
	defer func() { _ = body.Close() }()

	all, err := decodeAngelScripMaster(body)
	if err != nil {
		return fmt.Errorf("parse scrip master JSON: %w", err)
	}

	filtered := filterNSEEquityCash(all)
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "pt instruments sync: loaded %d rows, %d NSE cash equities (-EQ)\n", len(all), len(filtered))

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("postgres ping: %w", err)
	}

	entropy := ulid.Monotonic(rand.Reader, 0)
	var skipped int
	const batchSize = 500
	for start := 0; start < len(filtered); start += batchSize {
		end := start + batchSize
		if end > len(filtered) {
			end = len(filtered)
		}
		batch := &pgx.Batch{}
		for _, row := range filtered[start:end] {
			args, err := buildUpsertArgs(row, entropy)
			if err != nil {
				skipped++
				continue
			}
			batch.Queue(upsertInstrumentSQL, args...)
		}
		if batch.Len() == 0 {
			continue
		}
		br := pool.SendBatch(ctx, batch)
		for i := 0; i < batch.Len(); i++ {
			_, err := br.Exec()
			if err != nil {
				_ = br.Close()
				return fmt.Errorf("upsert batch at offset %d: %w", start+i, err)
			}
		}
		if err := br.Close(); err != nil {
			return fmt.Errorf("close batch: %w", err)
		}
	}

	if skipped > 0 {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "pt instruments sync: skipped %d rows with invalid numeric fields\n", skipped)
	}
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "pt instruments sync: upsert complete (%d rows)\n", len(filtered)-skipped)
	return nil
}

func buildUpsertArgs(row angelScripRow, entropy io.Reader) ([]any, error) {
	lot, err := strconv.Atoi(stringsTrimSpaceDefault(row.LotSize, "1"))
	if err != nil || lot < 1 {
		return nil, fmt.Errorf("lot size")
	}
	tick, err := strconv.ParseFloat(stringsTrimSpaceDefault(row.TickSize, "0.05"), 64)
	if err != nil || tick <= 0 {
		return nil, fmt.Errorf("tick size")
	}

	meta, err := json.Marshal(map[string]any{
		"angel": map[string]string{
			"token":  stringsTrimSpaceDefault(row.Token, ""),
			"name":   stringsTrimSpaceDefault(row.Name, ""),
			"symbol": stringsTrimSpaceDefault(row.Symbol, ""),
		},
	})
	if err != nil {
		return nil, err
	}

	id, err := ulid.New(ulid.Timestamp(time.Now()), entropy)
	if err != nil {
		return nil, err
	}

	sym := stringsTrimSpaceDefault(row.Symbol, "")
	return []any{
		id.String(),
		sym,
		"NSE",
		"NSE_EQ",
		"EQUITY",
		"EQ",
		lot,
		tick,
		meta,
	}, nil
}

func stringsTrimSpaceDefault(s, def string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	return s
}
