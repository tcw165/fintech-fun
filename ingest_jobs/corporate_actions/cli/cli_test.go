package cli

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
	cmd := New(func(req request.Request) error {
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
	if !got.called || got.req.Name != "" {
		t.Fatalf("got %+v, want empty Name so tools default to ingest", got.req)
	}
}

func TestParseForwardsFile(t *testing.T) {
	got := execute_cli(t, "parse", "--file", "tracker.txt")
	if got.err != nil {
		t.Fatalf("parse: %v", got.err)
	}
	if got.req.Name != "parse" || got.req.File != "tracker.txt" {
		t.Fatalf("got %+v", got.req)
	}
}

func TestParseRequiresFile(t *testing.T) {
	got := execute_cli(t, "parse")
	if got.err == nil || got.called {
		t.Fatal("expected required --file")
	}
	if !strings.Contains(got.err.Error(), "file") {
		t.Fatalf("got %v", got.err)
	}
}

func TestSearchFlags(t *testing.T) {
	got := execute_cli(t, "search", "--query", "LivePerson stock merger", "--limit", "3")
	if got.err != nil {
		t.Fatalf("search: %v", got.err)
	}
	if got.req.Name != "search" || got.req.Query != "LivePerson stock merger" || got.req.Limit != 3 {
		t.Fatalf("got %+v", got.req)
	}
}

func TestIngestDryRun(t *testing.T) {
	got := execute_cli(t, "ingest", "--file", "tracker.txt", "--dry-run")
	if got.err != nil {
		t.Fatalf("ingest: %v", got.err)
	}
	if got.req.Name != "ingest" || got.req.File != "tracker.txt" || !got.req.DryRun {
		t.Fatalf("got %+v", got.req)
	}
}

func TestIngestDryRunRequiresFile(t *testing.T) {
	got := execute_cli(t, "ingest", "--dry-run")
	if got.err == nil || got.called {
		t.Fatal("expected --dry-run to require --file")
	}
}

func TestFoldFlags(t *testing.T) {
	got := execute_cli(t, "fold", "--q", "square", "--qty", "10")
	if got.err != nil {
		t.Fatalf("fold: %v", got.err)
	}
	if got.req.Name != "fold" || got.req.Q != "square" || got.req.Qty != 10 {
		t.Fatalf("got %+v", got.req)
	}
}

func TestLeftoverPositionalRejected(t *testing.T) {
	got := execute_cli(t, "parse", "--file", "tracker.txt", "extra")
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
	got := execute_cli(t, "nope")
	if got.err == nil || got.called {
		t.Fatal("expected unknown command")
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
	if !strings.Contains(got.out, "Available Commands") || !strings.Contains(got.out, "parse") {
		t.Fatalf("help:\n%s", got.out)
	}
}

func TestHelpParse(t *testing.T) {
	got := execute_cli(t, "parse", "--help")
	if got.err != nil {
		t.Fatalf("parse help: %v", got.err)
	}
	if got.called {
		t.Fatal("help must not run parse")
	}
	if !strings.Contains(got.out, "--file") {
		t.Fatalf("parse help:\n%s", got.out)
	}
}

func TestHelpSearch(t *testing.T) {
	got := execute_cli(t, "help", "search")
	if got.err != nil {
		t.Fatalf("help search: %v", got.err)
	}
	if got.called {
		t.Fatal("help must not run search")
	}
	if !strings.Contains(got.out, "--limit") {
		t.Fatalf("help search:\n%s", got.out)
	}
}
