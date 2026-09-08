package di

import (
	"context"

	neo4jimpl "github.com/tcw165/fintech-fun/graph/clients/neo4j/impl"
	qdrantimpl "github.com/tcw165/fintech-fun/graph/clients/qdrant/impl"
	"github.com/tcw165/fintech-fun/graph/embed/lexical"
)

func needs_live_clients(name string) bool {
	if name == "" {
		return true
	}
	switch name {
	case "ping", "seed", "fold", "ingest", "search", "refresh":
		return true
	default:
		return false
	}
}

func wants_optional_clients(name string) bool {
	return name == "verify"
}

func NewFromEnv(ctx context.Context, name string) (*AppContext, error) {
	app := New(nil, nil, lexical.New(), nil)
	if !needs_live_clients(name) && !wants_optional_clients(name) {
		return app, nil
	}
	driver, err := neo4jimpl.OpenFromEnv()
	if err != nil {
		if needs_live_clients(name) {
			return nil, err
		}
		return app, nil
	}
	app.closer = driver.Close
	if err := driver.Verify(ctx); err != nil {
		if needs_live_clients(name) {
			_ = app.Close(ctx)
			return nil, err
		}
		return app, nil
	}
	app.graph_db = driver
	app.vector_db = qdrantimpl.NewHTTPFromEnv()
	return app, nil
}
