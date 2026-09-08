package parser

import (
	"strings"
	"testing"

	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/testdata"
	"github.com/tcw165/fintech-fun/graph"
)

func TestParseDateHeading(t *testing.T) {
	got, ok := ParseDateHeading("## September 4, 2026")
	if !ok || !got.Equal(graph.Date(2026, 9, 4)) {
		t.Fatalf("%v %v", got, ok)
	}
	got, ok = ParseDateHeading("November 17, 2025")
	if !ok || !got.Equal(graph.Date(2025, 11, 17)) {
		t.Fatalf("%v %v", got, ok)
	}
	if _, ok := ParseDateHeading("not a date"); ok {
		t.Fatal("expected miss")
	}
}

func TestParseCompanyTicker(t *testing.T) {
	company, ticker := ParseCompanyTicker("LivePerson (LPSN) performed a stock merger.")
	if company != "LivePerson" || ticker != "LPSN" {
		t.Fatalf("%q %q", company, ticker)
	}
	_, ticker = ParseCompanyTicker("Arqit Quantum Inc. Warrants (ARQQW) expired worthless.")
	if ticker != "ARQQW" {
		t.Fatalf("%q", ticker)
	}
	company, ticker = ParseCompanyTicker("No ticker here.")
	if company != "" || ticker != "" {
		t.Fatal("expected empty")
	}
}

func TestParseFixtureDedupesDays(t *testing.T) {
	rows := ParseTracker(testdata.TrackerSept2026)
	var sept4, sept3 int
	tickers := map[string]bool{}
	for _, row := range rows {
		if row.Date.Equal(graph.Date(2026, 9, 4)) {
			sept4++
			tickers[row.Ticker] = true
		}
		if row.Date.Equal(graph.Date(2026, 9, 3)) {
			sept3++
		}
	}
	if sept4 != 8 || sept3 != 5 {
		t.Fatalf("sept4=%d sept3=%d total=%d", sept4, sept3, len(rows))
	}
	for _, ticker := range []string{"NHPAP", "CLGN", "TANH", "JFB", "LPSN", "BTMCQ", "BTOG"} {
		if !tickers[ticker] {
			t.Fatalf("missing %s", ticker)
		}
	}
}

func TestParseFixtureKeepsLPSNAndAPGE(t *testing.T) {
	rows := ParseTracker(testdata.TrackerSept2026)
	var lpsn_ticker, apge_headline string
	var lpsn_date, apge_date bool
	for _, row := range rows {
		if row.Ticker == "LPSN" {
			lpsn_ticker = row.Ticker
			lpsn_date = row.Date.Equal(graph.Date(2026, 9, 4))
			if !strings.Contains(row.Headline, "0.4673") || !strings.Contains(row.Headline, "SOUN") {
				t.Fatalf("%+v", row)
			}
		}
		if row.Ticker == "APGE" {
			apge_headline = row.Headline
			apge_date = row.Date.Equal(graph.Date(2026, 9, 3))
		}
	}
	if lpsn_ticker == "" || !lpsn_date {
		t.Fatal("missing LPSN")
	}
	if !apge_date || !strings.Contains(apge_headline, "$135.11") {
		t.Fatalf("%q", apge_headline)
	}
}

func TestParseStripsSimpleHTML(t *testing.T) {
	html := `
    <h2>September 3, 2026</h2>
    <p>Apogee Therapeutics, Inc. (APGE) performed a cash merger.</p>
    <hr>
    <p>Amphenol (APH) performed a 2 for 1 Forward Split.</p>
    `
	rows := ParseTracker(html)
	if len(rows) != 2 || rows[0].Ticker != "APGE" || rows[1].Ticker != "APH" {
		t.Fatalf("%+v", rows)
	}
}

func TestParseTrackerPage(t *testing.T) {
	result := ParseTrackerPage(testdata.TrackerSept2026)
	if result.Status != "success" || result.Count != 13 {
		t.Fatalf("%+v", result)
	}
}
