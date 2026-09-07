package contract

import "testing"

func TestPayloadIncludesPaginationAndOmitsRefresh(t *testing.T) {
	out := Result{
		Written: 2, Created: 2, PageSHA256: "abc",
		Companies: []string{"Amphenol"}, Stocks: []string{"APH"},
		Overlap: 3, Suffix: 2, Pages: 1, HistoryLen: 5,
	}.Payload()
	if out["status"] != "ok" || out["written"] != 2 || out["overlap"] != 3 || out["suffix"] != 2 {
		t.Fatalf("%v", out)
	}
	if out["companies"] != 1 || out["refresh"] != nil || out["backfilled"] != nil {
		t.Fatalf("%v", out)
	}
}

func TestPayloadRefreshAndBackfill(t *testing.T) {
	out := Result{Refresh: true, WaitingPolicy: WaitingPolicy, Backfilled: true}.Payload()
	if out["refresh"] != true || out["waiting_policy"] != WaitingPolicy || out["backfilled"] != true {
		t.Fatalf("%v", out)
	}
}
