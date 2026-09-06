// Package impl is the Qdrant Client implementation. Child of //graph/clients/qdrant.
package impl

import "github.com/tcw165/fintech-fun/graph/clients/qdrant"

const URL = "http://localhost:6333"

type Recording struct {
	Collections []string
	Upserts     int
}

func (c *Recording) PutCollection(name string, body map[string]any) (map[string]any, error) {
	c.Collections = append(c.Collections, name)
	return map[string]any{"status": "ok"}, nil
}

func (c *Recording) Upsert(name string, body map[string]any) (map[string]any, error) {
	c.Upserts++
	return map[string]any{"status": "ok", "name": name, "body": body}, nil
}

func (c *Recording) Search(name string, body map[string]any) (map[string]any, error) {
	return map[string]any{"status": "ok", "name": name, "result": []any{}}, nil
}

var _ qdrant.Client = (*Recording)(nil)
