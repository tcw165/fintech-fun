package impl

import (
	"strings"

	"github.com/tcw165/fintech-fun/api_server/contract"
	"github.com/tcw165/fintech-fun/graph/clients/qdrant"
	"github.com/tcw165/fintech-fun/graph/embed"
)

// QdrantSearch embeds the query and ranks headlines through the injected client.
type QdrantSearch struct {
	VectorsClient qdrant.Client
	Embedder      embed.Embedder
}

func (s QdrantSearch) Search(query string) (contract.SearchResponse, error) {
	if s.VectorsClient == nil || s.Embedder == nil {
		return contract.SearchResponse{}, err_no_search
	}
	query = strings.Join(strings.Fields(query), " ")
	vecs, err := s.Embedder.Embed([]string{query})
	if err != nil {
		return contract.SearchResponse{}, err
	}
	hits, err := qdrant.SearchHeadlines(s.VectorsClient, vecs[0], 5)
	if err != nil {
		return contract.SearchResponse{}, err
	}
	return contract.SearchResponse{Status: "ok", Query: query, Hits: hits}, nil
}

const err_no_search fold_error = "search requires injected Qdrant and Embedder"
