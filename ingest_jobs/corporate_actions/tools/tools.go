// Package tools dispatches corporate_actions CLI commands. Child of ingest packages.
package tools

import (
	"fmt"
	"os"
	"strconv"
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
)

type Deps struct {
	Graph   neo4j.Client
	Vectors qdrant.Client
	Embed   embed.Embedder
	Agent   *agent.Agent
}

func Run(deps Deps, args []string) (any, error) {
	if len(args) == 0 {
		args = []string{"ingest"}
	}
	switch args[0] {
	case "parse":
		return run_parse(args)
	case "classify":
		return run_classify(args)
	case "plan":
		return run_plan(args)
	case "fetch":
		return run_fetch(args)
	case "gold":
		return run_gold(args)
	case "ingest":
		return run_ingest(deps, args)
	case "verify":
		return run_verify(deps)
	case "search":
		return run_search(deps, args)
	case "refresh":
		return run_refresh(deps)
	case "ping":
		return run_ping(deps)
	case "seed":
		return run_seed(deps)
	case "fold":
		return run_fold(deps, args)
	default:
		return nil, fmt.Errorf("unknown command %q\n\nsource: %s", args[0], hood_events.TrackerURL)
	}
}

func run_ingest(deps Deps, args []string) (any, error) {
	if err := require_clients(deps, "ingest"); err != nil {
		return nil, err
	}
	if len(args) > 1 {
		data, err := os.ReadFile(args[1])
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

func run_search(deps Deps, args []string) (any, error) {
	if deps.Vectors == nil {
		return nil, fmt.Errorf("search requires an injected Qdrant client")
	}
	if len(args) < 2 {
		return nil, fmt.Errorf("search requires a query")
	}
	query := strings.Join(args[1:], " ")
	embedder := deps.Embed
	if embedder == nil {
		embedder = lexical.New()
	}
	vecs, err := embedder.Embed([]string{query})
	if err != nil {
		return nil, err
	}
	hits, err := qdrant.SearchHeadlines(deps.Vectors, vecs[0], 5)
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
	if deps.Graph != nil {
		var bolt []map[string]any
		for _, want := range fold.GoldFold() {
			rows, err := neo4j.Fold(deps.Graph, want.Q, want.Qty)
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
	if err := neo4j.ApplyConstraints(deps.Graph); err != nil {
		return nil, err
	}
	if _, err := qdrant.EnsureCollection(deps.Vectors); err != nil {
		return nil, err
	}
	company, stock, events := examples.NFLX()
	if _, err := neo4j.IngestGraph(deps.Graph, company, stock, events); err != nil {
		return nil, err
	}
	if _, err := qdrant.UpsertEvents(deps.Vectors, events, nil); err != nil {
		return nil, err
	}
	rows, err := deps.Graph.Run(
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
	if err := neo4j.ApplyConstraints(deps.Graph); err != nil {
		return nil, err
	}
	if _, err := qdrant.EnsureCollection(deps.Vectors); err != nil {
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
		result, err := neo4j.IngestGraph(deps.Graph, fixture.Company, fixture.Stock, fixture.Events)
		if err != nil {
			return nil, err
		}
		if _, err := qdrant.UpsertEvents(deps.Vectors, fixture.Events, nil); err != nil {
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

func run_fold(deps Deps, args []string) (any, error) {
	if err := require_clients(deps, "fold"); err != nil {
		return nil, err
	}
	if len(args) < 3 {
		return nil, fmt.Errorf("fold requires <q> and <qty>")
	}
	qty, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return nil, err
	}
	rows, err := neo4j.Fold(deps.Graph, args[1], qty)
	if err != nil {
		return nil, err
	}
	return map[string]any{"status": "ok", "q": args[1], "qty": qty, "rows": rows}, nil
}

func run_parse(args []string) (any, error) {
	text, err := read_arg(args, 1)
	if err != nil {
		return nil, err
	}
	return parser.ParseTrackerPage(text), nil
}

func run_classify(args []string) (any, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("classify requires date and headline")
	}
	company, ticker := "", ""
	if len(args) > 3 {
		company = args[3]
	}
	if len(args) > 4 {
		ticker = args[4]
	}
	return classify.ClassifyHeadline(args[2], args[1], company, ticker)
}

func run_fetch(args []string) (any, error) {
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
	if len(args) > 1 {
		if err := os.WriteFile(args[1], []byte(result.Text), 0o644); err != nil {
			return nil, err
		}
		out["path"] = args[1]
		return out, nil
	}
	out["text"] = result.Text
	return out, nil
}

func run_gold(args []string) (any, error) {
	var text string
	if len(args) > 1 {
		data, err := os.ReadFile(args[1])
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

func run_plan(args []string) (any, error) {
	text, err := read_arg(args, 1)
	if err != nil {
		return nil, err
	}
	return ingest.PlanIngest(text), nil
}

func require_clients(deps Deps, cmd string) error {
	if deps.Graph == nil || deps.Vectors == nil || deps.Agent == nil {
		return fmt.Errorf("%s requires injected Neo4j, Qdrant, and ingest agent", cmd)
	}
	return nil
}

func read_arg(args []string, i int) (string, error) {
	if len(args) <= i {
		return "", fmt.Errorf("missing file argument")
	}
	data, err := os.ReadFile(args[i])
	if err != nil {
		return "", err
	}
	return string(data), nil
}
