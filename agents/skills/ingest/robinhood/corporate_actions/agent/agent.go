// Package agent is the custom corporate-actions ingest agent. Child of contract.
package agent

import (
	hood_events "github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/agent/contract"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/classify"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/dedup"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/ingest"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/parser"
	"github.com/tcw165/fintech-fun/graph"
	graphneo4j "github.com/tcw165/fintech-fun/graph/clients/neo4j"
	graphqdrant "github.com/tcw165/fintech-fun/graph/clients/qdrant"
	"github.com/tcw165/fintech-fun/graph/embed"
)

type Agent struct {
	graph_db  graphneo4j.Client
	vector_db graphqdrant.Client
	embedder  embed.Embedder
	fetcher   contract.Fetcher
}

func New(graph_db graphneo4j.Client, vector_db graphqdrant.Client, embedder embed.Embedder, fetcher contract.Fetcher) *Agent {
	if fetcher == nil {
		fetcher = LiveFetcher{}
	}
	return &Agent{graph_db: graph_db, vector_db: vector_db, embedder: embedder, fetcher: fetcher}
}

var _ contract.Agent = (*Agent)(nil)

func (a *Agent) IngestLive(url string) (contract.Result, error) {
	return a.ingest_live(url, false)
}

func (a *Agent) DryRunLive(url string) (contract.Result, error) {
	return a.ingest_live(url, true)
}

func (a *Agent) ingest_live(url string, dry_run bool) (contract.Result, error) {
	page, err := a.fetcher.Fetch(url)
	if err != nil {
		return contract.Result{}, err
	}
	return a.ingest_text(page.Text, dry_run)
}

func (a *Agent) Refresh(url string) (contract.Result, error) {
	result, err := a.IngestLive(url)
	if err != nil {
		return result, err
	}
	result.Refresh = true
	result.WaitingPolicy = contract.WaitingPolicy
	return result, nil
}

func (a *Agent) IngestText(text string) (contract.Result, error) {
	return a.ingest_text(text, false)
}

func (a *Agent) DryRunText(text string) (contract.Result, error) {
	return a.ingest_text(text, true)
}

func (a *Agent) ingest_text(text string, dry_run bool) (contract.Result, error) {
	hash := ingest.PageSHA256(text)
	result := contract.Result{PageSHA256: hash, DryRun: dry_run}
	if !dry_run {
		if err := graphneo4j.ApplyConstraints(a.graph_db); err != nil {
			return result, err
		}
	}
	var watermark graphneo4j.SourceWatermark
	if a.graph_db != nil {
		got, err := graphneo4j.ReadSource(a.graph_db, graph.IngestSourceCorporateActions)
		if err != nil {
			if !dry_run {
				return result, err
			}
		} else {
			watermark = got
		}
	}
	rows := dedup.Chronological(parser.ParseTracker(text))
	fingerprints := dedup.Fingerprints(rows)
	result.HistoryLen = len(fingerprints)
	if watermark.PageSHA256 != "" && watermark.PageSHA256 == hash {
		result.Unchanged = true
		result.Overlap = len(fingerprints)
		if !dry_run && len(watermark.HistoryPrefix) == 0 {
			if err := graphneo4j.UpsertSourceState(a.graph_db, graph.IngestSourceCorporateActions, hash, fingerprints); err != nil {
				return result, err
			}
			result.Backfilled = true
		}
		return result, nil
	}
	overlap := dedup.FindPrefixOverlap(watermark.HistoryPrefix, fingerprints)
	suffix := dedup.Suffix(rows, overlap)
	pages := dedup.Pages(suffix)
	result.Overlap = overlap
	result.Suffix = len(suffix)
	result.Pages = len(pages)
	if dry_run {
		return a.plan_pages(result, pages), nil
	}
	for _, page := range pages {
		stats, err := ingest.IngestClassified(a.graph_db, classify.ClassifyRows(page), a.vector_db, a.embedder)
		if err != nil {
			return result, err
		}
		result.Written += stats.Written
		result.Created += stats.Created
		result.Duplicates += stats.Duplicates
		result.Skipped += stats.Skipped
		result.Companies = append_unique(result.Companies, stats.Companies)
		result.Stocks = append_unique(result.Stocks, stats.Stocks)
	}
	if err := graphneo4j.UpsertSourceState(a.graph_db, graph.IngestSourceCorporateActions, hash, fingerprints); err != nil {
		return result, err
	}
	return result, nil
}

func (a *Agent) plan_pages(result contract.Result, pages [][]hood_events.TrackerRow) contract.Result {
	existing := map[string]bool{}
	if a.graph_db != nil {
		if ids, err := graphneo4j.ListEventIDs(a.graph_db); err == nil {
			existing = ids
		}
	}
	for _, page := range pages {
		for _, item := range classify.ClassifyRows(page) {
			_, _, event, ok := ingest.ToGraph(item)
			if !ok {
				result.Skipped++
				continue
			}
			if existing[event.ID()] {
				result.Duplicates++
				continue
			}
			result.Written++
			result.Created++
		}
	}
	return result
}

func append_unique(dst, src []string) []string {
	seen := map[string]bool{}
	for _, item := range dst {
		seen[item] = true
	}
	for _, item := range src {
		if seen[item] {
			continue
		}
		seen[item] = true
		dst = append(dst, item)
	}
	return dst
}
