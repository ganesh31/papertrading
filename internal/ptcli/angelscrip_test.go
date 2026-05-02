package ptcli

import (
	"strings"
	"testing"
)

func TestFilterNSEEquityCash(t *testing.T) {
	rows := []angelScripRow{
		{Symbol: "INFY-EQ", ExchSeg: "NSE"},
		{Symbol: "NIFTY24DEC25000CE", ExchSeg: "NFO"},
		{Symbol: "HDFCBANK-EQ", ExchSeg: "NSE"},
		{Symbol: "RELIANCE-EQ", ExchSeg: "BSE"},
	}
	got := filterNSEEquityCash(rows)
	if len(got) != 2 {
		t.Fatalf("want 2 rows, got %d: %+v", len(got), got)
	}
	if got[0].Symbol != "INFY-EQ" || got[1].Symbol != "HDFCBANK-EQ" {
		t.Fatalf("unexpected filter: %+v", got)
	}
}

func TestDecodeAngelScripMaster(t *testing.T) {
	const sample = `[{"token":"1","symbol":"INFY-EQ","name":"INFY","expiry":"","strike":"-1","lotsize":"1","instrumenttype":"","exch_seg":"NSE","tick_size":"5"}]`
	rows, err := decodeAngelScripMaster(strings.NewReader(sample))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Symbol != "INFY-EQ" {
		t.Fatalf("got %+v", rows)
	}
}
