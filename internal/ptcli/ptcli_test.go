package ptcli

import (
	"errors"
	"strings"
	"testing"
)

func TestVersionLine(t *testing.T) {
	s := versionLine()
	if !strings.HasPrefix(s, "pt ") {
		t.Fatalf("unexpected version line: %q", s)
	}
}

func TestInstrumentsSync_NotImplemented(t *testing.T) {
	rootCmd.SetArgs([]string{"instruments", "sync"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	err := rootCmd.Execute()
	if !errors.Is(err, ErrInstrumentsSyncNotImplemented) {
		t.Fatalf("want ErrInstrumentsSyncNotImplemented, got %v", err)
	}
}
