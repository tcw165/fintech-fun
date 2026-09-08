package tools

import (
	"os"
	"testing"

	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/agent"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/classify"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/ingest"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/parser"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/testdata"
	"github.com/tcw165/fintech-fun/graph/clients/neo4j"
	neo4jimpl "github.com/tcw165/fintech-fun/graph/clients/neo4j/impl"
	"github.com/tcw165/fintech-fun/graph/clients/qdrant"
	qdrantimpl "github.com/tcw165/fintech-fun/graph/clients/qdrant/impl"
	"github.com/tcw165/fintech-fun/graph/embed/lexical"
	"github.com/tcw165/fintech-fun/ingest_jobs/corporate_actions/request"
)

func test_deps(graph_db neo4j.Client, vector_db qdrant.Client) Deps {
	return Deps{
		GraphClient:   graph_db,
		VectorsClient: vector_db,
		Embedder:      lexical.New(),
		Agent:         agent.New(graph_db, vector_db, lexical.New(), nil),
	}
}

func TestPingUsesInjectedClients(t *testing.T) {
	graph_db := &neo4jimpl.Recording{}
	vector_db := &qdrantimpl.Recording{}
	payload, err := Run(test_deps(graph_db, vector_db), request.Request{Name: "ping"})
	if err != nil {
		t.Fatal(err)
	}
	out := payload.(map[string]any)
	if out["status"] != "ok" || out["event_id"] != "NFLX|2025-11-17|split" {
		t.Fatalf("%v", out)
	}
	if len(graph_db.Calls) == 0 || len(vector_db.Collections) != 1 || vector_db.Upserts != 1 {
		t.Fatalf("graph=%d collections=%v upserts=%d", len(graph_db.Calls), vector_db.Collections, vector_db.Upserts)
	}
}

func TestSeedWritesSixFixtures(t *testing.T) {
	graph_db := &neo4jimpl.Recording{}
	vector_db := &qdrantimpl.Recording{}
	payload, err := Run(test_deps(graph_db, vector_db), request.Request{Name: "seed"})
	if err != nil {
		t.Fatal(err)
	}
	out := payload.(map[string]any)
	if out["fixtures"] != 6 || out["events"] != 11 {
		t.Fatalf("%v", out)
	}
	if vector_db.Upserts != 6 {
		t.Fatalf("upserts %d", vector_db.Upserts)
	}
}

func TestIngestMissingFile(t *testing.T) {
	_, err := Run(test_deps(&neo4jimpl.Recording{}, &qdrantimpl.Recording{}), request.Request{Name: "ingest", File: "missing-file-for-error"})
	if err == nil {
		t.Fatal("expected missing file")
	}
}

func TestIngestFileUsesAgentPrefixDedup(t *testing.T) {
	path := write_temp(t, testdata.TrackerSept2026)
	payload, err := Run(test_deps(&neo4jimpl.Recording{}, &qdrantimpl.Recording{}), request.Request{Name: "ingest", File: path})
	if err != nil {
		t.Fatal(err)
	}
	out := payload.(map[string]any)
	if out["status"] != "ok" || out["unchanged"] != false {
		t.Fatalf("%v", out)
	}
	written, _ := out["written"].(int)
	if written < 12 {
		t.Fatalf("%v", out)
	}
	if _, ok := out["overlap"]; !ok {
		t.Fatalf("missing overlap: %v", out)
	}
}

func TestIngestDryRunPlansWithoutClients(t *testing.T) {
	path := write_temp(t, testdata.TrackerSept2026)
	payload, err := Run(Deps{}, request.Request{Name: "ingest", File: path, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	plan := payload.(ingest.PlanResult)
	if !plan.DryRun || plan.Written < 12 {
		t.Fatalf("%+v", plan)
	}
}

func TestFoldIssuesAccountQuery(t *testing.T) {
	graph_db := &neo4jimpl.Recording{}
	payload, err := Run(test_deps(graph_db, &qdrantimpl.Recording{}), request.Request{Name: "fold", Q: "square", Qty: 10})
	if err != nil {
		t.Fatal(err)
	}
	out := payload.(map[string]any)
	if out["q"] != "square" || out["qty"] != 10.0 {
		t.Fatalf("%v", out)
	}
	if len(graph_db.Calls) != 1 || graph_db.Calls[0].Params["q"] != "square" {
		t.Fatalf("%+v", graph_db.Calls)
	}
}

func TestRefreshRequiresClients(t *testing.T) {
	_, err := Run(Deps{}, request.Request{Name: "refresh"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSearchUsesInjectedQdrant(t *testing.T) {
	vector_db := &qdrantimpl.Recording{}
	payload, err := Run(
		Deps{VectorsClient: vector_db, Embedder: lexical.New()},
		request.Request{
			Name:  "search",
			Query: "LivePerson stock merger",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	out := payload.(map[string]any)
	if out["query"] != "LivePerson stock merger" {
		t.Fatalf("%v", out)
	}
}

func TestSearchCollapsesMultilineQuery(t *testing.T) {
	vector_db := &qdrantimpl.Recording{}
	payload, err := Run(
		Deps{VectorsClient: vector_db, Embedder: lexical.New()},
		request.Request{
			Name:  "search",
			Query: "LivePerson\nstock\nmerger",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	out := payload.(map[string]any)
	if out["query"] != "LivePerson stock merger" {
		t.Fatalf("%v", out)
	}
}

func TestClassifyMultilineHeadline(t *testing.T) {
	payload, err := Run(
		Deps{},
		request.Request{
			Name:     "classify",
			Date:     "2026-09-04",
			Headline: "LivePerson (LPSN) performed a stock merger.\nShareholders will receive 0.4673 new shares of SOUN for each old share of LPSN previously held.",
			Company:  "LivePerson",
			Ticker:   "LPSN",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	result := payload.(classify.ClassifyResult)
	if result.Status != "success" || len(result.Events) != 1 {
		t.Fatalf("%+v", result)
	}
	if result.Events[0].Kind != "now_different_stock" || result.Events[0].ShareMultiplier != 0.4673 || result.Events[0].YouNowHold != "SOUN" {
		t.Fatalf("%+v", result.Events[0])
	}
}

func TestVerifyMatchesNotionTables(t *testing.T) {
	payload, err := Run(Deps{}, request.Request{Name: "verify"})
	if err != nil {
		t.Fatal(err)
	}
	out := payload.(map[string]any)
	if out["passed"] != 9 {
		t.Fatalf("%v", out)
	}
}

func TestVerifyIncludesBoltWhenGraphSet(t *testing.T) {
	graph_db := &neo4jimpl.Recording{}
	payload, err := Run(Deps{GraphClient: graph_db}, request.Request{Name: "verify"})
	if err != nil {
		t.Fatal(err)
	}
	out := payload.(map[string]any)
	if out["bolt"] == nil || len(graph_db.Calls) == 0 || out["bolt_failed"] != 0 {
		t.Fatalf("%v calls=%d", out, len(graph_db.Calls))
	}
}

func TestVerifyBoltGoldMatchesFoldRows(t *testing.T) {
	graph_db := neo4jimpl.Func(func(_ string, params map[string]any) ([]map[string]any, error) {
		q, _ := params["q"].(string)
		for _, want := range []struct {
			q, company, ticker string
			qty, cash          float64
		}{
			{"nflx", "Netflix", "NFLX", 100, 0},
			{"mnts", "Momentus", "MNTS", 10, 0},
			{"apge", "Apogee", "APGE", 0, 1351.10},
			{"lpsn", "LivePerson", "LPSN", 46.73, 0},
			{"square", "Block", "XYZ", 10, 0},
			{"ftel", "GMEX Robotics", "GMEX", 10.0 / 16 / 8 / 7 / 9, 0},
		} {
			if q == want.q {
				return []map[string]any{{
					"company": want.company, "ticker_now": want.ticker,
					"qty_now": want.qty, "cash_received": want.cash,
				}}, nil
			}
		}
		return nil, nil
	})
	payload, err := Run(Deps{GraphClient: graph_db}, request.Request{Name: "verify"})
	if err != nil {
		t.Fatal(err)
	}
	if payload.(map[string]any)["bolt_failed"] != 0 {
		t.Fatalf("%v", payload)
	}
}

func TestVerifyBoltGoldFailsOnMismatch(t *testing.T) {
	graph_db := neo4jimpl.Func(func(string, map[string]any) ([]map[string]any, error) {
		return []map[string]any{{"company": "Wrong", "ticker_now": "NOPE", "qty_now": 1.0}}, nil
	})
	_, err := Run(Deps{GraphClient: graph_db}, request.Request{Name: "verify"})
	if err == nil {
		t.Fatal("expected bolt gold failure")
	}
}

func TestPingRequiresClients(t *testing.T) {
	_, err := Run(Deps{}, request.Request{Name: "ping"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseFixture(t *testing.T) {
	path := write_temp(t, testdata.TrackerSept2026)
	payload, err := Run(Deps{}, request.Request{Name: "parse", File: path})
	if err != nil {
		t.Fatal(err)
	}
	page := payload.(parser.ParseResult)
	if page.Count < 12 {
		t.Fatalf("%+v", page)
	}
}

func TestUnknownCommand(t *testing.T) {
	_, err := Run(Deps{}, request.Request{Name: "nope"})
	if err == nil {
		t.Fatal("expected unknown command")
	}
}

func write_temp(t *testing.T, text string) string {
	t.Helper()
	path := t.TempDir() + "/tracker.txt"
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
