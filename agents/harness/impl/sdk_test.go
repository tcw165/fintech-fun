package impl_test

import (
	"context"
	"testing"

	"github.com/tcw165/fintech-fun/agents/harness/contract"
	"github.com/tcw165/fintech-fun/agents/harness/impl"
	"github.com/tcw165/fintech-fun/agents/harness/skillmd"
)

func TestManifestFromSkillMarkdown(t *testing.T) {
	doc, err := skillmd.Parse([]byte(`---
name: corporate-actions
description: Ingest Robinhood corporate actions. Use when building the retail graph.
allowed-tools: parse ingest fold
metadata:
  default-tool: ingest
---
Build Company, Stock, Event.
`))
	if err != nil {
		t.Fatal(err)
	}
	manifest := impl.ManifestFrom(doc)
	if manifest.Name != "corporate-actions" || manifest.DefaultTool != "ingest" || len(manifest.AllowedTools) != 3 {
		t.Fatalf("%+v", manifest)
	}
}

func TestSDKDispatchesAllowedTool(t *testing.T) {
	var seen []string
	tools := map[string]contract.Tool{
		"parse": func(_ context.Context, req contract.Request) (contract.Result, error) {
			seen = req.Args
			return contract.Result{Payload: "parsed"}, nil
		},
	}
	got, err := impl.NewSDK().Run(context.Background(), contract.Manifest{
		Name: "corporate-actions", AllowedTools: []string{"parse", "ingest"},
	}, tools, contract.Request{Args: []string{"parse", "page.html"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Payload != "parsed" || len(seen) != 2 || seen[0] != "parse" {
		t.Fatalf("%+v %v", got, seen)
	}
}

func TestSDKDefaultsEmptyArgs(t *testing.T) {
	var seen []string
	tools := map[string]contract.Tool{
		"ingest": func(_ context.Context, req contract.Request) (contract.Result, error) {
			seen = req.Args
			return contract.Result{Payload: "ingested"}, nil
		},
	}
	got, err := impl.NewSDK().Run(context.Background(), contract.Manifest{
		Name: "corporate-actions", AllowedTools: []string{"ingest"}, DefaultTool: "ingest",
	}, tools, contract.Request{})
	if err != nil || got.Payload != "ingested" || len(seen) != 1 || seen[0] != "ingest" {
		t.Fatalf("%+v %v %v", got, seen, err)
	}
}

func TestSDKRejectsDisallowedTool(t *testing.T) {
	_, err := impl.NewSDK().Run(context.Background(), contract.Manifest{
		Name: "corporate-actions", AllowedTools: []string{"ingest"},
	}, map[string]contract.Tool{"drop": func(context.Context, contract.Request) (contract.Result, error) {
		return contract.Result{}, nil
	}}, contract.Request{Args: []string{"drop"}})
	if err == nil {
		t.Fatal("expected disallowed")
	}
}

func TestSDKRejectsUnknownTool(t *testing.T) {
	_, err := impl.NewSDK().Run(context.Background(), contract.Manifest{
		Name: "corporate-actions", DefaultTool: "ingest",
	}, map[string]contract.Tool{}, contract.Request{})
	if err == nil {
		t.Fatal("expected unknown")
	}
}
