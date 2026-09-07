// Command api_server serves the retail graph HTTP surface.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tcw165/fintech-fun/api_server/di"
)

func main() {
	if err := new_root(listen).Execute(); err != nil {
		os.Exit(1)
	}
}

func listen(addr string) error {
	app, err := di.NewFromEnv()
	if err != nil {
		return err
	}
	defer app.Close(context.Background())
	app.SetAddr(addr)
	log.Printf("listening on %s", app.Addr())
	if err := http.ListenAndServe(app.Addr(), app.Handler()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
