package examples

import "testing"

func TestAllHasTheSixNotionFixtures(t *testing.T) {
	all := All()
	if len(all) != 6 {
		t.Fatalf("got %d", len(all))
	}
	events := 0
	tickers := map[string]bool{}
	for _, fixture := range all {
		events += len(fixture.Events)
		tickers[fixture.Stock.Ticker] = true
	}
	if events != 11 || !tickers["NFLX"] || !tickers["XYZ"] || !tickers["GMEX"] {
		t.Fatalf("events=%d tickers=%v", events, tickers)
	}
}
