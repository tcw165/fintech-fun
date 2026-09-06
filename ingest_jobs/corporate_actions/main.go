// Command corporate_actions runs the Robinhood corporate-actions ingest skill.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tcw165/fintech-fun/agents/harness/contract"
	"github.com/tcw165/fintech-fun/agents/harness/impl"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/skill"
)

func main() {
	runner := impl.New()
	result, err := runner.Run(context.Background(), skill.Skill{}, contract.Request{Args: os.Args[1:]})
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
