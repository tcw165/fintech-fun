package impl

import (
	"testing"

	"github.com/tcw165/fintech-fun/graph/clients/neo4j"
)

func TestOpenRejectsEmptyURI(t *testing.T) {
	driver, err := Open("", User, Password)
	if err == nil {
		_ = driver.Close(t.Context())
		t.Fatal("expected error")
	}
}

func TestOpenFromEnvUsesDefaults(t *testing.T) {
	t.Setenv("NEO4J_URI", "")
	t.Setenv("NEO4J_USER", "")
	t.Setenv("NEO4J_PASSWORD", "")
	if URIFromEnv() != URI || UserFromEnv() != User || PasswordFromEnv() != Password {
		t.Fatalf("%s %s %s", URIFromEnv(), UserFromEnv(), PasswordFromEnv())
	}
	t.Setenv("NEO4J_URI", "bolt://neo4j:7687")
	if URIFromEnv() != "bolt://neo4j:7687" {
		t.Fatal(URIFromEnv())
	}
}

func TestDriverSatisfiesClient(t *testing.T) {
	var client neo4j.Client = &Driver{}
	if client == nil {
		t.Fatal("nil")
	}
}

func TestRecordingStillSatisfiesClient(t *testing.T) {
	var client neo4j.Client = &Recording{}
	rows, err := client.Run("RETURN 1", map[string]any{"k": 1})
	if err != nil || rows[0]["ok"] != true {
		t.Fatalf("%v %v", rows, err)
	}
}
