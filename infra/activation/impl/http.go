// Package impl is the Gap A HTTP checker. Child of //infra/activation/contract.
package impl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/tcw165/fintech-fun/graph/fold"
	"github.com/tcw165/fintech-fun/infra/activation/contract"
)

type HTTP struct {
	http_client *http.Client
}

func (h HTTP) client() *http.Client {
	if h.http_client != nil {
		return h.http_client
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
	query := url.QueryEscape(
		"LivePerson stock merger",
	)
	steps := []contract.Step{
		h.Health(api),
		h.require_ok(api+"/fold?q=square&qty=10", "fold"),
		h.require_ok(api+"/search?q="+query, "search"),
		h.require_ok(api+"/v1/graph?q=square", "graph"),
		h.require_ok(api+"/v1/graph/source", "source"),
		h.require_ok(api+"/v1/graph/fold?q=square&qty=10", "v1/fold"),
		h.require_ok(api+"/v1/graph/search?q=LPSN", "v1/search"),
	}
	report := contract.Report{Status: "ok", Steps: steps}
	if !report.AllOK() {
		report.Status = "failed"
	}
	return report
}

func (h HTTP) Gold(api string) contract.Report {
	var steps []contract.Step
	for _, want := range fold.GoldFold() {
		raw := fmt.Sprintf("%s/fold?q=%s&qty=%g", api, url.QueryEscape(want.Q), want.Qty)
		code, body, err := h.get(raw)
		step := contract.Step{Name: "gold:" + want.Q}
		if err != nil {
			step.Detail = err.Error()
			steps = append(steps, step)
			continue
		}
		if code != http.StatusOK {
			step.Detail = fmt.Sprintf("status=%d body=%v", code, body)
			steps = append(steps, step)
			continue
		}
		check := fold.MatchRows(as_rows(body["rows"]), want)
		step.OK = check.OK
		if !check.OK {
			step.Detail = fmt.Sprintf("%s qty=%v cash=%v", check.Error, check.GotQty, check.GotCash)
		}
		steps = append(steps, step)
	}
	report := contract.Report{Status: "ok", Steps: steps}
	if !report.AllOK() {
		report.Status = "failed"
	}
	return report
}

func as_rows(value any) []map[string]any {
	switch typed := value.(type) {
	case []map[string]any:
		return typed
	case []any:
		out := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if row, ok := item.(map[string]any); ok {
				out = append(out, row)
			}
		}
		return out
	default:
		return nil
	}
}

func (h HTTP) require_ok(raw_url, name string) contract.Step {
	code, body, err := h.get(raw_url)
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

func (h HTTP) get(raw_url string) (int, map[string]any, error) {
	resp, err := h.client().Get(raw_url)
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
