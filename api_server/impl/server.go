// Package impl is the HTTP server. Child of //api_server/contract.
package impl

import (
	"encoding/json"
	"net/http"

	"github.com/tcw165/fintech-fun/api_server/contract"
)

func New(handler contract.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(handler.Health())
	})
	return mux
}

type StaticOK struct{}

func (StaticOK) Health() contract.HealthResponse {
	return contract.HealthResponse{Status: "ok"}
}
