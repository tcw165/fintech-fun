// Command activation runs Gap A HTTP proofs against api_server.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tcw165/fintech-fun/infra/activation/contract"
	"github.com/tcw165/fintech-fun/infra/activation/impl"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: activation healthz|smoke|gold|prove <api-base-url>")
		os.Exit(2)
	}
	cmd, api := os.Args[1], strings.TrimRight(os.Args[2], "/")
	checker := impl.HTTP{}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	switch cmd {
	case contract.StepHealthz, "health":
		step := checker.Health(api)
		_ = enc.Encode(step)
		if !step.OK {
			os.Exit(2)
		}
	case contract.StepSmoke:
		report := checker.Smoke(api)
		_ = enc.Encode(report)
		if !report.AllOK() {
			os.Exit(2)
		}
	case contract.StepGold:
		report := checker.Gold(api)
		_ = enc.Encode(report)
		if !report.AllOK() {
			os.Exit(2)
		}
	case "prove":
		smoke := checker.Smoke(api)
		gold := checker.Gold(api)
		report := contract.Report{Status: "ok", Steps: append(append([]contract.Step{}, smoke.Steps...), gold.Steps...)}
		if !report.AllOK() {
			report.Status = "failed"
		}
		_ = enc.Encode(report)
		if !report.AllOK() {
			os.Exit(2)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", cmd)
		os.Exit(2)
	}
}
