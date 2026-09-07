// Package dedup fingerprints tracker rows and finds the uningested suffix.
// Child of hood_events. Pure data; no graph I/O.
package dedup

import (
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
)

const MaxPageRows = 32

func Fingerprint(row hood_events.TrackerRow) string {
	return row.Date.Format("2006-01-02") + "|" + row.Ticker + "|" + row.Headline
}

func Fingerprints(rows []hood_events.TrackerRow) []string {
	out := make([]string, len(rows))
	for i, row := range rows {
		out[i] = Fingerprint(row)
	}
	return out
}

// Chronological turns a newest-first page into oldest-first history.
// Days appear newest-first; rows within a day are reversed to match that order.
func Chronological(rows []hood_events.TrackerRow) []hood_events.TrackerRow {
	type day struct {
		key  string
		rows []hood_events.TrackerRow
	}
	var days []day
	index := map[string]int{}
	for _, row := range rows {
		key := row.Date.Format("2006-01-02")
		if i, ok := index[key]; ok {
			days[i].rows = append(days[i].rows, row)
			continue
		}
		index[key] = len(days)
		days = append(days, day{key: key, rows: []hood_events.TrackerRow{row}})
	}
	out := make([]hood_events.TrackerRow, 0, len(rows))
	for i := len(days) - 1; i >= 0; i-- {
		group := days[i].rows
		for j := len(group) - 1; j >= 0; j-- {
			out = append(out, group[j])
		}
	}
	return out
}

func FindPrefixOverlap(stored, fetched []string) int {
	n := len(stored)
	if len(fetched) < n {
		n = len(fetched)
	}
	i := 0
	for i < n && stored[i] == fetched[i] {
		i++
	}
	return i
}

func Suffix(rows []hood_events.TrackerRow, overlap int) []hood_events.TrackerRow {
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= len(rows) {
		return nil
	}
	return rows[overlap:]
}

// Pages splits chronological rows into ingest pages: one calendar day,
// further split when a day exceeds MaxPageRows.
func Pages(rows []hood_events.TrackerRow) [][]hood_events.TrackerRow {
	if len(rows) == 0 {
		return nil
	}
	var pages [][]hood_events.TrackerRow
	start := 0
	for i := 1; i <= len(rows); i++ {
		same_day := i < len(rows) && rows[i].Date.Equal(rows[start].Date)
		full := i-start >= MaxPageRows
		if same_day && !full {
			continue
		}
		pages = append(pages, rows[start:i])
		start = i
	}
	return pages
}
