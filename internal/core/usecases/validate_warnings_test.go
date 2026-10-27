package usecases

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// TestConnectivityPropagatesToAncestors: a system whose containers are wired
// up is not an orphan. In C4 an edge is normally drawn between containers, so
// without this every system in every project would warn — and a warning that
// always fires is one users learn to ignore.
func TestConnectivityPropagatesToAncestors(t *testing.T) {
	t.Parallel()

	model := &arch.SourceModel{Elements: []arch.ElementDecl{
		elem(arch.KindSystem, "payments", 1),
		{Kind: arch.KindContainer, Name: "api", Range: at(5), Parent: ref("system.payments", 6),
			Relations: []arch.RelationDecl{uses("db", "container.orders_db", 7)}},
		{Kind: arch.KindContainer, Name: "orders_db", Range: at(10), Parent: ref("system.payments", 11)},
		{Kind: arch.KindComponent, Name: "handler", Range: at(15), Parent: ref("container.api", 16)},
	}}

	res, _ := ResolveModel(model)
	diags := ValidateWarnings(model, res)

	for _, d := range diags {
		if d.Code == arch.CodeOrphanElement && d.Address == "system.payments" {
			t.Error("a system whose containers are connected was reported as an orphan")
		}
	}

	// The component is genuinely unconnected and should still warn.
	var sawHandler bool
	for _, d := range diags {
		if d.Code == arch.CodeOrphanElement && d.Address == "component.handler" {
			sawHandler = true
		}
	}
	if !sawHandler {
		t.Error("a genuinely unconnected component was not reported; the propagation is too broad")
	}
}

// TestExternalNeedsNoDocs: an external is a system the team does not own, so
// it has no prose of its own to point at. Prose it does declare must exist.
func TestExternalNeedsNoDocs(t *testing.T) {
	t.Parallel()

	withDocs := elem(arch.KindExternal, "vendor", 5)
	withDocs.Docs, withDocs.DocsVia = "docs/vendor.md", arch.DocsMissing
	withDocs.DocsCandidates = [2]string{"docs/vendor.md", "docs/vendor.md"}
	model := &arch.SourceModel{Elements: []arch.ElementDecl{
		elem(arch.KindExternal, "stripe", 1),
		elem(arch.KindSystem, "shop", 3),
		withDocs,
	}}

	res, _ := ResolveModel(model)
	diags := ValidateWarnings(model, res)

	got := map[arch.Address]string{}
	for _, d := range diags {
		if d.Code == arch.CodeMissingDocs || d.Code == arch.CodeDocsNotFound {
			got[d.Address] = d.Code
		}
	}
	want := map[arch.Address]string{
		"system.shop":     arch.CodeMissingDocs,
		"external.vendor": arch.CodeDocsNotFound,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("docs warnings = %v, want %v", got, want)
	}
}

// TestDocsDiagnostics: one case per outcome of docs resolution (feature 018,
// R9, R11). The source adapter resolved the paths; this only reports.
func TestDocsDiagnostics(t *testing.T) {
	t.Parallel()
	doc := func(via arch.DocsVia, docs, path string, cands [2]string) arch.ElementDecl {
		e := elem(arch.KindSystem, "s", 1)
		e.Range.File = "architecture/s.loko.hcl"
		e.AttrRanges = map[string]arch.SourceRange{"docs": at(2)}
		e.Docs, e.DocsVia, e.DocsPath, e.DocsCandidates = docs, via, path, cands
		return e
	}
	tests := []struct {
		name     string
		e        arch.ElementDecl
		code     string
		severity arch.Severity
		detail   []string
	}{
		{"file", doc(arch.DocsFile, "docs/s.md", "architecture/docs/s.md",
			[2]string{"architecture/docs/s.md", "docs/s.md"}), "", 0, nil},
		{"root", doc(arch.DocsRoot, "architecture/docs/s.md", "architecture/docs/s.md",
			[2]string{"architecture/architecture/docs/s.md", "architecture/docs/s.md"}),
			arch.CodeDocsRootRelative, arch.SeverityInfo, []string{`"architecture/docs/s.md"`, `write "docs/s.md" instead`}},
		{"both", doc(arch.DocsBoth, "docs/s.md", "architecture/docs/s.md",
			[2]string{"architecture/docs/s.md", "docs/s.md"}),
			arch.CodeDocsAmbiguous, arch.SeverityWarning, []string{"architecture/docs/s.md", "(docs/s.md)", "the first is used"}},
		{"missing", doc(arch.DocsMissing, "docs/s.md", "",
			[2]string{"architecture/docs/s.md", "docs/s.md"}),
			arch.CodeDocsNotFound, arch.SeverityWarning, []string{"architecture/docs/s.md (from the declaring file)", "docs/s.md (from the project root)"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got arch.Diagnostics
			for _, d := range warnDocs(&arch.SourceModel{Elements: []arch.ElementDecl{tt.e}}) {
				got = append(got, d)
			}
			if tt.code == "" {
				if len(got) != 0 {
					t.Fatalf("diagnostics %v, want none", got)
				}
				return
			}
			if len(got) != 1 || got[0].Code != tt.code || got[0].Severity != tt.severity || got[0].Range.StartLine != 2 {
				t.Fatalf("diagnostics %+v, want one %s", got, tt.code)
			}
			for _, want := range tt.detail {
				if !strings.Contains(got[0].Detail, want) {
					t.Errorf("detail %q lacks %q", got[0].Detail, want)
				}
			}
		})
	}
}

// TestPropagationTerminatesOnCycle: a containment cycle is an error reported
// elsewhere, but the warning pass must not hang on one.
func TestPropagationTerminatesOnCycle(t *testing.T) {
	t.Parallel()

	model := &arch.SourceModel{Elements: []arch.ElementDecl{
		{Kind: arch.KindContainer, Name: "a", Range: at(1),
			Relations: []arch.RelationDecl{uses("b", "container.b", 2)}},
		elem(arch.KindContainer, "b", 5),
	}}
	res := &Resolved{
		Target: map[arch.Address]arch.Address{"container.a.uses.b": "container.b"},
		Parent: map[arch.Address]arch.Address{
			"container.a": "container.b",
			"container.b": "container.a",
		},
	}

	done := make(chan struct{})
	go func() {
		ValidateWarnings(model, res)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("ValidateWarnings did not terminate on a containment cycle")
	}
}

// TestPlannedInstances is feature 017's instance.planned: no binding is
// expected yet, and a marker that outlives the infrastructure is reported.
func TestPlannedInstances(t *testing.T) {
	t.Parallel()
	model := &arch.SourceModel{Environments: []arch.EnvironmentDecl{{
		Name: "prod", Range: at(1),
		Instances: []arch.InstanceDecl{
			{Name: "planned", Range: at(2), Planned: true, PlannedRange: at(3)},
			{Name: "bound", Range: at(4), Planned: true, PlannedRange: at(5),
				Claims: []arch.ClaimDecl{{Kind: arch.ClaimTerraform, Address: "m.x", Range: at(6)}}},
			{Name: "unbound", Range: at(7)},
		},
	}}}
	res, _ := ResolveModel(model)
	got := map[string]arch.Diagnostic{}
	for _, d := range ValidateWarnings(model, res) {
		if d.Code == arch.CodeUnboundInstance || d.Code == arch.CodePlannedInstanceBound {
			got[string(d.Address)+" "+d.Code] = d
		}
	}
	if _, ok := got["deployment.prod.instance.planned "+arch.CodeUnboundInstance]; ok {
		t.Error("a planned instance without a binding draws unbound_instance")
	}
	if _, ok := got["deployment.prod.instance.unbound "+arch.CodeUnboundInstance]; !ok {
		t.Error("an unplanned instance without a binding draws no unbound_instance")
	}
	d, ok := got["deployment.prod.instance.bound "+arch.CodePlannedInstanceBound]
	if !ok || d.Range != at(5) {
		t.Errorf("planned and bound: %+v, want planned_instance_bound at the planned attribute", d)
	}
	if len(got) != 2 {
		t.Errorf("got %v", got)
	}
}
