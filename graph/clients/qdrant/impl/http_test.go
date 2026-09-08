package impl

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tcw165/fintech-fun/graph/clients/qdrant"
	"github.com/tcw165/fintech-fun/graph/examples"
)

func TestHTTPPutCollectionAndUpsert(t *testing.T) {
	var got_path string
	var got_body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got_path = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &got_body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","result":true}`))
	}))
	t.Cleanup(server.Close)

	client := NewHTTP(server.URL)
	if _, err := qdrant.EnsureCollection(client); err != nil {
		t.Fatal(err)
	}
	if got_path != "/collections/corporate_actions" {
		t.Fatalf("path %s", got_path)
	}
	vectors := got_body["vectors"].(map[string]any)["headline"].(map[string]any)
	if vectors["size"] != float64(qdrant.VectorSize) || vectors["distance"] != "Cosine" {
		t.Fatalf("%v", got_body)
	}

	_, _, events := examples.NFLX()
	if _, err := qdrant.UpsertEvents(client, events, nil); err != nil {
		t.Fatal(err)
	}
	if got_path != "/collections/corporate_actions/points" {
		t.Fatalf("path %s", got_path)
	}
	points := got_body["points"].([]any)
	if len(points) != 1 {
		t.Fatalf("%v", got_body)
	}
}

func TestHTTPSurfacesStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)
	_, err := NewHTTP(server.URL).PutCollection("x", map[string]any{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestURLFromEnv(t *testing.T) {
	t.Setenv("QDRANT_URL", "")
	if URLFromEnv() != URL {
		t.Fatal(URLFromEnv())
	}
	t.Setenv("QDRANT_URL", "http://qdrant:6333")
	if NewHTTPFromEnv().BaseURL() != "http://qdrant:6333" {
		t.Fatal(NewHTTPFromEnv().BaseURL())
	}
}
