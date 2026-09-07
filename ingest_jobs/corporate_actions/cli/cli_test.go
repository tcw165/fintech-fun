package cli

import (
	"strings"
	"testing"
)

func execute_cli(t *testing.T, args ...string) ([]string, error) {
	t.Helper()
	var got []string
	cmd := New(func(argv []string) error {
		got = append([]string(nil), argv...)
		return nil
	})
	cmd.SetArgs(args)
	err := cmd.Execute()
	return got, err
}

func TestDefaultIngest(t *testing.T) {
	got, err := execute_cli(t)
	if err != nil {
		t.Fatalf("empty: %v", err)
	}
	if got != nil {
		t.Fatalf("got %v, want nil so tools default to ingest", got)
	}
}

func TestParseForwardsFile(t *testing.T) {
	got, err := execute_cli(t, "parse", "tracker.txt")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if strings.Join(got, " ") != "parse tracker.txt" {
		t.Fatalf("got %v", got)
	}
}

func TestSearchJoinsThroughSkill(t *testing.T) {
	got, err := execute_cli(t, "search", "LivePerson", "stock", "merger")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if strings.Join(got, " ") != "search LivePerson stock merger" {
		t.Fatalf("got %v", got)
	}
}

func TestFoldRequiresQty(t *testing.T) {
	_, err := execute_cli(t, "fold", "square")
	if err == nil {
		t.Fatal("expected missing qty")
	}
}

func TestUnknownCommand(t *testing.T) {
	_, err := execute_cli(t, "nope")
	if err == nil {
		t.Fatal("expected unknown command")
	}
}
