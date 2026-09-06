package fold

import "testing"

func TestGoldFoldIsTheSixNotionRows(t *testing.T) {
	got := GoldFold()
	if len(got) != 6 {
		t.Fatalf("%d %+v", len(got), got)
	}
	if got[0].Q != "nflx" || got[5].Q != "ftel" {
		t.Fatalf("%+v", got)
	}
}

func TestMatchRowsSquare(t *testing.T) {
	rows := []map[string]any{{
		"company":       "Block",
		"ticker_now":    "XYZ",
		"qty_now":       10.0,
		"cash_received": 0.0,
		"series":        []any{map[string]any{"kind": "ticker_changed"}},
	}}
	check := MatchRows(rows, Expectation{Q: "square", Company: "Block", Ticker: "XYZ", QtyNow: 10, Series: 1})
	if !check.OK {
		t.Fatalf("%+v", check)
	}
}

func TestMatchRowsApgeCashAndEmptyIsUnresolved(t *testing.T) {
	rows := []map[string]any{{
		"company":       "Apogee",
		"ticker_now":    "APGE",
		"qty_now":       0,
		"cash_received": 1351.10,
	}}
	check := MatchRows(rows, Expectation{Q: "apge", Company: "Apogee", Ticker: "APGE", QtyNow: 0, Cash: 1351.10, Series: 1})
	if !check.OK {
		t.Fatalf("%+v", check)
	}
	if FoldShape(nil) || MatchRows(nil, Expectation{Q: "apge"}).Error != "unresolved" {
		t.Fatal("empty")
	}
}

func TestFoldShapeIgnoresRecordingStub(t *testing.T) {
	if FoldShape([]map[string]any{{"ok": true}}) {
		t.Fatal("recording stub is not a fold row")
	}
}
