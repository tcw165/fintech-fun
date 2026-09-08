// Package impl is the HTTP server. Child of //api_server/contract.
package impl

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/tcw165/fintech-fun/api_server/contract"
)

func New(handler contract.Handler, folder contract.Folder, searcher contract.Searcher, grapher contract.Grapher) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(handler.Health())
	})
	write_fold := func(w http.ResponseWriter, r *http.Request) {
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
	}
	mux.HandleFunc("GET /fold", write_fold)
	mux.HandleFunc("GET /v1/graph/fold", write_fold)
	write_search := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if searcher == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
			return
		}
		q := strings.Join(strings.Fields(r.URL.Query().Get("q")), " ")
		if q == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "bad_request"})
			return
		}
		out, err := searcher.Search(q)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(out)
	}
	mux.HandleFunc("GET /search", write_search)
	mux.HandleFunc("GET /v1/graph/search", write_search)
	mux.HandleFunc("GET /v1/graph", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if grapher == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
			return
		}
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		if q == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "bad_request"})
			return
		}
		out, err := grapher.Series(q)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(out)
	})
	mux.HandleFunc("GET /v1/graph/source", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if grapher == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
			return
		}
		out, err := grapher.Source()
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
