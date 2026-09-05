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
	qtyNow := qty
	cash := 0.0
	rows := make([]Row, 0, len(ordered))
	for _, event := range ordered {
		qtyAfter := graph.QtyAfter(qtyNow, event)
		cashThis := graph.EventCash(qtyNow, event)
		rows = append(rows, Row{
			Date:          event.Date,
			Kind:          event.Kind,
			QtyBefore:     qtyNow,
			QtyAfter:      qtyAfter,
			CashThisEvent: cashThis,
			NowHolds:      event.YouNowHold,
		})
		qtyNow = qtyAfter
		cash += cashThis
	}
	return Result{
		Company:      company.Name,
		TickerNow:    stock.Ticker,
		Status:       stock.Status,
		QtyStarted:   qty,
		QtyNow:       qtyNow,
		CashReceived: cash,
		Series:       rows,
	}
}
