package hclsource

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
	"github.com/madstone-tech/loko/internal/core/entities/authoring"
	"github.com/madstone-tech/loko/internal/core/usecases"
)

// The round-trip property (research R10, SC-002, SC-003): for any sequence of
// valid edits on hand-written source,
//
//  1. IR oracle: compiling the edited source gives exactly what an independent
//     model of the edit predicts, and
//  2. byte oracle: every line outside the edited declaration's span, in every
//     file, is unchanged.
//
// The model below shares no code with the editor: it applies each edit to a
// plain description of the architecture.

var (
	propSeed  = flag.Uint64("seed", 0, "run the round-trip property for this one seed")
	propCount = flag.Int("sequences", 2000, "number of seeded edit sequences")
)

const propMaxEdits = 12

type pElem struct {
	kind, parent, desc, owner, tech string
	tags                            []string
}

type pRel struct{ source, target, desc, tech string }

type pInst struct {
	env, name, of string
	groups        []string // placement path
	attrs         map[string]any
	claims        []string // kind=…;address=…
}

type pModel struct {
	elems  map[string]*pElem
	rels   map[string]*pRel
	envs   map[string][3]string // provider, account, region
	groups map[string]bool      // env/g1/g2
	insts  map[string]*pInst    // deployment.env.instance.name
	views  map[string]*pView    // declared views, by name
	prev   map[string][]string  // addresses an element was renamed away from
	next   int
}

// seedModel describes the fixture's compiled IR in the model's own terms.
func seedModel(ir *arch.IR) *pModel {
	m := &pModel{elems: map[string]*pElem{}, rels: map[string]*pRel{}, envs: map[string][3]string{},
		groups: map[string]bool{}, insts: map[string]*pInst{}, views: map[string]*pView{}, prev: map[string][]string{}}
	for _, e := range ir.Elements {
		m.elems[string(e.Address)] = &pElem{string(e.Kind), string(e.Parent), e.Description, e.Owner, e.Technology, e.Tags}
	}
	for _, r := range ir.Relationships {
		m.rels[string(r.Address)] = &pRel{string(r.Source), string(r.Target), r.Description, r.Technology}
	}
	for _, env := range ir.Environments {
		m.envs[env.Name] = [3]string{env.Provider, env.Account, env.Region}
		var walk func(path []string, gs []arch.Group)
		walk = func(path []string, gs []arch.Group) {
			for _, g := range gs {
				p := append(slices.Clone(path), g.Name)
				m.groups[env.Name+"/"+strings.Join(p, "/")] = true
				walk(p, g.Groups)
			}
		}
		walk(nil, env.Groups)
		for _, in := range env.Instances {
			pi := &pInst{env: env.Name, name: in.Name, of: string(in.Of), attrs: map[string]any{}}
			if in.PlacedIn != "" {
				pi.groups = strings.Split(strings.TrimPrefix(string(in.PlacedIn), string(env.Address)+".node."), ".")
			}
			for _, a := range in.Attributes {
				b, _ := json.Marshal(a.Value)
				var v any
				_ = json.Unmarshal(b, &v)
				pi.attrs[a.Key] = v
			}
			for _, c := range in.Claims {
				pi.claims = append(pi.claims, fmt.Sprintf("kind=%s;address=%s", c.Kind, c.Address))
			}
			m.insts[string(in.Address)] = pi
		}
	}
	for _, v := range ir.Views {
		m.views[v.Name] = &pView{include: addrList(v.Include), exclude: addrList(v.Exclude), tags: slices.Clone(v.Tags)}
	}
	return m
}

type pView struct{ include, exclude, tags []string }

func addrList(as []arch.Address) []string {
	out := make([]string, len(as))
	for i, a := range as {
		out[i] = string(a)
	}
	return out
}

// inView reports whether any view names a.
func (m *pModel) inView(a string) bool {
	for _, v := range m.views {
		if slices.Contains(v.include, a) || slices.Contains(v.exclude, a) {
			return true
		}
	}
	return false
}

// projection renders an architecture as sorted lines, the form both oracles compare.
func (m *pModel) projection() []string {
	var out []string
	for a, e := range m.elems {
		out = append(out, fmt.Sprintf("E %s %s parent=%s desc=%q owner=%q tech=%q tags=%v", a, e.kind, e.parent, e.desc, e.owner, e.tech, e.tags))
	}
	for a, r := range m.rels {
		out = append(out, fmt.Sprintf("R %s %s>%s desc=%q tech=%q", a, r.source, r.target, r.desc, r.tech))
	}
	for n, v := range m.envs {
		out = append(out, fmt.Sprintf("V %s %v", n, v))
	}
	for g := range m.groups {
		out = append(out, "G "+g)
	}
	for n, v := range m.views {
		out = append(out, fmt.Sprintf("W %s inc=%v exc=%v tags=%v", n, sorted(v.include), sorted(v.exclude), sorted(v.tags)))
	}
	for a, in := range m.insts {
		attrs, _ := json.Marshal(in.attrs)
		claims := slices.Sorted(slices.Values(in.claims))
		out = append(out, fmt.Sprintf("I %s of=%s in=%s attrs=%s claims=%v", a, in.of, strings.Join(in.groups, "/"), attrs, claims))
	}
	slices.Sort(out)
	return out
}

func projectIR(ir *arch.IR) []string { return seedModel(ir).projection() }

func compileRoot(t testing.TB, root string) *arch.IR {
	t.Helper()
	res, err := usecases.CompileArchitecture(context.Background(), New(), usecases.CompileRequest{Root: root})
	if err != nil || res.HasErrors() {
		t.Fatalf("does not compile: %v %v", err, res.Diags)
	}
	return usecases.BuildIR(res.Model, res.Resolved)
}

// step is one generated edit with what the model expects and where the
// byte oracle allows change.
type step struct {
	edit  authoring.Edit
	apply func(*pModel)
	span  []string // block path whose span may change
	add   bool     // insertion at the end of span's block (or of the file when span is nil)
}

func pick[T any](r *rand.Rand, xs []T) T { return xs[r.IntN(len(xs))] }

func sortedKeys[V any](m map[string]V) []string { return slices.Sorted(maps.Keys(m)) }

func (m *pModel) fresh(prefix string) string { m.next++; return fmt.Sprintf("%s%d", prefix, m.next) }

func (m *pModel) ofKind(kinds ...string) []string {
	var out []string
	for _, a := range sortedKeys(m.elems) {
		if slices.Contains(kinds, m.elems[a].kind) {
			out = append(out, a)
		}
	}
	return out
}

func elemPath(a string) []string { k, n, _ := strings.Cut(a, "."); return []string{k, n} }

func (in *pInst) path() []string {
	p := []string{"deployment", in.env}
	for _, g := range in.groups {
		p = append(p, "node", g)
	}
	return append(p, "instance", in.name)
}

// groupAddr is the address of group "env/g1/g2": deployment.env.node.g1.g2.
func groupAddr(g string) string {
	parts := strings.Split(g, "/")
	if len(parts) == 1 {
		return "deployment." + parts[0]
	}
	return "deployment." + parts[0] + ".node." + strings.Join(parts[1:], ".")
}

// groupPath is the block path of group "env/g1/g2", for declSpan.
func groupPath(g string) []string {
	parts := strings.Split(g, "/")
	p := []string{"deployment", parts[0]}
	for _, n := range parts[1:] {
		p = append(p, "node", n)
	}
	return p
}

// generate returns one valid edit for the model's current state.
func (m *pModel) generate(r *rand.Rand) step {
	for {
		if s, ok := m.try(r, r.IntN(14)); ok {
			return s
		}
	}
}

func (m *pModel) try(r *rand.Rand, op int) (step, bool) {
	words := []string{"Alpha", "Beta service", "gRPC", "Go", "payments-team", "", "x"}
	switch op {
	case 0: // add a top-level element
		kind := pick(r, []string{"system", "person", "external"})
		a := kind + "." + m.fresh("n")
		desc := pick(r, words)
		var set []authoring.Attr
		if desc != "" {
			set = append(set, set1("description", s(desc)))
		}
		return step{edit: authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetElement, Address: a, Set: set}, add: true,
			apply: func(m *pModel) { m.elems[a] = &pElem{kind: kind, desc: desc} }}, true
	case 1: // add a container or component under an existing parent
		kind, parentKind, attr := "container", "system", "system"
		if r.IntN(2) == 0 {
			kind, parentKind, attr = "component", "container", "container"
		}
		parents := m.ofKind(parentKind)
		if len(parents) == 0 {
			return step{}, false
		}
		p, a, tech := pick(r, parents), kind+"."+m.fresh("c"), pick(r, words)
		set := []authoring.Attr{set1(attr, rf(p))}
		if tech != "" {
			set = append(set, set1("technology", s(tech)))
		}
		return step{edit: authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetElement, Address: a, Set: set}, add: true,
			apply: func(m *pModel) { m.elems[a] = &pElem{kind: kind, parent: p, tech: tech} }}, true
	case 2, 3: // update an element: set or clear one attribute
		a := pick(r, sortedKeys(m.elems))
		attr := pick(r, []string{"description", "owner", "technology", "tags"})
		e := authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetElement, Address: a}
		val, tags := pick(r, words[:5]), []string{pick(r, []string{"edge", "core"}), "pci"}
		if r.IntN(3) == 0 {
			e.Clear = []string{attr}
		} else if attr == "tags" {
			e.Set = []authoring.Attr{set1("tags", l(tags...))}
		} else {
			e.Set = []authoring.Attr{set1(attr, s(val))}
		}
		clear := len(e.Clear) > 0
		return step{edit: e, span: elemPath(a), apply: func(m *pModel) {
			el := m.elems[a]
			switch attr {
			case "description":
				el.desc = cond(clear, "", val)
			case "owner":
				el.owner = cond(clear, "", val)
			case "technology":
				el.tech = cond(clear, "", val)
			case "tags":
				el.tags = nil
				if !clear {
					el.tags = slices.Sorted(slices.Values(tags))
				}
			}
		}}, true
	case 4: // add a relationship
		src, dst := pick(r, sortedKeys(m.elems)), pick(r, sortedKeys(m.elems))
		if src == dst {
			return step{}, false
		}
		local := m.fresh("u")
		a := src + ".uses." + local
		return step{edit: authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetRelationship, Address: a,
			Set: []authoring.Attr{set1("target", rf(dst))}}, span: elemPath(src), add: true,
			apply: func(m *pModel) { m.rels[a] = &pRel{source: src, target: dst} }}, true
	case 5: // update or remove a relationship
		if len(m.rels) == 0 {
			return step{}, false
		}
		a := pick(r, sortedKeys(m.rels))
		src, _, _ := strings.Cut(a, ".uses.")
		path := append(elemPath(src), "uses", a[strings.LastIndex(a, ".")+1:])
		if r.IntN(2) == 0 {
			return step{edit: authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetRelationship, Address: a}, span: path,
				apply: func(m *pModel) { delete(m.rels, a) }}, true
		}
		d := pick(r, words[:5])
		return step{edit: authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetRelationship, Address: a,
			Set: []authoring.Attr{set1("description", s(d))}}, span: path, apply: func(m *pModel) { m.rels[a].desc = d }}, true
	case 6: // remove an element nothing depends on
		var leaves []string
		for _, a := range sortedKeys(m.elems) {
			if m.removable(a) {
				leaves = append(leaves, a)
			}
		}
		if len(leaves) == 0 {
			return step{}, false
		}
		a := pick(r, leaves)
		return step{edit: authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetElement, Address: a}, span: elemPath(a),
			apply: func(m *pModel) {
				delete(m.elems, a)
				for ra, rel := range m.rels {
					if rel.source == a {
						delete(m.rels, ra)
					}
				}
			}}, true
	case 7: // add an environment, or update one's region
		if r.IntN(2) == 0 || len(m.envs) == 0 {
			n, region := m.fresh("env"), pick(r, []string{"eu-west-1", "us-east-1"})
			return step{edit: authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetEnvironment, Address: "deployment." + n,
				Set: []authoring.Attr{set1("region", s(region))}}, add: true,
				apply: func(m *pModel) { m.envs[n] = [3]string{"", "", region} }}, true
		}
		n, region := pick(r, sortedKeys(m.envs)), pick(r, []string{"ap-south-1", "eu-north-1"})
		return step{edit: authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetEnvironment, Address: "deployment." + n,
			Set: []authoring.Attr{set1("region", s(region))}}, span: []string{"deployment", n},
			apply: func(m *pModel) { v := m.envs[n]; v[2] = region; m.envs[n] = v }}, true
	case 8: // add a group inside an environment or a group
		if len(m.envs) == 0 {
			return step{}, false
		}
		parents := slices.Concat(sortedKeys(m.envs), sortedKeys(m.groups))
		parent, name := pick(r, parents), m.fresh("g")
		g := parent + "/" + name
		return step{edit: authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetGroup, Address: groupAddr(g)},
			span: groupPath(parent), add: true, apply: func(m *pModel) { m.groups[g] = true }}, true
	case 9: // add an instance of a container, directly or in a group
		ctrs := m.ofKind("container")
		if len(m.envs) == 0 || len(ctrs) == 0 {
			return step{}, false
		}
		place := pick(r, slices.Concat(sortedKeys(m.envs), sortedKeys(m.groups)))
		parts := strings.Split(place, "/")
		in := &pInst{env: parts[0], name: m.fresh("i"), of: pick(r, ctrs), groups: parts[1:], attrs: map[string]any{}}
		addr := groupAddr(place) + ".instance." + in.name
		id := "deployment." + in.env + ".instance." + in.name
		return step{edit: authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetInstance, Address: addr,
			Set: []authoring.Attr{set1("of", rf(in.of))}}, span: groupPath(place), add: true,
			apply: func(m *pModel) { m.insts[id] = in }}, true
	case 10: // set an instance attribute, or add a binding
		if len(m.insts) == 0 {
			return step{}, false
		}
		id := pick(r, sortedKeys(m.insts))
		in := m.insts[id]
		if r.IntN(2) == 0 {
			mem := float64(128 * (1 + r.IntN(8)))
			return step{edit: authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetInstance, Address: id,
				Set: []authoring.Attr{{Name: "attributes", Value: authoring.AttrValue{Kind: authoring.ValueMap,
					Map: []authoring.MapEntry{{Key: "memory", Value: authoring.AttrValue{Kind: authoring.ValueNumber, Num: mem}}}}}}},
				span: in.path(), apply: func(m *pModel) { m.insts[id].attrs = map[string]any{"memory": mem} }}, true
		}
		kind, res := pick(r, []string{"terraform", "cloudformation"}), "module."+m.fresh("r")
		return step{edit: authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetBinding, Address: id,
			Binding: authoring.BindingRef{Kind: kind}, Set: []authoring.Attr{set1("address", s(res))}}, span: in.path(), add: true,
			apply: func(m *pModel) {
				m.insts[id].claims = append(m.insts[id].claims, fmt.Sprintf("kind=%s;address=%s", kind, res))
			}}, true
	case 13: // add, update or remove a view
		return m.viewStep(r)
	case 12: // rename an element: a new name, maybe a new kind, or back to a former address
		return m.renameStep(r)
	case 11: // remove an instance
		if len(m.insts) == 0 {
			return step{}, false
		}
		id := pick(r, sortedKeys(m.insts))
		path := m.insts[id].path()
		return step{edit: authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetInstance, Address: id}, span: path,
			apply: func(m *pModel) { delete(m.insts, id) }}, true
	}
	return step{}, false
}

func (m *pModel) renameStep(r *rand.Rand) (step, bool) {
	a := pick(r, sortedKeys(m.elems))
	kind, _, _ := strings.Cut(a, ".")
	toKind := kind
	topLevel := kind == "system" || kind == "person" || kind == "external"
	if topLevel && !m.hasChildren(a) && r.IntN(3) == 0 {
		toKind = pick(r, []string{"system", "person", "external"})
	}
	to := toKind + "." + m.fresh("rn")
	if back := m.prev[a]; len(back) > 0 && r.IntN(2) == 0 {
		if b := back[len(back)-1]; m.elems[b] == nil {
			to = b
		}
	}
	if to == a {
		return step{}, false
	}
	return step{edit: authoring.Edit{Op: authoring.OpRename, Target: authoring.TargetElement, Address: a, To: to},
		apply: func(m *pModel) { m.rename(a, to) }}, true
}

func (m *pModel) hasChildren(a string) bool {
	for _, e := range m.elems {
		if e.parent == a {
			return true
		}
	}
	return false
}

// rename is the model's rename: every reference follows the element.
func (m *pModel) rename(a, b string) {
	el := m.elems[a]
	delete(m.elems, a)
	el.kind, _, _ = strings.Cut(b, ".")
	m.elems[b] = el
	for _, e := range m.elems {
		if e.parent == a {
			e.parent = b
		}
	}
	for ra, rel := range m.rels {
		if rel.target == a {
			rel.target = b
		}
		if rel.source == a {
			delete(m.rels, ra)
			rel.source = b
			m.rels[b+ra[len(a):]] = rel
		}
	}
	for _, in := range m.insts {
		if in.of == a {
			in.of = b
		}
	}
	for _, v := range m.views {
		for _, l := range [][]string{v.include, v.exclude} {
			for i := range l {
				if l[i] == a {
					l[i] = b
				}
			}
		}
	}
	hist := slices.DeleteFunc(append(slices.Clone(m.prev[a]), a), func(x string) bool { return x == b })
	delete(m.prev, a)
	m.prev[b] = hist
}

// checkRename is the byte oracle for a rename, which edits every declaration
// referring to the element: outside moved blocks, each changed line differs
// from the original only by the renamed address, or is the declaring label.
func checkRename(t testing.TB, seed uint64, i int, e authoring.Edit, plan authoring.Plan) {
	fk, fn, _ := strings.Cut(e.Address, ".")
	tk, tn, _ := strings.Cut(e.To, ".")
	for _, f := range plan.Files {
		o, n := significant(f.Old), significant(f.New)
		if len(o) != len(n) {
			t.Fatalf("seed %d edit %d rename %s → %s: %s changed shape:\n%s", seed, i, e.Address, e.To, f.Path, diffLines(o, n))
		}
		for j := range o {
			label := strings.Replace(strings.Replace(o[j], fk+" ", tk+" ", 1), `"`+fn+`"`, `"`+tn+`"`, 1)
			if o[j] != n[j] && replaceRef(o[j], e.Address, e.To) != n[j] && label != n[j] {
				t.Fatalf("seed %d edit %d rename %s → %s: %s changed more than references:\n- %s+ %s", seed, i, e.Address, e.To, f.Path, o[j], n[j])
			}
		}
	}
}

// significant drops moved blocks and blank lines: renames add and remove
// moved blocks, which the IR oracle and the rename tests check.
func significant(b []byte) []string {
	var out []string
	in := false
	for _, l := range lines(b) {
		switch {
		case l == "moved {\n":
			in = true
		case in && l == "}\n":
			in = false
		case !in && strings.TrimSpace(l) != "":
			out = append(out, l)
		}
	}
	return out
}

// replaceRef replaces whole-token occurrences of from with to.
func replaceRef(line, from, to string) string {
	ident := func(c byte) bool {
		return c == '_' || c == '-' || c == '.' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
	}
	var b strings.Builder
	for i := 0; i < len(line); {
		j := strings.Index(line[i:], from)
		if j < 0 {
			b.WriteString(line[i:])
			break
		}
		start, end := i+j, i+j+len(from)
		if (start == 0 || !ident(line[start-1])) && (end == len(line) || !ident(line[end]) || line[end] == '.') {
			b.WriteString(line[i:start] + to)
		} else {
			b.WriteString(line[i:end])
		}
		i = end
	}
	return b.String()
}

// removable: no children, no incoming relationships, no instances, no view.
func (m *pModel) removable(a string) bool {
	for _, e := range m.elems {
		if e.parent == a {
			return false
		}
	}
	for _, r := range m.rels {
		if r.target == a {
			return false
		}
	}
	for _, in := range m.insts {
		if in.of == a {
			return false
		}
	}
	return !m.inView(a)
}

func set1(n string, v authoring.AttrValue) authoring.Attr { return authoring.Attr{Name: n, Value: v} }
func rf(a string) authoring.AttrValue                     { return r(a) }

func cond(c bool, a, b string) string {
	if c {
		return a
	}
	return b
}

// runSequence applies up to propMaxEdits generated edits and checks both
// oracles after each. It returns a description of the first failure.
func runSequence(t testing.TB, seed uint64, root string) {
	r := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	m := seedModel(compileRoot(t, root))
	ed := NewEditor()
	n := 1 + r.IntN(propMaxEdits)
	for i := range n {
		st := m.generate(r)
		e, err := authoring.NewEdit(st.edit)
		if err != nil {
			t.Fatalf("seed %d edit %d: generator produced an invalid edit %+v: %v", seed, i, st.edit, err)
		}
		plan, err := ed.Plan(context.Background(), root, []authoring.Edit{e})
		if err != nil {
			t.Fatalf("seed %d edit %d %s %s %s: plan: %v", seed, i, e.Op, e.Target, e.Address, err)
		}
		if e.Op == authoring.OpRename {
			checkRename(t, seed, i, e, plan)
		} else {
			checkBytes(t, seed, i, e, st, plan)
		}
		commitStep(t, ed, root, plan, i == 0)
		st.apply(m)
		if got, want := projectIR(compileRoot(t, root)), m.projection(); !slices.Equal(got, want) {
			t.Fatalf("seed %d edit %d %s %s %s: IR oracle:\n%s", seed, i, e.Op, e.Target, e.Address, diffLines(want, got))
		}
	}
}

// checkBytes is the byte oracle: one file changed, and only within the span.
func checkBytes(t testing.TB, seed uint64, i int, e authoring.Edit, st step, plan authoring.Plan) {
	if len(plan.Files) != 1 {
		t.Fatalf("seed %d edit %d: %d files changed, want 1", seed, i, len(plan.Files))
	}
	f := plan.Files[0]
	if f.Old == nil {
		return
	}
	tt, ok := t.(*testing.T)
	if !ok {
		tt = &testing.T{}
	}
	a, b := len(lines(f.Old)), len(lines(f.Old))
	if st.span != nil {
		a, b = declSpan(tt, f.Old, st.span...)
		if st.add {
			a, b = b-1, b-1
		}
	}
	if e.Op == authoring.OpRemove && e.Target == authoring.TargetElement && strings.Contains(string(f.Old), "moved {") {
		// The removed element's moved history goes with it (dropMovesTo):
		// set moved blocks and blank lines aside, then require that only
		// the element's own lines are gone.
		if !slices.Equal(significant(cut(f.Old, a, b)), significant(f.New)) {
			t.Fatalf("seed %d edit %d remove %s: byte oracle (with history):\n%s", seed, i, e.Address,
				diffLines(significant(cut(f.Old, a, b)), significant(f.New)))
		}
		return
	}
	if !spanHolds(f.Old, f.New, a, b, e.Op == authoring.OpRemove) {
		t.Fatalf("seed %d edit %d %s %s %s: byte oracle: lines outside [%d,%d) of %s changed:\n%s",
			seed, i, e.Op, e.Target, e.Address, a, b, f.Path, diffLines(lines(f.Old), lines(f.New)))
	}
}

func spanHolds(old, new []byte, a, b int, slack bool) bool {
	o, n := lines(old), lines(new)
	try := [][2]int{{a, b}}
	if slack {
		try = append(try, [2]int{a - 1, b}, [2]int{a, b + 1})
	}
	for _, sp := range try {
		a, b := max(sp[0], 0), min(sp[1], len(o))
		suffix := len(o) - b
		if len(n) >= a+suffix && slices.Equal(n[:a], o[:a]) && slices.Equal(n[len(n)-suffix:], o[b:]) {
			return true
		}
	}
	return false
}

// commitStep writes the plan. The first edit of each sequence goes through
// Commit, checking that a commit changes exactly the planned files' hashes;
// later ones are written directly, which keeps 2,000 sequences fast.
func commitStep(t testing.TB, ed *Editor, root string, plan authoring.Plan, viaCommit bool) {
	if !viaCommit {
		for _, f := range plan.Files {
			if err := os.WriteFile(filepath.Join(root, f.Path), f.New, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return
	}
	before, _ := ed.Revision(context.Background(), root)
	if err := ed.Commit(context.Background(), root, plan, before); err != nil {
		t.Fatalf("commit: %v", err)
	}
	after, _ := ed.Revision(context.Background(), root)
	for _, fh := range after.Files() {
		old, _ := before.Hash(fh.Path)
		planned := slices.ContainsFunc(plan.Files, func(f authoring.FileContent) bool {
			return f.Path == fh.Path && string(f.Old) != string(f.New)
		})
		if (old != fh.SHA256) != planned {
			t.Fatalf("commit changed %s unexpectedly (planned=%v)", fh.Path, planned)
		}
	}
}

func diffLines(want, got []string) string {
	var b strings.Builder
	for _, w := range want {
		if !slices.Contains(got, w) {
			fmt.Fprintf(&b, "- %s", w)
			if !strings.HasSuffix(w, "\n") {
				b.WriteByte('\n')
			}
		}
	}
	for _, g := range got {
		if !slices.Contains(want, g) {
			fmt.Fprintf(&b, "+ %s", g)
			if !strings.HasSuffix(g, "\n") {
				b.WriteByte('\n')
			}
		}
	}
	return b.String()
}

// TestRoundTripProperty is SC-003. -seed=N replays one failing sequence.
func TestRoundTripProperty(t *testing.T) {
	t.Parallel()
	seeds := make([]uint64, 0, *propCount)
	for i := 1; i <= *propCount; i++ {
		seeds = append(seeds, uint64(i))
	}
	if *propSeed != 0 {
		seeds = []uint64{*propSeed}
	}
	if testing.Short() {
		seeds = seeds[:min(len(seeds), 100)]
	}
	for _, seed := range seeds {
		t.Run(fmt.Sprint(seed), func(t *testing.T) {
			t.Parallel()
			runSequence(t, seed, handwrittenCopy(t))
		})
	}
}

// FuzzApplyEdits drives the same property from fuzz input.
func FuzzApplyEdits(f *testing.F) {
	for seed := range uint64(20) {
		f.Add(seed + 1)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		runSequence(t, seed, handwrittenCopy(t))
	})
}

// cut returns src without lines [a, b).
func cut(src []byte, a, b int) []byte {
	ls := lines(src)
	return []byte(strings.Join(slices.Concat(ls[:a], ls[b:]), ""))
}

func sorted(l []string) []string { return slices.Sorted(slices.Values(l)) }

// viewStep adds a view over one or two elements, replaces a view's include
// list, sets or clears its exclude list, or removes it.
func (m *pModel) viewStep(r *rand.Rand) (step, bool) {
	elems := sortedKeys(m.elems)
	pickSome := func() []string {
		out := []string{pick(r, elems)}
		if b := pick(r, elems); b != out[0] && r.IntN(2) == 0 {
			out = append(out, b)
		}
		return out
	}
	refsOf := func(l []string) authoring.AttrValue {
		return authoring.AttrValue{Kind: authoring.ValueRefList, List: l}
	}
	if len(m.views) == 0 || r.IntN(3) == 0 {
		name, inc := m.fresh("v"), pickSome()
		set := []authoring.Attr{set1("include", refsOf(inc))}
		var tags []string
		if r.IntN(2) == 0 {
			tags = []string{"flow"}
			set = append(set, set1("tags", l(tags...)))
		}
		return step{edit: authoring.Edit{Op: authoring.OpAdd, Target: authoring.TargetView, Address: "view." + name, Set: set}, add: true,
			apply: func(m *pModel) { m.views[name] = &pView{include: inc, tags: tags} }}, true
	}
	name := pick(r, sortedKeys(m.views))
	e := authoring.Edit{Op: authoring.OpUpdate, Target: authoring.TargetView, Address: "view." + name}
	st := step{span: []string{"view", name}}
	switch r.IntN(3) {
	case 0:
		inc := pickSome()
		e.Set = []authoring.Attr{set1("include", refsOf(inc))}
		st.apply = func(m *pModel) { m.views[name].include = inc }
	case 1:
		if len(m.views[name].exclude) > 0 {
			e.Clear = []string{"exclude"}
			st.apply = func(m *pModel) { m.views[name].exclude = nil }
			break
		}
		exc := []string{pick(r, elems)}
		e.Set = []authoring.Attr{set1("exclude", refsOf(exc))}
		st.apply = func(m *pModel) { m.views[name].exclude = exc }
	default:
		e = authoring.Edit{Op: authoring.OpRemove, Target: authoring.TargetView, Address: "view." + name}
		st.apply = func(m *pModel) { delete(m.views, name) }
	}
	st.edit = e
	return st, true
}
