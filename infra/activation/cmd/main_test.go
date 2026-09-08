package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tcw165/fintech-fun/infra/activation/contract"
)

type stub_checker struct {
	api    string
	health contract.Step
	smoke  contract.Report
	gold   contract.Report
}

func (s *stub_checker) Health(api string) contract.Step {
	s.api = api
	return s.health
}

func (s *stub_checker) Smoke(api string) contract.Report {
	s.api = api
	return s.smoke
}

func (s *stub_checker) Gold(api string) contract.Report {
	s.api = api
	return s.gold
}

func execute_cmd(t *testing.T, checker contract.Checker, args ...string) (string, error) {
	t.Helper()
	stdout := &bytes.Buffer{}
	cli := cmd(checker, stdout)
	cli.SetArgs(args)
	err := cli.Execute()
	return stdout.String(), err
}

func TestHealthzOKAndAlias(t *testing.T) {
	checker := &stub_checker{health: contract.Step{Name: contract.StepHealthz, OK: true}}
	out, err := execute_cmd(
		t,
		checker,
		"health",
		"http://127.0.0.1:30080/",
	)
	if err != nil {
		t.Fatalf("health: %v", err)
	}
	if checker.api != "http://127.0.0.1:30080" {
		t.Fatalf("api=%q", checker.api)
	}
	var step contract.Step
	if err := json.Unmarshal([]byte(out), &step); err != nil {
		t.Fatalf("json: %v", err)
	}
	if !step.OK || step.Name != contract.StepHealthz {
		t.Fatalf("step=%+v", step)
	}
}

func TestSmokeFailureExits(t *testing.T) {
	checker := &stub_checker{smoke: contract.Report{Status: "failed", Steps: []contract.Step{{Name: "fold"}}}}
	out, err := execute_cmd(
		t,
		checker,
		"smoke",
		"http://example",
	)
	if err == nil {
		t.Fatal("expected smoke failure")
	}
	if !strings.Contains(out, `"status": "failed"`) {
		t.Fatalf("stdout=%s", out)
	}
}

func TestProveCombinesSteps(t *testing.T) {
	checker := &stub_checker{
		smoke: contract.Report{Status: "ok", Steps: []contract.Step{{Name: "fold", OK: true}}},
		gold:  contract.Report{Status: "ok", Steps: []contract.Step{{Name: "gold:square", OK: true}}},
	}
	out, err := execute_cmd(
		t,
		checker,
		"prove",
		"http://api",
	)
	if err != nil {
		t.Fatalf("prove: %v", err)
	}
	var report contract.Report
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("json: %v", err)
	}
	if report.Status != "ok" || len(report.Steps) != 2 {
		t.Fatalf("report=%+v", report)
	}
}

func TestMissingURL(t *testing.T) {
	_, err := execute_cmd(t, &stub_checker{}, "gold")
	if err == nil {
		t.Fatal("expected missing url")
	}
}

func TestUnknownCommand(t *testing.T) {
	_, err := execute_cmd(
		t,
		&stub_checker{},
		"nope",
		"http://api",
	)
	if err == nil {
		t.Fatal("expected unknown command")
	}
}

func TestExtraArgsRejected(t *testing.T) {
	_, err := execute_cmd(
		t,
		&stub_checker{},
		"gold",
		"http://api",
		"extra",
	)
	if err == nil {
		t.Fatal("expected extra arg rejection")
	}
}
