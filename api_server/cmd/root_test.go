package main

import (
	"testing"
)

func TestListenAddrPortEnvWins(t *testing.T) {
	if got := listen_addr("9090", "30080"); got != ":30080" {
		t.Fatalf("got %q", got)
	}
	if got := listen_addr("9090", ""); got != ":9090" {
		t.Fatalf("got %q", got)
	}
	if got := listen_addr(":9090", ""); got != ":9090" {
		t.Fatalf("got %q", got)
	}
}

func TestRootListenAndServeAlias(t *testing.T) {
	var got string
	cmd := new_root(func(addr string) error {
		got = addr
		return nil
	})
	cmd.SetArgs([]string{
		"--port",
		"9090",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("root: %v", err)
	}
	if got != ":9090" {
		t.Fatalf("root addr=%q", got)
	}
	got = ""
	cmd = new_root(func(addr string) error {
		got = addr
		return nil
	})
	cmd.SetArgs([]string{
		"serve",
		"--port",
		"7070",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("serve: %v", err)
	}
	if got != ":7070" {
		t.Fatalf("serve addr=%q", got)
	}
}

func TestUnknownCommand(t *testing.T) {
	cmd := new_root(func(string) error { return nil })
	cmd.SetArgs([]string{"nope"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected unknown command")
	}
}
