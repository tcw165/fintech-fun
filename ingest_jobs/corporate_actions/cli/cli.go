// Package cli is the Cobra tree for the corporate_actions job binary.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tcw165/fintech-fun/ingest_jobs/corporate_actions/request"
)

// Run dispatches one parsed Request.
type Run func(req request.Request) error

type tool_spec struct {
	name  string
	use   string
	short string
	flags func(*cobra.Command)
}

func New(run Run) *cobra.Command {
	root := &cobra.Command{
		Use:           "corporate_actions",
		Short:         "Robinhood corporate-actions ingest agent",
		Long:          "Parse, classify, and ingest Robinhood tracker headlines. Empty args run the custom ingest agent.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			req, err := read_request(cmd, "")
			if err != nil {
				return err
			}
			return run(req)
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
		{"parse", "parse", "Parse a saved tracker page", flags_file_required},
		{"classify", "classify", "Classify one headline", flags_classify},
		{"plan", "plan", "Plan ingest writes from a saved page", flags_file_required},
		{"fetch", "fetch", "Fetch the live tracker page", flags_fetch},
		{"gold", "gold", "Classify gold report from a file or live fetch", flags_file_optional},
		{"ingest", "ingest", "Prefix-dedup ingest via the custom agent", flags_ingest},
		{"verify", "verify", "Verify memory gold and optional Bolt gold", nil},
		{"search", "search", "Search Event headlines in Qdrant", flags_search},
		{"refresh", "refresh", "Live prefix-dedup ingest plus waiting-policy metadata", nil},
		{"ping", "ping", "Write NFLX fixture and read it back", nil},
		{"seed", "seed", "Seed Notion fixtures into Neo4j and Qdrant", nil},
		{"fold", "fold", "Fold what a holder has now", flags_fold},
	}
}

func new_tool_cmd(spec tool_spec, run Run) *cobra.Command {
	cmd := &cobra.Command{
		Use:   spec.use,
		Short: spec.short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			req, err := read_request(cmd, spec.name)
			if err != nil {
				return err
			}
			if spec.name == "ingest" && req.DryRun && req.File == "" {
				return fmt.Errorf("ingest --dry-run requires --file")
			}
			return run(req)
		},
	}
	if spec.flags != nil {
		spec.flags(cmd)
	}
	return cmd
}

func flags_file_required(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "tracker page file")
	_ = cmd.MarkFlagRequired("file")
}

func flags_file_optional(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "tracker page file")
}

func flags_ingest(cmd *cobra.Command) {
	cmd.Flags().StringP("file", "f", "", "tracker page file")
	cmd.Flags().Bool("dry-run", false, "plan writes without touching Neo4j or Qdrant")
}

func flags_fetch(cmd *cobra.Command) {
	cmd.Flags().StringP("out", "o", "", "write tracker page to this path")
}

func flags_search(cmd *cobra.Command) {
	cmd.Flags().StringP("query", "q", "", "headline search query")
	cmd.Flags().IntP("limit", "n", 5, "max hits")
	_ = cmd.MarkFlagRequired("query")
}

func flags_fold(cmd *cobra.Command) {
	cmd.Flags().String("q", "", "holder query")
	cmd.Flags().Float64("qty", 0, "share quantity")
	_ = cmd.MarkFlagRequired("q")
	_ = cmd.MarkFlagRequired("qty")
}

func flags_classify(cmd *cobra.Command) {
	cmd.Flags().String("date", "", "headline date YYYY-MM-DD")
	cmd.Flags().String("headline", "", "headline text")
	cmd.Flags().String("company", "", "company name")
	cmd.Flags().String("ticker", "", "ticker")
	_ = cmd.MarkFlagRequired("date")
	_ = cmd.MarkFlagRequired("headline")
}

func read_request(cmd *cobra.Command, name string) (request.Request, error) {
	req := request.Request{Name: name}
	var err error
	if cmd.Flags().Lookup("file") != nil {
		req.File, err = cmd.Flags().GetString("file")
		if err != nil {
			return req, err
		}
	}
	if cmd.Flags().Lookup("out") != nil {
		req.Out, err = cmd.Flags().GetString("out")
		if err != nil {
			return req, err
		}
	}
	if cmd.Flags().Lookup("query") != nil {
		req.Query, err = cmd.Flags().GetString("query")
		if err != nil {
			return req, err
		}
	}
	if cmd.Flags().Lookup("limit") != nil {
		req.Limit, err = cmd.Flags().GetInt("limit")
		if err != nil {
			return req, err
		}
	}
	if cmd.Flags().Lookup("dry-run") != nil {
		req.DryRun, err = cmd.Flags().GetBool("dry-run")
		if err != nil {
			return req, err
		}
	}
	if cmd.Flags().Lookup("qty") != nil {
		req.Qty, err = cmd.Flags().GetFloat64("qty")
		if err != nil {
			return req, err
		}
	}
	if cmd.Flags().Lookup("q") != nil {
		req.Q, err = cmd.Flags().GetString("q")
		if err != nil {
			return req, err
		}
	}
	if cmd.Flags().Lookup("date") != nil {
		req.Date, err = cmd.Flags().GetString("date")
		if err != nil {
			return req, err
		}
	}
	if cmd.Flags().Lookup("headline") != nil {
		req.Headline, err = cmd.Flags().GetString("headline")
		if err != nil {
			return req, err
		}
	}
	if cmd.Flags().Lookup("company") != nil {
		req.Company, err = cmd.Flags().GetString("company")
		if err != nil {
			return req, err
		}
	}
	if cmd.Flags().Lookup("ticker") != nil {
		req.Ticker, err = cmd.Flags().GetString("ticker")
		if err != nil {
			return req, err
		}
	}
	return req, nil
}
