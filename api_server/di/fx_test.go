package di

import (
	"context"
	"testing"

	"go.uber.org/fx"
)

func TestModuleBuilds(t *testing.T) {
	app := fx.New(
		Module(),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("module: %v", err)
	}
}

func TestBootWiresEmbedder(t *testing.T) {
	app, err := Boot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = app.Close(context.Background())
	})
	if app.Embedder() == nil {
		t.Fatal("missing embedder")
	}
	if app.VectorsClient() == nil {
		t.Fatal("missing qdrant")
	}
	if app.Addr() != ":8080" {
		t.Fatalf("addr = %q", app.Addr())
	}
}
