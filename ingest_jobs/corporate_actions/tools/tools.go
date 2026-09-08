// Package tools dispatches corporate_actions job commands. Child of ingest packages.
package tools

import (
	"fmt"
	"os"

	hood_events "github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/agent"
	"github.com/tcw165/fintech-fun/graph/clients/neo4j"
	"github.com/tcw165/fintech-fun/graph/clients/qdrant"
	"github.com/tcw165/fintech-fun/graph/embed"
	"github.com/tcw165/fintech-fun/graph/examples"
	"github.com/tcw165/fintech-fun/graph/fold"
	"github.com/tcw165/fintech-fun/ingest_jobs/corporate_actions/cli_params"
)

type Deps struct {
	GraphClient   neo4j.Client
	VectorsClient qdrant.Client
	Embedder      embed.Embedder
	Agent         *agent.Agent
}

func Run(deps Deps, req cli_params.CliParams) (any, error) {
	if req.Name == "" {
		req.Name = "ingest"
	}
	switch req.Name {
	case "ingest":
		return run_ingest(deps, req)
	case "verify":
		return run_verify(deps)
	case "refresh":
		return run_refresh(deps)
	case "ping":
		return run_ping(deps)
	case "seed":
		return run_seed(deps)
	default:
		return nil, fmt.Errorf("unknown command %q\n\nsource: %s", req.Name, hood_events.TrackerURL)
	}
}

func run_ingest(deps Deps, req cli_params.CliParams) (any, error) {
	ingest_agent := deps.Agent
	if ingest_agent == nil {
		ingest_agent = agent.New(
			deps.GraphClient,
			deps.VectorsClient,
			deps.Embedder,
			nil,
		)
	}
	if req.DryRun {
		return run_dry_run(ingest_agent, req)
	}
	if err := require_clients(deps, "ingest"); err != nil {
		return nil, err
	}
	if req.File != "" {
		data, err := os.ReadFile(req.File)
		if err != nil {
			return nil, err
		}
		result, err := ingest_agent.IngestText(string(data))
		if err != nil {
			return nil, err
		}
		return result.Payload(), nil
	}
	result, err := ingest_agent.IngestLive("")
	if err != nil {
		return nil, err
	}
	return result.Payload(), nil
}

func run_dry_run(ingest_agent *agent.Agent, req cli_params.CliParams) (any, error) {
	if req.File != "" {
		data, err := os.ReadFile(req.File)
		if err != nil {
			return nil, err
		}
		result, err := ingest_agent.DryRunText(string(data))
		if err != nil {
			return nil, err
		}
		return result.Payload(), nil
	}
	result, err := ingest_agent.DryRunLive("")
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

func require_clients(deps Deps, cmd string) error {
	if deps.GraphClient == nil || deps.VectorsClient == nil || deps.Agent == nil {
		return fmt.Errorf("%s requires injected Neo4j, Qdrant, and ingest agent", cmd)
	}
	return nil
}
