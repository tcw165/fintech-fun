package graph

import (
	"testing"
)

func TestEventKindClosedSet(t *testing.T) {
	want := map[EventKind]bool{
		"ticker_changed": true, "name_changed": true, "split": true,
		"reverse_split": true, "now_different_stock": true, "cashed_out": true,
		"extra_stock": true, "worthless": true, "waiting": true,
	}
	if len(AllEventKinds()) != len(want) {
		t.Fatalf("got %d kinds", len(AllEventKinds()))
	}
	for _, kind := range AllEventKinds() {
		if !want[kind] {
			t.Fatalf("unexpected kind %q", kind)
		}
	}
}

func TestStockStatusClosedSet(t *testing.T) {
	for _, status := range []StockStatus{StatusTradeable, StatusOTC, StatusGone, StatusWorthless} {
		if status == "" {
			t.Fatal("empty status")
		}
	}
}

func TestEventIDIsTickerDateKind(t *testing.T) {
	event := Event{
		Date:            Date(2025, 11, 17),
		Kind:            KindSplit,
		Headline:        "Netflix 10-for-1",
		HappenedTo:      "NFLX",
		ShareMultiplier: 10,
	}
	if event.ID() != "NFLX|2025-11-17|split" {
		t.Fatalf("got %q", event.ID())
	}
}

func TestOnlyReverseSplitDivides(t *testing.T) {
	for _, kind := range AllEventKinds() {
		if kind == KindReverseSplit {
			if Multiplies(kind) {
				t.Fatal("reverse_split should divide")
			}
			continue
		}
		if !Multiplies(kind) {
			t.Fatalf("%s should multiply", kind)
		}
	}
}

func TestSplitMultiplies(t *testing.T) {
	event := Event{Kind: KindSplit, ShareMultiplier: 10}
	if QtyAfter(10, event) != 100 {
		t.Fatalf("got %v", QtyAfter(10, event))
	}
	if EventCash(10, event) != 0 {
		t.Fatalf("got cash %v", EventCash(10, event))
	}
}

func TestReverseSplitDivides(t *testing.T) {
	event := Event{Kind: KindReverseSplit, ShareMultiplier: 14}
	if QtyAfter(140, event) != 10 {
		t.Fatalf("got %v", QtyAfter(140, event))
	}
}

func TestCashedOutZerosQtyAndPaysCash(t *testing.T) {
	event := Event{Kind: KindCashedOut, ShareMultiplier: 0, CashPerShare: 135.11, CanTrade: false}
	if QtyAfter(10, event) != 0 {
		t.Fatalf("qty %v", QtyAfter(10, event))
	}
	if got := EventCash(10, event); got < 1351.099 || got > 1351.101 {
		t.Fatalf("cash %v", got)
	}
}

