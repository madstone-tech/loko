package hclsource

import (
	"slices"
	"testing"
)

// TestDecodeWarningControls covers feature 017's two attributes:
// instance.planned and project.ignore_warnings.
func TestDecodeWarningControls(t *testing.T) {
	t.Parallel()
	model, diags := loadSample(t, map[string]string{"arch.loko.hcl": `project "p" {
  ignore_warnings = ["missing_docs", "orphan_element"]
}

deployment "prod" {
  instance "api" {
    of      = container.api
    planned = true
  }
  instance "web" {
    of = container.web
  }
}
`})
	if diags.HasErrors() {
		t.Fatalf("errors: %v", diags)
	}
	if got := model.Project.IgnoreWarnings; !slices.Equal(got, []string{"missing_docs", "orphan_element"}) {
		t.Errorf("IgnoreWarnings = %v", got)
	}
	if r := model.Project.IgnoreWarningsRange; r.StartLine != 2 {
		t.Errorf("IgnoreWarningsRange = %+v", r)
	}
	inst := model.Environments[0].Instances
	if !inst[0].Planned || inst[0].PlannedRange.StartLine != 8 {
		t.Errorf("api: Planned %v at %+v", inst[0].Planned, inst[0].PlannedRange)
	}
	if inst[1].Planned {
		t.Error("web is planned")
	}
}
