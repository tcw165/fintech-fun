package di

import "go.uber.org/fx"

// Module is the ingest-job Fx graph. Empty until providers are registered.
func Module() fx.Option {
	return fx.Options()
}
