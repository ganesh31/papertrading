package marketstatus

import (
	"testing"
	"time"
)

func TestNSEEQSessionWeekend(t *testing.T) {
	// 2026-05-09 Saturday 10:00 IST
	ts := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	if g := NSEEQSession(ts, nil); g != "CLOSED" {
		t.Fatalf("Saturday: got %q want CLOSED", g)
	}
}

func TestNSEEQSessionHoliday(t *testing.T) {
	cal := &Calendar{nseEQ: map[string]struct{}{"2026-05-11": {}}}
	// Monday 2026-05-11 10:00 IST
	ts := time.Date(2026, 5, 11, 4, 30, 0, 0, time.UTC) // ~10:00 IST
	if g := NSEEQSession(ts, cal); g != "CLOSED" {
		t.Fatalf("holiday: got %q want CLOSED", g)
	}
}

func TestNSEEQSessionOpenWindow(t *testing.T) {
	loc := ISTLocation()
	// Monday 2026-05-18 10:00 IST — regular session
	openTS := time.Date(2026, 5, 18, 10, 0, 0, 0, loc)
	if g := NSEEQSession(openTS, nil); g != "OPEN" {
		t.Fatalf("Monday mid-session: got %q want OPEN", g)
	}
}
