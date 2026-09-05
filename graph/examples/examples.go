// Package examples is Notion fixture data. Child of //graph (models).
package examples

import "github.com/tcw165/fintech-fun/graph"

func NFLX() (graph.Company, graph.Stock, []graph.Event) {
	return graph.Company{Name: "Netflix"},
		graph.Stock{Ticker: "NFLX", Status: graph.StatusTradeable},
		[]graph.Event{{
			Date: graph.Date(2025, 11, 17), Kind: graph.KindSplit, Headline: "Netflix 10-for-1",
			HappenedTo: "NFLX", ShareMultiplier: 10, KeepFractionals: true, CanTrade: true,
		}}
}

func MNTS() (graph.Company, graph.Stock, []graph.Event) {
	return graph.Company{Name: "Momentus"},
		graph.Stock{Ticker: "MNTS", Status: graph.StatusTradeable},
		[]graph.Event{{
			Date: graph.Date(2024, 8, 1), Kind: graph.KindReverseSplit, Headline: "Momentus (MNTS) 1-for-14",
			HappenedTo: "MNTS", ShareMultiplier: 14, KeepFractionals: false, CanTrade: true,
		}}
}

func APGE() (graph.Company, graph.Stock, []graph.Event) {
	return graph.Company{Name: "Apogee"},
		graph.Stock{Ticker: "APGE", Status: graph.StatusGone},
		[]graph.Event{{
			Date: graph.Date(2026, 9, 3), Kind: graph.KindCashedOut, Headline: "Apogee (APGE) cash merger 135.11",
			HappenedTo: "APGE", ShareMultiplier: 0, CashPerShare: 135.11, KeepFractionals: true, CanTrade: false,
		}}
}

func Square() (graph.Company, graph.Stock, []graph.Event) {
	return graph.Company{Name: "Block", AlsoKnownAs: []string{"Square"}},
		graph.Stock{Ticker: "XYZ", FormerTickers: []string{"SQ"}, Status: graph.StatusTradeable},
		[]graph.Event{{
			Date: graph.Date(2025, 1, 21), Kind: graph.KindTickerChanged, Headline: "Block (SQ) → XYZ",
			HappenedTo: "XYZ", ShareMultiplier: 1, KeepFractionals: true, CanTrade: true,
		}}
}

func LPSN() (graph.Company, graph.Stock, []graph.Event) {
	return graph.Company{Name: "LivePerson"},
		graph.Stock{Ticker: "LPSN", Status: graph.StatusGone},
		[]graph.Event{{
			Date: graph.Date(2026, 9, 4), Kind: graph.KindNowDifferentStock, Headline: "LivePerson (LPSN) → 0.4673 SOUN",
			HappenedTo: "LPSN", ShareMultiplier: 0.4673, YouNowHold: "SOUN", KeepFractionals: true, CanTrade: false,
		}}
}

func FTEL() (graph.Company, graph.Stock, []graph.Event) {
	return graph.Company{Name: "GMEX Robotics", AlsoKnownAs: []string{"Fitell"}},
		graph.Stock{Ticker: "GMEX", FormerTickers: []string{"FTEL"}, Status: graph.StatusTradeable},
		[]graph.Event{
			{Date: graph.Date(2025, 9, 23), Kind: graph.KindReverseSplit, Headline: "Fitell (FTEL) 1-for-16", HappenedTo: "GMEX", ShareMultiplier: 16, KeepFractionals: true, CanTrade: true},
			{Date: graph.Date(2026, 1, 8), Kind: graph.KindReverseSplit, Headline: "Fitell (FTEL) 1-for-8", HappenedTo: "GMEX", ShareMultiplier: 8, KeepFractionals: true, CanTrade: true},
			{Date: graph.Date(2026, 3, 12), Kind: graph.KindTickerChanged, Headline: "FTEL → GMEX", HappenedTo: "GMEX", ShareMultiplier: 1, KeepFractionals: true, CanTrade: true},
			{Date: graph.Date(2026, 3, 12), Kind: graph.KindNameChanged, Headline: "Fitell renamed to GMEX Robotics", HappenedTo: "GMEX", ShareMultiplier: 1, KeepFractionals: true, CanTrade: true},
			{Date: graph.Date(2026, 5, 1), Kind: graph.KindReverseSplit, Headline: "GMEX Robotics 1-for-7", HappenedTo: "GMEX", ShareMultiplier: 7, KeepFractionals: true, CanTrade: true},
			{Date: graph.Date(2026, 7, 2), Kind: graph.KindReverseSplit, Headline: "GMEX Robotics 1-for-9", HappenedTo: "GMEX", ShareMultiplier: 9, KeepFractionals: true, CanTrade: true},
		}
}
