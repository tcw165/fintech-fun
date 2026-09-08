package di

import "context"

func needs_live_clients(name string) bool {
	return name == "" || name == "ingest"
}

func wants_optional_clients(name string) bool {
	return name == "dry-run"
}

func NewFromEnv(
	ctx context.Context,
	name string,
) (*AppContext, error) {
	return Boot(ctx, name)
}
