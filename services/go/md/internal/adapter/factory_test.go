package adapter

import (
	"context"
	"errors"
	"testing"
)

func TestKindFromEnv_Default(t *testing.T) {
	t.Setenv("MD_ADAPTER", "")
	k, err := KindFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if k != KindNSEReplay {
		t.Fatalf("want %s, got %s", KindNSEReplay, k)
	}
}

func TestKindFromEnv_AngelLive(t *testing.T) {
	t.Setenv("MD_ADAPTER", "angel_live")
	k, err := KindFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if k != KindAngelLive {
		t.Fatalf("want %s, got %s", KindAngelLive, k)
	}
}

func TestKindFromEnv_Unknown(t *testing.T) {
	t.Setenv("MD_ADAPTER", "bogus")
	_, err := KindFromEnv()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAngelLiveAdapter_Run_NotConfigured(t *testing.T) {
	var a AngelLiveAdapter
	ctx := context.Background()
	err := a.Run(ctx, nil)
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("want ErrNotConfigured, got %v", err)
	}
}

func TestNSEReplayAdapter_Run_ContextCancel(t *testing.T) {
	a, err := NewBroker(KindNSEReplay, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = a.Run(ctx, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}
