package impl

import (
	"github.com/tcw165/fintech-fun/api_server/contract"
	"github.com/tcw165/fintech-fun/graph/clients/neo4j"
)

// Neo4jFold runs Notion FoldCypher through the injected client.
type Neo4jFold struct {
	GraphClient neo4j.Client
}

func (n Neo4jFold) Fold(q string, qty float64) (contract.FoldResponse, error) {
	if n.GraphClient == nil {
		return contract.FoldResponse{}, err_no_graph
	}
	rows, err := neo4j.Fold(n.GraphClient, q, qty)
	if err != nil {
		return contract.FoldResponse{}, err
	}
	return contract.FoldResponse{Status: "ok", Q: q, Qty: qty, Rows: rows}, nil
}

type fold_error string

func (e fold_error) Error() string { return string(e) }

const err_no_graph fold_error = "fold requires an injected Neo4j client"
