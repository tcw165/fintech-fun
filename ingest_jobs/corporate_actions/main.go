// Command corporate_actions runs the Robinhood corporate-actions ingest job.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/tcw165/fintech-fun/ingest_jobs/corporate_actions/cli"
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
	root := cli.New(func(req request.Request) error {
		return run(ctx, req)
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

func run(ctx context.Context, req request.Request) error {
	app, err := di.NewFromEnv(ctx, req.Name)
	if err != nil {
		return err
	}
	defer app.Close(ctx)
	payload, err := tools.Run(tools.Deps{
		GraphClient:   app.GraphClient(),
		VectorsClient: app.VectorsClient(),
		Embedder:      app.Embedder(),
		Agent:         app.Agent(),
	}, req)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		return &exit_error{error: err, code: 1}
	}
	return nil
}
