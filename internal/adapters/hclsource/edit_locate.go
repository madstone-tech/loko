package hclsource

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/madstone-tech/loko/internal/core/entities/authoring"
)

// located is one declaration found in the workspace.
type located struct {
	path   string
	file   *hclwrite.File
	block  *hclwrite.Block
	parent *hclwrite.Body // the body that holds block
	depth  int            // 0 for a top-level block
}

// child returns this declaration's direct child block of the given type and
// first label ("" matches any label).
func (l located) child(typ, label string) (located, bool) {
	for _, b := range l.block.Body().Blocks() {
		if b.Type() == typ && (label == "" || firstLabel(b) == label) {
			return located{path: l.path, file: l.file, block: b, parent: l.block.Body(), depth: l.depth + 1}, true
		}
	}
	return located{}, false
}

func firstLabel(b *hclwrite.Block) string {
	if ls := b.Labels(); len(ls) > 0 {
		return ls[0]
	}
	return ""
}

// findTop finds a top-level block in any source file, in path order.
func (w *workspace) findTop(typ, label string) (located, bool) {
	for _, p := range w.paths {
		f, err := w.file(p)
		if err != nil {
			continue // recorded in w.broken; reported if nothing is found
		}
		for _, b := range f.Body().Blocks() {
			if b.Type() == typ && (label == "" || firstLabel(b) == label) {
				return located{path: p, file: f, block: b, parent: f.Body()}, true
			}
		}
	}
	return located{}, false
}

// locate finds the declaration an edit addresses.
func (w *workspace) locate(t authoring.TargetKind, address string, binding authoring.BindingRef) (located, bool) {
	parts, ok := authoring.SplitAddress(t, address)
	if t == authoring.TargetViewEntry {
		return w.findTop(blockView, viewName(address))
	}
	if !ok {
		return located{}, false
	}
	switch t {
	case authoring.TargetElement:
		kind, name := splitElement(parts.Element)
		return w.findTop(kind, name)
	case authoring.TargetRelationship:
		kind, name := splitElement(parts.Element)
		el, ok := w.findTop(kind, name)
		if !ok {
			return located{}, false
		}
		return el.child(blockUses, parts.Local)
	case authoring.TargetEnvironment:
		return w.findTop(blockDeployment, parts.Env)
	case authoring.TargetGroup:
		return w.group(parts.Env, parts.Groups)
	case authoring.TargetInstance:
		return w.instance(parts.Env, parts.Local)
	case authoring.TargetBinding:
		in, ok := w.instance(parts.Env, parts.Local)
		if !ok {
			return located{}, false
		}
		return in.binding(binding)
	}
	return located{}, false
}

// group walks an environment's node path; an empty path is the environment.
func (w *workspace) group(env string, path []string) (located, bool) {
	at, ok := w.findTop(blockDeployment, env)
	for _, g := range path {
		if !ok {
			break
		}
		at, ok = at.child(blockNode, g)
	}
	return at, ok
}

// instance finds an instance anywhere inside an environment: its identity
// does not include the placement path (FR-024 of feature 013).
func (w *workspace) instance(env, name string) (located, bool) {
	at, ok := w.findTop(blockDeployment, env)
	if !ok {
		return located{}, false
	}
	return findInstance(at, name)
}

func findInstance(at located, name string) (located, bool) {
	if in, ok := at.child(blockInstance, name); ok {
		return in, true
	}
	for _, b := range at.block.Body().Blocks() {
		if b.Type() != blockNode {
			continue
		}
		g := located{path: at.path, file: at.file, block: b, parent: at.block.Body(), depth: at.depth + 1}
		if in, ok := findInstance(g, name); ok {
			return in, true
		}
	}
	return located{}, false
}

// binding picks the Index-th binding of the given kind inside an instance.
func (l located) binding(ref authoring.BindingRef) (located, bool) {
	n := 0
	for _, b := range l.block.Body().Blocks() {
		if b.Type() != blockBinding || firstLabel(b) != ref.Kind {
			continue
		}
		if n == ref.Index {
			return located{path: l.path, file: l.file, block: b, parent: l.block.Body(), depth: l.depth + 1}, true
		}
		n++
	}
	return located{}, false
}

func splitElement(addr string) (kind, name string) {
	for i := range len(addr) {
		if addr[i] == '.' {
			return addr[:i], addr[i+1:]
		}
	}
	return addr, ""
}

func viewName(addr string) string {
	_, name := splitElement(addr)
	return name
}

// notFound is the refusal for a missing target. If some file could not be
// parsed, the target may be in it, so that is said instead.
func (w *workspace) notFound(address string) error {
	for _, p := range w.paths {
		if err, ok := w.broken[p]; ok {
			return refuse(authoring.ReasonCompileErrors, "%s was not found, and %v", address, err)
		}
	}
	return refuse(authoring.ReasonNotFound, "%s is not declared", address)
}

func refuse(reason, format string, args ...any) error {
	return &authoring.EditError{Reason: reason, Detail: fmt.Sprintf(format, args...)}
}
