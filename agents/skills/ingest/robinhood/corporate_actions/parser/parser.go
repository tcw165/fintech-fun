// Package parser splits tracker text into dated rows. Child of hood_events.
package parser

import (
	"regexp"
	"strings"
	"time"

	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
)

var (
	date_heading  = regexp.MustCompile(`(?m)^(?:#{1,3}\s+)?(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d{1,2}),\s+(\d{4})\s*$`)
	ticker_prefix = regexp.MustCompile(`^(.+?)\s+\(([A-Z0-9][A-Z0-9.^]*)\)\s+`)
	hr_split      = regexp.MustCompile(`\n\s*---+\s*\n`)
	html_heading  = regexp.MustCompile(`(?i)</?h[1-6][^>]*>`)
	html_hr       = regexp.MustCompile(`(?i)<hr\s*/?>`)
	html_p_open    = regexp.MustCompile(`(?i)<p[^>]*>`)
	html_p_close   = regexp.MustCompile(`(?i)</p>`)
	html_br       = regexp.MustCompile(`(?i)<br\s*/?>`)
	html_tag      = regexp.MustCompile(`<[^>]+>`)
	whitespace   = regexp.MustCompile(`\s+`)
)

var months = map[string]time.Month{
	"January": 1, "February": 2, "March": 3, "April": 4, "May": 5, "June": 6,
	"July": 7, "August": 8, "September": 9, "October": 10, "November": 11, "December": 12,
}

func ParseDateHeading(text string) (time.Time, bool) {
	match := date_heading.FindStringSubmatch(strings.TrimSpace(text))
	if match == nil {
		return time.Time{}, false
	}
	return parse_month_day_year(match[1], match[2], match[3]), true
}

func ParseCompanyTicker(headline string) (string, string) {
	match := ticker_prefix.FindStringSubmatch(strings.TrimSpace(headline))
	if match == nil {
		return "", ""
	}
	return strings.TrimSpace(match[1]), match[2]
}

func normalize(text string) string {
	text = html_heading.ReplaceAllString(text, "\n")
	text = html_hr.ReplaceAllString(text, "\n---\n")
	text = html_p_open.ReplaceAllString(text, "\n")
	text = html_p_close.ReplaceAllString(text, "\n")
	text = html_br.ReplaceAllString(text, "\n")
	text = html_tag.ReplaceAllString(text, "")
	return strings.ReplaceAll(text, "\r\n", "\n")
}

func parse_month_day_year(month, day, year string) time.Time {
	var d, y int
	_, _ = parse_int(day, &d)
	_, _ = parse_int(year, &y)
	return time.Date(y, months[month], d, 0, 0, 0, 0, time.UTC)
}

func parse_int(s string, dest *int) (int, error) {
	n := 0
	for _, r := range s {
		n = n*10 + int(r-'0')
	}
	*dest = n
	return n, nil
}

func ParseTracker(text string) []hood_events.TrackerRow {
	normalized := normalize(text)
	matches := date_heading.FindAllStringSubmatchIndex(normalized, -1)
	var rows []hood_events.TrackerRow
	seen := map[string]bool{}
	for i, loc := range matches {
		groups := date_heading.FindStringSubmatch(normalized[loc[0]:loc[1]])
		on := parse_month_day_year(groups[1], groups[2], groups[3])
		start := loc[1]
		end := len(normalized)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		body := strings.TrimSpace(normalized[start:end])
		if body == "" {
			continue
		}
		chunks := split_non_empty(hr_split.Split(body, -1))
		if len(chunks) == 1 && strings.Contains(chunks[0], "\n\n") {
			chunks = split_non_empty(regexp.MustCompile(`\n{2,}`).Split(chunks[0], -1))
		}
		for _, headline := range chunks {
			headline = strings.TrimSpace(whitespace.ReplaceAllString(headline, " "))
			if headline == "" || date_heading.MatchString(headline) {
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

func split_non_empty(parts []string) []string {
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
