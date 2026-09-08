// Command api_server serves the retail graph HTTP surface.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tcw165/fintech-fun/api_server/di"
	"go.uber.org/fx"
)

func main() {
	if err := new_root(listen).Execute(); err != nil {
		os.Exit(1)
	}
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
