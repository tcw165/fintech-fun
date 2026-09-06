package neo4j

import "github.com/tcw165/fintech-fun/graph"

func ApplyConstraints(client Client) error {
	for _, statement := range Constraints {
		if _, err := client.Run(statement, map[string]any{}); err != nil {
			return err
		}
	}
	return nil
}

func issuerParams(company graph.Company, stock graph.Stock) map[string]any {
	aliases := company.AlsoKnownAs
	if aliases == nil {
		aliases = []string{}
	}
	former := stock.FormerTickers
	if former == nil {
		former = []string{}
	}
	return map[string]any{
		"ticker":         stock.Ticker,
		"former_tickers": former,
		"status":         string(stock.Status),
		"company_name":   company.Name,
		"also_known_as":  aliases,
	}
}

func UpsertCompany(client Client, company graph.Company) ([]map[string]any, error) {
	return client.Run(UpsertIssuerCypher(), issuerParams(company, graph.Stock{Ticker: company.Name}))
}

func UpsertStock(client Client, company graph.Company, stock graph.Stock) ([]map[string]any, error) {
	return client.Run(UpsertIssuerCypher(), issuerParams(company, stock))
}

func UpsertEvent(client Client, event graph.Event) ([]map[string]any, error) {
	var youNowHold any
	if event.YouNowHold != "" {
		youNowHold = event.YouNowHold
	}
	return client.Run(UpsertEventCypher(), map[string]any{
		"id":               event.ID(),
		"date":             event.Date.Format("2006-01-02"),
		"kind":             string(event.Kind),
		"headline":         event.Headline,
		"share_multiplier": event.ShareMultiplier,
		"cash_per_share":   event.CashPerShare,
		"keep_fractionals": event.KeepFractionals,
		"can_trade":        event.CanTrade,
		"happened_to":      event.HappenedTo,
		"you_now_hold":     youNowHold,
	})
}

type IngestResult struct {
	Company string
	Ticker  string
	Events  int
}

func IngestGraph(client Client, company graph.Company, stock graph.Stock, events []graph.Event) (IngestResult, error) {
	if _, err := UpsertStock(client, company, stock); err != nil {
		return IngestResult{}, err
	}
	for _, event := range events {
		if _, err := UpsertEvent(client, event); err != nil {
			return IngestResult{}, err
		}
	}
	return IngestResult{Company: company.Name, Ticker: stock.Ticker, Events: len(events)}, nil
}
