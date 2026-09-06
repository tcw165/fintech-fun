package impl

import (
	"github.com/tcw165/fintech-fun/api_server/contract"
	"github.com/tcw165/fintech-fun/graph/clients/qdrant"
	"github.com/tcw165/fintech-fun/graph/embed"
)

// QdrantSearch embeds the query and ranks headlines through the injected client.
type QdrantSearch struct {
	Vectors qdrant.Client
	Embed   embed.Embedder
}

func (s QdrantSearch) Search(query string) (contract.SearchResponse, error) {
	if s.Vectors == nil || s.Embed == nil {
		return contract.SearchResponse{}, errNoSearch
	}
	vecs, err := s.Embed.Embed([]string{query})
	if err != nil {
		return contract.SearchResponse{}, err
	}
	hits, err := qdrant.SearchHeadlines(s.Vectors, vecs[0], 5)
	if err != nil {
		return contract.SearchResponse{}, err
	}
	return contract.SearchResponse{Status: "ok", Query: query, Hits: hits}, nil
}

const errNoSearch foldError = "search requires injected Qdrant and Embedder"
