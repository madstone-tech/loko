package hclsource

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

func s(v string) authoring.AttrValue { return authoring.AttrValue{Kind: authoring.ValueString, Str: v} }
func r(v string) authoring.AttrValue { return authoring.AttrValue{Kind: authoring.ValueRef, Ref: v} }
func l(v ...string) authoring.AttrValue {
	return authoring.AttrValue{Kind: authoring.ValueList, List: v}
}
func set(n string, v authoring.AttrValue) authoring.Attr { return authoring.Attr{Name: n, Value: v} }

func mustEdit(t *testing.T, e authoring.Edit) authoring.Edit {
	t.Helper()
	out, err := authoring.NewEdit(e)
	if err != nil {
		t.Fatalf("NewEdit: %v", err)
	}
	return out
}

// lines splits keeping every line's terminator, so joining is lossless.
func lines(b []byte) []string {
	ls := strings.SplitAfter(string(b), "\n")
	if ls[len(ls)-1] == "" {
		ls = ls[:len(ls)-1]
	}
	return ls
}

// declSpan is the test's own oracle for research R1 § Declaration span: the
// 0-based [start, end) line range of the block at path, extended upward over
// directly attached comment lines. path is block type/label pairs from the top.
func declSpan(t *testing.T, src []byte, path ...string) (int, int) {
	t.Helper()
	f, diags := hclsyntax.ParseConfig(src, "x", hcl.InitialPos)
	if diags.HasErrors() {
		t.Fatal(diags)
	}
	body := f.Body.(*hclsyntax.Body)
	var b *hclsyntax.Block
	for i := 0; i < len(path); i += 2 {
		b = nil
		for _, c := range body.Blocks {
			if c.Type == path[i] && (path[i+1] == "" || (len(c.Labels) > 0 && c.Labels[0] == path[i+1])) {
				b = c
				break
			}
		}
		if b == nil {
			t.Fatalf("no block %v", path[:i+2])
		}
		body = b.Body
	}
	ls := lines(src)
	start := b.TypeRange.Start.Line - 1
	for start > 0 {
		prev := strings.TrimSpace(ls[start-1])
		if !(strings.HasPrefix(prev, "#") || strings.HasPrefix(prev, "//") || strings.HasPrefix(prev, "/*") || strings.HasSuffix(prev, "*/")) {
			break
		}
		start--
	}
	return start, b.CloseBraceRange.End.Line
}

// assertOnlySpanChanged: new = old[:a] + X + old[b:]. With slack, the span may
// also take one adjacent blank line (a removal).
func assertOnlySpanChanged(t *testing.T, old, new []byte, a, b int, slack bool) []string {
	t.Helper()
	o, n := lines(old), lines(new)
	try := [][2]int{{a, b}}
	if slack {
		try = append(try, [2]int{a - 1, b}, [2]int{a, b + 1})
	}
	for _, sp := range try {
		a, b := max(sp[0], 0), min(sp[1], len(o))
		suffix := len(o) - b
		if len(n) >= a+suffix && slices.Equal(n[:a], o[:a]) && slices.Equal(n[len(n)-suffix:], o[b:]) {
			return n[a : len(n)-suffix]
		}
	}
	t.Fatalf("lines outside the edited span [%d,%d) changed:\n--- old\n%s--- new\n%s", a, b, old, new)
	return nil
}

// assertCanonical: inserted text, dedented, is exactly what the formatter writes.
func assertCanonical(t *testing.T, inserted []string) {
	t.Helper()
	text := strings.TrimLeft(strings.Join(inserted, ""), "\n")
	indent := len(text) - len(strings.TrimLeft(text, " "))
	var b strings.Builder
	for _, ln := range strings.SplitAfter(text, "\n") {
		b.WriteString(strings.TrimPrefix(ln, strings.Repeat(" ", indent)))
	}
	if got := string(hclwrite.Format([]byte(b.String()))); got != b.String() {
		t.Errorf("new declaration is not canonical (FR-014):\n%s\nformatter:\n%s", b.String(), got)
	}
}

type planCase struct {
	name    string
	edit    authoring.Edit
	file    string   // the one file that should change
	span    []string // block path whose span may change; nil = insertion
	insert  []string // add: the parent block path, or nil for end of file
	compile bool     // the result must compile
}

func planOne(t *testing.T, root string, e authoring.Edit) authoring.Plan {
	t.Helper()
	p, err := NewEditor().Plan(t.Context(), root, []authoring.Edit{e})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	return p
}

func changedFile(t *testing.T, p authoring.Plan, want string) authoring.FileContent {
	t.Helper()
	var changed []string
	var out authoring.FileContent
	for _, f := range p.Files {
		if string(f.Old) != string(f.New) {
			changed = append(changed, f.Path)
			out = f
		}
	}
	if len(changed) != 1 || changed[0] != want {
		t.Fatalf("changed files = %v, want [%s]", changed, want)
	}
	return out
}

func compiles(t *testing.T, root string, p authoring.Plan) {
	t.Helper()
	_, diags, err := New().LoadOverlay(t.Context(), root, p.Files)
	if err != nil || diags.HasErrors() {
		t.Fatalf("planned source does not parse: %v %v", err, diags)
	}
}

func TestPlanAdd(t *testing.T) {
	t.Parallel()
	cases := []planCase{
		{name: "top-level element goes to the project file", edit: authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetElement,
			Address: "system.ledger", Set: []authoring.Attr{set("description", s("Ledger")), set("tags", l("core", "money"))}}, file: "main.loko.hcl"},
		{name: "container goes beside its system", edit: authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetElement,
			Address: "container.cache", Set: []authoring.Attr{set("system", r("system.shop")), set("technology", s("Redis"))}}, file: "main.loko.hcl"},
		{name: "relationship goes inside its source", edit: authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetRelationship,
			Address: "container.api.uses.web", Set: []authoring.Attr{set("target", r("container.web"))}},
			file: "main.loko.hcl", insert: []string{"container", "api"}},
		{name: "environment goes to the project file", edit: authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetEnvironment,
			Address: "deployment.dev", Set: []authoring.Attr{set("region", s("eu-west-1"))}}, file: "main.loko.hcl"},
		{name: "nested group", edit: authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetGroup,
			Address: "deployment.prod.node.vpc.subnet-b"}, file: "deploy.loko.hcl", insert: []string{"deployment", "prod", "node", "vpc"}},
		{name: "instance in a nested group", edit: authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetInstance,
			Address: "deployment.prod.node.vpc.subnet-a.instance.web", Set: []authoring.Attr{set("of", r("container.web"))}},
			file: "deploy.loko.hcl", insert: []string{"deployment", "prod", "node", "vpc", "node", "subnet-a"}},
		{name: "binding", edit: authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetBinding,
			Address: "deployment.prod.instance.gateway", Binding: authoring.BindingRef{Kind: "cloudformation"},
			Set: []authoring.Attr{set("address", s("Stack/Gateway"))}},
			file: "deploy.loko.hcl", insert: []string{"deployment", "prod", "instance", "gateway"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := handwrittenCopy(t)
			p := planOne(t, root, mustEdit(t, tc.edit))
			f := changedFile(t, p, tc.file)
			a := len(lines(f.Old))
			if tc.insert != nil {
				_, end := declSpan(t, f.Old, tc.insert...)
				a = end - 1 // before the parent's closing brace line
			}
			inserted := assertOnlySpanChanged(t, f.Old, f.New, a, a, false)
			assertCanonical(t, inserted)
			compiles(t, root, p)
		})
	}
}

func TestPlanAddToNewFile(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	p := planOne(t, root, mustEdit(t, authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetElement,
		Address: "system.extra", File: "more/extra.loko.hcl"}))
	if len(p.Files) != 1 || p.Files[0].Path != "more/extra.loko.hcl" || p.Files[0].Old != nil {
		t.Fatalf("plan: %+v", p.Files)
	}
	assertCanonical(t, lines(p.Files[0].New))
	compiles(t, root, p)
}

func TestPlanUpdate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		edit authoring.Edit
		file string
		path []string
		want []string // substrings of the new span
	}{
		{"existing attribute keeps its spacing", authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetElement,
			Address: "system.shop", Set: []authoring.Attr{set("owner", s("payments"))}},
			"main.loko.hcl", []string{"system", "shop"}, []string{`  owner   = "payments"` + "\n", `description="Storefront"   # trailing comment`}},
		{"new attribute in a block with nested blocks", authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetElement,
			Address: "container.api", Set: []authoring.Attr{set("tags", l("core"))}},
			"main.loko.hcl", []string{"container", "api"}, []string{`  tags = ["core"]` + "\n", "// comment between attributes"}},
		{"reference written as a traversal", authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetElement,
			Address: "container.gateway", Set: []authoring.Attr{set("system", r("system.shop"))}},
			"main.loko.hcl", []string{"container", "gateway"}, []string{"  system = system.shop\n"}},
		{"clear", authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetElement,
			Address: "container.web", Clear: []string{"tags"}},
			"main.loko.hcl", []string{"container", "web"}, []string{"technology    = \"React\"\n\n  uses"}},
		{"single-line relationship is expanded", authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetRelationship,
			Address: "person.customer.uses.browse", Set: []authoring.Attr{set("description", s("Shops"))}},
			"main.loko.hcl", []string{"person", "customer"}, []string{"  uses \"browse\" {\n    target = container.web\n    description = \"Shops\"\n  }\n", "# nested comment"}},
		{"environment", authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetEnvironment,
			Address: "deployment.prod", Set: []authoring.Attr{set("region", s("eu-west-1"))}},
			"deploy.loko.hcl", []string{"deployment", "prod"}, []string{`  region   = "eu-west-1"`, "# inner comment"}},
		{"instance attributes map", authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetInstance,
			Address: "deployment.prod.instance.api", Set: []authoring.Attr{{Name: "attributes", Value: authoring.AttrValue{
				Kind: authoring.ValueMap, Map: []authoring.MapEntry{{Key: "memory", Value: authoring.AttrValue{Kind: authoring.ValueNumber, Num: 1024}}}}}}},
			"deploy.loko.hcl", []string{"deployment", "prod", "node", "vpc", "node", "subnet-a", "instance", "api"}, []string{"memory = 1024"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := handwrittenCopy(t)
			p := planOne(t, root, mustEdit(t, tc.edit))
			f := changedFile(t, p, tc.file)
			a, b := declSpan(t, f.Old, tc.path...)
			span := strings.Join(assertOnlySpanChanged(t, f.Old, f.New, a, b, false), "")
			for _, w := range tc.want {
				if !strings.Contains(span, w) {
					t.Errorf("edited span lacks %q:\n%s", w, span)
				}
			}
			compiles(t, root, p)
		})
	}
}

func TestPlanRemove(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		edit authoring.Edit
		file string
		path []string
		gone []string
		kept []string
	}{
		{"attached comment goes with the block", authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetElement, Address: "external.bank"},
			"main.loko.hcl", []string{"external", "bank"}, []string{"attached comment for bank"}, nil},
		{"attached block comment goes with the block", authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetElement, Address: "container.web"},
			"main.loko.hcl", []string{"container", "web"}, []string{"A block comment"}, []string{"Owned by commerce."}},
		{"separated comment stays", authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetElement, Address: "system.payments"},
			"main.loko.hcl", []string{"system", "payments"}, nil, []string{"belongs to no declaration"}},
		{"relationship", authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetRelationship, Address: "container.api.uses.charge"},
			"main.loko.hcl", []string{"container", "api", "uses", "charge"}, nil, nil},
		{"instance", authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetInstance, Address: "deployment.prod.instance.gateway"},
			"deploy.loko.hcl", []string{"deployment", "prod", "instance", "gateway"}, nil, nil},
		{"binding", authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetBinding, Address: "deployment.prod.instance.api",
			Binding: authoring.BindingRef{Kind: "terraform"}}, "deploy.loko.hcl",
			[]string{"deployment", "prod", "node", "vpc", "node", "subnet-a", "instance", "api", "binding", "terraform"}, nil, nil},
		{"group", authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetGroup, Address: "deployment.prod.node.vpc.subnet-a"},
			"deploy.loko.hcl", []string{"deployment", "prod", "node", "vpc", "node", "subnet-a"}, []string{"# inner comment"}, nil},
		{"environment", authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetEnvironment, Address: "deployment.prod"},
			"deploy.loko.hcl", []string{"deployment", "prod"}, []string{"Deployment, hand-written"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := handwrittenCopy(t)
			p := planOne(t, root, mustEdit(t, tc.edit))
			f := changedFile(t, p, tc.file)
			a, b := declSpan(t, f.Old, tc.path...)
			if rest := assertOnlySpanChanged(t, f.Old, f.New, a, b, true); len(rest) != 0 {
				t.Errorf("removal left %q", rest)
			}
			if strings.Contains(string(f.New), "\n\n\n\n") && !strings.Contains(string(f.Old), "\n\n\n\n") {
				t.Error("removal left a doubled blank line")
			}
			for _, g := range tc.gone {
				if strings.Contains(string(f.New), g) {
					t.Errorf("%q should have been removed", g)
				}
			}
			for _, k := range tc.kept {
				if !strings.Contains(string(f.New), k) {
					t.Errorf("%q should have survived", k)
				}
			}
		})
	}
}

func TestPlanRefusals(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	tests := []struct {
		edit   authoring.Edit
		reason string
	}{
		{authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetElement, Address: "container.nope"}, authoring.ReasonNotFound},
		{authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetElement, Address: "system.nope", Clear: []string{"owner"}}, authoring.ReasonNotFound},
		{authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetElement, Address: "system.shop"}, authoring.ReasonAddressInUse},
		{authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetRelationship, Address: "container.api.uses.charge",
			Set: []authoring.Attr{set("target", r("container.web"))}}, authoring.ReasonAddressInUse},
		{authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetGroup, Address: "deployment.qa.node.x"}, authoring.ReasonNotFound},
		// Built without NewEdit: the editor must refuse on its own (R12).
		{authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetElement, Address: "system.x", File: "../escape.loko.hcl"}, authoring.ReasonPathRefused},
		{authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetElement, Address: "system.x", File: "notes.md"}, authoring.ReasonPathRefused},
	}
	for i, tt := range tests {
		_, err := NewEditor().Plan(t.Context(), root, []authoring.Edit{tt.edit})
		var ee *authoring.EditError
		if !errors.As(err, &ee) || ee.Reason != tt.reason {
			t.Errorf("case %d (%s %s): err = %v, want %s", i, tt.edit.Op, tt.edit.Address, err, tt.reason)
		}
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "escape.loko.hcl")); err == nil {
		t.Error("planning wrote a file")
	}
}

func TestPlanBatchAppliesInOrder(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	edits := []authoring.Edit{
		mustEdit(t, authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetElement, Address: "container.cache",
			Set: []authoring.Attr{set("system", r("system.shop"))}}),
		mustEdit(t, authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetRelationship, Address: "container.cache.uses.db",
			Set: []authoring.Attr{set("target", r("container.api"))}}),
		mustEdit(t, authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetElement, Address: "container.cache",
			Set: []authoring.Attr{set("technology", s("Redis"))}}),
	}
	p, err := NewEditor().Plan(t.Context(), root, edits)
	if err != nil {
		t.Fatal(err)
	}
	compiles(t, root, p)
	f := changedFile(t, p, "main.loko.hcl")
	if !strings.Contains(string(f.New), "container \"cache\" {\n  system = system.shop\n\n  uses \"db\" {\n    target = container.api\n  }\n  technology = \"Redis\"\n}\n") {
		t.Errorf("batch result:\n%s", f.New)
	}
	for _, path := range []string{"deploy.loko.hcl", "views.loko.hcl"} {
		for _, pf := range p.Files {
			if pf.Path == path {
				t.Errorf("untouched file %s is in the plan", path)
			}
		}
	}
}

// TestPlanSameValueIsNoOp is FR-021: setting attributes to their current
// values, however they are spaced, leaves the file byte-identical.
func TestPlanSameValueIsNoOp(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	p := planOne(t, root, mustEdit(t, authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetElement, Address: "container.web",
		Set: []authoring.Attr{set("tags", l("edge", "public")), set("technology", s("React")), set("system", r("system.shop"))}}))
	if p.Changed() {
		for _, f := range p.Files {
			t.Errorf("%s changed:\n%s", f.Path, unifiedDiff(f.Path, f.Old, f.New))
		}
	}
}

// TestPlanRemoveLastNestedBlock: removing the last block in a body takes the
// blank line above it, so no blank line is left before the closing brace.
func TestPlanRemoveLastNestedBlock(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	p := planOne(t, root, mustEdit(t, authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetRelationship, Address: "container.api.uses.charge"}))
	f := changedFile(t, p, "main.loko.hcl")
	if !strings.Contains(string(f.New), "  technology = \"Go\"\n}\n") {
		t.Errorf("a blank line was left before the closing brace:\n%s", unifiedDiff(f.Path, f.Old, f.New))
	}
	p = planOne(t, root, mustEdit(t, authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetInstance, Address: "deployment.prod.instance.gateway"}))
	if f = changedFile(t, p, "deploy.loko.hcl"); strings.Contains(string(f.New), "\n\n}") {
		t.Errorf("a blank line was left before the closing brace:\n%s", unifiedDiff(f.Path, f.Old, f.New))
	}
}

// TestPlanRemoveLastNestedBlockDeep: the removed block's indentation shares
// leading spaces with the closing brace below it.
func TestPlanRemoveLastNestedBlockDeep(t *testing.T) {
	t.Parallel()
	root := handwrittenCopy(t)
	p, err := NewEditor().Plan(t.Context(), root, []authoring.Edit{
		mustEdit(t, authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetInstance,
			Address: "deployment.prod.node.vpc.subnet-a.instance.web", Set: []authoring.Attr{set("of", r("container.web"))}}),
		mustEdit(t, authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetInstance, Address: "deployment.prod.instance.web"}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Files) != 1 {
		t.Fatalf("plan files: %d", len(p.Files))
	}
	if f := p.Files[0]; string(f.New) != string(f.Old) {
		t.Errorf("adding then removing the last block must restore the file:\n%s", unifiedDiff(f.Path, f.Old, f.New))
	}
}

func rl(a ...string) authoring.AttrValue {
	return authoring.AttrValue{Kind: authoring.ValueRefList, List: a}
}

func TestPlanViews(t *testing.T) {
	t.Parallel()
	t.Run("add goes to the project file, canonical", func(t *testing.T) {
		t.Parallel()
		root := handwrittenCopy(t)
		p := planOne(t, root, mustEdit(t, authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetView, Address: "view.storefront",
			Set: []authoring.Attr{set("include", rl("system.shop", "container.gateway")), set("tags", l("edge"))}}))
		f := changedFile(t, p, "main.loko.hcl")
		inserted := assertOnlySpanChanged(t, f.Old, f.New, len(lines(f.Old)), len(lines(f.Old)), false)
		assertCanonical(t, inserted)
		if !strings.Contains(strings.Join(inserted, ""), "include = [system.shop, container.gateway]") {
			t.Errorf("include must be unquoted references:\n%s", strings.Join(inserted, ""))
		}
		compiles(t, root, p)
	})
	for name, tc := range map[string]struct {
		edit authoring.Edit
		want string
	}{
		"update include": {authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetView, Address: "view.payments-path",
			Set: []authoring.Attr{set("include", rl("system.payments"))}}, "include = [system.payments]"},
		"set exclude": {authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetView, Address: "view.payments-path",
			Set: []authoring.Attr{set("exclude", rl("container.gateway"))}}, "exclude = [container.gateway]"},
		"clear include": {authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetView, Address: "view.payments-path",
			Clear: []string{"include"}}, "view \"payments-path\" {\n}"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := handwrittenCopy(t)
			p := planOne(t, root, mustEdit(t, tc.edit))
			f := changedFile(t, p, "views.loko.hcl")
			a, b := declSpan(t, f.Old, "view", "payments-path")
			span := strings.Join(assertOnlySpanChanged(t, f.Old, f.New, a, b, false), "")
			if !strings.Contains(span, tc.want) {
				t.Errorf("view span lacks %q:\n%s", tc.want, span)
			}
			compiles(t, root, p)
		})
	}
	t.Run("remove", func(t *testing.T) {
		t.Parallel()
		root := handwrittenCopy(t)
		p := planOne(t, root, mustEdit(t, authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetView, Address: "view.payments-path"}))
		if f := changedFile(t, p, "views.loko.hcl"); strings.TrimSpace(string(f.New)) != "" {
			t.Errorf("view not removed:\n%s", f.New)
		}
	})
}
