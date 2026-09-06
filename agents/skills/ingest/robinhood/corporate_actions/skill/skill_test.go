package skill

import (
	"context"
	"testing"

	"github.com/tcw165/fintech-fun/agents/harness/contract"
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

func TestPingRequiresClients(t *testing.T) {
	_, err := Skill{}.Run(context.Background(), contract.Request{Args: []string{"ping"}})
	if err == nil {
		t.Fatal("expected error")
	}
}
