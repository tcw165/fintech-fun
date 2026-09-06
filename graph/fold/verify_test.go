package fold

import "testing"

func TestVerifyExamplesMatchNotion(t *testing.T) {
	checks := VerifyExamples()
	if len(checks) != 9 {
		t.Fatalf("%d", len(checks))
	}
	for _, check := range checks {
		if !check.OK {
			t.Fatalf("%+v", check)
		}
	}
}

func TestResolveSquareAliases(t *testing.T) {
	for _, q := range []string{"square", "block", "sq", "xyz"} {
		fixture, ok := Resolve(q)
		if !ok || fixture.Stock.Ticker != "XYZ" {
			t.Fatalf("%s → %+v", q, fixture)
		}
	}
}
