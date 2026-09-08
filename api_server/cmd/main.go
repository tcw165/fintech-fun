// Command api_server serves the retail graph HTTP surface.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tcw165/fintech-fun/api_server/di"
	"go.uber.org/fx"
)

func main() {
	if err := cmd(listen).Execute(); err != nil {
		os.Exit(1)
	}
}

func cmd(listen func(addr string) error) *cobra.Command {
	port := "8080"
	run_listen := func(*cobra.Command, []string) error {
		return listen(listen_addr(port, os.Getenv("PORT")))
	}
	cmd := &cobra.Command{
		Use:           "api_server",
		Short:         "Retail graph HTTP surface",
		Long:          "Serve /healthz, /fold, /search, and /v1/graph. No args starts the server so the k8s Deployment stays unchanged.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE:          run_listen,
	}
	cmd.PersistentFlags().StringVar(&port, "port", "8080", "listen port (PORT env wins if set)")
	cmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Listen for HTTP (default)",
		Args:  cobra.NoArgs,
		RunE:  run_listen,
	})
	cmd.CompletionOptions.DisableDefaultCmd = true
	return cmd
}

func listen_addr(port_flag, port_env string) string {
	port := port_flag
	if port_env != "" {
		port = port_env
	}
	if port == "" {
		port = "8080"
	}
	return ":" + strings.TrimPrefix(port, ":")
}

func listen(addr string) error {
	fx_app := fx.New(
		di.Module(),
		fx.Invoke(func(
			lc fx.Lifecycle,
			app *di.AppContext,
		) {
			app.SetAddr(addr)
			server := &http.Server{
				Addr:    app.Addr(),
				Handler: app.Handler(),
			}
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						err := server.ListenAndServe()
						if err != nil && err != http.ErrServerClosed {
							log.Print(err)
						}
					}()
					log.Printf("listening on %s", app.Addr())
					return nil
				},
				OnStop: server.Shutdown,
			})
		}),
		fx.NopLogger,
	)
	if err := fx_app.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	fx_app.Run()
	return fx_app.Err()
}
