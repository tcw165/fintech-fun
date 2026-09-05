package fold

import (
	"testing"

	"github.com/tcw165/fintech-fun/graph"
	"github.com/tcw165/fintech-fun/graph/examples"
)

func TestFoldNFLX(t *testing.T) {
	c, s, e := examples.NFLX()
	result := Fold(c, s, e, 10)
	if result.Company != "Netflix" || result.TickerNow != "NFLX" {
		t.Fatalf("%+v", result)
	}
	if result.QtyNow != 100 || result.CashReceived != 0 {
		t.Fatalf("qty=%v cash=%v", result.QtyNow, result.CashReceived)
	}
	if result.Series[0].Kind != graph.KindSplit || result.Series[0].QtyBefore != 10 || result.Series[0].QtyAfter != 100 {
		t.Fatalf("%+v", result.Series[0])
	}
}

func TestFoldMNTS(t *testing.T) {
	c, s, e := examples.MNTS()
	result := Fold(c, s, e, 140)
	if result.QtyNow != 10 || result.Series[0].Kind != graph.KindReverseSplit {
		t.Fatalf("%+v", result)
	}
}

func TestFoldAPGE(t *testing.T) {
	c, s, e := examples.APGE()
	result := Fold(c, s, e, 10)
	if result.Status != graph.StatusGone || result.QtyNow != 0 {
		t.Fatalf("%+v", result)
	}
	if result.CashReceived < 1351.099 || result.CashReceived > 1351.101 {
		t.Fatalf("cash %v", result.CashReceived)
	}
}

func TestFoldSquare(t *testing.T) {
	c, s, e := examples.Square()
	result := Fold(c, s, e, 10)
	if result.Company != "Block" || result.TickerNow != "XYZ" || result.QtyNow != 10 {
		t.Fatalf("%+v", result)
	}
}

func TestFoldLPSN(t *testing.T) {
	c, s, e := examples.LPSN()
	result := Fold(c, s, e, 100)
	if result.TickerNow != "LPSN" || result.Series[0].NowHolds != "SOUN" {
		t.Fatalf("%+v", result)
	}
	if result.QtyNow < 46.729 || result.QtyNow > 46.731 {
		t.Fatalf("qty %v", result.QtyNow)
	}
}

func TestFoldFTEL(t *testing.T) {
	c, s, e := examples.FTEL()
	result := Fold(c, s, e, 10)
	if result.Company != "GMEX Robotics" || result.TickerNow != "GMEX" {
		t.Fatalf("%+v", result)
	}
	want := []graph.EventKind{graph.KindReverseSplit, graph.KindReverseSplit, graph.KindNameChanged, graph.KindTickerChanged, graph.KindReverseSplit, graph.KindReverseSplit}
	if len(result.Series) != len(want) {
		t.Fatalf("series %v", result.Series)
	}
	for i, kind := range want {
		if result.Series[i].Kind != kind {
			t.Fatalf("row %d got %s want %s", i, result.Series[i].Kind, kind)
		}
	}
	wantQty := 10.0 / 16 / 8 / 7 / 9
	if result.QtyNow < wantQty-1e-12 || result.QtyNow > wantQty+1e-12 {
		t.Fatalf("qty %v want %v", result.QtyNow, wantQty)
	}
	if !result.Series[2].Date.Equal(graph.Date(2026, 3, 12)) {
		t.Fatalf("date %v", result.Series[2].Date)
	}
}
