// Package cli is the Cobra tree for the corporate_actions job binary.
package cli

import (
	"github.com/spf13/cobra"
)

// Run dispatches the selected tool as Request.Args, command name first.
type Run func(args []string) error

type tool_spec struct {
	name  string
	use   string
	short string
	args  cobra.PositionalArgs
}

func New(run Run) *cobra.Command {
	root := &cobra.Command{
		Use:           "corporate_actions",
		Short:         "Robinhood corporate-actions ingest agent",
		Long:          "Parse, classify, and ingest Robinhood tracker headlines. Empty args run the custom ingest agent.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			return run(nil)
		},
	}
	root.CompletionOptions.DisableDefaultCmd = true
	for _, spec := range tools() {
		root.AddCommand(new_tool_cmd(spec, run))
	}
	return root
}

func tools() []tool_spec {
	return []tool_spec{
		{"parse", "parse <file>", "Parse a saved tracker page", cobra.ExactArgs(1)},
		{"classify", "classify <YYYY-MM-DD> <headline> [company] [ticker]", "Classify one headline", cobra.RangeArgs(2, 4)},
		{"plan", "plan <file>", "Plan ingest writes from a saved page", cobra.ExactArgs(1)},
		{"fetch", "fetch [outfile]", "Fetch the live tracker page", cobra.MaximumNArgs(1)},
		{"gold", "gold [file]", "Classify gold report from a file or live fetch", cobra.MaximumNArgs(1)},
		{"ingest", "ingest [file]", "Prefix-dedup ingest via the custom agent", cobra.MaximumNArgs(1)},
		{"verify", "verify", "Verify memory gold and optional Bolt gold", cobra.NoArgs},
		{"search", "search <query>", "Search Event headlines in Qdrant", cobra.MinimumNArgs(1)},
		{"refresh", "refresh", "Live prefix-dedup ingest plus waiting-policy metadata", cobra.NoArgs},
		{"ping", "ping", "Write NFLX fixture and read it back", cobra.NoArgs},
		{"seed", "seed", "Seed Notion fixtures into Neo4j and Qdrant", cobra.NoArgs},
		{"fold", "fold <q> <qty>", "Fold what a holder has now", cobra.ExactArgs(2)},
	}
}

func new_tool_cmd(spec tool_spec, run Run) *cobra.Command {
	return &cobra.Command{
		Use:   spec.use,
		Short: spec.short,
		Args:  spec.args,
		RunE: func(_ *cobra.Command, args []string) error {
			return run(append([]string{spec.name}, args...))
		},
	}
}
