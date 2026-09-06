// Package skill implements the harness Skill for Robinhood corporate actions ingest.
package skill

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/tcw165/fintech-fun/agents/harness/contract"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/classify"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/fetch"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/ingest"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/parser"
	"github.com/tcw165/fintech-fun/graph/clients/neo4j"
	"github.com/tcw165/fintech-fun/graph/clients/qdrant"
	"github.com/tcw165/fintech-fun/graph/examples"
)

const name = "ingest/robinhood/corporate_actions"

// Skill is the ingest facade. Graph and Vectors are injected by the job binary.
type Skill struct {
	Graph   neo4j.Client
	Vectors qdrant.Client
}

func (Skill) Name() string { return name }

func (s Skill) Run(ctx context.Context, req contract.Request) (contract.Result, error) {
	if len(req.Args) < 1 {
		return contract.Result{}, fmt.Errorf("usage:\n  corporate_actions parse <file>\n  corporate_actions classify <YYYY-MM-DD> <headline> [company] [ticker]\n  corporate_actions plan <file>\n  corporate_actions fetch [outfile]\n  corporate_actions gold [file]\n  corporate_actions ingest [file]\n  corporate_actions ping\n  corporate_actions seed\n  corporate_actions fold <q> <qty>\n\nsource: %s", hood_events.TrackerURL)
	}
	var (
		out any
		err error
	)
	switch req.Args[0] {
	case "parse":
		out, err = runParse(req.Args)
	case "classify":
		out, err = runClassify(req.Args)
	case "plan":
		out, err = runPlan(req.Args)
	case "fetch":
		out, err = runFetch(req.Args)
	case "gold":
		out, err = runGold(req.Args)
	case "ingest":
		out, err = s.runIngest(req.Args)
	case "ping":
		out, err = s.runPing()
	case "seed":
		out, err = s.runSeed()
	case "fold":
		out, err = s.runFold(req.Args)
	default:
		return contract.Result{}, fmt.Errorf("unknown command %q", req.Args[0])
	}
	if err != nil {
		return contract.Result{}, err
	}
	return contract.Result{Payload: out}, nil
}

func (s Skill) runIngest(args []string) (any, error) {
	if err := s.requireClients("ingest"); err != nil {
		return nil, err
	}
	text, err := trackerText(args)
	if err != nil {
		return nil, err
	}
	if err := neo4j.ApplyConstraints(s.Graph); err != nil {
		return nil, err
	}
	stats, err := ingest.IngestText(s.Graph, text, s.Vectors)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"status":     "ok",
		"written":    stats.Written,
		"skipped":    stats.Skipped,
		"companies":  len(stats.Companies),
		"stocks":     len(stats.Stocks),
		"company_names": stats.Companies,
		"tickers":    stats.Stocks,
	}, nil
}

func trackerText(args []string) (string, error) {
	if len(args) > 1 {
		data, err := os.ReadFile(args[1])
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	page, err := fetch.FetchTracker(nil, "")
	if err != nil {
		return "", err
	}
	return page.Text, nil
}

func (s Skill) runPing() (any, error) {
	if s.Graph == nil || s.Vectors == nil {
		return nil, fmt.Errorf("ping requires injected Neo4j and Qdrant clients")
	}
	if err := neo4j.ApplyConstraints(s.Graph); err != nil {
		return nil, err
	}
	if _, err := qdrant.EnsureCollection(s.Vectors); err != nil {
		return nil, err
	}
	company, stock, events := examples.NFLX()
	if _, err := neo4j.IngestGraph(s.Graph, company, stock, events); err != nil {
		return nil, err
	}
	if _, err := qdrant.UpsertEvents(s.Vectors, events, nil); err != nil {
		return nil, err
	}
	rows, err := s.Graph.Run(
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

func (s Skill) requireClients(cmd string) error {
	if s.Graph == nil || s.Vectors == nil {
		return fmt.Errorf("%s requires injected Neo4j and Qdrant clients", cmd)
	}
	return nil
}

func (s Skill) runSeed() (any, error) {
	if err := s.requireClients("seed"); err != nil {
		return nil, err
	}
	if err := neo4j.ApplyConstraints(s.Graph); err != nil {
		return nil, err
	}
	if _, err := qdrant.EnsureCollection(s.Vectors); err != nil {
		return nil, err
	}
	type row struct {
		Company string `json:"company"`
		Ticker  string `json:"ticker"`
		Count   int    `json:"events"`
	}
	var events []row
	seeded := 0
	eventCount := 0
	for _, fixture := range examples.All() {
		result, err := neo4j.IngestGraph(s.Graph, fixture.Company, fixture.Stock, fixture.Events)
		if err != nil {
			return nil, err
		}
		if _, err := qdrant.UpsertEvents(s.Vectors, fixture.Events, nil); err != nil {
			return nil, err
		}
		seeded++
		eventCount += result.Events
		events = append(events, row{Company: result.Company, Ticker: result.Ticker, Count: result.Events})
	}
	return map[string]any{
		"status":   "ok",
		"fixtures": seeded,
		"events":   eventCount,
		"rows":     events,
	}, nil
}

func (s Skill) runFold(args []string) (any, error) {
	if err := s.requireClients("fold"); err != nil {
		return nil, err
	}
	if len(args) < 3 {
		return nil, fmt.Errorf("fold requires <q> and <qty>")
	}
	qty, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return nil, err
	}
	rows, err := neo4j.Fold(s.Graph, args[1], qty)
	if err != nil {
		return nil, err
	}
	return map[string]any{"status": "ok", "q": args[1], "qty": qty, "rows": rows}, nil
}

func runParse(args []string) (any, error) {
	text, err := readArg(args, 1)
	if err != nil {
		return nil, err
	}
	return parser.ParseTrackerPage(text), nil
}

func runClassify(args []string) (any, error) {
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

func runFetch(args []string) (any, error) {
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

func runGold(args []string) (any, error) {
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

func runPlan(args []string) (any, error) {
	text, err := readArg(args, 1)
	if err != nil {
		return nil, err
	}
	return ingest.PlanIngest(text), nil
}

func readArg(args []string, i int) (string, error) {
	if len(args) <= i {
		return "", fmt.Errorf("missing file argument")
	}
	data, err := os.ReadFile(args[i])
	if err != nil {
		return "", err
	}
	return string(data), nil
}
