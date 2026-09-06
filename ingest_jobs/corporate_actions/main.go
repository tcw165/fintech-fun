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
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/skill"
	neo4jimpl "github.com/tcw165/fintech-fun/graph/clients/neo4j/impl"
	qdrantimpl "github.com/tcw165/fintech-fun/graph/clients/qdrant/impl"
)

func needsLiveClients(args []string) bool {
	if len(args) == 0 {
		return true
	}
	switch args[0] {
	case "ping", "seed", "fold", "ingest", "search", "refresh":
		return true
	default:
		return false
	}
}

func wantsOptionalClients(args []string) bool {
	return len(args) > 0 && args[0] == "verify"
}

func main() {
	ctx := context.Background()
	doc, err := skillmd.Parse(hood_events.SkillMarkdown)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	manifest := impl.ManifestFrom(doc)
	s := skill.Skill{}
	args := os.Args[1:]
	if needsLiveClients(args) || wantsOptionalClients(args) {
		driver, err := neo4jimpl.OpenFromEnv()
		if err != nil {
			if needsLiveClients(args) {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
		} else {
			defer driver.Close(ctx)
			if err := driver.Verify(ctx); err != nil {
				if needsLiveClients(args) {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(2)
				}
			} else {
				s.Graph = driver
				s.Vectors = qdrantimpl.NewHTTPFromEnv()
			}
		}
	}
	result, err := impl.NewSDK().Run(ctx, manifest, s.Tools(), contract.Request{Args: args})
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
