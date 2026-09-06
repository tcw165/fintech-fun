package impl_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tcw165/fintech-fun/api_server/contract"
	"github.com/tcw165/fintech-fun/api_server/impl"
	neo4jimpl "github.com/tcw165/fintech-fun/graph/clients/neo4j/impl"
)

func TestHealthz(t *testing.T) {
	srv := httptest.NewServer(impl.New(impl.StaticOK{}, nil))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body contract.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("status = %q", body.Status)
	}
}

type stubFold struct {
	q    string
	qty  float64
	resp contract.FoldResponse
	err  error
}

func (s *stubFold) Fold(q string, qty float64) (contract.FoldResponse, error) {
	s.q, s.qty = q, qty
	return s.resp, s.err
}

func TestFoldQuery(t *testing.T) {
	stub := &stubFold{resp: contract.FoldResponse{
		Status: "ok", Q: "square", Qty: 10,
		Rows: []map[string]any{{"company": "Block", "ticker_now": "XYZ"}},
	}}
	srv := httptest.NewServer(impl.New(impl.StaticOK{}, stub))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/fold?q=square&qty=10")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body contract.FoldResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if stub.q != "square" || stub.qty != 10 || body.Rows[0]["company"] != "Block" {
		t.Fatalf("%+v %+v", stub, body)
	}
}

func TestFoldUnavailable(t *testing.T) {
	srv := httptest.NewServer(impl.New(impl.StaticOK{}, nil))
	t.Cleanup(srv.Close)
	resp, err := http.Get(srv.URL + "/fold?q=square&qty=10")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestFoldBadRequest(t *testing.T) {
	srv := httptest.NewServer(impl.New(impl.StaticOK{}, &stubFold{}))
	t.Cleanup(srv.Close)
	resp, err := http.Get(srv.URL + "/fold?q=&qty=x")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestNeo4jFoldIssuesCypher(t *testing.T) {
	rec := &neo4jimpl.Recording{}
	out, err := impl.Neo4jFold{Graph: rec}.Fold("square", 10)
	if err != nil {
		t.Fatal(err)
	}
	if out.Q != "square" || out.Qty != 10 || len(rec.Calls) != 1 || rec.Calls[0].Params["q"] != "square" {
		t.Fatalf("%+v %+v", out, rec.Calls)
	}
}
