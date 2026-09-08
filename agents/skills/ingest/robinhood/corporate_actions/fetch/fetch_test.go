package fetch

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const fixture = `<html><h2>September 4, 2026</h2><p>LivePerson (LPSN) performed a stock merger.</p></html>`

func TestFetchTrackerReadsHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Fatal("missing UA")
		}
		_, _ = w.Write([]byte(fixture))
	}))
	t.Cleanup(server.Close)
	result, err := FetchTracker(server.Client(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != 200 || !strings.Contains(result.Text, "LivePerson") {
		t.Fatalf("%+v", result)
	}
}

func TestFetchTrackerRetriesThenFails(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)
	if _, err := FetchTracker(server.Client(), server.URL); err == nil {
		t.Fatal("expected error")
	}
	if hits != max_attempts {
		t.Fatalf("hits %d", hits)
	}
}
