// Command crawler reads Robinhood Corporate Actions Tracker text.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tcw165/fintech-fun/ingest/agents/robinhood/corporate_actions"
	"github.com/tcw165/fintech-fun/ingest/agents/robinhood/corporate_actions/classify"
	"github.com/tcw165/fintech-fun/ingest/agents/robinhood/corporate_actions/ingest"
	"github.com/tcw165/fintech-fun/ingest/agents/robinhood/corporate_actions/parser"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var (
		out any
		err error
	)
	switch os.Args[1] {
	case "parse":
		out = parser.ParseTrackerPage(readArg(2))
	case "classify":
		if len(os.Args) < 5 {
			usage()
			os.Exit(2)
		}
		company, ticker := "", ""
		if len(os.Args) > 5 {
			company = os.Args[5]
		}
		if len(os.Args) > 6 {
			ticker = os.Args[6]
		}
		out, err = classify.ClassifyHeadline(os.Args[3], os.Args[2], company, ticker)
	case "plan":
		out = ingest.PlanIngest(readArg(2))
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func readArg(i int) string {
	if len(os.Args) <= i {
		usage()
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[i])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return string(data)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage:\n  crawler parse <file>\n  crawler classify <YYYY-MM-DD> <headline> [company] [ticker]\n  crawler plan <file>\n\nsource: %s\n", hood_events.TrackerURL)
}
