package adapter

import (
	"fmt"
	"os"
	"strings"

	"github.com/ganesh/papertrading/services/go/md/internal/replay"
)

// KindFromEnv returns KindNSEReplay when MD_ADAPTER is unset or empty.
func KindFromEnv() (Kind, error) {
	v := strings.TrimSpace(os.Getenv("MD_ADAPTER"))
	if v == "" {
		return KindNSEReplay, nil
	}
	switch strings.ToLower(v) {
	case "nse_replay":
		return KindNSEReplay, nil
	case "angel_live":
		return KindAngelLive, nil
	default:
		return "", fmt.Errorf("md: unknown MD_ADAPTER %q (want nse_replay or angel_live)", v)
	}
}

// NewBroker returns the adapter for kind; coord is used only for KindNSEReplay (may be nil).
func NewBroker(kind Kind, coord *replay.Coordinator) (BrokerAdapter, error) {
	switch kind {
	case KindNSEReplay:
		return NSEReplayAdapter{Coord: coord}, nil
	case KindAngelLive:
		return AngelLiveAdapter{}, nil
	default:
		return nil, fmt.Errorf("md: unsupported adapter %q", kind)
	}
}

// NewFromEnv constructs the adapter selected by MD_ADAPTER (replay coordinator must be wired via NewBroker when DB-backed replay is needed).
func NewFromEnv() (BrokerAdapter, error) {
	k, err := KindFromEnv()
	if err != nil {
		return nil, err
	}
	return NewBroker(k, nil)
}
