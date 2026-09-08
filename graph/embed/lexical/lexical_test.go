package lexical

import (
	"math"
	"testing"

	"github.com/tcw165/fintech-fun/graph/embed"
	"github.com/tcw165/fintech-fun/graph/examples"
)

func cosine(a, b []float64) float64 {
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func TestEmbedIs384AndNonZero(t *testing.T) {
	vecs, err := New().Embed([]string{"LivePerson stock merger"})
	if err != nil || len(vecs[0]) != embed.Size {
		t.Fatalf("%v %v", vecs, err)
	}
	var sum float64
	for _, v := range vecs[0] {
		sum += math.Abs(v)
	}
	if sum == 0 {
		t.Fatal("zero vector")
	}
}

func TestLivePersonQueryRanksLPSN(t *testing.T) {
	_, _, lpsn := examples.LPSN()
	_, _, nflx := examples.NFLX()
	query := "LivePerson stock merger"
	vecs, err := New().Embed([]string{query, lpsn[0].Headline, nflx[0].Headline})
	if err != nil {
		t.Fatal(err)
	}
	if cosine(vecs[0], vecs[1]) <= cosine(vecs[0], vecs[2]) {
		t.Fatalf("lpsn=%v nflx=%v", cosine(vecs[0], vecs[1]), cosine(vecs[0], vecs[2]))
	}
}

func TestEmbedCollapsesMultilineQuery(t *testing.T) {
	vecs, err := New().Embed([]string{
		"LivePerson stock merger",
		"LivePerson\nstock\nmerger",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cosine(vecs[0], vecs[1]) < 0.999 {
		t.Fatalf("newline query should match spaced query: %v", cosine(vecs[0], vecs[1]))
	}
}
