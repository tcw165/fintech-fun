package qdrant

import (
	"crypto/sha1"
	"fmt"

	"github.com/tcw165/fintech-fun/graph"
)

const (
	Collection = "corporate_actions"
	VectorSize = 384
	VectorName = "headline"
)

func CollectionBody() map[string]any {
	return map[string]any{
		"vectors": map[string]any{
			VectorName: map[string]any{"size": VectorSize, "distance": "Cosine"},
		},
	}
}

func ZeroVector() []float64 {
	return make([]float64, VectorSize)
}

func PointID(event graph.Event) string {
	ns := []byte{0x6b, 0xa7, 0xb8, 0x11, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	h := sha1.New()
	h.Write(ns)
	h.Write([]byte("fintech-fun:" + event.ID()))
	sum := h.Sum(nil)[:16]
	sum[6] = (sum[6] & 0x0f) | 0x50
	sum[8] = (sum[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", sum[0:4], sum[4:6], sum[6:8], sum[8:10], sum[10:16])
}

func EventPayload(event graph.Event) map[string]any {
	var you_now_hold any
	if event.YouNowHold != "" {
		you_now_hold = event.YouNowHold
	}
	return map[string]any{
		"event_id":         event.ID(),
		"date":             event.Date.Format("2006-01-02"),
		"kind":             string(event.Kind),
		"headline":         event.Headline,
		"happened_to":      event.HappenedTo,
		"you_now_hold":     you_now_hold,
		"share_multiplier": event.ShareMultiplier,
		"cash_per_share":   event.CashPerShare,
	}
}

func EventPoint(event graph.Event, vector []float64) map[string]any {
	values := vector
	if values == nil {
		values = ZeroVector()
	}
	if len(values) != VectorSize {
		panic(fmt.Sprintf("vector must have %d dims, got %d", VectorSize, len(values)))
	}
	return map[string]any{
		"id":      PointID(event),
		"vector":  map[string]any{VectorName: values},
		"payload": EventPayload(event),
	}
}

func UpsertBody(events []graph.Event, vectors [][]float64) map[string]any {
	if vectors != nil && len(vectors) != len(events) {
		panic("vectors length must match events")
	}
	points := make([]map[string]any, 0, len(events))
	for i, event := range events {
		var vector []float64
		if vectors != nil {
			vector = vectors[i]
		}
		points = append(points, EventPoint(event, vector))
	}
	return map[string]any{"points": points}
}

func SearchBody(vector []float64, limit int) map[string]any {
	if limit <= 0 {
		limit = 5
	}
	if len(vector) != VectorSize {
		panic(fmt.Sprintf("vector must have %d dims, got %d", VectorSize, len(vector)))
	}
	return map[string]any{
		"vector":       map[string]any{"name": VectorName, "vector": vector},
		"limit":        limit,
		"with_payload": true,
	}
}

func SearchHeadlines(client Client, vector []float64, limit int) (map[string]any, error) {
	return client.Search(Collection, SearchBody(vector, limit))
}

func EnsureCollection(client Client) (map[string]any, error) {
	return client.PutCollection(Collection, CollectionBody())
}

func UpsertEvents(client Client, events []graph.Event, vectors [][]float64) (map[string]any, error) {
	return client.Upsert(Collection, UpsertBody(events, vectors))
}
