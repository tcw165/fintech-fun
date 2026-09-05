// Package graph is the retail investor graph: Company, Stock, Event.
//
// Matches the Notion page "Retail investor graph". Cash is a field on Event,
// not a node. Event identity is (happened_to ticker, date, kind).
package graph

import (
	"fmt"
	"time"
)

type EventKind string

const (
	KindTickerChanged     EventKind = "ticker_changed"
	KindNameChanged       EventKind = "name_changed"
	KindSplit             EventKind = "split"
	KindReverseSplit      EventKind = "reverse_split"
	KindNowDifferentStock EventKind = "now_different_stock"
	KindCashedOut         EventKind = "cashed_out"
	KindExtraStock        EventKind = "extra_stock"
	KindWorthless         EventKind = "worthless"
	KindWaiting           EventKind = "waiting"
)

func AllEventKinds() []EventKind {
	return []EventKind{
		KindTickerChanged,
		KindNameChanged,
		KindSplit,
		KindReverseSplit,
		KindNowDifferentStock,
		KindCashedOut,
		KindExtraStock,
		KindWorthless,
		KindWaiting,
	}
}

type StockStatus string

const (
	StatusTradeable StockStatus = "tradeable"
	StatusOTC       StockStatus = "otc"
	StatusGone      StockStatus = "gone"
	StatusWorthless StockStatus = "worthless"
)

type Company struct {
	Name         string
	AlsoKnownAs  []string
}

type Stock struct {
	Ticker         string
	FormerTickers  []string
	Status         StockStatus
}

type Event struct {
	Date             time.Time
	Kind             EventKind
	Headline         string
	HappenedTo       string
	ShareMultiplier  float64
	CashPerShare     float64
	KeepFractionals  bool
	CanTrade         bool
	YouNowHold       string
}

func (e Event) ID() string {
	return fmt.Sprintf("%s|%s|%s", e.HappenedTo, e.Date.Format("2006-01-02"), e.Kind)
}

func QtyAfter(qtyBefore float64, event Event) float64 {
	if event.Kind == KindReverseSplit {
		return qtyBefore / event.ShareMultiplier
	}
	return qtyBefore * event.ShareMultiplier
}

func EventCash(qtyBefore float64, event Event) float64 {
	return qtyBefore * event.CashPerShare
}

func Multiplies(kind EventKind) bool {
	return kind != KindReverseSplit
}

// Date is a UTC calendar day. Children use this instead of repeating time.Date.
func Date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
