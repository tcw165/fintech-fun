package agent

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/agent/contract"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/dedup"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/parser"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/testdata"
	qdrantimpl "github.com/tcw165/fintech-fun/graph/clients/qdrant/impl"
	"github.com/tcw165/fintech-fun/graph/embed/lexical"
)

type memory_store struct {
	hash                string
	prefix              []string
	ids                 map[string]bool
	event_upserts       int
	source_state_writes int
}

func new_memory_store() *memory_store {
	return &memory_store{ids: map[string]bool{}}
}

func (s *memory_store) Run(cypher string, params map[string]any) ([]map[string]any, error) {
	if strings.Contains(cypher, "MATCH (e:Event) RETURN e.id AS id") {
		var rows []map[string]any
		for id := range s.ids {
			rows = append(rows, map[string]any{"id": id})
		}
		return rows, nil
	}
	if strings.Contains(cypher, "AS history_prefix") && strings.Contains(cypher, "RETURN") && !strings.Contains(cypher, "SET") {
		return []map[string]any{{
			"page_sha256":    s.hash,
			"history_prefix": append([]string(nil), s.prefix...),
		}}, nil
	}
	if strings.Contains(cypher, "s.history_prefix = $history_prefix") {
		s.hash, _ = params["page_sha256"].(string)
		if prefix, ok := params["history_prefix"].([]string); ok {
			s.prefix = append([]string(nil), prefix...)
		}
		s.source_state_writes++
		return []map[string]any{{"ok": true}}, nil
	}
	if strings.Contains(cypher, "MERGE (e:Event") {
		if id, ok := params["id"].(string); ok {
			s.ids[id] = true
		}
		s.event_upserts++
	}
	return []map[string]any{{"ok": true}}, nil
}

type stub_fetcher struct {
	text string
}

func (s stub_fetcher) Fetch(string) (contract.FetchedPage, error) {
	return contract.FetchedPage{Status: 200, URL: "https://example.test/tracker", Bytes: len(s.text), Text: s.text}, nil
}

func TestIngestTextWritesFixtureAndPersistsPrefix(t *testing.T) {
	store := new_memory_store()
	result, err := New(store, &qdrantimpl.Recording{}, lexical.New(), nil).IngestText(testdata.TrackerSept2026)
	if err != nil {
		t.Fatal(err)
	}
	want_prefix := dedup.Fingerprints(dedup.Chronological(parser.ParseTracker(testdata.TrackerSept2026)))
	if result.Unchanged || result.Written < 12 || result.Pages != 2 || result.HistoryLen != len(want_prefix) {
		t.Fatalf("%+v", result)
	}
	if store.source_state_writes != 1 || len(store.prefix) != len(want_prefix) || store.prefix[0] != want_prefix[0] {
		t.Fatalf("prefix %v want %v writes=%d", store.prefix, want_prefix, store.source_state_writes)
	}
}

func TestIngestTextSkipsUnchangedHash(t *testing.T) {
	store := new_memory_store()
	agent := New(store, &qdrantimpl.Recording{}, lexical.New(), nil)
	if _, err := agent.IngestText(testdata.TrackerSept2026); err != nil {
		t.Fatal(err)
	}
	upserts := store.event_upserts
	writes := store.source_state_writes
	result, err := agent.IngestText(testdata.TrackerSept2026)
	if err != nil || !result.Unchanged || result.Written != 0 || result.Backfilled || store.event_upserts != upserts || store.source_state_writes != writes {
		t.Fatalf("%+v upserts=%d writes=%d err=%v", result, store.event_upserts, store.source_state_writes, err)
	}
}

func TestIngestTextBackfillsEmptyPrefixOnSameHash(t *testing.T) {
	store := new_memory_store()
	agent := New(store, nil, nil, nil)
	if _, err := agent.IngestText(testdata.TrackerSept2026); err != nil {
		t.Fatal(err)
	}
	store.prefix = nil
	store.source_state_writes = 0
	result, err := agent.IngestText(testdata.TrackerSept2026)
	if err != nil || !result.Unchanged || !result.Backfilled || store.source_state_writes != 1 || len(store.prefix) == 0 {
		t.Fatalf("%+v prefix=%d writes=%d err=%v", result, len(store.prefix), store.source_state_writes, err)
	}
}

func TestIngestTextIngestsOnlyNewDaySuffix(t *testing.T) {
	store := new_memory_store()
	agent := New(store, &qdrantimpl.Recording{}, lexical.New(), nil)
	if _, err := agent.IngestText(testdata.TrackerSept2026); err != nil {
		t.Fatal(err)
	}
	upserts := store.event_upserts
	page := strings.Replace(testdata.TrackerSept2026, "## September 4, 2026", "## September 5, 2026\n\nNetflix (NFLX) performed a 10 for 1 Forward Split.\n\n---\n\nBlock (XYZ) performed a ticker change to XYZ.\n\n## September 4, 2026", 1)
	result, err := agent.IngestText(page)
	if err != nil {
		t.Fatal(err)
	}
	if result.Unchanged || result.Suffix != 2 || result.Pages != 1 || result.Created < 1 {
		t.Fatalf("%+v", result)
	}
	if store.event_upserts <= upserts {
		t.Fatalf("expected new event writes, before=%d after=%d", upserts, store.event_upserts)
	}
	if store.event_upserts-upserts > 2 {
		t.Fatalf("ingested more than suffix: delta=%d", store.event_upserts-upserts)
	}
}

func TestIngestTextPaginatesLargeDay(t *testing.T) {
	var b strings.Builder
	b.WriteString("## September 6, 2026\n\n")
	for i := 0; i < dedup.MaxPageRows+3; i++ {
		fmt.Fprintf(&b, "Pager Co (PG%02d) performed a 2 for 1 Forward Split.\n\n---\n\n", i)
	}
	store := new_memory_store()
	result, err := New(store, nil, nil, nil).IngestText(b.String())
	if err != nil {
		t.Fatal(err)
	}
	if result.Pages != 2 || result.Suffix != dedup.MaxPageRows+3 || result.Created != dedup.MaxPageRows+3 {
		t.Fatalf("%+v", result)
	}
}

func TestDryRunTextDoesNotWrite(t *testing.T) {
	store := new_memory_store()
	result, err := New(store, &qdrantimpl.Recording{}, lexical.New(), nil).DryRunText(testdata.TrackerSept2026)
	if err != nil {
		t.Fatal(err)
	}
	if !result.DryRun || result.Written < 12 || result.Unchanged {
		t.Fatalf("%+v", result)
	}
	if store.event_upserts != 0 || store.source_state_writes != 0 {
		t.Fatalf("writes upserts=%d source=%d", store.event_upserts, store.source_state_writes)
	}
}

func TestDryRunTextWorksOffline(t *testing.T) {
	result, err := New(nil, nil, nil, nil).DryRunText(testdata.TrackerSept2026)
	if err != nil || !result.DryRun || result.Written < 12 {
		t.Fatalf("%+v %v", result, err)
	}
}

func TestDryRunLiveUsesFetcher(t *testing.T) {
	store := new_memory_store()
	result, err := New(
		store,
		&qdrantimpl.Recording{},
		lexical.New(),
		stub_fetcher{text: testdata.TrackerSept2026},
	).DryRunLive("")
	if err != nil || !result.DryRun || result.Written < 12 || store.event_upserts != 0 {
		t.Fatalf("%+v upserts=%d err=%v", result, store.event_upserts, err)
	}
}

func TestRefreshUsesFetcherAndWaitingPolicy(t *testing.T) {
	store := new_memory_store()
	result, err := New(store, &qdrantimpl.Recording{}, lexical.New(), stub_fetcher{text: testdata.TrackerSept2026}).Refresh("")
	if err != nil || !result.Refresh || result.WaitingPolicy != contract.WaitingPolicy || result.Written < 12 {
		t.Fatalf("%+v %v", result, err)
	}
}
