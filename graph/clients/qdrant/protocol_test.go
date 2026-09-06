package qdrant

import (
	"testing"

	"github.com/tcw165/fintech-fun/graph/examples"
)

type stubClient struct {
	collections []string
	upserts     int
}

func (c *stubClient) PutCollection(name string, body map[string]any) (map[string]any, error) {
	c.collections = append(c.collections, name)
	return map[string]any{"status": "ok"}, nil
}

func (c *stubClient) Upsert(name string, body map[string]any) (map[string]any, error) {
	c.upserts++
	return map[string]any{"status": "ok", "name": name, "body": body}, nil
}

func (c *stubClient) Search(name string, body map[string]any) (map[string]any, error) {
	return map[string]any{"status": "ok", "name": name, "body": body}, nil
}

func TestCollectionIsNamedCosine384(t *testing.T) {
	body := CollectionBody()
	vectors := body["vectors"].(map[string]any)[VectorName].(map[string]any)
	if vectors["size"] != VectorSize || vectors["distance"] != "Cosine" {
		t.Fatalf("%v", vectors)
	}
}

func TestZeroVectorMatchesSize(t *testing.T) {
	if len(ZeroVector()) != VectorSize {
		t.Fatal(len(ZeroVector()))
	}
}

func TestPointIDIsStableUUID5(t *testing.T) {
	_, _, events := examples.NFLX()
	first := PointID(events[0])
	if first != PointID(events[0]) || len(first) != 36 {
		t.Fatalf("%q", first)
	}
	_, _, other := examples.LPSN()
	if PointID(events[0]) == PointID(other[0]) {
		t.Fatal("ids collided")
	}
}

func TestEventPayloadRetailFields(t *testing.T) {
	_, _, events := examples.LPSN()
	payload := EventPayload(events[0])
	if payload["kind"] != "now_different_stock" || payload["happened_to"] != "LPSN" || payload["you_now_hold"] != "SOUN" {
		t.Fatalf("%v", payload)
	}
}

func TestEventPointRejectsWrongDim(t *testing.T) {
	_, _, events := examples.NFLX()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	EventPoint(events[0], []float64{0, 1})
}

func TestEnsureAndUpsertUseCorporateActions(t *testing.T) {
	client := &stubClient{}
	if _, err := EnsureCollection(client); err != nil {
		t.Fatal(err)
	}
	_, _, events := examples.NFLX()
	if _, err := UpsertEvents(client, events, nil); err != nil {
		t.Fatal(err)
	}
	if len(client.collections) != 1 || client.collections[0] != Collection || client.upserts != 1 {
		t.Fatalf("%+v", client)
	}
}
