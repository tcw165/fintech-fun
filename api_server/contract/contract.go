// Package contract defines the api_server protocol. Pure data and interfaces.
package contract

type HealthResponse struct {
	Status string `json:"status"`
}

type FoldResponse struct {
	Status string           `json:"status"`
	Q      string           `json:"q"`
	Qty    float64          `json:"qty"`
	Rows   []map[string]any `json:"rows,omitempty"`
}

type Handler interface {
	Health() HealthResponse
}

// Folder is the account-screen fold. HTTP impl depends on this, not Bolt.
type Folder interface {
	Fold(q string, qty float64) (FoldResponse, error)
}

type SearchResponse struct {
	Status string `json:"status"`
	Query  string `json:"query"`
	Hits   any    `json:"hits,omitempty"`
}

// Searcher ranks Event.headline in Qdrant. HTTP impl depends on this, not REST.
type Searcher interface {
	Search(query string) (SearchResponse, error)
}

type GraphResponse struct {
	Status string           `json:"status"`
	Q      string           `json:"q"`
	Rows   []map[string]any `json:"rows,omitempty"`
}

type SourceResponse struct {
	Status     string `json:"status"`
	ID         string `json:"id"`
	PageSHA256 string `json:"page_sha256,omitempty"`
	FetchedAt  any    `json:"fetched_at,omitempty"`
}

// Grapher surfaces the constructed retail graph and ingest watermark.
type Grapher interface {
	Series(q string) (GraphResponse, error)
	Source() (SourceResponse, error)
}
