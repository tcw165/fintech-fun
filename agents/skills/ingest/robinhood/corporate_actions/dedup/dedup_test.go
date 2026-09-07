package dedup

import (
	"strings"
	"testing"
	"time"

	hood_events "github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/parser"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/testdata"
)

func row(day, ticker, headline string) hood_events.TrackerRow {
	return hood_events.TrackerRow{
		Date:     date(day),
		Ticker:   ticker,
		Headline: headline,
		Company:  ticker,
	}
}

func date(day string) time.Time {
	t, err := time.Parse("2006-01-02", day)
	if err != nil {
		panic(err)
	}
	return t
}

func TestFingerprintUsesDateTickerHeadline(t *testing.T) {
	got := Fingerprint(row("2026-09-04", "LPSN", "LivePerson (LPSN) merger"))
	if got != "2026-09-04|LPSN|LivePerson (LPSN) merger" {
		t.Fatalf("%s", got)
	}
}

func TestChronologicalReversesNewestFirstPage(t *testing.T) {
	rows := parser.ParseTracker(testdata.TrackerSept2026)
	if len(rows) < 12 {
		t.Fatalf("rows %d", len(rows))
	}
	if rows[0].Date.Format("2006-01-02") != "2026-09-04" {
		t.Fatalf("page should start newest-first: %s", rows[0].Date.Format("2006-01-02"))
	}
	chrono := Chronological(rows)
	if chrono[0].Date.Format("2006-01-02") != "2026-09-03" {
		t.Fatalf("chrono start %s", chrono[0].Date.Format("2006-01-02"))
	}
	if chrono[len(chrono)-1].Date.Format("2006-01-02") != "2026-09-04" {
		t.Fatalf("chrono end %s", chrono[len(chrono)-1].Date.Format("2006-01-02"))
	}
	if chrono[0].Ticker != "NTZ" {
		t.Fatalf("oldest day last row should lead after reverse: %s %s", chrono[0].Ticker, chrono[0].Headline)
	}
	if chrono[len(chrono)-1].Ticker != "NHPAP" {
		t.Fatalf("newest day first row should trail after reverse: %s", chrono[len(chrono)-1].Ticker)
	}
}

func TestFindPrefixOverlapAndSuffixOnNewDay(t *testing.T) {
	old_rows := Chronological(parser.ParseTracker(testdata.TrackerSept2026))
	stored := Fingerprints(old_rows)
	new_day := []hood_events.TrackerRow{
		row("2026-09-05", "NFLX", "Netflix (NFLX) performed a 10 for 1 Forward Split."),
		row("2026-09-05", "XYZ", "Block (XYZ) performed a ticker change."),
	}
	// Newest-first page: Sept 5, then the original Sept 4 / Sept 3 body.
	page := testdata.TrackerSept2026
	page = strings.Replace(page, "## September 4, 2026", "## September 5, 2026\n\n"+new_day[0].Headline+"\n\n---\n\n"+new_day[1].Headline+"\n\n## September 4, 2026", 1)
	fetched := Chronological(parser.ParseTracker(page))
	overlap := FindPrefixOverlap(stored, Fingerprints(fetched))
	if overlap != len(stored) {
		t.Fatalf("overlap %d stored %d fetched %d", overlap, len(stored), len(fetched))
	}
	suffix := Suffix(fetched, overlap)
	if len(suffix) != 2 || suffix[0].Ticker != "XYZ" || suffix[1].Ticker != "NFLX" {
		t.Fatalf("suffix %+v", suffix)
	}
}

func TestFindPrefixOverlapBreaksOnMidHistoryEdit(t *testing.T) {
	stored := []string{"2026-09-03|APH|a", "2026-09-03|APGE|b", "2026-09-04|LPSN|c"}
	fetched := []string{"2026-09-03|APH|a", "2026-09-03|APGE|EDITED", "2026-09-04|LPSN|c"}
	overlap := FindPrefixOverlap(stored, fetched)
	if overlap != 1 {
		t.Fatalf("overlap %d", overlap)
	}
	rows := []hood_events.TrackerRow{
		row("2026-09-03", "APH", "a"),
		row("2026-09-03", "APGE", "EDITED"),
		row("2026-09-04", "LPSN", "c"),
	}
	suffix := Suffix(rows, overlap)
	if len(suffix) != 2 || suffix[0].Ticker != "APGE" || suffix[1].Ticker != "LPSN" {
		t.Fatalf("%+v", suffix)
	}
}

func TestPagesOneDayPerPageAndCapsSize(t *testing.T) {
	var rows []hood_events.TrackerRow
	for i := 0; i < 5; i++ {
		rows = append(rows, row("2026-09-03", "A", "old"))
	}
	for i := 0; i < MaxPageRows+3; i++ {
		rows = append(rows, row("2026-09-04", "B", "new"))
	}
	pages := Pages(rows)
	if len(pages) != 3 {
		t.Fatalf("pages %d", len(pages))
	}
	if len(pages[0]) != 5 || !pages[0][0].Date.Equal(date("2026-09-03")) {
		t.Fatalf("day1 %+v", pages[0])
	}
	if len(pages[1]) != MaxPageRows || !pages[1][0].Date.Equal(date("2026-09-04")) {
		t.Fatalf("day2 page1 %d", len(pages[1]))
	}
	if len(pages[2]) != 3 {
		t.Fatalf("day2 page2 %d", len(pages[2]))
	}
}

func TestEmptyPagesAndFullOverlapSuffix(t *testing.T) {
	if Pages(nil) != nil {
		t.Fatal("empty pages")
	}
	rows := []hood_events.TrackerRow{row("2026-09-03", "APH", "a")}
	if Suffix(rows, 1) != nil || Suffix(rows, 2) != nil {
		t.Fatal("full overlap should be empty")
	}
}
