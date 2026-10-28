package viewmodel

import (
	"reflect"
	"testing"
)

func navLeaf(addr string) NavNode { return NavNode{Ref: LinkRef{Address: addr, Name: addr}} }

func navParent(addr string, children ...NavNode) NavNode {
	n := navLeaf(addr)
	n.Children, n.HasChildren = children, len(children) > 0
	return n
}

func navTree() []NavNode {
	return []NavNode{
		navLeaf("person.p"),
		navParent("system.a",
			navParent("container.a1", navLeaf("component.a1.x"), navLeaf("component.a1.y")),
			navParent("container.a2", navLeaf("component.a2.z"))),
		navParent("system.b", navParent("container.b1", navLeaf("component.b1.w"))),
	}
}

func TestNavFor(t *testing.T) {
	t.Parallel()
	tree := navTree()
	before := navTree()

	top := NavFor(tree, "")
	if len(top) != 3 || top[1].Children != nil || !top[1].HasChildren || top[0].HasChildren {
		t.Fatalf("top level = %+v", top)
	}

	got := NavFor(tree, "container.a1")
	a := got[1]
	if !a.Open || a.Current || len(a.Children) != 2 {
		t.Fatalf("system.a = %+v, want open with its two containers", a)
	}
	a1, a2 := a.Children[0], a.Children[1]
	if !a1.Current || !a1.Open || len(a1.Children) != 2 {
		t.Errorf("container.a1 = %+v, want current, open, with its components", a1)
	}
	if a2.Children != nil || !a2.HasChildren || a2.Open {
		t.Errorf("container.a2 = %+v, want collapsed with HasChildren", a2)
	}
	if got[2].Children != nil || got[2].Open {
		t.Errorf("system.b = %+v, want collapsed", got[2])
	}

	leaf := NavFor(tree, "component.a1.y")
	y := leaf[1].Children[0].Children[1]
	if !y.Current || y.HasChildren {
		t.Errorf("component.a1.y = %+v, want current leaf", y)
	}
	if !leaf[1].Children[0].Open || leaf[1].Children[1].Children != nil {
		t.Errorf("path not opened or sibling container expanded: %+v", leaf[1])
	}

	if !reflect.DeepEqual(tree, before) {
		t.Error("NavFor mutated its input")
	}
}
