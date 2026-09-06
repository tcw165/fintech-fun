package impl

import (
	"github.com/tcw165/fintech-fun/api_server/contract"
	"github.com/tcw165/fintech-fun/graph"
	"github.com/tcw165/fintech-fun/graph/clients/neo4j"
)

// Neo4jGraph surfaces Company/Stock/Event series and the ingest watermark.
type Neo4jGraph struct {
	Graph neo4j.Client
}

func (n Neo4jGraph) Series(q string) (contract.GraphResponse, error) {
	if n.Graph == nil {
		return contract.GraphResponse{}, errNoGraph
	}
	rows, err := neo4j.Series(n.Graph, q)
	if err != nil {
		return contract.GraphResponse{}, err
	}
	return contract.GraphResponse{Status: "ok", Q: q, Rows: rows}, nil
}

func (n Neo4jGraph) Source() (contract.SourceResponse, error) {
	if n.Graph == nil {
		return contract.SourceResponse{}, errNoGraph
	}
	watermark, err := neo4j.ReadSource(n.Graph, graph.IngestSourceCorporateActions)
	if err != nil {
		return contract.SourceResponse{}, err
	}
	return contract.SourceResponse{
		Status:     "ok",
		ID:         watermark.ID,
		PageSHA256: watermark.PageSHA256,
		FetchedAt:  watermark.FetchedAt,
	}, nil
}
