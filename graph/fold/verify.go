package fold

import (
	"math"
	"strings"

	"github.com/tcw165/fintech-fun/graph/examples"
)

type Expectation struct {
	Q        string  `json:"q"`
	Qty      float64 `json:"qty"`
	Company  string  `json:"company"`
	Ticker   string  `json:"ticker"`
	QtyNow   float64 `json:"qty_now"`
	Cash     float64 `json:"cash"`
	Series   int     `json:"series"`
}

func Expectations() []Expectation {
	ftelQty := 10.0 / 16 / 8 / 7 / 9
	return []Expectation{
		{Q: "nflx", Qty: 10, Company: "Netflix", Ticker: "NFLX", QtyNow: 100, Series: 1},
		{Q: "mnts", Qty: 140, Company: "Momentus", Ticker: "MNTS", QtyNow: 10, Series: 1},
		{Q: "apge", Qty: 10, Company: "Apogee", Ticker: "APGE", QtyNow: 0, Cash: 1351.10, Series: 1},
		{Q: "lpsn", Qty: 100, Company: "LivePerson", Ticker: "LPSN", QtyNow: 46.73, Series: 1},
		{Q: "square", Qty: 10, Company: "Block", Ticker: "XYZ", QtyNow: 10, Series: 1},
		{Q: "block", Qty: 10, Company: "Block", Ticker: "XYZ", QtyNow: 10, Series: 1},
		{Q: "sq", Qty: 10, Company: "Block", Ticker: "XYZ", QtyNow: 10, Series: 1},
		{Q: "xyz", Qty: 10, Company: "Block", Ticker: "XYZ", QtyNow: 10, Series: 1},
		{Q: "ftel", Qty: 10, Company: "GMEX Robotics", Ticker: "GMEX", QtyNow: ftelQty, Series: 6},
	}
}

func Resolve(q string) (examples.Fixture, bool) {
	needle := strings.ToLower(q)
	for _, fixture := range examples.All() {
		if strings.Contains(strings.ToLower(fixture.Company.Name), needle) {
			return fixture, true
		}
		for _, alias := range fixture.Company.AlsoKnownAs {
			if strings.Contains(strings.ToLower(alias), needle) {
				return fixture, true
			}
		}
		if strings.ToLower(fixture.Stock.Ticker) == needle {
			return fixture, true
		}
		for _, former := range fixture.Stock.FormerTickers {
			if strings.ToLower(former) == needle {
				return fixture, true
			}
		}
	}
	return examples.Fixture{}, false
}

type Check struct {
	Expectation
	OK     bool   `json:"ok"`
	GotQty float64 `json:"got_qty"`
	GotCash float64 `json:"got_cash"`
	Error  string `json:"error,omitempty"`
}

func VerifyExamples() []Check {
	var out []Check
	for _, want := range Expectations() {
		check := Check{Expectation: want}
		fixture, ok := Resolve(want.Q)
		if !ok {
			check.Error = "unresolved"
			out = append(out, check)
			continue
		}
		result := Fold(fixture.Company, fixture.Stock, fixture.Events, want.Qty)
		check.GotQty = result.QtyNow
		check.GotCash = result.CashReceived
		if result.Company != want.Company || result.TickerNow != want.Ticker {
			check.Error = "identity"
		} else if len(result.Series) != want.Series {
			check.Error = "series"
		} else if math.Abs(result.QtyNow-want.QtyNow) > 0.005 {
			check.Error = "qty"
		} else if math.Abs(result.CashReceived-want.Cash) > 0.02 {
			check.Error = "cash"
		} else {
			check.OK = true
		}
		out = append(out, check)
	}
	return out
}
