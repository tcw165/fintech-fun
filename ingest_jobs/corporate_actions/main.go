// Command corporate_actions runs the Robinhood corporate-actions ingest job.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/tcw165/fintech-fun/agents/harness/contract"
	"github.com/tcw165/fintech-fun/agents/harness/impl"
	"github.com/tcw165/fintech-fun/agents/harness/skillmd"
	hood_events "github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
	"github.com/tcw165/fintech-fun/ingest_jobs/corporate_actions/cli"
	"github.com/tcw165/fintech-fun/ingest_jobs/di"
)

type exit_error struct {
	error
	code int
}

func main() {
	ctx := context.Background()
	doc, err := skillmd.Parse(hood_events.SkillMarkdown)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	manifest := impl.ManifestFrom(doc)
	root := cli.New(func(args []string) error {
		return run(ctx, manifest, args)
	})
	if err := root.Execute(); err != nil {
		var exit_err *exit_error
		if errors.As(err, &exit_err) {
			os.Exit(exit_err.code)
		}
		os.Exit(2)
	}
}

func run(ctx context.Context, manifest contract.Manifest, args []string) error {
	app, err := di.NewFromEnv(ctx, args)
	if err != nil {
		return err
	}
	defer app.Close(ctx)
	if di.UsesAgent(args) {
		return run_agent(app, args)
	}
	return run_skill(ctx, app, manifest, args)
}

func run_agent(app *di.AppContext, args []string) error {
	ingest_agent := app.Agent()
	var payload map[string]any
	switch {
	case len(args) == 0 || args[0] == "ingest":
		if len(args) > 1 {
			data, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}
			result, err := ingest_agent.IngestText(string(data))
			if err != nil {
				return err
			}
			payload = result.Payload()
		} else {
			result, err := ingest_agent.IngestLive("")
			if err != nil {
				return err
			}
			payload = result.Payload()
		}
	case args[0] == "refresh":
		result, err := ingest_agent.Refresh("")
		if err != nil {
			return err
		}
		payload = result.Payload()
	default:
		return fmt.Errorf("unknown agent command %q", args[0])
	}
	return write_payload(payload)
}

func run_skill(ctx context.Context, app *di.AppContext, manifest contract.Manifest, args []string) error {
	result, err := impl.NewSDK().Run(ctx, manifest, app.Skill().Tools(), contract.Request{Args: args})
	if err != nil {
		return err
	}
	return write_payload(result.Payload)
}

func write_payload(payload any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		return &exit_error{error: err, code: 1}
	}
	return nil
}
