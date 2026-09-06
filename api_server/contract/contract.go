// Package contract defines the api_server protocol. Pure data and interfaces.
package contract

type HealthResponse struct {
	Status string `json:"status"`
}

type Handler interface {
	Health() HealthResponse
}
