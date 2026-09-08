package di

import (
	"context"

	"github.com/tcw165/fintech-fun/graph/clients/neo4j"
	neo4jimpl "github.com/tcw165/fintech-fun/graph/clients/neo4j/impl"
	"github.com/tcw165/fintech-fun/graph/clients/qdrant"
	qdrantimpl "github.com/tcw165/fintech-fun/graph/clients/qdrant/impl"
	"github.com/tcw165/fintech-fun/graph/embed"
	"github.com/tcw165/fintech-fun/graph/embed/lexical"
	"go.uber.org/fx"
)

type boot_request struct {
	ctx  context.Context
	name string
}

// Module is the ingest-job Fx graph. name is request.Request.Name (empty = ingest).
func Module(
	ctx context.Context,
	name string,
) fx.Option {
	if ctx == nil {
		ctx = context.Background()
	}
	return fx.Module(
		"ingest_jobs",
		fx.Supply(boot_request{
			ctx:  ctx,
			name: name,
		}),
		fx.Provide(
			provide_embedder,
			provide_graph_db,
			provide_vector_db,
			provide_app,
		),
	)
}

func provide_embedder() embed.Embedder {
	return lexical.New()
}

func provide_graph_db(
	boot boot_request,
	lc fx.Lifecycle,
) (neo4j.Client, error) {
	if !needs_live_clients(boot.name) && !wants_optional_clients(boot.name) {
		return nil, nil
	}
	driver, err := neo4jimpl.OpenFromEnv()
	if err != nil {
		if needs_live_clients(boot.name) {
			return nil, err
		}
		return nil, nil
	}
	if err := driver.Verify(boot.ctx); err != nil {
		_ = driver.Close(boot.ctx)
		if needs_live_clients(boot.name) {
			return nil, err
		}
		return nil, nil
	}
	lc.Append(fx.Hook{
		OnStop: driver.Close,
	})
	return driver, nil
}

func provide_vector_db(
	graph_db neo4j.Client,
) qdrant.Client {
	if graph_db == nil {
		return nil
	}
	return qdrantimpl.NewHTTPFromEnv()
}

func provide_app(
	graph_db neo4j.Client,
	vector_db qdrant.Client,
	embedder embed.Embedder,
) *AppContext {
	return New(
		graph_db,
		vector_db,
		embedder,
		nil,
	)
}

// ClientName maps a parsed Request onto the Fx graph key.
// ingest --dry-run uses the offline plan graph.
func ClientName(
	name string,
	dry_run bool,
) string {
	if dry_run && (name == "" || name == "ingest") {
		return "plan"
	}
	return name
}

// Boot starts the ingest Fx graph for one CLI Request.Name.
func Boot(
	ctx context.Context,
	name string,
) (*AppContext, error) {
	var app *AppContext
	fx_app := fx.New(
		Module(ctx, name),
		fx.Populate(&app),
		fx.NopLogger,
	)
	if err := fx_app.Err(); err != nil {
		return nil, err
	}
	if err := fx_app.Start(ctx); err != nil {
		return nil, err
	}
	app.closer = fx_app.Stop
	return app, nil
}
