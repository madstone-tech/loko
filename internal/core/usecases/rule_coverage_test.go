package usecases

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/viewmodel"
)

// TestValidationRuleCoverage is the mechanical check behind SC-004.
//
// Without it, "every validation rule has a fixture" is a claim in a document.
// The test walks every diagnostic code the product can emit and fails if any
// is never produced by a test in this package or by a golden fixture — so a
// rule cannot ship, or be quietly dropped, without someone noticing.
func TestValidationRuleCoverage(t *testing.T) {
	t.Parallel()

	produced := map[string]bool{}
	for _, code := range codesFromCoreTests(t) {
		produced[code] = true
	}
	for _, code := range codesFromGoldenFixtures(t) {
		produced[code] = true
	}

	var missing []string
	for _, code := range arch.AllCodes {
		if !produced[code] {
			missing = append(missing, code)
		}
	}
	sort.Strings(missing)

	if len(missing) > 0 {
		t.Errorf("no test or fixture produces these diagnostics: %s\n"+
			"Every code in arch.AllCodes needs a case; see contracts/diagnostics.schema.json.",
			strings.Join(missing, ", "))
	}
}

// codesFromCoreTests exercises the core rules directly and collects what they
// emit. These are the rules that belong to this package.
func codesFromCoreTests(t *testing.T) []string {
	t.Helper()

	var all arch.Diagnostics
	collect := func(d arch.Diagnostics) { all = append(all, d...) }

	// Duplicate declaration, unresolved reference, wrong reference kind.
	_, d := ResolveModel(&arch.SourceModel{Elements: []arch.ElementDecl{
		elem(arch.KindContainer, "dupe", 1),
		elem(arch.KindContainer, "dupe", 2),
		elem(arch.KindSystem, "s", 3),
		{Kind: arch.KindContainer, Name: "bad", Range: at(4), Parent: ref("system.missing", 5)},
		{Kind: arch.KindComponent, Name: "wrongkind", Range: at(6), Parent: ref("system.s", 7)},
	}})
	collect(d)

	// Wrong parent kind (absent parent) and containment cycle.
	cyc := &Resolved{Parent: map[arch.Address]arch.Address{
		"container.a": "container.b", "container.b": "container.a",
	}}
	collect(ValidateStructure(&arch.SourceModel{Elements: []arch.ElementDecl{
		elem(arch.KindContainer, "a", 1), elem(arch.KindContainer, "b", 2),
		elem(arch.KindComponent, "noparent", 3),
	}}, cyc))

	// Duplicate instance name, which collides across two different nodes
	// because instance addresses omit the placement path.
	_, dupInst := ResolveModel(&arch.SourceModel{Environments: []arch.EnvironmentDecl{{
		Name: "prod", Range: at(1),
		Groups: []arch.GroupDecl{
			{Name: "vpc-a", Range: at(2), Instances: []arch.InstanceDecl{{Name: "api", Range: at(3)}}},
			{Name: "vpc-b", Range: at(4), Instances: []arch.InstanceDecl{{Name: "api", Range: at(5)}}},
		},
	}}})
	collect(dupInst)

	// Version constraint.
	collect(ValidateProjectVersion(&arch.SourceModel{Project: arch.ProjectDecl{
		Declared: true, Version: ">= 99.0", VersionRange: at(1),
	}}, "1.0.0"))

	// Duplicate claim and binding arity.
	collect(ValidateDeployment(&arch.SourceModel{Environments: []arch.EnvironmentDecl{{
		Name: "prod", Range: at(1),
		Instances: []arch.InstanceDecl{
			{Name: "a", Range: at(2), Claims: []arch.ClaimDecl{{Kind: arch.ClaimTerraform, Address: "m.x", Range: at(3)}}},
			{Name: "b", Range: at(4), Claims: []arch.ClaimDecl{
				{Kind: arch.ClaimTerraform, Address: "m.x", Range: at(5)},
				{Kind: arch.ClaimTerraform, Range: at(6)},
			}},
		},
	}}}, &Resolved{}))

	// Every warning: an orphaned, undocumented, self-relating element inside
	// an empty system, plus an unbound instance and a missing prose file.
	warnModel := &arch.SourceModel{
		Elements: []arch.ElementDecl{
			elem(arch.KindSystem, "empty", 1),
			elem(arch.KindContainer, "orphan", 2),
			{Kind: arch.KindContainer, Name: "selfish", Range: at(3),
				Docs:       "does-not-exist.md",
				AttrRanges: map[string]arch.SourceRange{"docs": at(4)},
				Relations:  []arch.RelationDecl{{LocalName: "me", Target: ref("container.selfish", 5), Range: at(5)}}},
		},
		Environments: []arch.EnvironmentDecl{{
			Name: "prod", Range: at(10),
			Instances: []arch.InstanceDecl{{Name: "unbound", Range: at(11)}},
		}},
	}
	warnRes, _ := ResolveModel(warnModel)
	collect(ValidateWarnings(warnModel, warnRes, t.TempDir()))

	// Render stage (feature 014): two artifacts whose paths differ only by
	// case.
	collect(checkCollisions([]viewmodel.Artifact{
		{Path: "element/container/Api.html", Owner: "container.Api"},
		{Path: "element/container/api.html", Owner: "container.api"},
	}, Provenance{}))

	// A declared view that selects nothing, and one that shadows a derived
	// view.
	_, viewDiags := ResolveViews(declaredIR(
		arch.View{Address: "view.landscape", Name: "landscape", Include: []arch.Address{"person.p"}},
		arch.View{Address: "view.none", Name: "none", Include: []arch.Address{"system.shop"}, Exclude: []arch.Address{"system.shop"}},
	), Provenance{})
	collect(viewDiags)

	// Renames (feature 015): moved from a declared address, to nowhere, and
	// the same address moved twice.
	collect(ValidateMoved(movedModel(
		moved("system.s", "system.c", 1), moved("system.gone", "system.nowhere", 2), moved("system.gone", "system.s", 3),
	)))

	// A malformed theme override.
	collect(arch.Diagnostics{themeDiagnostic(&ThemeError{File: "templates/partials.gohtml", Message: "bad"})})

	out := make([]string, 0, len(all))
	for _, x := range all {
		out = append(out, x.Code)
	}
	return out
}

// codesFromGoldenFixtures reads the parse-stage codes the adapter's fixtures
// produce. Core cannot import the adapter, so the fixtures' recorded output is
// read from disk instead of regenerated.
func codesFromGoldenFixtures(t *testing.T) []string {
	t.Helper()

	root := filepath.Join("..", "..", "adapters", "hclsource", "testdata", "golden")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("reading golden fixtures at %s: %v", root, err)
	}

	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		b, readErr := os.ReadFile(filepath.Join(root, e.Name(), "expected_diagnostics.json"))
		if readErr != nil {
			continue
		}
		var diags []struct {
			Code string `json:"code"`
		}
		if json.Unmarshal(b, &diags) != nil {
			continue
		}
		for _, d := range diags {
			out = append(out, d.Code)
		}
	}
	return out
}
