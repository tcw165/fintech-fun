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

func issuer_params(company graph.Company, stock graph.Stock) map[string]any {
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
	return client.Run(UpsertIssuerCypher(), issuer_params(company, graph.Stock{Ticker: company.Name}))
}

func UpsertStock(client Client, company graph.Company, stock graph.Stock) ([]map[string]any, error) {
	return client.Run(UpsertIssuerCypher(), issuer_params(company, stock))
}

func UpsertEvent(client Client, event graph.Event) ([]map[string]any, error) {
	var you_now_hold any
	if event.YouNowHold != "" {
		you_now_hold = event.YouNowHold
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
		"you_now_hold":     you_now_hold,
	})
}

type IngestResult struct {
	Company string
	Ticker  string
	Events  int
}

type SourceWatermark struct {
	ID            string
	PageSHA256    string
	FetchedAt     any
	HistoryPrefix []string
}

func ListEventIDs(client Client) (map[string]bool, error) {
	rows, err := client.Run(ListEventIDsCypher(), map[string]any{})
	if err != nil {
		return nil, err
	}
	ids := map[string]bool{}
	for _, row := range rows {
		if id, ok := row["id"].(string); ok && id != "" {
			ids[id] = true
		}
	}
	return ids, nil
}

func ReadSource(client Client, id string) (SourceWatermark, error) {
	rows, err := client.Run(ReadSourceCypher(), map[string]any{"id": id})
	if err != nil {
		return SourceWatermark{ID: id}, err
	}
	out := SourceWatermark{ID: id}
	if len(rows) == 0 {
		return out, nil
	}
	if hash, ok := rows[0]["page_sha256"].(string); ok {
		out.PageSHA256 = hash
	}
	out.FetchedAt = rows[0]["fetched_at"]
	out.HistoryPrefix = as_string_slice(rows[0]["history_prefix"])
	return out, nil
}

func as_string_slice(value any) []string {
	switch items := value.(type) {
	case []string:
		return append([]string(nil), items...)
	case []any:
		out := make([]string, 0, len(items))
		for _, item := range items {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func UpsertSource(client Client, id, page_sha256 string) error {
	_, err := client.Run(UpsertSourceCypher(), map[string]any{
		"id":           id,
		"page_sha256":  page_sha256,
	})
	return err
}

func UpsertSourceState(client Client, id, page_sha256 string, history_prefix []string) error {
	if history_prefix == nil {
		history_prefix = []string{}
	}
	_, err := client.Run(UpsertSourceStateCypher(), map[string]any{
		"id":             id,
		"page_sha256":    page_sha256,
		"history_prefix": history_prefix,
	})
	return err
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
