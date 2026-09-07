// Command activation runs Gap A HTTP proofs against api_server.
package main

import (
	"os"

	"github.com/tcw165/fintech-fun/infra/activation/impl"
)

func main() {
	if err := new_root(impl.HTTP{}, os.Stdout).Execute(); err != nil {
		os.Exit(2)
	}
}
