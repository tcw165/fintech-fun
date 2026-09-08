package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tcw165/fintech-fun/ingest_jobs/corporate_actions/request"
)

type exec_result struct {
	req    request.Request
	called bool
	out    string
	err    error
}

func execute_cli(t *testing.T, args ...string) exec_result {
	t.Helper()
	var got exec_result
	cmd := new_root(func(req request.Request) error {
		got.called = true
		got.req = req
		return nil
	})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	got.err = cmd.Execute()
	got.out = buf.String()
	return got
}

func TestDefaultIngest(t *testing.T) {
	got := execute_cli(t)
	if got.err != nil {
		t.Fatalf("empty: %v", got.err)
	}
	if !got.called || got.req.Name != "ingest" {
		t.Fatalf("got %+v, want ingest", got.req)
	}
}

func TestFileAndDryRunFlags(t *testing.T) {
	got := execute_cli(
		t,
		"--file",
		"tracker.txt",
		"--dry-run",
	)
	if got.err != nil {
		t.Fatalf("flags: %v", got.err)
	}
	if got.req.Name != "ingest" || got.req.File != "tracker.txt" || !got.req.DryRun {
		t.Fatalf("got %+v", got.req)
	}
}

func TestDryRunWithoutFile(t *testing.T) {
	got := execute_cli(t, "--dry-run")
	if got.err != nil {
		t.Fatalf("dry-run: %v", got.err)
	}
	if !got.called || !got.req.DryRun || got.req.File != "" {
		t.Fatalf("got %+v", got.req)
	}
}

func TestLeftoverPositionalRejected(t *testing.T) {
	got := execute_cli(
		t,
		"--file",
		"tracker.txt",
		"extra",
	)
	if got.err == nil || got.called {
		t.Fatal("expected leftover positional to fail")
	}
}

func TestRefresh(t *testing.T) {
	got := execute_cli(t, "refresh")
	if got.err != nil {
		t.Fatalf("refresh: %v", got.err)
	}
	if got.req.Name != "refresh" {
		t.Fatalf("got %+v", got.req)
	}
}

func TestUnknownCommand(t *testing.T) {
	got := execute_cli(t, "parse")
	if got.err == nil || got.called {
		t.Fatal("expected unknown command")
	}
}

func TestLogDryRunIsMinimal(t *testing.T) {
	var buf bytes.Buffer
	err := log_dry_run(
		&buf,
		map[string]any{
			"written":   12,
			"skipped":   2,
			"unchanged": false,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if got != "dry-run would_write=12 skipped=2 unchanged=false\n" {
		t.Fatalf("%q", got)
	}
	if strings.Contains(got, "{") {
		t.Fatalf("not minimal: %q", got)
	}
}

func TestHelpRoot(t *testing.T) {
	got := execute_cli(t, "--help")
	if got.err != nil {
		t.Fatalf("help: %v", got.err)
	}
	if got.called {
		t.Fatal("help must not run ingest")
	}
	if !strings.Contains(got.out, "--dry-run") || !strings.Contains(got.out, "--file") {
		t.Fatalf("help:\n%s", got.out)
	}
	if strings.Contains(got.out, "  parse") {
		t.Fatalf("dropped toolkit command still listed:\n%s", got.out)
	}
}
