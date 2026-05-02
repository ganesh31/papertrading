package replay

import (
	"reflect"
	"testing"
)

func TestNormalizedSymbols(t *testing.T) {
	got := normalizedSymbols([]string{" infy ", "RELIANCE-EQ", "tcs"})
	want := []string{"INFY-EQ", "RELIANCE-EQ", "TCS-EQ"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}
