package neo4j

import (
	"strings"
	"testing"

	"github.com/tcw165/fintech-fun/graph/examples"
)

type stubClient struct {
	calls []struct {
		cypher string
		params map[string]any
	}
}

func (s *stubClient) Run(cypher string, params map[string]any) ([]map[string]any, error) {
	s.calls = append(s.calls, struct {
		cypher string
		params map[string]any
	}{cypher, params})
	return []map[string]any{{"ok": true}}, nil
}

func TestConstraintsCoverStockEventAndIndexes(t *testing.T) {
	joined := strings.Join(Constraints, "\n")
	for _, needle := range []string{"stock_ticker", "event_id", "company_name"} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("missing %s", needle)
		}
	}
	if strings.Contains(joined, "REQUIRE c.name IS UNIQUE") {
		t.Fatal("company name must not be unique")
	}
}

func TestApplyConstraints(t *testing.T) {
	var stub stubClient
	if err := ApplyConstraints(&stub); err != nil {
		t.Fatal(err)
	}
	if len(stub.calls) != len(Constraints) {
		t.Fatalf("got %d calls", len(stub.calls))
	}
}

func TestUpsertCypherUsesMergeAndRetailEdges(t *testing.T) {
	if !strings.Contains(UpsertCompanyCypher(), "MERGE (c:Company {name: $name})") {
		t.Fatal(UpsertCompanyCypher())
	}
	if !strings.Contains(UpsertStockCypher(), "MERGE (c)-[:ISSUES]->(s)") {
		t.Fatal(UpsertStockCypher())
	}
	event := UpsertEventCypher()
	for _, needle := range []string{"MERGE (e)-[:HAPPENED_TO]->(s)", "YOU_NOW_HOLD", "ON CREATE SET"} {
		if !strings.Contains(event, needle) {
			t.Fatalf("missing %s", needle)
		}
	}
}

func TestIngestNFLX(t *testing.T) {
	var stub stubClient
	c, s, e := examples.NFLX()
	result, err := IngestGraph(&stub, c, s, e)
	if err != nil {
		t.Fatal(err)
	}
	if result != (IngestResult{Company: "Netflix", Ticker: "NFLX", Events: 1}) {
		t.Fatalf("%+v", result)
	}
	var found bool
	for _, call := range stub.calls {
		if call.params["kind"] == "split" {
			found = true
			if call.params["id"] != "NFLX|2025-11-17|split" || call.params["share_multiplier"] != 10.0 {
				t.Fatalf("%v", call.params)
			}
			if call.params["you_now_hold"] != nil {
				t.Fatalf("you_now_hold %v", call.params["you_now_hold"])
			}
		}
	}
	if !found {
		t.Fatal("missing split event")
	}
}

func TestIngestLPSNSetsYouNowHold(t *testing.T) {
	var stub stubClient
	c, s, e := examples.LPSN()
	if _, err := IngestGraph(&stub, c, s, e); err != nil {
		t.Fatal(err)
	}
	for _, call := range stub.calls {
		if call.params["kind"] == "now_different_stock" {
			if call.params["you_now_hold"] != "SOUN" || call.params["happened_to"] != "LPSN" {
				t.Fatalf("%v", call.params)
			}
			return
		}
	}
	t.Fatal("missing event")
}

func TestIngestAPGEStoresCash(t *testing.T) {
	var stub stubClient
	c, s, e := examples.APGE()
	if _, err := IngestGraph(&stub, c, s, e); err != nil {
		t.Fatal(err)
	}
	for _, call := range stub.calls {
		if call.params["kind"] == "cashed_out" {
			if call.params["cash_per_share"] != 135.11 || call.params["can_trade"] != false {
				t.Fatalf("%v", call.params)
			}
			return
		}
	}
	t.Fatal("missing event")
}
