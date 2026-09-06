// Package qdrant is the Qdrant contract: Client + collection/point protocol. No HTTP client.
package qdrant

type Client interface {
	PutCollection(name string, body map[string]any) (map[string]any, error)
	Upsert(name string, body map[string]any) (map[string]any, error)
	Search(name string, body map[string]any) (map[string]any, error)
}
