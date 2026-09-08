// Package fold is the account-screen query over graph models. Child of //graph.
package fold

import (
	"sort"
	"time"

	"github.com/tcw165/fintech-fun/graph"
)

type Row struct {
	Date          time.Time
	Kind          graph.EventKind
	QtyBefore     float64
	QtyAfter      float64
	CashThisEvent float64
	NowHolds      string
}

type Result struct {
	Company      string
	TickerNow    string
	Status       graph.StockStatus
	QtyStarted   float64
	QtyNow       float64
	CashReceived float64
	Series       []Row
}

func Fold(company graph.Company, stock graph.Stock, events []graph.Event, qty float64) Result {
	ordered := append([]graph.Event(nil), events...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Date.Equal(ordered[j].Date) {
			return ordered[i].Kind < ordered[j].Kind
		}
		return ordered[i].Date.Before(ordered[j].Date)
	})
	qty_now := qty
	cash := 0.0
	rows := make([]Row, 0, len(ordered))
	for _, event := range ordered {
		qty_after := graph.QtyAfter(qty_now, event)
		cash_this := graph.EventCash(qty_now, event)
		rows = append(rows, Row{
			Date:          event.Date,
			Kind:          event.Kind,
			QtyBefore:     qty_now,
			QtyAfter:      qty_after,
			CashThisEvent: cash_this,
			NowHolds:      event.YouNowHold,
		})
		qty_now = qty_after
		cash += cash_this
	}
	return Result{
		Company:      company.Name,
		TickerNow:    stock.Ticker,
		Status:       stock.Status,
		QtyStarted:   qty,
		QtyNow:       qty_now,
		CashReceived: cash,
		Series:       rows,
	}
}
