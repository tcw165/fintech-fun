package di

import (
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
