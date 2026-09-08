// Command corporate_actions runs the Robinhood corporate-actions ingest job.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"

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
	root := cli.New(func(args []string) error {
		return run(ctx, request_from_args(args))
	})
	if err := root.Execute(); err != nil {
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

// request_from_args is a temporary shim until the Cobra flag PR reads Request via Get*.
func request_from_args(args []string) request.Request {
	if len(args) == 0 {
		return request.Request{}
	}
	req := request.Request{Name: args[0]}
	switch req.Name {
	case "parse", "plan", "ingest", "gold":
		if len(args) > 1 {
			req.File = args[1]
		}
	case "fetch":
		if len(args) > 1 {
			req.Out = args[1]
		}
	case "search":
		req.Query = strings.Join(args[1:], " ")
		req.Limit = 5
	case "fold":
		if len(args) > 1 {
			req.Q = args[1]
		}
		if len(args) > 2 {
			qty, err := strconv.ParseFloat(args[2], 64)
			if err == nil {
				req.Qty = qty
			}
		}
	case "classify":
		if len(args) > 1 {
			req.Date = args[1]
		}
		if len(args) > 2 {
			req.Headline = args[2]
		}
		if len(args) > 3 {
			req.Company = args[3]
		}
		if len(args) > 4 {
			req.Ticker = args[4]
		}
	}
	return req
}
