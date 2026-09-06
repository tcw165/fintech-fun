package ingest

import (
	"testing"

	"github.com/tcw165/fintech-fun/graph"
	neo4jimpl "github.com/tcw165/fintech-fun/graph/clients/neo4j/impl"
	qdrantimpl "github.com/tcw165/fintech-fun/graph/clients/qdrant/impl"
	"github.com/tcw165/fintech-fun/graph/embed/lexical"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/testdata"
)

func classified(kind, headline, company, ticker string) hood_events.ClassifiedEvent {
	return hood_events.ClassifiedEvent{
		Date: graph.Date(2026, 9, 4), Kind: kind, Headline: headline, Company: company, Ticker: ticker,
		ShareMultiplier: 2, KeepFractionals: true, CanTrade: true,
	}
}

func TestToGraphLPSN(t *testing.T) {
	item := classified("now_different_stock", "LivePerson (LPSN) performed a stock merger.", "LivePerson", "LPSN")
	item.ShareMultiplier = 0.4673
	item.CanTrade = false
	item.YouNowHold = "SOUN"
	company, stock, event, ok := ToGraph(item)
	if !ok || company.Name != "LivePerson" || stock.Ticker != "LPSN" || stock.Status != graph.StatusGone {
		t.Fatalf("%+v %+v %+v", company, stock, event)
	}
	if event.Kind != graph.KindNowDifferentStock || event.YouNowHold != "SOUN" {
		t.Fatalf("%+v", event)
	}
}

func TestToGraphTickerChangeOmitsYouNowHold(t *testing.T) {
	item := classified("ticker_changed", "Block (SQ) → XYZ", "Block", "SQ")
	item.ShareMultiplier = 1
	item.YouNowHold = "XYZ"
	company, stock, event, ok := ToGraph(item)
	if !ok || stock.Ticker != "XYZ" || len(stock.FormerTickers) != 1 || stock.FormerTickers[0] != "SQ" {
		t.Fatalf("%+v %+v", company, stock)
	}
	if event.HappenedTo != "XYZ" || event.YouNowHold != "" {
		t.Fatalf("%+v", event)
	}
}

func TestToGraphNameChangeAliases(t *testing.T) {
	item := classified("name_changed", "Bit Origin changed its corporate name to Sangrix.", "Bit Origin Limited Class A", "SGRX")
	item.ShareMultiplier = 1
	item.NewName = "Sangrix"
	company, _, event, ok := ToGraph(item)
	if !ok || company.Name != "Sangrix" || len(company.AlsoKnownAs) != 1 || event.Kind != graph.KindNameChanged || event.YouNowHold != "" {
		t.Fatalf("%+v %+v", company, event)
	}
}

func TestToGraphSkips(t *testing.T) {
	item := classified("", "CUSIP change", "", "")
	item.Skip = true
	if _, _, _, ok := ToGraph(item); ok {
		t.Fatal("expected skip")
	}
	item = classified("split", "x", "Co", "")
	if _, _, _, ok := ToGraph(item); ok {
		t.Fatal("expected missing ticker")
	}
}

func TestIngestClassifiedWritesGraphAndQdrant(t *testing.T) {
	run := &neo4jimpl.Recording{}
	vectors := &qdrantimpl.Recording{}
	skip := classified("", "CUSIP change", "", "")
	skip.Skip = true
	stats, err := IngestClassified(run, []hood_events.ClassifiedEvent{
		classified("split", "Amphenol (APH) performed a 2 for 1 Forward Split.", "Amphenol", "APH"),
		skip,
	}, vectors, lexical.New())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Written != 1 || stats.Skipped != 1 || len(vectors.Collections) != 1 || vectors.Upserts != 1 || len(run.Calls) == 0 {
		t.Fatalf("stats=%+v q=%+v n=%d", stats, vectors, len(run.Calls))
	}
}

func TestIngestTextWritesFixture(t *testing.T) {
	run := &neo4jimpl.Recording{}
	stats, err := IngestText(run, testdata.TrackerSept2026, &qdrantimpl.Recording{}, lexical.New())
	if err != nil || stats.Written < 12 || stats.Skipped < 0 || len(stats.Stocks) == 0 {
		t.Fatalf("%+v %v", stats, err)
	}
}

func TestPlanIngestIsDryRun(t *testing.T) {
	result := PlanIngest(testdata.TrackerSept2026)
	if !result.DryRun || result.Written < 12 {
		t.Fatalf("%+v", result)
	}
	kinds := map[string]bool{}
	var lpsn PlannedEvent
	for _, item := range result.Events {
		kinds[item.Kind] = true
		if item.Ticker == "LPSN" {
			lpsn = item
		}
	}
	if !kinds["now_different_stock"] || !kinds["cashed_out"] || lpsn.YouNowHold != "SOUN" {
		t.Fatalf("%+v %+v", kinds, lpsn)
	}
}
