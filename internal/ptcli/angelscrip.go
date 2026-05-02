package ptcli

import (
	"encoding/json"
	"io"
	"strings"
)

const angelMasterURL = "https://margincalculator.angelbroking.com/OpenAPI_File/files/OpenAPIScripMaster.json"

// angelScripRow mirrors Angel OpenAPIScripMaster.json entries (unknown fields ignored).
type angelScripRow struct {
	Token          string `json:"token"`
	Symbol         string `json:"symbol"`
	Name           string `json:"name"`
	Expiry         string `json:"expiry"`
	Strike         string `json:"strike"`
	LotSize        string `json:"lotsize"`
	InstrumentType string `json:"instrumenttype"`
	ExchSeg        string `json:"exch_seg"`
	TickSize       string `json:"tick_size"`
}

func decodeAngelScripMaster(r io.Reader) ([]angelScripRow, error) {
	dec := json.NewDecoder(r)
	var rows []angelScripRow
	if err := dec.Decode(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// filterNSEEquityCash keeps NSE cash equities (Angel convention: exch_seg NSE + symbol suffix -EQ).
func filterNSEEquityCash(rows []angelScripRow) []angelScripRow {
	out := make([]angelScripRow, 0, len(rows)/8)
	for _, row := range rows {
		if isNSEEquityCash(row) {
			out = append(out, row)
		}
	}
	return out
}

func isNSEEquityCash(r angelScripRow) bool {
	if strings.ToUpper(strings.TrimSpace(r.ExchSeg)) != "NSE" {
		return false
	}
	sym := strings.TrimSpace(r.Symbol)
	return strings.HasSuffix(sym, "-EQ")
}
