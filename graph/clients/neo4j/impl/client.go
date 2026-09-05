// Package impl is the Neo4j Client implementation. Child of //graph/clients/neo4j.
package impl

import "github.com/tcw165/fintech-fun/graph/clients/neo4j"

const (
	URI      = "bolt://localhost:7687"
	User     = "neo4j"
	Password = "fintechfun"
)

// Func adapts a callback to neo4j.Client. The official driver lands here later.
type Func func(cypher string, params map[string]any) ([]map[string]any, error)

func (f Func) Run(cypher string, params map[string]any) ([]map[string]any, error) {
	return f(cypher, params)
}

type Call struct {
	Cypher string
	Params map[string]any
}

type Recording struct {
	Calls []Call
}

func (r *Recording) Run(cypher string, params map[string]any) ([]map[string]any, error) {
	r.Calls = append(r.Calls, Call{Cypher: cypher, Params: params})
	return []map[string]any{{"ok": true}}, nil
}

var _ neo4j.Client = Func(nil)
var _ neo4j.Client = (*Recording)(nil)
