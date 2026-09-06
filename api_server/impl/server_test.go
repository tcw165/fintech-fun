package impl_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tcw165/fintech-fun/api_server/contract"
	"github.com/tcw165/fintech-fun/api_server/impl"
	neo4jimpl "github.com/tcw165/fintech-fun/graph/clients/neo4j/impl"
	qdrantimpl "github.com/tcw165/fintech-fun/graph/clients/qdrant/impl"
	"github.com/tcw165/fintech-fun/graph/embed/lexical"
)

func TestHealthz(t *testing.T) {
	srv := httptest.NewServer(impl.New(impl.StaticOK{}, nil, nil))
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
	srv := httptest.NewServer(impl.New(impl.StaticOK{}, stub, nil))
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
	srv := httptest.NewServer(impl.New(impl.StaticOK{}, nil, nil))
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
	srv := httptest.NewServer(impl.New(impl.StaticOK{}, &stubFold{}, nil))
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

type stubSearch struct {
	q    string
	resp contract.SearchResponse
}

func (s *stubSearch) Search(query string) (contract.SearchResponse, error) {
	s.q = query
	return s.resp, nil
}

func TestSearchQuery(t *testing.T) {
	stub := &stubSearch{resp: contract.SearchResponse{Status: "ok", Query: "LivePerson stock merger", Hits: []any{"LPSN"}}}
	srv := httptest.NewServer(impl.New(impl.StaticOK{}, nil, stub))
	t.Cleanup(srv.Close)
	resp, err := http.Get(srv.URL + "/search?q=LivePerson+stock+merger")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body contract.SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if stub.q != "LivePerson stock merger" || body.Query != stub.q {
		t.Fatalf("%+v %+v", stub, body)
	}
}

func TestSearchUnavailable(t *testing.T) {
	srv := httptest.NewServer(impl.New(impl.StaticOK{}, nil, nil))
	t.Cleanup(srv.Close)
	resp, err := http.Get(srv.URL + "/search?q=LPSN")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestQdrantSearchUsesRecording(t *testing.T) {
	vectors := &qdrantimpl.Recording{}
	out, err := impl.QdrantSearch{Vectors: vectors, Embed: lexical.New()}.Search("LivePerson stock merger")
	if err != nil {
		t.Fatal(err)
	}
	if out.Query != "LivePerson stock merger" || out.Status != "ok" || out.Hits == nil {
		t.Fatalf("%+v", out)
	}
}
