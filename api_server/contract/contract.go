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
