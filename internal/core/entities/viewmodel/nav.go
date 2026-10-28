package viewmodel

// NavNode is one entry of the site's Architecture tree (feature 021).
type NavNode struct {
	// Ref carries no technology or description, to keep every page small.
	Ref      LinkRef   `json:"ref"`
	Children []NavNode `json:"children,omitempty"`
	// HasChildren survives pruning, so a collapsed node can still show ▸.
	HasChildren bool `json:"hasChildren,omitempty"`
	// Current and Open are set only in a copy returned by NavFor.
	Current bool `json:"current,omitempty"`
	Open    bool `json:"open,omitempty"`
}

// NavFor returns the tree a page shows (spec FR-002): every top-level node,
// the children of each ancestor of address, and the children of address
// itself. Other branches are cut, keeping HasChildren. address is marked
// Current and its ancestors and itself Open. An empty address gives the top
// level only. The input is never modified.
func NavFor(nav []NavNode, address string) []NavNode {
	out := make([]NavNode, len(nav))
	for i, n := range nav {
		out[i] = prune(n, address)
	}
	return out
}

// prune copies n, keeping its children only when address is n or lies below it.
func prune(n NavNode, address string) NavNode {
	c := NavNode{Ref: n.Ref, HasChildren: n.HasChildren}
	if address == "" || !contains(n, address) {
		return c
	}
	c.Open = true
	c.Current = n.Ref.Address == address
	for _, child := range n.Children {
		c.Children = append(c.Children, prune(child, address))
	}
	return c
}

func contains(n NavNode, address string) bool {
	if n.Ref.Address == address {
		return true
	}
	for _, c := range n.Children {
		if contains(c, address) {
			return true
		}
	}
	return false
}
