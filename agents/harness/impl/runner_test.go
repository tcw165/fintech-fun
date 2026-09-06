package impl_test

import (
	"context"
	"testing"

	"github.com/tcw165/fintech-fun/agents/harness/contract"
	"github.com/tcw165/fintech-fun/agents/harness/impl"
)

type fakeSkill struct {
	called bool
}

func (f *fakeSkill) Name() string { return "fake" }

func (f *fakeSkill) Run(ctx context.Context, req contract.Request) (contract.Result, error) {
	f.called = true
	return contract.Result{Payload: req.Args}, nil
}

func TestRunnerForwardsToSkill(t *testing.T) {
	skill := &fakeSkill{}
	runner := impl.New()
	got, err := runner.Run(context.Background(), skill, contract.Request{Args: []string{"parse", "x.txt"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !skill.called {
		t.Fatal("expected skill.Run to be called")
	}
	payload, ok := got.Payload.([]string)
	if !ok || len(payload) != 2 || payload[0] != "parse" {
		t.Fatalf("payload = %#v", got.Payload)
	}
}
