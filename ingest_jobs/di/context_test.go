package di

import (
	"context"
	"errors"
	"testing"

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

func TestNewMapsClients(t *testing.T) {
	graph_db := &neo4jimpl.Recording{}
	vector_db := &qdrantimpl.Recording{}
	embedder := lexical.New()
	app := New(graph_db, vector_db, embedder, nil)
	if app.graph_db != graph_db || app.vector_db != vector_db || app.embedder == nil {
		t.Fatalf("clients not stored: %+v", app)
	}
	if app.GraphClient() != graph_db || app.VectorsClient() != vector_db || app.Embedder() == nil {
		t.Fatalf("accessors: %+v", app)
	}
	if app.Agent() == nil {
		t.Fatal("missing agent")
	}
}

func TestNilContextAccessors(t *testing.T) {
	var app *AppContext
	if app.GraphClient() != nil || app.VectorsClient() != nil || app.Embedder() != nil {
		t.Fatalf("nil accessors: %+v", app)
	}
	if app.Agent() == nil {
		t.Fatal("nil context should still build an agent")
	}
}

func TestCommandGating(t *testing.T) {
	cases := []struct {
		name     string
		live     bool
		optional bool
	}{
		{"", true, false},
		{"parse", false, false},
		{"classify", false, false},
		{"plan", false, false},
		{"fetch", false, false},
		{"gold", false, false},
		{"ingest", true, false},
		{"refresh", true, false},
		{"ping", true, false},
		{"seed", true, false},
		{"fold", true, false},
		{"search", true, false},
		{"verify", false, true},
		{"dry-run", false, true},
	}
	for _, tc := range cases {
		if got := needs_live_clients(tc.name); got != tc.live {
			t.Fatalf("needs_live_clients(%q) = %v, want %v", tc.name, got, tc.live)
		}
		if got := wants_optional_clients(tc.name); got != tc.optional {
			t.Fatalf("wants_optional_clients(%q) = %v, want %v", tc.name, got, tc.optional)
		}
	}
}

func TestNewFromEnvOfflineLeavesClientsNil(t *testing.T) {
	for _, name := range []string{"parse", "classify", "plan", "fetch", "gold"} {
		app, err := NewFromEnv(context.Background(), name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if app.graph_db != nil || app.vector_db != nil {
			t.Fatalf("%s clients set: %+v", name, app)
		}
		if app.embedder == nil {
			t.Fatalf("%s missing embedder", name)
		}
		if app.GraphClient() != nil || app.VectorsClient() != nil {
			t.Fatalf("%s clients set via accessors", name)
		}
	}
}
