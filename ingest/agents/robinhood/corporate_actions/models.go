// Package hood_events is tracker models. Pure data. Children depend on this.
package hood_events

import "time"

type TrackerRow struct {
	Date     time.Time
	Headline string
	Company  string
	Ticker   string
}
