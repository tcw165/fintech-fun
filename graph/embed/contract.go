// Package embed is the headline embedding contract. 384-d, no model weights.
package embed

const Size = 384

type Embedder interface {
	Embed(texts []string) ([][]float64, error)
}
