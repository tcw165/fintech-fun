package di

import (
	"context"
	"testing"

	"go.uber.org/fx"
)

func TestModuleBuilds(t *testing.T) {
	app := fx.New(
		Module(context.Background(), "dry-run"),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("module: %v", err)
	}
}

func TestBootDryRunLeavesClientsNilWhenOffline(t *testing.T) {
	app, err := Boot(context.Background(), "dry-run")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = app.Close(context.Background())
	})
	if app.GraphClient() != nil || app.VectorsClient() != nil {
		t.Fatalf("clients set: %+v", app)
	}
	if app.Embedder() == nil {
		t.Fatal("missing embedder")
	}
}

func TestClientNameDryRun(t *testing.T) {
	if got := ClientName(true); got != "dry-run" {
		t.Fatalf("dry-run = %q", got)
	}
	if got := ClientName(false); got != "ingest" {
		t.Fatalf("ingest = %q", got)
	}
}
