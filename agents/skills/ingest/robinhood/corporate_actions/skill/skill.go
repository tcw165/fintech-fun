// Package skill implements the harness Skill for Robinhood corporate actions ingest.
package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tcw165/fintech-fun/agents/harness/contract"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/classify"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/ingest"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/parser"
)

const name = "ingest/robinhood/corporate_actions"

type Skill struct{}

func (Skill) Name() string { return name }

func (Skill) Run(ctx context.Context, req contract.Request) (contract.Result, error) {
	if len(req.Args) < 1 {
		return contract.Result{}, fmt.Errorf("usage:\n  corporate_actions parse <file>\n  corporate_actions classify <YYYY-MM-DD> <headline> [company] [ticker]\n  corporate_actions plan <file>\n\nsource: %s", hood_events.TrackerURL)
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
	default:
		return contract.Result{}, fmt.Errorf("unknown command %q", req.Args[0])
	}
	if err != nil {
		return contract.Result{}, err
	}
	return contract.Result{Payload: out}, nil
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
