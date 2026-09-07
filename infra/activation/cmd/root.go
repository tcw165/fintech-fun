package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tcw165/fintech-fun/infra/activation/contract"
)

func new_root(checker contract.Checker, stdout io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:           "activation",
		Short:         "Gap A HTTP proofs against api_server",
		Long:          "Run healthz, smoke, gold, or prove checks against an api_server base URL.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(
		new_step_cmd("healthz", "GET /healthz", []string{"health"}, cobra.ExactArgs(1), func(api string) error {
			step := checker.Health(api)
			if err := write_json(stdout, step); err != nil {
				return err
			}
			if !step.OK {
				return fmt.Errorf("healthz failed")
			}
			return nil
		}),
		new_step_cmd("smoke", "GET /healthz /fold /search /v1/graph", nil, cobra.ExactArgs(1), func(api string) error {
			return write_report(stdout, checker.Smoke(api))
		}),
		new_step_cmd("gold", "HTTP fold rows match Notion gold tables", nil, cobra.ExactArgs(1), func(api string) error {
			return write_report(stdout, checker.Gold(api))
		}),
		new_step_cmd("prove", "smoke + gold combined", nil, cobra.ExactArgs(1), func(api string) error {
			smoke := checker.Smoke(api)
			gold := checker.Gold(api)
			report := contract.Report{
				Status: "ok",
				Steps:  append(append([]contract.Step{}, smoke.Steps...), gold.Steps...),
			}
			if !report.AllOK() {
				report.Status = "failed"
			}
			return write_report(stdout, report)
		}),
	)
	return root
}

func new_step_cmd(name, short string, aliases []string, args cobra.PositionalArgs, run func(api string) error) *cobra.Command {
	return &cobra.Command{
		Use:     name + " <api-base-url>",
		Short:   short,
		Aliases: aliases,
		Args:    args,
		RunE: func(_ *cobra.Command, argv []string) error {
			return run(strings.TrimRight(argv[0], "/"))
		},
	}
}

func write_report(stdout io.Writer, report contract.Report) error {
	if err := write_json(stdout, report); err != nil {
		return err
	}
	if !report.AllOK() {
		return fmt.Errorf("%s failed", report.Status)
	}
	return nil
}

func write_json(stdout io.Writer, value any) error {
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}
