package classify

import (
	"testing"

	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/parser"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/testdata"
	"github.com/tcw165/fintech-fun/graph"
)

func row(headline, ticker, company string) hood_events.TrackerRow {
	return hood_events.TrackerRow{Date: graph.Date(2026, 9, 4), Headline: headline, Company: company, Ticker: ticker}
}

func TestClassifyLPSN(t *testing.T) {
	events := ClassifyRow(row(
		"LivePerson (LPSN) performed a stock merger. Shareholders will receive 0.4673 new shares of SOUN for each old share of LPSN previously held.",
		"LPSN", "LivePerson",
	))
	if len(events) != 1 || events[0].Kind != "now_different_stock" || events[0].ShareMultiplier != 0.4673 || events[0].YouNowHold != "SOUN" || events[0].CanTrade {
		t.Fatalf("%+v", events)
	}
}

func TestClassifyAPGE(t *testing.T) {
	events := ClassifyRow(row(
		"Apogee Therapeutics, Inc. (APGE) performed a cash merger. This means that shares were removed and shareholders will receive $135.11 per share in cash.",
		"APGE", "",
	))
	if events[0].Kind != "cashed_out" || events[0].CashPerShare != 135.11 || events[0].ShareMultiplier != 0 {
		t.Fatalf("%+v", events)
	}
}

func TestClassifySplits(t *testing.T) {
	split := ClassifyRow(row("Amphenol (APH) performed a 2 for 1 Forward Split.", "APH", ""))[0]
	reverse := ClassifyRow(row("CollPlant Biotechnologies (CLGN) performed a 1 for 10 Reverse Split. Fractional shares resulting from the split will be retained.", "CLGN", ""))[0]
	if split.Kind != "split" || split.ShareMultiplier != 2 {
		t.Fatalf("%+v", split)
	}
	if reverse.Kind != "reverse_split" || reverse.ShareMultiplier != 10 || !reverse.KeepFractionals {
		t.Fatalf("%+v", reverse)
	}
}

func TestClassifyReversePlusTicker(t *testing.T) {
	events := ClassifyRow(row(
		"Nuburu, Inc. (BURU) performed a 1 for 40 Reverse Split, and changed its ticker to BURUD. Fractional shares resulting from the split will be paid as cash in lieu.",
		"BURU", "",
	))
	if len(events) != 2 || events[0].Kind != "reverse_split" || events[1].Kind != "ticker_changed" {
		t.Fatalf("%+v", events)
	}
	if events[0].ShareMultiplier != 40 || events[0].KeepFractionals || events[1].YouNowHold != "BURUD" {
		t.Fatalf("%+v", events)
	}
}

func TestClassifyTickerAndName(t *testing.T) {
	events := ClassifyRow(row(
		"Bit Origin Limited Class A (BTOG) performed a ticker change to SGRX, and changed its corporate name to Sangrix.",
		"BTOG", "",
	))
	if len(events) != 2 || events[0].Kind != "ticker_changed" || events[1].Kind != "name_changed" {
		t.Fatalf("%+v", events)
	}
	if events[0].YouNowHold != "SGRX" || events[1].NewName != "Sangrix" {
		t.Fatalf("%+v", events)
	}
}

func TestClassifyWorthlessWaitingSpinoff(t *testing.T) {
	worthless := ClassifyRow(row("Arqit Quantum Inc. Warrants (ARQQW) expired worthless.", "ARQQW", ""))[0]
	waiting := ClassifyRow(row("Natuzzi (NTZ) was delisted pending a stock liquidation. Details are still pending.", "NTZ", ""))[0]
	extra := ClassifyRow(row(
		"KLX Energy (KLXE) performed a spinoff of KLX Energy Services Holdings, Inc. Rights (KLXER). For every 1 share of KLXE held, shareholders will receive 1 share of KLXER.",
		"KLXE", "",
	))[0]
	if worthless.Kind != "worthless" || waiting.Kind != "waiting" || extra.Kind != "extra_stock" || extra.YouNowHold != "KLXER" || extra.ShareMultiplier != 1 {
		t.Fatalf("%+v %+v %+v", worthless, waiting, extra)
	}
}

func TestClassifyDropsCUSIPOnly(t *testing.T) {
	events := ClassifyRow(row("Kurv Yield Premium Strategy Netflix (NFLX) ETF (NFLP) performed a CUSIP change.", "NFLP", ""))
	if !events[0].Skip {
		t.Fatalf("%+v", events)
	}
}

func TestClassifyFixture(t *testing.T) {
	classified := ClassifyRows(parser.ParseTracker(testdata.TrackerSept2026))
	kinds := map[string]bool{}
	for _, event := range classified {
		if !event.Skip {
			kinds[event.Ticker+"|"+event.Kind] = true
		}
	}
	for _, key := range []string{"LPSN|now_different_stock", "APGE|cashed_out", "CLGN|reverse_split", "APH|split", "BTOG|ticker_changed", "BTOG|name_changed"} {
		if !kinds[key] {
			t.Fatalf("missing %s in %v", key, kinds)
		}
	}
}

func TestClassifyLiveMustHits(t *testing.T) {
	cases := []struct {
		headline string
		kind     string
		x        float64
		hold     string
	}{
		{"Netflix, Inc. (NFLX) performed a forward split. This means that customers will now have 10 shares for every 1 share of NFLX previously held.", "split", 10, ""},
		{"Momentus Inc. (MNTS) performed a 1 for 14 Reverse Split. This means shareholders will now hold 1 share of MNTS for every 14 shares of MNTS previously held. Fractional shares from the reverse split will be paid as cash in lieu.", "reverse_split", 14, ""},
		{"Apogee Therapeutics, Inc. (APGE) performed a cash merger. This means that shares were removed and shareholders will receive $135.11 per share in cash.", "cashed_out", 0, ""},
		{"LivePerson (LPSN) performed a stock merger. Shareholders will receive 0.4673 new shares of SOUN for each old share of LPSN previously held.", "now_different_stock", 0.4673, "SOUN"},
		{"Block, Inc. (SQ) performed a symbol change to XYZ.", "ticker_changed", 1, "XYZ"},
	}
	for _, tc := range cases {
		events := ClassifyRow(row(tc.headline, "", ""))
		if len(events) == 0 || events[0].Kind != tc.kind || events[0].ShareMultiplier != tc.x || events[0].YouNowHold != tc.hold {
			t.Fatalf("%s → %+v", tc.kind, events)
		}
	}
	ftel := ClassifyRow(row("Fitell Corporation Class A Ordinary Shares (FTEL) performed a ticker change to GMEX, and changed its corporate name to GMEX Robotics.", "FTEL", "Fitell"))
	if len(ftel) != 2 || ftel[0].Kind != "ticker_changed" || ftel[0].YouNowHold != "GMEX" || ftel[1].Kind != "name_changed" || ftel[1].NewName != "GMEX Robotics" {
		t.Fatalf("%+v", ftel)
	}
}

func TestGoldReportCountsSkips(t *testing.T) {
	rows := parser.ParseTracker(testdata.TrackerSept2026)
	rows = append(rows, row("Fitell Corporation (FTEL) performed a 1:1 CUSIP change.", "FTEL", "Fitell"))
	report := Report(rows)
	if report.Written == 0 || report.SkipWhy["cusip"] != 1 {
		t.Fatalf("%+v", report)
	}
}

func TestClassifyHeadline(t *testing.T) {
	result, err := ClassifyHeadline(
		"Apogee Therapeutics, Inc. (APGE) performed a cash merger. Shareholders will receive $135.11 per share in cash.",
		"2026-09-03",
		"Apogee Therapeutics, Inc.",
		"APGE",
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Events[0].Kind != "cashed_out" || result.Events[0].CashPerShare != 135.11 {
		t.Fatalf("%+v", result)
	}
}
