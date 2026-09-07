// Package agent is the custom corporate-actions ingest agent. Child of contract.
package agent

import (
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
	page, err := a.fetcher.Fetch(url)
	if err != nil {
		return contract.Result{}, err
	}
	return a.IngestText(page.Text)
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
	hash := ingest.PageSHA256(text)
	result := contract.Result{PageSHA256: hash}
	if err := graphneo4j.ApplyConstraints(a.graph_db); err != nil {
		return result, err
	}
	watermark, err := graphneo4j.ReadSource(a.graph_db, graph.IngestSourceCorporateActions)
	if err != nil {
		return result, err
	}
	rows := dedup.Chronological(parser.ParseTracker(text))
	fingerprints := dedup.Fingerprints(rows)
	result.HistoryLen = len(fingerprints)
	if watermark.PageSHA256 != "" && watermark.PageSHA256 == hash {
		result.Unchanged = true
		result.Overlap = len(fingerprints)
		if len(watermark.HistoryPrefix) == 0 {
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
