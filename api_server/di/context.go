// Package di wires api_server process dependencies. Runnable composition root.
package di

import (
	"context"
	"net/http"

	"github.com/tcw165/fintech-fun/api_server/contract"
	"github.com/tcw165/fintech-fun/api_server/impl"
	"github.com/tcw165/fintech-fun/graph/clients/neo4j"
	"github.com/tcw165/fintech-fun/graph/clients/qdrant"
	"github.com/tcw165/fintech-fun/graph/embed"
)

type AppContext struct {
	graph_db  neo4j.Client
	vector_db qdrant.Client
	embedder  embed.Embedder
	addr      string
	closer    func(context.Context) error
}

func New(graph_db neo4j.Client, vector_db qdrant.Client, embedder embed.Embedder, closer func(context.Context) error) *AppContext {
	return &AppContext{
		graph_db:  graph_db,
		vector_db: vector_db,
		embedder:  embedder,
		addr:      ":8080",
		closer:    closer,
	}
}

func (c *AppContext) Close(ctx context.Context) error {
	if c == nil || c.closer == nil {
		return nil
	}
	return c.closer(ctx)
}

func (c *AppContext) Addr() string {
	if c == nil || c.addr == "" {
		return ":8080"
	}
	return c.addr
}

func (c *AppContext) Handler() http.Handler {
	var folder contract.Folder
	var grapher contract.Grapher
	if c != nil && c.graph_db != nil {
		folder = impl.Neo4jFold{Graph: c.graph_db}
		grapher = impl.Neo4jGraph{Graph: c.graph_db}
	}
	var searcher contract.Searcher
	if c != nil && c.vector_db != nil && c.embedder != nil {
		searcher = impl.QdrantSearch{Vectors: c.vector_db, Embed: c.embedder}
	}
	return impl.New(impl.StaticOK{}, folder, searcher, grapher)
}
