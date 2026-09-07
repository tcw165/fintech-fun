package di

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tcw165/fintech-fun/api_server/contract"
	neo4jimpl "github.com/tcw165/fintech-fun/graph/clients/neo4j/impl"
	qdrantimpl "github.com/tcw165/fintech-fun/graph/clients/qdrant/impl"
	"github.com/tcw165/fintech-fun/graph/embed/lexical"
)

func TestCloseNilAndEmpty(t *testing.T) {
	var app *AppContext
	if err := app.Close(context.Background()); err != nil {
		t.Fatalf("nil close: %v", err)
	}
	if err := New(nil, nil, nil, nil).Close(context.Background()); err != nil {
		t.Fatalf("empty close: %v", err)
	}
}

func TestCloseCallsCloser(t *testing.T) {
	called := false
	app := New(nil, nil, nil, func(context.Context) error {
		called = true
		return nil
	})
	if err := app.Close(context.Background()); err != nil || !called {
		t.Fatalf("close: %v called=%v", err, called)
	}
}

func TestClosePropagatesError(t *testing.T) {
	want := errors.New("boom")
	app := New(nil, nil, nil, func(context.Context) error { return want })
	if err := app.Close(context.Background()); err != want {
		t.Fatalf("got %v", err)
	}
}

func TestNewHandlerHealthz(t *testing.T) {
	graph_db := &neo4jimpl.Recording{}
	vector_db := &qdrantimpl.Recording{}
	embedder := lexical.New()
	app := New(graph_db, vector_db, embedder, nil)
	if app.graph_db != graph_db || app.vector_db != vector_db || app.embedder == nil {
		t.Fatalf("clients not stored: %+v", app)
	}
	srv := httptest.NewServer(app.Handler())
	t.Cleanup(srv.Close)
	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body contract.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ok" {
		t.Fatalf("status = %q", body.Status)
	}
}

func TestHandlerFoldUnavailableWithoutGraph(t *testing.T) {
	app := New(nil, &qdrantimpl.Recording{}, lexical.New(), nil)
	srv := httptest.NewServer(app.Handler())
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

func TestHandlerFoldUsesGraphDB(t *testing.T) {
	graph_db := &neo4jimpl.Recording{}
	app := New(graph_db, &qdrantimpl.Recording{}, lexical.New(), nil)
	srv := httptest.NewServer(app.Handler())
	t.Cleanup(srv.Close)
	resp, err := http.Get(srv.URL + "/fold?q=square&qty=10")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if len(graph_db.Calls) != 1 || graph_db.Calls[0].Params["q"] != "square" {
		t.Fatalf("calls = %+v", graph_db.Calls)
	}
}

func TestAddrDefault(t *testing.T) {
	if got := New(nil, nil, nil, nil).Addr(); got != ":8080" {
		t.Fatalf("addr = %q", got)
	}
	var app *AppContext
	if got := app.Addr(); got != ":8080" {
		t.Fatalf("nil addr = %q", got)
	}
}
