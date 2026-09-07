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

func TestNewSkillMapsClients(t *testing.T) {
	graph_db := &neo4jimpl.Recording{}
	vector_db := &qdrantimpl.Recording{}
	embedder := lexical.New()
	app := New(graph_db, vector_db, embedder, nil)
	if app.graph_db != graph_db || app.vector_db != vector_db || app.embedder == nil {
		t.Fatalf("clients not stored: %+v", app)
	}
	s := app.Skill()
	if s.Graph != graph_db || s.Vectors != vector_db {
		t.Fatalf("skill clients: %+v", s)
	}
	ingest_agent := app.Agent()
	if ingest_agent == nil {
		t.Fatal("missing agent")
	}
}

func TestSkillNilContext(t *testing.T) {
	var app *AppContext
	s := app.Skill()
	if s.Graph != nil || s.Vectors != nil {
		t.Fatalf("nil skill: %+v", s)
	}
	if app.Agent() == nil {
		t.Fatal("nil context should still build an agent")
	}
}

func TestCommandGating(t *testing.T) {
	cases := []struct {
		args     []string
		live     bool
		optional bool
	}{
		{nil, true, false},
		{[]string{}, true, false},
		{[]string{"parse"}, false, false},
		{[]string{"classify"}, false, false},
		{[]string{"plan"}, false, false},
		{[]string{"fetch"}, false, false},
		{[]string{"gold"}, false, false},
		{[]string{"ingest"}, true, false},
		{[]string{"refresh"}, true, false},
		{[]string{"ping"}, true, false},
		{[]string{"seed"}, true, false},
		{[]string{"fold"}, true, false},
		{[]string{"search"}, true, false},
		{[]string{"verify"}, false, true},
	}
	for _, tc := range cases {
		want_agent := tc.live && (len(tc.args) == 0 || tc.args[0] == "ingest" || tc.args[0] == "refresh")
		if got := UsesAgent(tc.args); got != want_agent {
			t.Fatalf("uses_agent(%v) = %v, want %v", tc.args, got, want_agent)
		}
	}
	for _, tc := range cases {
		if got := needs_live_clients(tc.args); got != tc.live {
			t.Fatalf("needs_live_clients(%v) = %v, want %v", tc.args, got, tc.live)
		}
		if got := wants_optional_clients(tc.args); got != tc.optional {
			t.Fatalf("wants_optional_clients(%v) = %v, want %v", tc.args, got, tc.optional)
		}
	}
}

func TestNewFromEnvOfflineLeavesClientsNil(t *testing.T) {
	for _, args := range [][]string{{"parse"}, {"classify"}, {"plan"}, {"fetch"}, {"gold"}} {
		app, err := NewFromEnv(context.Background(), args)
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if app.graph_db != nil || app.vector_db != nil {
			t.Fatalf("%v clients set: %+v", args, app)
		}
		if app.embedder == nil {
			t.Fatalf("%v missing embedder", args)
		}
		if app.Skill().Graph != nil || app.Skill().Vectors != nil {
			t.Fatalf("%v skill clients set", args)
		}
	}
}
