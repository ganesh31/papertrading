package marketstatus

import (
	"time"
)

// ISTLocation returns Asia/Kolkata or a fixed IST offset fallback.
func ISTLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return time.FixedZone("IST", 5*3600+1800)
	}
	return loc
}

// NSEEQSession returns PREOPEN / OPEN / POSTCLOSE / CLOSED for NSE equity cash.
// Weekends are CLOSED. Holidays are not loaded until Phase 1.11 (market-hours package).
func NSEEQSession(now time.Time) string {
	t := now.In(ISTLocation())
	wd := t.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return "CLOSED"
	}
	min := t.Hour()*60 + t.Minute()
	// Pre-open order collection ~09:00–09:08 IST (coarse).
	switch {
	case min >= 9*60 && min < 9*60+8:
		return "PREOPEN"
	case min >= 9*60+8 && min < 9*60+15:
		return "CLOSED"
	case min >= 9*60+15 && min < 15*60+30:
		return "OPEN"
	case min >= 15*60+30 && min < 16*60:
		return "POSTCLOSE"
	default:
		return "CLOSED"
	}
}
