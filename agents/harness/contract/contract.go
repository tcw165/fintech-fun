// Package contract defines the agent harness protocol. Pure data and interfaces.
package contract

import "context"

type Request struct {
	Args []string
}

type Result struct {
	Payload any
}

type Skill interface {
	Name() string
	Run(ctx context.Context, req Request) (Result, error)
}

type Runner interface {
	Run(ctx context.Context, skill Skill, req Request) (Result, error)
}
