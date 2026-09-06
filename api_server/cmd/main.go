// Command api_server serves the retail graph HTTP surface.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tcw165/fintech-fun/api_server/impl"
)

func main() {
	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, impl.New(impl.StaticOK{})); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
