// Command corporate_actions runs the Robinhood corporate-actions ingest skill.
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
		return run_skill(ctx, manifest, args)
	})
	if err := root.Execute(); err != nil {
		var exit_err *exit_error
		if errors.As(err, &exit_err) {
			os.Exit(exit_err.code)
		}
		os.Exit(2)
	}
}

func run_skill(ctx context.Context, manifest contract.Manifest, args []string) error {
	app, err := di.NewFromEnv(ctx, args)
	if err != nil {
		return err
	}
	defer app.Close(ctx)
	result, err := impl.NewSDK().Run(ctx, manifest, app.Skill().Tools(), contract.Request{Args: args})
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result.Payload); err != nil {
		return &exit_error{error: err, code: 1}
	}
	return nil
}
