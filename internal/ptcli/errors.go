package ptcli

import "errors"

// ErrInstrumentsSyncNotImplemented is replaced by real ingestion in P1-T03.
var ErrInstrumentsSyncNotImplemented = errors.New("pt instruments sync: not implemented yet (complete P1-T03)")
