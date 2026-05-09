package marketstatus

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

// Calendar holds NSE_EQ holiday calendar dates as YYYY-MM-DD strings in IST.
type Calendar struct {
	nseEQ map[string]struct{}
}

type holidaysFile struct {
	NSEQ []string `json:"nse_eq"`
}

// TryLoadCalendar reads path and builds a Calendar. Missing file → nil, nil.
func TryLoadCalendar(path string) (*Calendar, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var hf holidaysFile
	if err := json.Unmarshal(raw, &hf); err != nil {
		return nil, err
	}
	m := make(map[string]struct{}, len(hf.NSEQ))
	for _, d := range hf.NSEQ {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		m[d] = struct{}{}
	}
	return &Calendar{nseEQ: m}, nil
}

// IsNSEEQHoliday reports whether the IST calendar date of t is a configured holiday.
// HolidayCount returns how many NSE_EQ off-days are configured (for startup logs).
func (c *Calendar) HolidayCount() int {
	if c == nil {
		return 0
	}
	return len(c.nseEQ)
}

func (c *Calendar) IsNSEEQHoliday(t time.Time) bool {
	if c == nil || len(c.nseEQ) == 0 {
		return false
	}
	local := t.In(ISTLocation())
	key := local.Format("2006-01-02")
	_, ok := c.nseEQ[key]
	return ok
}
