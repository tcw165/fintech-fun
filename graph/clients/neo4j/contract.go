// Package neo4j is the Neo4j contract: Client + Cypher. No driver.
package neo4j

type Client interface {
	Run(cypher string, params map[string]any) ([]map[string]any, error)
}
