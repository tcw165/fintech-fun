package di

import (
	"log"
	"os"

	neo4jimpl "github.com/tcw165/fintech-fun/graph/clients/neo4j/impl"
	qdrantimpl "github.com/tcw165/fintech-fun/graph/clients/qdrant/impl"
	"github.com/tcw165/fintech-fun/graph/embed/lexical"
)

func NewFromEnv() (*AppContext, error) {
	addr := ":8080"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	app := New(nil, qdrantimpl.NewHTTPFromEnv(), lexical.New(), nil)
	app.addr = addr
	if driver, err := neo4jimpl.OpenFromEnv(); err != nil {
		log.Printf("neo4j unavailable: %v", err)
	} else {
		app.graph_db = driver
		app.closer = driver.Close
	}
	return app, nil
}
