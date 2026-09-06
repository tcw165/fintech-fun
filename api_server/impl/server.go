// Package impl is the HTTP server. Child of //api_server/contract.
package impl

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/tcw165/fintech-fun/api_server/contract"
)

func New(handler contract.Handler, folder contract.Folder) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(handler.Health())
	})
	mux.HandleFunc("GET /fold", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if folder == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
			return
		}
		q := r.URL.Query().Get("q")
		qty, err := strconv.ParseFloat(r.URL.Query().Get("qty"), 64)
		if q == "" || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "bad_request"})
			return
		}
		out, err := folder.Fold(q, qty)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(out)
	})
	return mux
}

type StaticOK struct{}

func (StaticOK) Health() contract.HealthResponse {
	return contract.HealthResponse{Status: "ok"}
}
