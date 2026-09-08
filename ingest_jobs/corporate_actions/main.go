// Command corporate_actions runs the Robinhood corporate-actions ingest job.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/tcw165/fintech-fun/ingest_jobs/corporate_actions/cli_params"
	"github.com/tcw165/fintech-fun/ingest_jobs/corporate_actions/tools"
	"github.com/tcw165/fintech-fun/ingest_jobs/di"
)

type exit_error struct {
	error
	code int
}

type run_func func(req cli_params.CliParams) error

func main() {
	ctx := context.Background()
	if err := cmd(func(req cli_params.CliParams) error {
		app, err := di.Boot(
			ctx,
			di.ClientName(req.DryRun),
		)
		if err != nil {
			return err
		}
		defer app.Close(ctx)
		payload, err := tools.Run(
			tools.Deps{
				GraphClient:   app.GraphClient(),
				VectorsClient: app.VectorsClient(),
				Embedder:      app.Embedder(),
				Agent:         app.Agent(),
			},
			req,
		)
		if err != nil {
			return err
		}
		return write_result(os.Stdout, req.DryRun, payload)
	}).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		var exit_err *exit_error
		if errors.As(err, &exit_err) {
			os.Exit(exit_err.code)
		}
		os.Exit(2)
	}
}

func cmd(run run_func) *cobra.Command {
	var dry_run bool
	var file string
	cmd := &cobra.Command{
		Use:           "corporate_actions",
		Short:         "Robinhood corporate-actions ingest",
		Long:          "Fetch or read a tracker page, classify headlines, and ingest. --dry-run runs the same path without writing to Neo4j or Qdrant.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			return run(cli_params.CliParams{
				File:   file,
				DryRun: dry_run,
			})
		},
	}
	cmd.Flags().BoolVar(
		&dry_run,
		"dry-run",
		false,
		"run E2E without writing to Neo4j or Qdrant",
	)
	cmd.Flags().StringVarP(
		&file,
		"file",
		"f",
		"",
		"tracker page file (default: fetch live)",
	)
	cmd.CompletionOptions.DisableDefaultCmd = true
	return cmd
}

func write_result(stdout io.Writer, dry_run bool, payload any) error {
	if dry_run {
		return log_dry_run(stdout, payload)
	}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		return &exit_error{error: err, code: 1}
	}
	return nil
}

func log_dry_run(stdout io.Writer, payload any) error {
	out, _ := payload.(map[string]any)
	if out == nil {
		fmt.Fprintln(stdout, "dry-run")
		return nil
	}
	fmt.Fprintf(
		stdout,
		"dry-run would_write=%v skipped=%v unchanged=%v\n",
		out["written"],
		out["skipped"],
		out["unchanged"],
	)
	return nil
}
