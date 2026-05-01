package adapter

import (
	"fmt"
	"os"
	"strings"
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

// NewFromEnv constructs the adapter selected by MD_ADAPTER.
func NewFromEnv() (BrokerAdapter, error) {
	k, err := KindFromEnv()
	if err != nil {
		return nil, err
	}
	switch k {
	case KindNSEReplay:
		return NSEReplayAdapter{}, nil
	case KindAngelLive:
		return AngelLiveAdapter{}, nil
	default:
		return nil, fmt.Errorf("md: unsupported adapter %q", k)
	}
}
