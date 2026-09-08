// Package contract is the ingest-agent protocol. Pure data and interfaces.
package contract

const WaitingPolicy = "waiting Events keep id (ticker|date|waiting). A later cashed_out or now_different_stock is a new Event on a later date, not an in-place update of the waiting row."

type FetchedPage struct {
	Status int
	URL    string
	Bytes  int
	Text   string
}

type Fetcher interface {
	Fetch(url string) (FetchedPage, error)
}

type Result struct {
	Written       int
	Created       int
	Duplicates    int
	Skipped       int
	Unchanged     bool
	DryRun        bool
	PageSHA256    string
	Companies     []string
	Stocks        []string
	Overlap       int
	Suffix        int
	Pages         int
	HistoryLen    int
	Backfilled    bool
	Refresh       bool
	WaitingPolicy string
}

func (r Result) Payload() map[string]any {
	out := map[string]any{
		"status":        "ok",
		"written":       r.Written,
		"created":       r.Created,
		"duplicates":    r.Duplicates,
		"skipped":       r.Skipped,
		"unchanged":     r.Unchanged,
		"page_sha256":   r.PageSHA256,
		"companies":     len(r.Companies),
		"stocks":        len(r.Stocks),
		"company_names": r.Companies,
		"tickers":       r.Stocks,
		"overlap":       r.Overlap,
		"suffix":        r.Suffix,
		"pages":         r.Pages,
		"history_len":   r.HistoryLen,
	}
	if r.DryRun {
		out["dry_run"] = true
	}
	if r.Backfilled {
		out["backfilled"] = true
	}
	if r.Refresh {
		out["refresh"] = true
		out["waiting_policy"] = r.WaitingPolicy
	}
	return out
}

type Agent interface {
	IngestText(text string) (Result, error)
	IngestLive(url string) (Result, error)
	Refresh(url string) (Result, error)
}
