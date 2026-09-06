// Command api_server serves the retail graph HTTP surface.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tcw165/fintech-fun/api_server/contract"
	"github.com/tcw165/fintech-fun/api_server/impl"
	neo4jimpl "github.com/tcw165/fintech-fun/graph/clients/neo4j/impl"
	qdrantimpl "github.com/tcw165/fintech-fun/graph/clients/qdrant/impl"
	"github.com/tcw165/fintech-fun/graph/embed/lexical"
)

func main() {
	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	var folder contract.Folder
	if driver, err := neo4jimpl.OpenFromEnv(); err != nil {
		log.Printf("neo4j unavailable: %v", err)
	} else {
		defer driver.Close(context.Background())
		folder = impl.Neo4jFold{Graph: driver}
	}
	searcher := impl.QdrantSearch{Vectors: qdrantimpl.NewHTTPFromEnv(), Embed: lexical.New()}
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, impl.New(impl.StaticOK{}, folder, searcher)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
