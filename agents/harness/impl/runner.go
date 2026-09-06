// Package impl is the default harness runner. Child of //agents/harness/contract.
package impl

import (
	"context"

	"github.com/tcw165/fintech-fun/agents/harness/contract"
)

type Runner struct{}

func New() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, skill contract.Skill, req contract.Request) (contract.Result, error) {
	return skill.Run(ctx, req)
}

var _ contract.Runner = (*Runner)(nil)
