package classify

import (
	"sort"
	"strings"

	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
)

// Skip reasons we drop on purpose. CUSIP-only is the main one.
// Also dropped: reorganizations with two deliverable tickers, escrow CUSIPs,
// and remaining unclassified English. waiting→later cashed_out is a refresh concern.
const SkipNotes = `Dropped on purpose:
- CUSIP-only / 1:1 CUSIP change (no ticker, split, or name)
- Multi-name reorganizations that mint two new tickers
- Escrow / contingent CUSIP rights
- Unclassified English that is not a closed Event.kind
`

func SkipReason(headline string) string {
	lower := strings.ToLower(headline)
	if strings.Contains(lower, "cusip") && !strings.Contains(lower, "ticker") && !strings.Contains(lower, "split") && !strings.Contains(lower, "name") {
		return "cusip"
	}
	if strings.Contains(lower, "reorganization") {
		return "reorganization"
	}
	if strings.Contains(lower, "escrow cusip") {
		return "escrow"
	}
	return "unclassified"
}

type GoldReport struct {
	Status    string         `json:"status"`
	Rows      int            `json:"rows"`
	Events    int            `json:"events"`
	Written   int            `json:"written"`
	Skipped   int            `json:"skipped"`
	SkipRate  float64        `json:"skip_rate"`
	Kinds     map[string]int `json:"kinds"`
	SkipWhy   map[string]int `json:"skip_reasons"`
	SkipNotes string         `json:"skip_notes"`
}

func Report(rows []hood_events.TrackerRow) GoldReport {
	events := ClassifyRows(rows)
	kinds := map[string]int{}
	why := map[string]int{}
	skipped := 0
	for _, event := range events {
		if event.Skip || event.Kind == "" {
			skipped++
			why[SkipReason(event.Headline)]++
			continue
		}
		kinds[event.Kind]++
	}
	rate := 0.0
	if len(events) > 0 {
		rate = float64(skipped) / float64(len(events))
	}
	return GoldReport{
		Status: "ok", Rows: len(rows), Events: len(events),
		Written: len(events) - skipped, Skipped: skipped, SkipRate: rate,
		Kinds: kinds, SkipWhy: why, SkipNotes: SkipNotes,
	}
}

func SortedKindNames(kinds map[string]int) []string {
	names := make([]string, 0, len(kinds))
	for name := range kinds {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func MustHitLive(row hood_events.TrackerRow, wantKind string) bool {
	for _, event := range ClassifyRow(row) {
		if !event.Skip && event.Kind == wantKind {
			return true
		}
	}
	return false
}
