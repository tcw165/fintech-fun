// Command corporate_actions runs the Robinhood corporate-actions ingest job.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/tcw165/fintech-fun/ingest_jobs/corporate_actions/request"
	"github.com/tcw165/fintech-fun/ingest_jobs/corporate_actions/tools"
	"github.com/tcw165/fintech-fun/ingest_jobs/di"
)

type exit_error struct {
	error
	code int
}

func main() {
	ctx := context.Background()
	root := new_root(func(req request.Request) error {
		app, err := di.Boot(
			ctx,
			di.ClientName(req.Name, req.DryRun),
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
	})
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		var exit_err *exit_error
		if errors.As(err, &exit_err) {
			os.Exit(exit_err.code)
		}
		os.Exit(2)
	}
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
