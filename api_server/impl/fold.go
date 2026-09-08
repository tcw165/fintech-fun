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
		return contract.FoldResponse{}, errNoGraph
	}
	rows, err := neo4j.Fold(n.GraphClient, q, qty)
	if err != nil {
		return contract.FoldResponse{}, err
	}
	return contract.FoldResponse{Status: "ok", Q: q, Qty: qty, Rows: rows}, nil
}

type foldError string

func (e foldError) Error() string { return string(e) }

const errNoGraph foldError = "fold requires an injected Neo4j client"
