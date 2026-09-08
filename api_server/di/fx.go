package di

import (
	"context"
	"log"

	"github.com/tcw165/fintech-fun/graph/clients/neo4j"
	neo4jimpl "github.com/tcw165/fintech-fun/graph/clients/neo4j/impl"
	"github.com/tcw165/fintech-fun/graph/clients/qdrant"
	qdrantimpl "github.com/tcw165/fintech-fun/graph/clients/qdrant/impl"
	"github.com/tcw165/fintech-fun/graph/embed"
	"github.com/tcw165/fintech-fun/graph/embed/lexical"
	"go.uber.org/fx"
)

// Module is the api_server Fx graph. Clients are decided at boot.
func Module() fx.Option {
	return fx.Module(
		"api_server",
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
	lc fx.Lifecycle,
) neo4j.Client {
	driver, err := neo4jimpl.OpenFromEnv()
	if err != nil {
		log.Printf("neo4j unavailable: %v", err)
		return nil
	}
	lc.Append(fx.Hook{
		OnStop: driver.Close,
	})
	return driver
}

func provide_vector_db() qdrant.Client {
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

// Boot starts the api_server Fx graph.
func Boot(ctx context.Context) (*AppContext, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var app *AppContext
	fx_app := fx.New(
		Module(),
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
