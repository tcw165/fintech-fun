// Package di wires ingest job process dependencies. Runnable composition root.
package di

import (
	"context"

	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/agent"
	"github.com/tcw165/fintech-fun/graph/clients/neo4j"
	"github.com/tcw165/fintech-fun/graph/clients/qdrant"
	"github.com/tcw165/fintech-fun/graph/embed"
)

type AppContext struct {
	graph_db  neo4j.Client
	vector_db qdrant.Client
	embedder  embed.Embedder
	closer    func(context.Context) error
}

func New(graph_db neo4j.Client, vector_db qdrant.Client, embedder embed.Embedder, closer func(context.Context) error) *AppContext {
	return &AppContext{
		graph_db:  graph_db,
		vector_db: vector_db,
		embedder:  embedder,
		closer:    closer,
	}
}

func (c *AppContext) Close(ctx context.Context) error {
	if c == nil || c.closer == nil {
		return nil
	}
	return c.closer(ctx)
}

func (c *AppContext) GraphClient() neo4j.Client {
	if c == nil {
		return nil
	}
	return c.graph_db
}

func (c *AppContext) VectorsClient() qdrant.Client {
	if c == nil {
		return nil
	}
	return c.vector_db
}

func (c *AppContext) Embedder() embed.Embedder {
	if c == nil {
		return nil
	}
	return c.embedder
}

func (c *AppContext) Agent() *agent.Agent {
	if c == nil {
		return agent.New(nil, nil, nil, nil)
	}
	return agent.New(c.graph_db, c.vector_db, c.embedder, agent.LiveFetcher{})
}
