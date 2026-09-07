// Command corporate_actions runs the Robinhood corporate-actions ingest skill.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tcw165/fintech-fun/agents/harness/contract"
	"github.com/tcw165/fintech-fun/agents/harness/impl"
	"github.com/tcw165/fintech-fun/agents/harness/skillmd"
	hood_events "github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
	"github.com/tcw165/fintech-fun/ingest_jobs/di"
)

func main() {
	ctx := context.Background()
	doc, err := skillmd.Parse(hood_events.SkillMarkdown)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	manifest := impl.ManifestFrom(doc)
	args := os.Args[1:]
	app, err := di.NewFromEnv(ctx, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer app.Close(ctx)
	result, err := impl.NewSDK().Run(ctx, manifest, app.Skill().Tools(), contract.Request{Args: args})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result.Payload); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
