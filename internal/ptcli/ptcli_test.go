package ptcli

import (
	"strings"
	"testing"
)

func TestVersionLine(t *testing.T) {
	s := versionLine()
	if !strings.HasPrefix(s, "pt ") {
		t.Fatalf("unexpected version line: %q", s)
	}
}
