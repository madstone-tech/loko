package usecases

import (
	"fmt"
	"slices"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// visible resolves the named addresses to themselves; anything else is
// outside the view.
func visible(addrs ...arch.Address) resolveFunc {
	return func(a arch.Address) []string {
		if slices.Contains(addrs, a) {
			return []string{string(a)}
		}
		return nil
	}
}

func kindRel(src arch.Address, local string, dst arch.Address, kind string, tags ...string) arch.Relationship {
	r := rl(src, local, dst, local, "")
	r.Kind, r.Tags = kind, tags
	return r
}

func kindSummary(es []viewmodel.Edge) []string {
	var out []string
	for _, e := range es {
		out = append(out, fmt.Sprintf("%s>%s async=%v dashed=%v tags=%v n=%d", e.Source, e.Target, e.Style.Async, e.Style.Dashed, e.Tags, len(e.Relationships)))
	}
	return out
}

func TestLiftTrigger(t *testing.T) {
	t.Parallel()
	sync, queue, ext := ctr("sync"), ctr("queue"), ctr("elsewhere")
	got, _ := liftEdges([]arch.Relationship{kindRel(sync, "consume", queue, arch.RelTrigger)}, visible(sync, queue))
	if s := kindSummary(got); !slices.Equal(s, []string{"container.queue>container.sync async=false dashed=false tags=[] n=1"}) {
		t.Errorf("a trigger is drawn from the trigger: %v", s)
	}
	// A trigger declared on sync and a sync relationship authored queue → sync
	// are drawn the same way, so they merge.
	merged, _ := liftEdges([]arch.Relationship{
		kindRel(sync, "consume", queue, arch.RelTrigger),
		kindRel(queue, "notify", sync, ""),
	}, visible(sync, queue))
	if len(merged) != 1 || len(merged[0].Relationships) != 2 {
		t.Errorf("same drawn direction merges: %v", kindSummary(merged))
	}
	// The trigger's own end is outside the view: a crossing edge from the
	// outside marker to the invoked element.
	cross, outside := liftEdges([]arch.Relationship{kindRel(sync, "consume", ext, arch.RelTrigger)}, visible(sync))
	if !outside || len(cross) != 1 || cross[0].Source != viewmodel.OutsideNodeID || cross[0].Target != string(sync) {
		t.Errorf("crossing trigger: %v", kindSummary(cross))
	}
}

func TestLiftAsync(t *testing.T) {
	t.Parallel()
	a, b, c := ctr("a"), ctr("b"), ctr("c")
	allAsync, _ := liftEdges([]arch.Relationship{kindRel(a, "x", b, arch.RelAsync, "write"), kindRel(a, "y", b, arch.RelAsync, "audit", "write")}, visible(a, b))
	if len(allAsync) != 1 || !allAsync[0].Style.Async || fmt.Sprint(allAsync[0].Tags) != "[audit write]" {
		t.Errorf("all-async merged edge, tag union: %v", kindSummary(allAsync))
	}
	mixed, _ := liftEdges([]arch.Relationship{kindRel(a, "x", b, arch.RelAsync), kindRel(a, "y", b, "")}, visible(a, b))
	if mixed[0].Style.Async {
		t.Errorf("a mix of async and sync is drawn solid: %v", kindSummary(mixed))
	}
	single, _ := liftEdges([]arch.Relationship{kindRel(a, "x", b, "", "read")}, visible(a, b))
	if fmt.Sprint(single[0].Tags) != "[read]" {
		t.Errorf("single edge tags: %v", kindSummary(single))
	}
	crossing, _ := liftEdges([]arch.Relationship{kindRel(a, "x", c, arch.RelAsync)}, visible(a))
	if !crossing[0].Style.Dashed || crossing[0].Style.Async {
		t.Errorf("crossing edges keep Dashed and are never Async: %v", kindSummary(crossing))
	}
}
