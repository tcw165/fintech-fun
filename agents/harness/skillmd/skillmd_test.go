package skillmd

import (
	"strings"
	"testing"
)

const sample = `---
name: corporate-actions
description: Ingest Robinhood corporate actions into the retail graph. Use when building or refreshing Company, Stock, and Event.
license: Apache-2.0
allowed-tools: parse classify plan fetch gold ingest verify search refresh ping seed fold
metadata:
  default-tool: ingest
  source: robinhood/corporate_actions
---
# Corporate actions

Fetch the tracker, classify headlines, MERGE Events.
`

func TestParseSample(t *testing.T) {
	doc, err := Parse([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Name != "corporate-actions" || doc.License != "Apache-2.0" {
		t.Fatalf("%+v", doc)
	}
	if !strings.Contains(doc.Description, "retail graph") {
		t.Fatalf("description %q", doc.Description)
	}
	if doc.DefaultTool() != "ingest" || !doc.Allows("fold") || doc.Allows("drop") {
		t.Fatalf("tools %+v default %s", doc.AllowedTools, doc.DefaultTool())
	}
	if doc.Metadata["source"] != "robinhood/corporate_actions" {
		t.Fatalf("metadata %+v", doc.Metadata)
	}
	if !strings.Contains(doc.Body, "MERGE Events") {
		t.Fatalf("body %q", doc.Body)
	}
}

func TestParseRejectsMissingFrontmatter(t *testing.T) {
	if _, err := Parse([]byte("# just markdown\n")); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseRejectsBadName(t *testing.T) {
	for _, name := range []string{"Corporate-Actions", "-ingest", "ingest-", "in--gest", "has_underscore"} {
		raw := "---\nname: " + name + "\ndescription: x\n---\nbody\n"
		if _, err := Parse([]byte(raw)); err == nil {
			t.Fatalf("expected reject %q", name)
		}
	}
}

func TestParseRequiresDescription(t *testing.T) {
	if _, err := Parse([]byte("---\nname: ingest\n---\n")); err == nil {
		t.Fatal("expected error")
	}
}

func TestEmptyAllowedToolsAllowsAny(t *testing.T) {
	doc, err := Parse([]byte("---\nname: ingest\ndescription: Do the thing. Use when testing.\n---\nbody\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !doc.Allows("anything") || doc.DefaultTool() != "" {
		t.Fatalf("%+v", doc)
	}
}

func TestDefaultToolFallsBackToFirstAllowed(t *testing.T) {
	doc, err := Parse([]byte("---\nname: ingest\ndescription: Do the thing. Use when testing.\nallowed-tools: parse ingest\n---\n"))
	if err != nil {
		t.Fatal(err)
	}
	if doc.DefaultTool() != "parse" {
		t.Fatalf("%s", doc.DefaultTool())
	}
}
