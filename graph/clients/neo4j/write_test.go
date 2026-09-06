package neo4j

import (
	"strings"
	"testing"

	"github.com/tcw165/fintech-fun/graph"
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

func TestFoldCypherResolvesAliasAndFormerTicker(t *testing.T) {
	query := FoldCypher()
	for _, needle := range []string{"also_known_as", "former_tickers", "reverse_split", "reduce("} {
		if !strings.Contains(query, needle) {
			t.Fatalf("missing %s", needle)
		}
	}
}

func TestSeriesCypherResolvesAliasAndFormerTicker(t *testing.T) {
	query := SeriesCypher()
	for _, needle := range []string{"also_known_as", "former_tickers", "collect({"} {
		if !strings.Contains(query, needle) {
			t.Fatalf("missing %s", needle)
		}
	}
}

func TestFoldRunsAccountQuery(t *testing.T) {
	var stub stubClient
	if _, err := Fold(&stub, "square", 10); err != nil {
		t.Fatal(err)
	}
	if len(stub.calls) != 1 || stub.calls[0].params["q"] != "square" || stub.calls[0].params["qty"] != 10.0 {
		t.Fatalf("%+v", stub.calls)
	}
}

func TestSeriesRunsGraphQuery(t *testing.T) {
	var stub stubClient
	if _, err := Series(&stub, "square"); err != nil {
		t.Fatal(err)
	}
	if len(stub.calls) != 1 || stub.calls[0].params["q"] != "square" {
		t.Fatalf("%+v", stub.calls)
	}
}

func TestConstraintsCoverStockEventAndIndexes(t *testing.T) {
	joined := strings.Join(Constraints, "\n")
	for _, needle := range []string{"stock_ticker", "event_id", "ingest_source_id", "company_name"} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("missing %s", needle)
		}
	}
	if strings.Contains(joined, "REQUIRE c.name IS UNIQUE") {
		t.Fatal("company name must not be unique")
	}
}

func TestListEventIDsAndReadSource(t *testing.T) {
	var stub stubClient
	ids, err := ListEventIDs(&stub)
	if err != nil || len(ids) != 0 {
		t.Fatalf("%v %v", ids, err)
	}
	wm, err := ReadSource(&stub, graph.IngestSourceCorporateActions)
	if err != nil || wm.PageSHA256 != "" || wm.ID != graph.IngestSourceCorporateActions {
		t.Fatalf("%+v %v", wm, err)
	}
	if err := UpsertSource(&stub, graph.IngestSourceCorporateActions, "abc"); err != nil {
		t.Fatal(err)
	}
	if len(stub.calls) != 3 {
		t.Fatalf("calls %d", len(stub.calls))
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
	issuer := UpsertIssuerCypher()
	if strings.Contains(issuer, "MERGE (c:Company {name:") {
		t.Fatal("company name must not be a MERGE key")
	}
	for _, needle := range []string{"MERGE (s:Stock {ticker: $ticker})", "[:ISSUES]->(s)", "$company_name"} {
		if !strings.Contains(issuer, needle) {
			t.Fatalf("missing %s", needle)
		}
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
