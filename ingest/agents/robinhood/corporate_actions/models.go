// Package hood_events is tracker models. Pure data. Children depend on this.
package hood_events

import "time"

const TrackerURL = "https://robinhood.com/us/en/support/articles/corporate-actions-tracker/"

type TrackerRow struct {
	Date     time.Time
	Headline string
	Company  string
	Ticker   string
}

type ClassifiedEvent struct {
	Date            time.Time
	Kind            string
	Headline        string
	Company         string
	Ticker          string
	ShareMultiplier float64
	CashPerShare    float64
	KeepFractionals bool
	CanTrade        bool
	YouNowHold      string
	NewName         string
	Skip            bool
}
