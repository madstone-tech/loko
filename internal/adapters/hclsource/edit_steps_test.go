package hclsource

import (
	"errors"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

const stepSource = `project "p" {}

system "shop" {}

container "web" {
  system = system.shop

  uses "orders" {
    target = container.api
  }
}

container "api" {
  system = system.shop
}

view "empty" {}

view "flow" {
  title = "Flow"

  # The first call.
  step "order" {
    from = container.web
    to   = container.api
  }

  step "again" {
    from = container.web
    to   = container.api
  }
}
`

func stepEdit(t *testing.T, e authoring.Edit) string {
	t.Helper()
	root := writeTree(t, map[string]string{"main.loko.hcl": stepSource})
	return string(changedFile(t, planOne(t, root, mustEdit(t, e)), "main.loko.hcl").New)
}

func addStep(label, before string) authoring.Edit {
	return authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetStep, Address: "view.flow.step." + label,
		Before: before, Set: []authoring.Attr{set("from", r("container.api")), set("to", r("container.web"))}}
}

// TestEditStepsAppend: an added step goes after the others, in canonical form.
func TestEditStepsAppend(t *testing.T) {
	t.Parallel()
	got := stepEdit(t, addStep("reply", ""))
	want := stepSource[:len(stepSource)-2] + `
  step "reply" {
    from = container.api
    to   = container.web
  }
}
`
	if got != want {
		t.Errorf("append:\n%s\nwant:\n%s", got, want)
	}
	first := stepEdit(t, authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetStep, Address: "view.empty.step.s",
		Set: []authoring.Attr{set("from", r("container.web")), set("to", r("container.api"))}})
	if want := "view \"empty\" {\n  step \"s\" {\n    from = container.web\n    to   = container.api\n  }\n}\n"; !contains(first, want) {
		t.Errorf("first step of an empty view:\n%s", first)
	}
}

// TestEditStepsInsertBefore: before places the step ahead of that sibling and
// its comment, and leaves every other byte alone.
func TestEditStepsInsertBefore(t *testing.T) {
	t.Parallel()
	got := stepEdit(t, addStep("start", "order"))
	want := `view "flow" {
  title = "Flow"

  step "start" {
    from = container.api
    to   = container.web
  }

  # The first call.
  step "order" {`
	if !contains(got, want) {
		t.Errorf("insert before order:\n%s", got)
	}
	if mid := stepEdit(t, addStep("mid", "again")); !contains(mid, "    to   = container.api\n  }\n\n  step \"mid\" {\n    from = container.api\n    to   = container.web\n  }\n\n  step \"again\" {") {
		t.Errorf("insert before again:\n%s", mid)
	}
	root := writeTree(t, map[string]string{"main.loko.hcl": stepSource})
	_, err := NewEditor().Plan(t.Context(), root, []authoring.Edit{mustEdit(t, addStep("x", "nope"))})
	var ee *authoring.EditError
	if !errors.As(err, &ee) || ee.Reason != authoring.ReasonNotFound {
		t.Errorf("before an unknown step: %v", err)
	}
}

// TestEditStepsUpdateRemove: a step is updated and removed in place.
func TestEditStepsUpdateRemove(t *testing.T) {
	t.Parallel()
	got := stepEdit(t, authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetStep, Address: "view.flow.step.again",
		Set: []authoring.Attr{set("description", s("Retry"))}})
	if !contains(got, "  step \"again\" {\n    from        = container.web\n    to          = container.api\n    description = \"Retry\"\n  }") {
		t.Errorf("update:\n%s", got)
	}
	gone := stepEdit(t, authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetStep, Address: "view.flow.step.order"})
	if contains(gone, `step "order"`) || contains(gone, "The first call") || !contains(gone, "step \"again\"") {
		t.Errorf("remove:\n%s", gone)
	}
}
