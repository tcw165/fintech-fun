package contract

import "testing"

func TestNodePorts(t *testing.T) {
	got := NodePorts("192.168.49.2")
	if got.API != "http://192.168.49.2:30080" || got.Bolt != "bolt://192.168.49.2:30687" || got.Qdrant != "http://192.168.49.2:30333" {
		t.Fatalf("%+v", got)
	}
	if NodePorts("").API != "http://127.0.0.1:30080" {
		t.Fatal(NodePorts(""))
	}
}

func TestReportAllOK(t *testing.T) {
	if (Report{}).AllOK() {
		t.Fatal("empty")
	}
	ok := Report{Steps: []Step{{Name: StepHealthz, OK: true}, {Name: StepSmoke, OK: true}}}
	if !ok.AllOK() {
		t.Fatal("want ok")
	}
	fail := Report{Steps: []Step{{Name: StepHealthz, OK: true}, {Name: StepGold, OK: false}}}
	if fail.AllOK() {
		t.Fatal("want fail")
	}
}
