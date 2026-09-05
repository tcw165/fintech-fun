// Package classify maps tracker rows to Event.kind. Child of hood_events.
package classify

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tcw165/fintech-fun/ingest/agents/robinhood/corporate_actions"
)

const tickerPat = `[A-Z0-9]+(?:\.[A-Z0-9]+)?\^?`

var (
	ratioRE         = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s+for\s+(\d+(?:\.\d+)?)`)
	cashRE          = regexp.MustCompile(`\$([0-9]+(?:\.[0-9]+)?)`)
	newSharesRE     = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s+new shares of\s+(` + tickerPat + `)`)
	receiveSharesRE = regexp.MustCompile(`(?i)receive\s+(\d+(?:\.\d+)?)\s+shares? of\s+(` + tickerPat + `)`)
	tickerToRE      = regexp.MustCompile(`(?i)(?:ticker change to|changed its ticker to|ticker to)\s+(` + tickerPat + `)`)
	nameToRE        = regexp.MustCompile(`(?i)(?:corporate name to|renamed to|name change(?: and ticker change)? to)\s+([^.(]+)`)
	spinoffTickerRE = regexp.MustCompile(`(?i)spinoff of\s+(?:.+?\()?(` + tickerPat + `)\)?`)
)

func parseRatio(headline string) (float64, float64, bool) {
	match := ratioRE.FindStringSubmatch(headline)
	if match == nil {
		return 0, 0, false
	}
	left, _ := strconv.ParseFloat(match[1], 64)
	right, _ := strconv.ParseFloat(match[2], 64)
	return left, right, true
}

func parseCash(headline string) float64 {
	match := cashRE.FindStringSubmatch(headline)
	if match == nil {
		return 0
	}
	n, _ := strconv.ParseFloat(match[1], 64)
	return n
}

func keepFractionals(headline string) bool {
	lower := strings.ToLower(headline)
	return !strings.Contains(lower, "cash in lieu") && !strings.Contains(lower, "paid as cash")
}

func forwardMultiplier(headline string) float64 {
	if left, right, ok := parseRatio(headline); ok {
		return left / right
	}
	lower := strings.ToLower(headline)
	if strings.Contains(lower, "10 shares for every 1") || strings.Contains(lower, "10-for-1") {
		return 10
	}
	return 1
}

func reverseMultiplier(headline string) float64 {
	left, right, ok := parseRatio(headline)
	if !ok {
		return 1
	}
	return right / left
}

func skipEvent(row hood_events.TrackerRow, kind string) hood_events.ClassifiedEvent {
	return hood_events.ClassifiedEvent{
		Date: row.Date, Kind: kind, Headline: row.Headline, Company: row.Company, Ticker: row.Ticker,
		ShareMultiplier: 1, Skip: true,
	}
}

func ClassifyRow(row hood_events.TrackerRow) []hood_events.ClassifiedEvent {
	headline := row.Headline
	lower := strings.ToLower(headline)
	if strings.Contains(lower, "cusip change") && !strings.Contains(lower, "ticker") && !strings.Contains(lower, "split") && !strings.Contains(lower, "name") {
		return []hood_events.ClassifiedEvent{skipEvent(row, "")}
	}
	keep := keepFractionals(headline)
	base := hood_events.ClassifiedEvent{
		Date: row.Date, Headline: headline, Company: row.Company, Ticker: row.Ticker,
		KeepFractionals: keep, CanTrade: true,
	}

	if strings.Contains(lower, "expired worthless") || strings.Contains(lower, "declared worthless") {
		base.Kind = "worthless"
		base.ShareMultiplier = 0
		base.CanTrade = false
		return []hood_events.ClassifiedEvent{base}
	}
	if strings.Contains(lower, "delisted pending") || strings.Contains(lower, "details are still pending") {
		base.Kind = "waiting"
		base.ShareMultiplier = 1
		base.CanTrade = false
		return []hood_events.ClassifiedEvent{base}
	}
	if strings.Contains(lower, "cash merger") || strings.Contains(lower, "was liquidated") {
		base.Kind = "cashed_out"
		base.ShareMultiplier = 0
		base.CashPerShare = parseCash(headline)
		base.CanTrade = false
		return []hood_events.ClassifiedEvent{base}
	}
	if strings.Contains(lower, "stock merger") {
		base.Kind = "now_different_stock"
		base.CanTrade = false
		if match := newSharesRE.FindStringSubmatch(headline); match != nil {
			base.ShareMultiplier, _ = strconv.ParseFloat(match[1], 64)
			base.YouNowHold = match[2]
		}
		return []hood_events.ClassifiedEvent{base}
	}
	if strings.Contains(lower, "spinoff") {
		base.Kind = "extra_stock"
		base.ShareMultiplier = 1
		if match := receiveSharesRE.FindStringSubmatch(headline); match != nil {
			base.ShareMultiplier, _ = strconv.ParseFloat(match[1], 64)
			base.YouNowHold = match[2]
		} else if match := newSharesRE.FindStringSubmatch(headline); match != nil {
			base.ShareMultiplier, _ = strconv.ParseFloat(match[1], 64)
			base.YouNowHold = match[2]
		} else if match := spinoffTickerRE.FindStringSubmatch(headline); match != nil {
			base.YouNowHold = match[1]
		}
		return []hood_events.ClassifiedEvent{base}
	}
	if strings.Contains(lower, "reverse split") {
		base.Kind = "reverse_split"
		base.ShareMultiplier = reverseMultiplier(headline)
		events := []hood_events.ClassifiedEvent{base}
		if match := tickerToRE.FindStringSubmatch(headline); match != nil {
			events[0].YouNowHold = match[1]
			change := base
			change.Kind = "ticker_changed"
			change.ShareMultiplier = 1
			change.YouNowHold = match[1]
			events = append(events, change)
		}
		return events
	}
	if strings.Contains(lower, "forward split") {
		base.Kind = "split"
		base.ShareMultiplier = forwardMultiplier(headline)
		return []hood_events.ClassifiedEvent{base}
	}

	var events []hood_events.ClassifiedEvent
	if match := tickerToRE.FindStringSubmatch(headline); match != nil {
		item := base
		item.Kind = "ticker_changed"
		item.ShareMultiplier = 1
		item.YouNowHold = match[1]
		events = append(events, item)
	}
	if match := nameToRE.FindStringSubmatch(headline); match != nil {
		item := base
		item.Kind = "name_changed"
		item.ShareMultiplier = 1
		item.NewName = strings.TrimRight(strings.TrimSpace(match[1]), ".")
		events = append(events, item)
	}
	if len(events) > 0 {
		return events
	}
	return []hood_events.ClassifiedEvent{skipEvent(row, "")}
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

func ClassifyHeadline(headline, dateText, company, ticker string) (ClassifyResult, error) {
	on, err := time.Parse("2006-01-02", dateText)
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
