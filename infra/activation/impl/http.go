// Package impl is the Gap A HTTP checker. Child of //infra/activation/contract.
package impl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/tcw165/fintech-fun/infra/activation/contract"
)

type HTTP struct {
	Client *http.Client
}

func (h HTTP) client() *http.Client {
	if h.Client != nil {
		return h.Client
	}
	return &http.Client{Timeout: 15 * time.Second}
}

func (h HTTP) Health(api string) contract.Step {
	code, body, err := h.get(api + "/healthz")
	step := contract.Step{Name: contract.StepHealthz}
	if err != nil {
		step.Detail = err.Error()
		return step
	}
	if code != http.StatusOK || fmt.Sprint(body["status"]) != "ok" {
		step.Detail = fmt.Sprintf("status=%d body=%v", code, body)
		return step
	}
	step.OK = true
	return step
}

func (h HTTP) Smoke(api string) contract.Report {
	query := url.QueryEscape("LivePerson stock merger")
	steps := []contract.Step{
		h.Health(api),
		h.requireOK(api+"/fold?q=square&qty=10", "fold"),
		h.requireOK(api+"/search?q="+query, "search"),
		h.requireOK(api+"/v1/graph?q=square", "graph"),
		h.requireOK(api+"/v1/graph/source", "source"),
		h.requireOK(api+"/v1/graph/fold?q=square&qty=10", "v1/fold"),
		h.requireOK(api+"/v1/graph/search?q=LPSN", "v1/search"),
	}
	report := contract.Report{Status: "ok", Steps: steps}
	if !report.AllOK() {
		report.Status = "failed"
	}
	return report
}

func (h HTTP) Gold(api string) contract.Report {
	return contract.Report{Status: "unimplemented"}
}

func (h HTTP) requireOK(rawURL, name string) contract.Step {
	code, body, err := h.get(rawURL)
	step := contract.Step{Name: name}
	if err != nil {
		step.Detail = err.Error()
		return step
	}
	if code != http.StatusOK || fmt.Sprint(body["status"]) != "ok" {
		step.Detail = fmt.Sprintf("status=%d body=%v", code, body)
		return step
	}
	step.OK = true
	return step
}

func (h HTTP) get(rawURL string) (int, map[string]any, error) {
	resp, err := h.client().Get(rawURL)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, body, nil
}

var _ contract.Checker = HTTP{}
