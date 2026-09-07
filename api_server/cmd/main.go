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
	app, err := di.NewFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer app.Close(context.Background())
	log.Printf("listening on %s", app.Addr())
	if err := http.ListenAndServe(app.Addr(), app.Handler()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
