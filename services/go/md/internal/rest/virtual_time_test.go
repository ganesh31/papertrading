package rest

import (
	"testing"
	"time"
)

func TestParseVirtualTime_QueryPlusCorruptedToSpace(t *testing.T) {
	// Simulates url.Values.Get after ParseQuery turns + into space in offset.
	raw := "2026-04-24T09:45:00 05:30"
	got, err := parseVirtualTime(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 4, 24, 9, 45, 0, 0, time.FixedZone("IST", 5*3600+1800))
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParseVirtualTime_EncodedPlus(t *testing.T) {
	got, err := parseVirtualTime("2026-04-24T09:45:00+05:30")
	if err != nil {
		t.Fatal(err)
	}
	if got.Year() != 2026 || got.Month() != 4 || got.Day() != 24 {
		t.Fatalf("got %v", got)
	}
}
