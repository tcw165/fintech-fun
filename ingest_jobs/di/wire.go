package di

import "context"

func needs_live_clients(name string) bool {
	if name == "" {
		return true
	}
	switch name {
	case "ping", "seed", "fold", "ingest", "search", "refresh":
		return true
	default:
		return false
	}
}

func wants_optional_clients(name string) bool {
	return name == "verify"
}

func NewFromEnv(
	ctx context.Context,
	name string,
) (*AppContext, error) {
	return Boot(ctx, name)
}
