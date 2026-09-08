// Package tools dispatches corporate_actions CLI commands. Child of ingest packages.
package tools

import (
	"fmt"
	"os"
	"strings"

	hood_events "github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/agent"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/classify"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/fetch"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/ingest"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/parser"
	"github.com/tcw165/fintech-fun/graph/clients/neo4j"
	"github.com/tcw165/fintech-fun/graph/clients/qdrant"
	"github.com/tcw165/fintech-fun/graph/embed"
	"github.com/tcw165/fintech-fun/graph/embed/lexical"
	"github.com/tcw165/fintech-fun/graph/examples"
	"github.com/tcw165/fintech-fun/graph/fold"
	"github.com/tcw165/fintech-fun/ingest_jobs/corporate_actions/request"
)

type Deps struct {
	GraphClient   neo4j.Client
	VectorsClient qdrant.Client
	Embedder      embed.Embedder
	Agent         *agent.Agent
}

func Run(deps Deps, req request.Request) (any, error) {
	if req.Name == "" {
		req.Name = "ingest"
	}
	switch req.Name {
	case "parse":
		return run_parse(req)
	case "classify":
		return run_classify(req)
	case "plan":
		return run_plan(req)
	case "fetch":
		return run_fetch(req)
	case "gold":
		return run_gold(req)
	case "ingest":
		return run_ingest(deps, req)
	case "verify":
		return run_verify(deps)
	case "search":
		return run_search(deps, req)
	case "refresh":
		return run_refresh(deps)
	case "ping":
		return run_ping(deps)
	case "seed":
		return run_seed(deps)
	case "fold":
		return run_fold(deps, req)
	default:
		return nil, fmt.Errorf("unknown command %q\n\nsource: %s", req.Name, hood_events.TrackerURL)
	}
}

func run_ingest(deps Deps, req request.Request) (any, error) {
	if req.DryRun {
		return run_plan(req)
	}
	if err := require_clients(deps, "ingest"); err != nil {
		return nil, err
	}
	if req.File != "" {
		data, err := os.ReadFile(req.File)
		if err != nil {
			return nil, err
		}
		result, err := deps.Agent.IngestText(string(data))
		if err != nil {
			return nil, err
		}
		return result.Payload(), nil
	}
	result, err := deps.Agent.IngestLive("")
	if err != nil {
		return nil, err
	}
	return result.Payload(), nil
}

func run_refresh(deps Deps) (any, error) {
	if err := require_clients(deps, "refresh"); err != nil {
		return nil, err
	}
	result, err := deps.Agent.Refresh("")
	if err != nil {
		return nil, err
	}
	return result.Payload(), nil
}

func run_search(deps Deps, req request.Request) (any, error) {
	if deps.VectorsClient == nil {
		return nil, fmt.Errorf("search requires an injected Qdrant client")
	}
	query := strings.Join(strings.Fields(req.Query), " ")
	if query == "" {
		return nil, fmt.Errorf("search requires a query")
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 5
	}
	embedder := deps.Embedder
	if embedder == nil {
		embedder = lexical.New()
	}
	vecs, err := embedder.Embed([]string{query})
	if err != nil {
		return nil, err
	}
	hits, err := qdrant.SearchHeadlines(deps.VectorsClient, vecs[0], limit)
	if err != nil {
		return nil, err
	}
	return map[string]any{"status": "ok", "query": query, "hits": hits}, nil
}

func run_verify(deps Deps) (any, error) {
	memory := fold.VerifyExamples()
	ok := 0
	for _, check := range memory {
		if check.OK {
			ok++
		}
	}
	out := map[string]any{"status": "ok", "passed": ok, "checks": memory}
	bolt_failed := 0
	if deps.GraphClient != nil {
		var bolt []map[string]any
		for _, want := range fold.GoldFold() {
			rows, err := neo4j.Fold(deps.GraphClient, want.Q, want.Qty)
			item := map[string]any{"q": want.Q, "qty": want.Qty}
			if err != nil {
				item["error"] = err.Error()
				bolt_failed++
			} else if !fold.FoldShape(rows) {
				item["skipped"] = "not_fold"
				item["rows"] = rows
			} else {
				check := fold.MatchRows(rows, want)
				item["ok"] = check.OK
				item["got_qty"] = check.GotQty
				item["got_cash"] = check.GotCash
				if !check.OK {
					item["error"] = check.Error
					bolt_failed++
				}
			}
			bolt = append(bolt, item)
		}
		out["bolt"] = bolt
		out["bolt_failed"] = bolt_failed
	}
	if ok != len(memory) {
		return out, fmt.Errorf("verify failed: %d/%d", ok, len(memory))
	}
	if bolt_failed > 0 {
		return out, fmt.Errorf("verify bolt gold failed: %d", bolt_failed)
	}
	return out, nil
}

func run_ping(deps Deps) (any, error) {
	if err := require_clients(deps, "ping"); err != nil {
		return nil, err
	}
	if err := neo4j.ApplyConstraints(deps.GraphClient); err != nil {
		return nil, err
	}
	if _, err := qdrant.EnsureCollection(deps.VectorsClient); err != nil {
		return nil, err
	}
	company, stock, events := examples.NFLX()
	if _, err := neo4j.IngestGraph(deps.GraphClient, company, stock, events); err != nil {
		return nil, err
	}
	if _, err := qdrant.UpsertEvents(deps.VectorsClient, events, nil); err != nil {
		return nil, err
	}
	rows, err := deps.GraphClient.Run(
		"MATCH (e:Event {id: $id}) RETURN e.headline AS headline, e.kind AS kind",
		map[string]any{"id": events[0].ID()},
	)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"status":      "ok",
		"constraints": len(neo4j.Constraints),
		"collection":  qdrant.Collection,
		"event_id":    events[0].ID(),
		"read":        rows,
	}, nil
}

func run_seed(deps Deps) (any, error) {
	if err := require_clients(deps, "seed"); err != nil {
		return nil, err
	}
	if err := neo4j.ApplyConstraints(deps.GraphClient); err != nil {
		return nil, err
	}
	if _, err := qdrant.EnsureCollection(deps.VectorsClient); err != nil {
		return nil, err
	}
	type row struct {
		Company string `json:"company"`
		Ticker  string `json:"ticker"`
		Count   int    `json:"events"`
	}
	var events []row
	seeded := 0
	event_count := 0
	for _, fixture := range examples.All() {
		result, err := neo4j.IngestGraph(deps.GraphClient, fixture.Company, fixture.Stock, fixture.Events)
		if err != nil {
			return nil, err
		}
		if _, err := qdrant.UpsertEvents(deps.VectorsClient, fixture.Events, nil); err != nil {
			return nil, err
		}
		seeded++
		event_count += result.Events
		events = append(events, row{Company: result.Company, Ticker: result.Ticker, Count: result.Events})
	}
	return map[string]any{
		"status":   "ok",
		"fixtures": seeded,
		"events":   event_count,
		"rows":     events,
	}, nil
}

func run_fold(deps Deps, req request.Request) (any, error) {
	if err := require_clients(deps, "fold"); err != nil {
		return nil, err
	}
	if req.Q == "" {
		return nil, fmt.Errorf("fold requires --q and --qty")
	}
	rows, err := neo4j.Fold(deps.GraphClient, req.Q, req.Qty)
	if err != nil {
		return nil, err
	}
	return map[string]any{"status": "ok", "q": req.Q, "qty": req.Qty, "rows": rows}, nil
}

func run_parse(req request.Request) (any, error) {
	text, err := read_file(req.File)
	if err != nil {
		return nil, err
	}
	return parser.ParseTrackerPage(text), nil
}

func run_classify(req request.Request) (any, error) {
	headline := strings.Join(strings.Fields(req.Headline), " ")
	if req.Date == "" || headline == "" {
		return nil, fmt.Errorf("classify requires date and headline")
	}
	return classify.ClassifyHeadline(
		headline,
		req.Date,
		req.Company,
		req.Ticker,
	)
}

func run_fetch(req request.Request) (any, error) {
	result, err := fetch.FetchTracker(nil, "")
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"status":      "ok",
		"url":         result.URL,
		"bytes":       result.Bytes,
		"status_code": result.Status,
	}
	if req.Out != "" {
		if err := os.WriteFile(req.Out, []byte(result.Text), 0o644); err != nil {
			return nil, err
		}
		out["path"] = req.Out
		return out, nil
	}
	out["text"] = result.Text
	return out, nil
}

func run_gold(req request.Request) (any, error) {
	var text string
	if req.File != "" {
		data, err := os.ReadFile(req.File)
		if err != nil {
			return nil, err
		}
		text = string(data)
	} else {
		page, err := fetch.FetchTracker(nil, "")
		if err != nil {
			return nil, err
		}
		text = page.Text
	}
	return classify.Report(parser.ParseTracker(text)), nil
}

func run_plan(req request.Request) (any, error) {
	text, err := read_file(req.File)
	if err != nil {
		return nil, err
	}
	return ingest.PlanIngest(text), nil
}

func require_clients(deps Deps, cmd string) error {
	if deps.GraphClient == nil || deps.VectorsClient == nil || deps.Agent == nil {
		return fmt.Errorf("%s requires injected Neo4j, Qdrant, and ingest agent", cmd)
	}
	return nil
}

func read_file(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("missing file argument")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
