// Package classify maps tracker rows to Event.kind. Child of hood_events.
package classify

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
)

const ticker_pat = `[A-Z0-9]+(?:\.[A-Z0-9]+)?\^?`

var (
	ratio_re         = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*-?\s*for\s*-?\s*(\d+(?:\.\d+)?)`)
	cash_re          = regexp.MustCompile(`\$([0-9]+(?:\.[0-9]+)?)`)
	new_shares_re     = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s+new shares of\s+(` + ticker_pat + `)`)
	receive_shares_re = regexp.MustCompile(`(?i)receive\s+(\d+(?:\.\d+)?)\s+shares? of\s+(` + ticker_pat + `)`)
	shares_of_re      = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s+shares? of\s+(` + ticker_pat + `)`)
	ticker_to_re      = regexp.MustCompile(`(?i)(?:ticker change to|changed its ticker(?: symbol)? to|ticker symbol to|symbol change to|ticker to)\s+(` + ticker_pat + `)`)
	name_to_re        = regexp.MustCompile(`(?i)(?:corporate name to|renamed to|name change(?: and ticker change)? to)\s+([^.(]+)`)
	spinoff_ticker_re = regexp.MustCompile(`(?i)spin-?off of\s+(?:.+?\()?(` + ticker_pat + `)\)?`)
)

func parse_ratio(headline string) (float64, float64, bool) {
	match := ratio_re.FindStringSubmatch(headline)
	if match == nil {
		return 0, 0, false
	}
	left, _ := strconv.ParseFloat(match[1], 64)
	right, _ := strconv.ParseFloat(match[2], 64)
	return left, right, true
}

func parse_cash(headline string) float64 {
	match := cash_re.FindStringSubmatch(headline)
	if match == nil {
		return 0
	}
	n, _ := strconv.ParseFloat(match[1], 64)
	return n
}

func keep_fractionals(headline string) bool {
	lower := strings.ToLower(headline)
	return !strings.Contains(lower, "cash in lieu") && !strings.Contains(lower, "paid as cash")
}

func forward_multiplier(headline string) float64 {
	if left, right, ok := parse_ratio(headline); ok {
		return left / right
	}
	lower := strings.ToLower(headline)
	if strings.Contains(lower, "10 shares for every 1") || strings.Contains(lower, "10-for-1") {
		return 10
	}
	return 1
}

func reverse_multiplier(headline string) float64 {
	left, right, ok := parse_ratio(headline)
	if !ok {
		return 1
	}
	return right / left
}

func skip_event(row hood_events.TrackerRow, kind string) hood_events.ClassifiedEvent {
	return hood_events.ClassifiedEvent{
		Date: row.Date, Kind: kind, Headline: row.Headline, Company: row.Company, Ticker: row.Ticker,
		ShareMultiplier: 1, Skip: true,
	}
}

func pending_terms(lower string) bool {
	switch {
	case strings.Contains(lower, "delisted pending"),
		strings.Contains(lower, "delisted to otc"),
		strings.Contains(lower, "still pending"),
		strings.Contains(lower, "untradeable"),
		strings.Contains(lower, "tbd"),
		strings.Contains(lower, "tba"):
		return true
	default:
		return false
	}
}

func has_share_consideration(headline string) bool {
	return new_shares_re.MatchString(headline) || receive_shares_re.MatchString(headline) || shares_of_re.MatchString(headline)
}

func apply_share_exchange(base *hood_events.ClassifiedEvent, headline string) {
	if match := new_shares_re.FindStringSubmatch(headline); match != nil {
		base.ShareMultiplier, _ = strconv.ParseFloat(match[1], 64)
		base.YouNowHold = match[2]
		return
	}
	if match := receive_shares_re.FindStringSubmatch(headline); match != nil {
		base.ShareMultiplier, _ = strconv.ParseFloat(match[1], 64)
		base.YouNowHold = match[2]
		return
	}
	if match := shares_of_re.FindStringSubmatch(headline); match != nil {
		base.ShareMultiplier, _ = strconv.ParseFloat(match[1], 64)
		base.YouNowHold = match[2]
	}
}

func ClassifyRow(row hood_events.TrackerRow) []hood_events.ClassifiedEvent {
	headline := row.Headline
	lower := strings.ToLower(headline)
	if strings.Contains(lower, "cusip change") && !strings.Contains(lower, "ticker") && !strings.Contains(lower, "split") && !strings.Contains(lower, "name") {
		return []hood_events.ClassifiedEvent{skip_event(row, "")}
	}
	if strings.Contains(lower, "partial liquidation") {
		return []hood_events.ClassifiedEvent{skip_event(row, "")}
	}
	keep := keep_fractionals(headline)
	base := hood_events.ClassifiedEvent{
		Date: row.Date, Headline: headline, Company: row.Company, Ticker: row.Ticker,
		KeepFractionals: keep, CanTrade: true,
	}

	if strings.Contains(lower, "expired worthless") || strings.Contains(lower, "declared worthless") || strings.Contains(lower, "deemed worthless") || strings.Contains(lower, "redeemed at $0") || strings.Contains(lower, "rights are no longer trading") {
		base.Kind = "worthless"
		base.ShareMultiplier = 0
		base.CanTrade = false
		return []hood_events.ClassifiedEvent{base}
	}
	if pending_terms(lower) {
		base.Kind = "waiting"
		base.ShareMultiplier = 1
		base.CanTrade = false
		return []hood_events.ClassifiedEvent{base}
	}
	if strings.Contains(lower, "cash merger") || strings.Contains(lower, "was liquidated") || strings.Contains(lower, "performed a liquidation") || strings.Contains(lower, "liquidation at $") || (strings.Contains(lower, "was acquired") && parse_cash(headline) > 0 && !has_share_consideration(headline)) {
		base.Kind = "cashed_out"
		base.ShareMultiplier = 0
		base.CashPerShare = parse_cash(headline)
		base.CanTrade = false
		return []hood_events.ClassifiedEvent{base}
	}
	if strings.Contains(lower, "stock merger") || (strings.Contains(lower, "was acquired") && has_share_consideration(headline)) {
		base.Kind = "now_different_stock"
		base.CanTrade = false
		base.CashPerShare = parse_cash(headline)
		apply_share_exchange(&base, headline)
		return []hood_events.ClassifiedEvent{base}
	}
	if strings.Contains(lower, "was acquired") {
		base.Kind = "waiting"
		base.ShareMultiplier = 1
		base.CanTrade = false
		return []hood_events.ClassifiedEvent{base}
	}
	if strings.Contains(lower, "spinoff") || strings.Contains(lower, "spin-off") {
		base.Kind = "extra_stock"
		base.ShareMultiplier = 1
		apply_share_exchange(&base, headline)
		if base.YouNowHold == "" {
			if match := spinoff_ticker_re.FindStringSubmatch(headline); match != nil {
				base.YouNowHold = match[1]
			}
		}
		return []hood_events.ClassifiedEvent{base}
	}
	split_text := strings.ReplaceAll(lower, "stock split", "split")
	if strings.Contains(split_text, "reverse split") {
		base.Kind = "reverse_split"
		base.ShareMultiplier = reverse_multiplier(headline)
		events := []hood_events.ClassifiedEvent{base}
		if match := ticker_to_re.FindStringSubmatch(headline); match != nil {
			events[0].YouNowHold = match[1]
			change := base
			change.Kind = "ticker_changed"
			change.ShareMultiplier = 1
			change.YouNowHold = match[1]
			events = append(events, change)
		}
		return events
	}
	if strings.Contains(split_text, "forward split") {
		base.Kind = "split"
		base.ShareMultiplier = forward_multiplier(headline)
		return []hood_events.ClassifiedEvent{base}
	}

	var events []hood_events.ClassifiedEvent
	if match := ticker_to_re.FindStringSubmatch(headline); match != nil {
		item := base
		item.Kind = "ticker_changed"
		item.ShareMultiplier = 1
		item.YouNowHold = match[1]
		events = append(events, item)
	}
	if match := name_to_re.FindStringSubmatch(headline); match != nil {
		item := base
		item.Kind = "name_changed"
		item.ShareMultiplier = 1
		item.NewName = strings.TrimRight(strings.TrimSpace(match[1]), ".")
		events = append(events, item)
	}
	if len(events) > 0 {
		return events
	}
	return []hood_events.ClassifiedEvent{skip_event(row, "")}
}

func ClassifyRows(rows []hood_events.TrackerRow) []hood_events.ClassifiedEvent {
	var out []hood_events.ClassifiedEvent
	for _, row := range rows {
		out = append(out, ClassifyRow(row)...)
	}
	return out
}

type ClassifiedView struct {
	Kind            string  `json:"kind"`
	Skip            bool    `json:"skip"`
	ShareMultiplier float64 `json:"share_multiplier"`
	CashPerShare    float64 `json:"cash_per_share"`
	YouNowHold      string  `json:"you_now_hold"`
	NewName         string  `json:"new_name"`
}

type ClassifyResult struct {
	Status string           `json:"status"`
	Events []ClassifiedView `json:"events"`
}

func ClassifyHeadline(headline, date_text, company, ticker string) (ClassifyResult, error) {
	on, err := time.Parse("2006-01-02", date_text)
	if err != nil {
		return ClassifyResult{}, err
	}
	events := ClassifyRow(hood_events.TrackerRow{Date: on, Headline: headline, Company: company, Ticker: ticker})
	views := make([]ClassifiedView, 0, len(events))
	for _, event := range events {
		views = append(views, ClassifiedView{
			Kind: event.Kind, Skip: event.Skip, ShareMultiplier: event.ShareMultiplier,
			CashPerShare: event.CashPerShare, YouNowHold: event.YouNowHold, NewName: event.NewName,
		})
	}
	return ClassifyResult{Status: "success", Events: views}, nil
}
