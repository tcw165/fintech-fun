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
	root := cmd(func(addr string) error {
		got = addr
		return nil
	})
	root.SetArgs([]string{
		"--port",
		"9090",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("root: %v", err)
	}
	if got != ":9090" {
		t.Fatalf("root addr=%q", got)
	}
	got = ""
	serve := cmd(func(addr string) error {
		got = addr
		return nil
	})
	serve.SetArgs([]string{
		"serve",
		"--port",
		"7070",
	})
	if err := serve.Execute(); err != nil {
		t.Fatalf("serve: %v", err)
	}
	if got != ":7070" {
		t.Fatalf("serve addr=%q", got)
	}
}

func TestUnknownCommand(t *testing.T) {
	unknown := cmd(func(string) error { return nil })
	unknown.SetArgs([]string{"nope"})
	if err := unknown.Execute(); err == nil {
		t.Fatal("expected unknown command")
	}
}
