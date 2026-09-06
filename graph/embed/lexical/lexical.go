// Package lexical is a 384-d signed-hash embedder. Child of //graph/embed.
package lexical

import (
	"hash/fnv"
	"math"
	"strings"
	"unicode"

	"github.com/tcw165/fintech-fun/graph/embed"
)

type Embedder struct{}

func New() Embedder { return Embedder{} }

func (Embedder) Embed(texts []string) ([][]float64, error) {
	out := make([][]float64, len(texts))
	for i, text := range texts {
		out[i] = vector(text)
	}
	return out, nil
}

func vector(text string) []float64 {
	vals := make([]float64, embed.Size)
	for _, token := range tokens(text) {
		sum := fnv.New64a()
		_, _ = sum.Write([]byte(token))
		h := sum.Sum64()
		for dim := 0; dim < embed.Size; dim++ {
			bit := (h >> (uint(dim) % 64)) & 1
			if bit == 1 {
				vals[dim]++
			} else {
				vals[dim]--
			}
			h = h*6364136223846793005 + 1
		}
	}
	var norm float64
	for _, v := range vals {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if norm == 0 {
		return vals
	}
	for i := range vals {
		vals[i] /= norm
	}
	return vals
}

func tokens(text string) []string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field) > 1 {
			out = append(out, field)
		}
	}
	return out
}

var _ embed.Embedder = Embedder{}
