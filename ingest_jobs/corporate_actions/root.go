package main

import (
	"github.com/spf13/cobra"
	"github.com/tcw165/fintech-fun/ingest_jobs/corporate_actions/request"
)

type run_func func(req request.Request) error

func new_root(run run_func) *cobra.Command {
	var dry_run bool
	var file string
	run_ingest := func(*cobra.Command, []string) error {
		return run(request.Request{
			Name:   "ingest",
			File:   file,
			DryRun: dry_run,
		})
	}
	root := &cobra.Command{
		Use:           "corporate_actions",
		Short:         "Robinhood corporate-actions ingest",
		Long:          "Fetch or read a tracker page, classify headlines, and ingest. --dry-run runs the same path without writing to Neo4j or Qdrant.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE:          run_ingest,
	}
	root.Flags().BoolVar(
		&dry_run,
		"dry-run",
		false,
		"run E2E without writing to Neo4j or Qdrant",
	)
	root.Flags().StringVarP(
		&file,
		"file",
		"f",
		"",
		"tracker page file (default: fetch live)",
	)
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(
		new_named_cmd("refresh", "Live prefix-dedup ingest plus waiting-policy metadata", run),
		new_named_cmd("ping", "Write NFLX fixture and read it back", run),
		new_named_cmd("seed", "Seed Notion fixtures into Neo4j and Qdrant", run),
		new_named_cmd("verify", "Verify memory gold and optional Bolt gold", run),
	)
	return root
}

func new_named_cmd(name, short string, run run_func) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			return run(request.Request{Name: name})
		},
	}
}
