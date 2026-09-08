package impl_test

import (
	"net/http/httptest"
	"testing"

	"github.com/tcw165/fintech-fun/api_server/contract"
	apiimpl "github.com/tcw165/fintech-fun/api_server/impl"
	"github.com/tcw165/fintech-fun/infra/activation/impl"
)

type stub_fold struct{}

func (stub_fold) Fold(q string, qty float64) (contract.FoldResponse, error) {
	row := map[string]any{"company": "Block", "ticker_now": "XYZ", "qty_now": 10.0, "cash_received": 0.0}
	switch q {
	case "nflx":
		row = map[string]any{"company": "Netflix", "ticker_now": "NFLX", "qty_now": 100.0, "cash_received": 0.0}
	case "mnts":
		row = map[string]any{"company": "Momentus", "ticker_now": "MNTS", "qty_now": 10.0, "cash_received": 0.0}
	case "apge":
		row = map[string]any{"company": "Apogee", "ticker_now": "APGE", "qty_now": 0.0, "cash_received": 1351.10}
	case "lpsn":
		row = map[string]any{"company": "LivePerson", "ticker_now": "LPSN", "qty_now": 46.73, "cash_received": 0.0}
	case "ftel":
		row = map[string]any{"company": "GMEX Robotics", "ticker_now": "GMEX", "qty_now": 10.0 / 16 / 8 / 7 / 9, "cash_received": 0.0}
	}
	return contract.FoldResponse{Status: "ok", Q: q, Qty: qty, Rows: []map[string]any{row}}, nil
}

type stub_search struct{}

func (stub_search) Search(query string) (contract.SearchResponse, error) {
	return contract.SearchResponse{Status: "ok", Query: query, Hits: []any{"LPSN"}}, nil
}

type stub_graph struct{}

func (stub_graph) Series(q string) (contract.GraphResponse, error) {
	return contract.GraphResponse{Status: "ok", Q: q, Rows: []map[string]any{{"company": "Block"}}}, nil
}

func (stub_graph) Source() (contract.SourceResponse, error) {
	return contract.SourceResponse{Status: "ok", ID: "robinhood/corporate_actions", PageSHA256: "abc"}, nil
}

func TestSmokeAgainstStubAPI(t *testing.T) {
	srv := httptest.NewServer(apiimpl.New(apiimpl.StaticOK{}, stub_fold{}, stub_search{}, stub_graph{}))
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

func TestGoldAgainstStubAPI(t *testing.T) {
	srv := httptest.NewServer(apiimpl.New(apiimpl.StaticOK{}, stub_fold{}, stub_search{}, stub_graph{}))
	t.Cleanup(srv.Close)
	report := impl.HTTP{}.Gold(srv.URL)
	if !report.AllOK() || len(report.Steps) != 6 {
		t.Fatalf("%+v", report)
	}
}
