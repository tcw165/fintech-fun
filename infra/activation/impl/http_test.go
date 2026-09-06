package impl_test

import (
	"net/http/httptest"
	"testing"

	"github.com/tcw165/fintech-fun/api_server/contract"
	apiimpl "github.com/tcw165/fintech-fun/api_server/impl"
	"github.com/tcw165/fintech-fun/infra/activation/impl"
)

type stubFold struct{}

func (stubFold) Fold(q string, qty float64) (contract.FoldResponse, error) {
	return contract.FoldResponse{Status: "ok", Q: q, Qty: qty, Rows: []map[string]any{{"company": "Block"}}}, nil
}

type stubSearch struct{}

func (stubSearch) Search(query string) (contract.SearchResponse, error) {
	return contract.SearchResponse{Status: "ok", Query: query, Hits: []any{"LPSN"}}, nil
}

type stubGraph struct{}

func (stubGraph) Series(q string) (contract.GraphResponse, error) {
	return contract.GraphResponse{Status: "ok", Q: q, Rows: []map[string]any{{"company": "Block"}}}, nil
}

func (stubGraph) Source() (contract.SourceResponse, error) {
	return contract.SourceResponse{Status: "ok", ID: "robinhood/corporate_actions", PageSHA256: "abc"}, nil
}

func TestSmokeAgainstStubAPI(t *testing.T) {
	srv := httptest.NewServer(apiimpl.New(apiimpl.StaticOK{}, stubFold{}, stubSearch{}, stubGraph{}))
	t.Cleanup(srv.Close)
	report := impl.HTTP{}.Smoke(srv.URL)
	if !report.AllOK() {
		t.Fatalf("%+v", report)
	}
}

func TestHealthFailsWhenDown(t *testing.T) {
	step := impl.HTTP{}.Health("http://127.0.0.1:1")
	if step.OK {
		t.Fatalf("%+v", step)
	}
}
