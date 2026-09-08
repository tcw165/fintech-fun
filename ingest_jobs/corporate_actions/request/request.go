// Package request is the corporate_actions command contract. Pure data.
package request

// Request is one parsed CLI invocation. Empty Name means default ingest.
type Request struct {
	Name     string
	File     string
	Out      string
	Query    string
	Limit    int
	Q        string
	Qty      float64
	Date     string
	Headline string
	Company  string
	Ticker   string
	DryRun   bool
}
