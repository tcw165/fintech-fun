package di

import "context"

func NewFromEnv() (*AppContext, error) {
	return Boot(context.Background())
}
