package skill

import (
	"context"
	"testing"

	"github.com/tcw165/fintech-fun/agents/harness/contract"
	"github.com/tcw165/fintech-fun/agents/harness/skillmd"
	hood_events "github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
	neo4jimpl "github.com/tcw165/fintech-fun/graph/clients/neo4j/impl"
	qdrantimpl "github.com/tcw165/fintech-fun/graph/clients/qdrant/impl"
)

func TestPingUsesInjectedClients(t *testing.T) {
	graph := &neo4jimpl.Recording{}
	vectors := &qdrantimpl.Recording{}
	result, err := Skill{Graph: graph, Vectors: vectors}.Run(context.Background(), contract.Request{Args: []string{"ping"}})
	if err != nil {
		t.Fatal(err)
	}
	payload := result.Payload.(map[string]any)
	if payload["status"] != "ok" || payload["event_id"] != "NFLX|2025-11-17|split" {
		t.Fatalf("%v", payload)
	}
	if len(graph.Calls) == 0 || len(vectors.Collections) != 1 || vectors.Upserts != 1 {
		t.Fatalf("graph=%d collections=%v upserts=%d", len(graph.Calls), vectors.Collections, vectors.Upserts)
	}
	foundRead := false
	for _, call := range graph.Calls {
		if call.Params["id"] == "NFLX|2025-11-17|split" && call.Cypher != "" {
			foundRead = true
		}
	}
	if !foundRead {
		t.Fatalf("%+v", graph.Calls)
	}
}

func TestSeedWritesSixFixtures(t *testing.T) {
	graph := &neo4jimpl.Recording{}
	vectors := &qdrantimpl.Recording{}
	result, err := Skill{Graph: graph, Vectors: vectors}.Run(context.Background(), contract.Request{Args: []string{"seed"}})
	if err != nil {
		t.Fatal(err)
	}
	payload := result.Payload.(map[string]any)
	if payload["fixtures"] != 6 || payload["events"] != 11 {
		t.Fatalf("%v", payload)
	}
	if vectors.Upserts != 6 {
		t.Fatalf("upserts %d", vectors.Upserts)
	}
}

func TestIngestWritesFixtureThroughSkill(t *testing.T) {
	graph := &neo4jimpl.Recording{}
	vectors := &qdrantimpl.Recording{}
	_, err := Skill{Graph: graph, Vectors: vectors}.Run(context.Background(), contract.Request{Args: []string{
		"ingest", "missing-file-for-error",
	}})
	if err == nil {
		t.Fatal("expected missing file")
	}
}

func TestFoldIssuesAccountQuery(t *testing.T) {
	graph := &neo4jimpl.Recording{}
	vectors := &qdrantimpl.Recording{}
	result, err := Skill{Graph: graph, Vectors: vectors}.Run(context.Background(), contract.Request{Args: []string{"fold", "square", "10"}})
	if err != nil {
		t.Fatal(err)
	}
	payload := result.Payload.(map[string]any)
	if payload["q"] != "square" || payload["qty"] != 10.0 {
		t.Fatalf("%v", payload)
	}
	if len(graph.Calls) != 1 || graph.Calls[0].Params["q"] != "square" {
		t.Fatalf("%+v", graph.Calls)
	}
}

func TestRefreshRequiresClients(t *testing.T) {
	_, err := Skill{}.Run(context.Background(), contract.Request{Args: []string{"refresh"}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSearchUsesInjectedQdrant(t *testing.T) {
	vectors := &qdrantimpl.Recording{}
	result, err := Skill{Vectors: vectors}.Run(context.Background(), contract.Request{Args: []string{"search", "LivePerson", "stock", "merger"}})
	if err != nil {
		t.Fatal(err)
	}
	payload := result.Payload.(map[string]any)
	if payload["query"] != "LivePerson stock merger" {
		t.Fatalf("%v", payload)
	}
}

func TestVerifyMatchesNotionTables(t *testing.T) {
	result, err := Skill{}.Run(context.Background(), contract.Request{Args: []string{"verify"}})
	if err != nil {
		t.Fatal(err)
	}
	payload := result.Payload.(map[string]any)
	if payload["passed"] != 9 {
		t.Fatalf("%v", payload)
	}
}

func TestVerifyIncludesBoltWhenGraphSet(t *testing.T) {
	graph := &neo4jimpl.Recording{}
	result, err := Skill{Graph: graph}.Run(context.Background(), contract.Request{Args: []string{"verify"}})
	if err != nil {
		t.Fatal(err)
	}
	payload := result.Payload.(map[string]any)
	if payload["bolt"] == nil || len(graph.Calls) == 0 || payload["bolt_failed"] != 0 {
		t.Fatalf("%v calls=%d", payload, len(graph.Calls))
	}
}

func TestVerifyBoltGoldMatchesFoldRows(t *testing.T) {
	graph := neo4jimpl.Func(func(_ string, params map[string]any) ([]map[string]any, error) {
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
	result, err := Skill{Graph: graph}.Run(context.Background(), contract.Request{Args: []string{"verify"}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Payload.(map[string]any)["bolt_failed"] != 0 {
		t.Fatalf("%v", result.Payload)
	}
}

func TestVerifyBoltGoldFailsOnMismatch(t *testing.T) {
	graph := neo4jimpl.Func(func(string, map[string]any) ([]map[string]any, error) {
		return []map[string]any{{"company": "Wrong", "ticker_now": "NOPE", "qty_now": 1.0}}, nil
	})
	_, err := Skill{Graph: graph}.Run(context.Background(), contract.Request{Args: []string{"verify"}})
	if err == nil {
		t.Fatal("expected bolt gold failure")
	}
}

func TestPingRequiresClients(t *testing.T) {
	_, err := Skill{}.Run(context.Background(), contract.Request{Args: []string{"ping"}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestToolsMatchEmbeddedSkillMarkdown(t *testing.T) {
	doc, err := skillmd.Parse(hood_events.SkillMarkdown)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Name != "corporate-actions" || doc.DefaultTool() != "ingest" {
		t.Fatalf("%+v", doc)
	}
	tools := Skill{}.Tools()
	if len(tools) != len(doc.AllowedTools) {
		t.Fatalf("tools %d allowed %d", len(tools), len(doc.AllowedTools))
	}
	for _, name := range doc.AllowedTools {
		if tools[name] == nil || !doc.Allows(name) {
			t.Fatalf("missing tool %s", name)
		}
	}
}
