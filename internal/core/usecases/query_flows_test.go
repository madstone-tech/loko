package usecases

import (
	"testing"
)

// TestQueryFlows: the steps an element takes part in, directly or through an
// element it contains, sorted by view then number (feature 019, R10).
func TestQueryFlows(t *testing.T) {
	t.Parallel()
	ir := dynamicIR(
		st("browse", "person.customer", "container.web", "", ""),
		st("order", "container.web", "container.api", "", ""),
		st("store", "component.repo", "container.db", "", ""),
	)
	got := QueryIR(ir, QueryRequest{Kind: "flows", Address: "container.api"})
	if !got.OK || len(got.Flows) != 2 {
		t.Fatalf("flows %+v (error %+v)", got.Flows, got.Error)
	}
	want := []FlowHit{
		{View: "view.checkout", Number: 2, Step: "order", Role: "to"},
		{View: "view.checkout", Number: 3, Step: "store", Role: "from"},
	}
	for i, w := range want {
		if got.Flows[i] != w {
			t.Errorf("flow %d = %+v, want %+v", i, got.Flows[i], w)
		}
	}
	if none := QueryIR(ir, QueryRequest{Kind: "flows", Address: "external.payments"}); !none.OK || len(none.Flows) != 0 {
		t.Errorf("payments takes no step: %+v", none)
	}
	if bad := QueryIR(ir, QueryRequest{Kind: "flows", Address: "container.nope"}); bad.Error == nil || bad.Error.Reason != "not_found" {
		t.Errorf("unknown element: %+v", bad)
	}
}
