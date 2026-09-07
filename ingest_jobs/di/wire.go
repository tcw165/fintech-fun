package di

import (
	"context"

	neo4jimpl "github.com/tcw165/fintech-fun/graph/clients/neo4j/impl"
	qdrantimpl "github.com/tcw165/fintech-fun/graph/clients/qdrant/impl"
	"github.com/tcw165/fintech-fun/graph/embed/lexical"
)

func needs_live_clients(args []string) bool {
	if len(args) == 0 {
		return true
	}
	switch args[0] {
	case "ping", "seed", "fold", "ingest", "search", "refresh":
		return true
	default:
		return false
	}
}

func wants_optional_clients(args []string) bool {
	return len(args) > 0 && args[0] == "verify"
}

func NewFromEnv(ctx context.Context, args []string) (*AppContext, error) {
	app := New(nil, nil, lexical.New(), nil)
	if !needs_live_clients(args) && !wants_optional_clients(args) {
		return app, nil
	}
	driver, err := neo4jimpl.OpenFromEnv()
	if err != nil {
		if needs_live_clients(args) {
			return nil, err
		}
		return app, nil
	}
	app.closer = driver.Close
	if err := driver.Verify(ctx); err != nil {
		if needs_live_clients(args) {
			_ = app.Close(ctx)
			return nil, err
		}
		return app, nil
	}
	app.graph_db = driver
	app.vector_db = qdrantimpl.NewHTTPFromEnv()
	return app, nil
}
