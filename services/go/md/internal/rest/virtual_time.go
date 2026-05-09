package rest

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Go's url.ParseQuery treats '+' as space in query values, so clients sending
// virtualTime=2026-04-24T09:45:00+05:30 often arrive as "...T09:45:00 05:30".
// Repair that offset pattern before time.Parse.
var virtualTimeSpaceBeforeOffset = regexp.MustCompile(
	`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(?::\d{2}(?:\.\d+)?)?) (\d{2}:\d{2})$`,
)

func repairVirtualTimeFromQuery(s string) string {
	s = strings.TrimSpace(s)
	if m := virtualTimeSpaceBeforeOffset.FindStringSubmatch(s); len(m) == 3 {
		return m[1] + "+" + m[2]
	}
	return s
}

func parseVirtualTime(v string) (time.Time, error) {
	v = repairVirtualTimeFromQuery(strings.TrimSpace(v))
	if v == "" {
		return time.Time{}, fmt.Errorf("empty virtualTime")
	}
	t, err := time.Parse(time.RFC3339Nano, v)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse(time.RFC3339, v)
	if err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid RFC3339: %w", err)
}
