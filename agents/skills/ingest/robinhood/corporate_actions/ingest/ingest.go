// Package ingest maps classified rows onto graph contracts. Child of hood_events.
package ingest

import (
	"strings"

	"github.com/tcw165/fintech-fun/graph"
	graphneo4j "github.com/tcw165/fintech-fun/graph/clients/neo4j"
	graphqdrant "github.com/tcw165/fintech-fun/graph/clients/qdrant"
	"github.com/tcw165/fintech-fun/graph/embed"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/classify"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/parser"
)

func youNowHoldKinds() map[graph.EventKind]bool {
	return map[graph.EventKind]bool{
		graph.KindNowDifferentStock: true,
		graph.KindExtraStock:        true,
	}
}

func stockStatus(classified hood_events.ClassifiedEvent) graph.StockStatus {
	headline := strings.ToLower(classified.Headline)
	if classified.Kind == "worthless" {
		return graph.StatusWorthless
	}
	if classified.Kind == "cashed_out" || classified.Kind == "now_different_stock" {
		return graph.StatusGone
	}
	if strings.Contains(headline, "otc") {
		return graph.StatusOTC
	}
	return graph.StatusTradeable
}

func ToGraph(classified hood_events.ClassifiedEvent) (graph.Company, graph.Stock, graph.Event, bool) {
	if classified.Skip || classified.Kind == "" || classified.Ticker == "" {
		return graph.Company{}, graph.Stock{}, graph.Event{}, false
	}
	kind := graph.EventKind(classified.Kind)
	currentTicker := classified.Ticker
	var former []string
	if kind == graph.KindTickerChanged && classified.YouNowHold != "" {
		former = []string{classified.Ticker}
		currentTicker = classified.YouNowHold
	}
	name := classified.Company
	if classified.NewName != "" {
		name = classified.NewName
	}
	if name == "" {
		name = classified.Ticker
	}
	var aliases []string
	if classified.NewName != "" && classified.Company != "" && classified.Company != classified.NewName {
		aliases = []string{classified.Company}
	}
	youNowHold := ""
	if youNowHoldKinds()[kind] {
		youNowHold = classified.YouNowHold
	}
	return graph.Company{Name: name, AlsoKnownAs: aliases},
		graph.Stock{Ticker: currentTicker, FormerTickers: former, Status: stockStatus(classified)},
		graph.Event{
			Date:            classified.Date,
			Kind:            kind,
			Headline:        classified.Headline,
			HappenedTo:      currentTicker,
			ShareMultiplier: classified.ShareMultiplier,
			CashPerShare:    classified.CashPerShare,
			KeepFractionals: classified.KeepFractionals,
			CanTrade:        classified.CanTrade,
			YouNowHold:      youNowHold,
		}, true
}

type IngestStats struct {
	Written   int      `json:"written"`
	Skipped   int      `json:"skipped"`
	Companies []string `json:"companies"`
	Stocks    []string `json:"stocks"`
}

func IngestClassified(graphClient graphneo4j.Client, events []hood_events.ClassifiedEvent, vectors graphqdrant.Client, embedder embed.Embedder) (IngestStats, error) {
	stats := IngestStats{}
	seenCompany := map[string]bool{}
	seenStock := map[string]bool{}
	var graphEvents []graph.Event
	for _, classified := range events {
		company, stock, event, ok := ToGraph(classified)
		if !ok {
			stats.Skipped++
			continue
		}
		if _, err := graphneo4j.IngestGraph(graphClient, company, stock, []graph.Event{event}); err != nil {
			return stats, err
		}
		graphEvents = append(graphEvents, event)
		stats.Written++
		if !seenCompany[company.Name] {
			seenCompany[company.Name] = true
			stats.Companies = append(stats.Companies, company.Name)
		}
		if !seenStock[stock.Ticker] {
			seenStock[stock.Ticker] = true
			stats.Stocks = append(stats.Stocks, stock.Ticker)
		}
	}
	if vectors != nil && len(graphEvents) > 0 {
		if _, err := graphqdrant.EnsureCollection(vectors); err != nil {
			return stats, err
		}
		var vecs [][]float64
		if embedder != nil {
			texts := make([]string, len(graphEvents))
			for i, event := range graphEvents {
				texts[i] = event.Headline
			}
			embedded, err := embedder.Embed(texts)
			if err != nil {
				return stats, err
			}
			vecs = embedded
		}
		if _, err := graphqdrant.UpsertEvents(vectors, graphEvents, vecs); err != nil {
			return stats, err
		}
	}
	return stats, nil
}

type PlannedEvent struct {
	Company       string   `json:"company"`
	AlsoKnownAs   []string `json:"also_known_as"`
	Ticker        string   `json:"ticker"`
	FormerTickers []string `json:"former_tickers"`
	Status        string   `json:"status"`
	EventID       string   `json:"event_id"`
	Kind          string   `json:"kind"`
	YouNowHold    string   `json:"you_now_hold"`
}

type PlanResult struct {
	Status  string         `json:"status"`
	DryRun  bool           `json:"dry_run"`
	Written int            `json:"written"`
	Skipped int            `json:"skipped"`
	Events  []PlannedEvent `json:"events"`
}

func PlanIngest(text string) PlanResult {
	classified := classify.ClassifyRows(parser.ParseTracker(text))
	var planned []PlannedEvent
	skipped := 0
	for _, item := range classified {
		company, stock, event, ok := ToGraph(item)
		if !ok {
			skipped++
			continue
		}
		planned = append(planned, PlannedEvent{
			Company: company.Name, AlsoKnownAs: company.AlsoKnownAs,
			Ticker: stock.Ticker, FormerTickers: stock.FormerTickers,
			Status: string(stock.Status), EventID: event.ID(),
			Kind: string(event.Kind), YouNowHold: event.YouNowHold,
		})
	}
	return PlanResult{Status: "success", DryRun: true, Written: len(planned), Skipped: skipped, Events: planned}
}

func IngestText(graphClient graphneo4j.Client, text string, vectors graphqdrant.Client, embedder embed.Embedder) (IngestStats, error) {
	return IngestClassified(graphClient, classify.ClassifyRows(parser.ParseTracker(text)), vectors, embedder)
}
