package replay

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ganesh/papertrading/services/go/md/internal/ticksynth"
)

// Request is one replay run (HTTP body or REPLAY_* env).
type Request struct {
	Date        string   `json:"date"`
	Symbols     []string `json:"symbols"`
	Speed       float64  `json:"speed"`
	TicksPerBar int      `json:"ticksPerBar"`
	SessionID   string   `json:"sessionId"`
	TickSize    float64  `json:"tickSize"`
	SpreadTicks float64  `json:"spreadTicks"`
	SymbolsCSV  string   `json:"-"` // env comma-separated
}

func (r Request) ticksynthConfig() ticksynth.Config {
	return ticksynth.Config{
		TicksPerBar: r.TicksPerBar,
		TickSize:    r.TickSize,
		SpreadTicks: r.SpreadTicks,
	}
}

const (
	defaultSpeed       = 100.0
	defaultTicksPerBar = 10
	defaultSessionID   = "default"
)

func normalizedSymbols(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, raw := range in {
		s := strings.TrimSpace(strings.ToUpper(raw))
		if s == "" {
			continue
		}
		if !strings.HasSuffix(s, "-EQ") {
			s = s + "-EQ"
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func parseSymbolsCSV(s string) []string {
	var parts []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return normalizedSymbols(parts)
}

func parseDateIST(day string) (time.Time, time.Time, error) {
	day = strings.TrimSpace(day)
	if day == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("replay: empty date")
	}
	t, err := time.ParseInLocation("2006-01-02", day, istLoc())
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("replay: date %q: %w", day, err)
	}
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, istLoc())
	end := start.Add(24 * time.Hour)
	return start, end, nil
}

func istLoc() *time.Location {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return time.FixedZone("IST", 5*3600+1800)
	}
	return loc
}

// EnvConfig returns a replay request from REPLAY_DATE + REPLAY_SYMBOLS, or nil if unset.
func EnvConfig() (*Request, error) {
	date := strings.TrimSpace(os.Getenv("REPLAY_DATE"))
	syms := strings.TrimSpace(os.Getenv("REPLAY_SYMBOLS"))
	if date == "" || syms == "" {
		return nil, nil
	}
	req := &Request{
		Date:       date,
		SymbolsCSV: syms,
		Speed:      defaultSpeed,
		SessionID:  strings.TrimSpace(os.Getenv("REPLAY_SESSION_ID")),
		TicksPerBar: defaultTicksPerBar,
		TickSize:    0,
		SpreadTicks: 0,
	}
	if req.SessionID == "" {
		req.SessionID = defaultSessionID
	}
	if v := strings.TrimSpace(os.Getenv("REPLAY_SPEED")); v != "" {
		x, err := strconv.ParseFloat(v, 64)
		if err != nil || x <= 0 {
			return nil, fmt.Errorf("replay: invalid REPLAY_SPEED %q", v)
		}
		req.Speed = x
	}
	if v := strings.TrimSpace(os.Getenv("REPLAY_TICKS_PER_BAR")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 4 {
			return nil, fmt.Errorf("replay: invalid REPLAY_TICKS_PER_BAR %q (need >= 4)", v)
		}
		req.TicksPerBar = n
	}
	if v := strings.TrimSpace(os.Getenv("REPLAY_TICK_SIZE")); v != "" {
		x, err := strconv.ParseFloat(v, 64)
		if err != nil || x <= 0 {
			return nil, fmt.Errorf("replay: invalid REPLAY_TICK_SIZE %q", v)
		}
		req.TickSize = x
	}
	if v := strings.TrimSpace(os.Getenv("REPLAY_SPREAD_TICKS")); v != "" {
		x, err := strconv.ParseFloat(v, 64)
		if err != nil || x <= 0 {
			return nil, fmt.Errorf("replay: invalid REPLAY_SPREAD_TICKS %q", v)
		}
		req.SpreadTicks = x
	}
	req.Symbols = parseSymbolsCSV(syms)
	if len(req.Symbols) == 0 {
		return nil, fmt.Errorf("replay: REPLAY_SYMBOLS produced no symbols")
	}
	return req, nil
}

func mergeDefaults(r *Request) error {
	if r.Speed <= 0 {
		r.Speed = defaultSpeed
	}
	if r.TicksPerBar == 0 {
		r.TicksPerBar = defaultTicksPerBar
	}
	if strings.TrimSpace(r.SessionID) == "" {
		r.SessionID = defaultSessionID
	}
	r.Symbols = normalizedSymbols(r.Symbols)
	if len(r.Symbols) == 0 {
		return fmt.Errorf("replay: no symbols")
	}
	if strings.TrimSpace(r.Date) == "" {
		return fmt.Errorf("replay: no date")
	}
	return nil
}
