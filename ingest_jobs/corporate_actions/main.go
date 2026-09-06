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
	neo4jimpl "github.com/tcw165/fintech-fun/graph/clients/neo4j/impl"
	qdrantimpl "github.com/tcw165/fintech-fun/graph/clients/qdrant/impl"
)

func needsLiveClients(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "ping", "seed", "fold":
		return true
	default:
		return false
	}
}

func main() {
	ctx := context.Background()
	s := skill.Skill{}
	if needsLiveClients(os.Args[1:]) {
		driver, err := neo4jimpl.OpenFromEnv()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		defer driver.Close(ctx)
		if err := driver.Verify(ctx); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		s.Graph = driver
		s.Vectors = qdrantimpl.NewHTTPFromEnv()
	}
	runner := impl.New()
	result, err := runner.Run(ctx, s, contract.Request{Args: os.Args[1:]})
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
