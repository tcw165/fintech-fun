// Package parser splits tracker text into dated rows. Child of hood_events.
package parser

import (
	"regexp"
	"strings"
	"time"

	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
)

var (
	dateHeading  = regexp.MustCompile(`(?m)^(?:#{1,3}\s+)?(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d{1,2}),\s+(\d{4})\s*$`)
	tickerPrefix = regexp.MustCompile(`^(.+?)\s+\(([A-Z0-9][A-Z0-9.^]*)\)\s+`)
	hrSplit      = regexp.MustCompile(`\n\s*---+\s*\n`)
	htmlHeading  = regexp.MustCompile(`(?i)</?h[1-6][^>]*>`)
	htmlHR       = regexp.MustCompile(`(?i)<hr\s*/?>`)
	htmlPOpen    = regexp.MustCompile(`(?i)<p[^>]*>`)
	htmlPClose   = regexp.MustCompile(`(?i)</p>`)
	htmlBR       = regexp.MustCompile(`(?i)<br\s*/?>`)
	htmlTag      = regexp.MustCompile(`<[^>]+>`)
	whitespace   = regexp.MustCompile(`\s+`)
)

var months = map[string]time.Month{
	"January": 1, "February": 2, "March": 3, "April": 4, "May": 5, "June": 6,
	"July": 7, "August": 8, "September": 9, "October": 10, "November": 11, "December": 12,
}

func ParseDateHeading(text string) (time.Time, bool) {
	match := dateHeading.FindStringSubmatch(strings.TrimSpace(text))
	if match == nil {
		return time.Time{}, false
	}
	return parseMonthDayYear(match[1], match[2], match[3]), true
}

func ParseCompanyTicker(headline string) (string, string) {
	match := tickerPrefix.FindStringSubmatch(strings.TrimSpace(headline))
	if match == nil {
		return "", ""
	}
	return strings.TrimSpace(match[1]), match[2]
}

func normalize(text string) string {
	text = htmlHeading.ReplaceAllString(text, "\n")
	text = htmlHR.ReplaceAllString(text, "\n---\n")
	text = htmlPOpen.ReplaceAllString(text, "\n")
	text = htmlPClose.ReplaceAllString(text, "\n")
	text = htmlBR.ReplaceAllString(text, "\n")
	text = htmlTag.ReplaceAllString(text, "")
	return strings.ReplaceAll(text, "\r\n", "\n")
}

func parseMonthDayYear(month, day, year string) time.Time {
	var d, y int
	_, _ = parseInt(day, &d)
	_, _ = parseInt(year, &y)
	return time.Date(y, months[month], d, 0, 0, 0, 0, time.UTC)
}

func parseInt(s string, dest *int) (int, error) {
	n := 0
	for _, r := range s {
		n = n*10 + int(r-'0')
	}
	*dest = n
	return n, nil
}

func ParseTracker(text string) []hood_events.TrackerRow {
	normalized := normalize(text)
	matches := dateHeading.FindAllStringSubmatchIndex(normalized, -1)
	var rows []hood_events.TrackerRow
	seen := map[string]bool{}
	for i, loc := range matches {
		groups := dateHeading.FindStringSubmatch(normalized[loc[0]:loc[1]])
		on := parseMonthDayYear(groups[1], groups[2], groups[3])
		start := loc[1]
		end := len(normalized)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		body := strings.TrimSpace(normalized[start:end])
		if body == "" {
			continue
		}
		chunks := splitNonEmpty(hrSplit.Split(body, -1))
		if len(chunks) == 1 && strings.Contains(chunks[0], "\n\n") {
			chunks = splitNonEmpty(regexp.MustCompile(`\n{2,}`).Split(chunks[0], -1))
		}
		for _, headline := range chunks {
			headline = strings.TrimSpace(whitespace.ReplaceAllString(headline, " "))
			if headline == "" || dateHeading.MatchString(headline) {
				continue
			}
			key := on.Format("2006-01-02") + "|" + headline
			if seen[key] {
				continue
			}
			seen[key] = true
			company, ticker := ParseCompanyTicker(headline)
			rows = append(rows, hood_events.TrackerRow{Date: on, Headline: headline, Company: company, Ticker: ticker})
		}
	}
	return rows
}

func splitNonEmpty(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

type ParseResult struct {
	Status string         `json:"status"`
	Count  int            `json:"count"`
	Rows   []hood_events.TrackerRow `json:"rows"`
}

func ParseTrackerPage(text string) ParseResult {
	rows := ParseTracker(text)
	return ParseResult{Status: "success", Count: len(rows), Rows: rows}
}
