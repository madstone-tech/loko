package viewmodel

import (
	"fmt"
	"sort"
)

// NodeRole says what a node draws.
type NodeRole string

const (
	RoleElement  NodeRole = "element"
	RoleSubject  NodeRole = "subject" // the boundary box of a system, container or environment view
	RoleGroup    NodeRole = "group"   // a deployment placement group
	RoleInstance NodeRole = "instance"
	RoleOutside  NodeRole = "outside" // the single synthetic end of every boundary connection
)

// OutsideNodeID is the ID of a view's synthetic outside node.
const OutsideNodeID = "outside"

// Node is one drawn element, group, instance, subject boundary, or the
// outside marker.
type Node struct {
	ID          string   `json:"id"`
	Address     string   `json:"address,omitempty"`
	Role        NodeRole `json:"role"`
	Kind        string   `json:"kind,omitempty"`
	Label       string   `json:"label"`
	Technology  string   `json:"technology,omitempty"`
	Description string   `json:"description,omitempty"`
	// Parent is the enclosing node's ID; empty at top level.
	Parent string `json:"parent,omitempty"`
	Style  Style  `json:"style"`
	// Link is the site page of the element drawn, produced only by Paths.
	Link string `json:"link,omitempty"`
}

// Edge is one drawn connection, merged from one or more relationships.
type Edge struct {
	ID            string    `json:"id"`
	Source        string    `json:"source"`
	Target        string    `json:"target"`
	Label         string    `json:"label,omitempty"`
	Technology    string    `json:"technology,omitempty"`
	Relationships []string  `json:"relationships"`
	Crossing      bool      `json:"crossing,omitempty"`
	Style         EdgeStyle `json:"style"`
}

// EdgeIDFor is "<source>--<target>", with "--self" appended for self-loops.
func EdgeIDFor(source, target string) string {
	if source == target {
		return source + "--" + target + "--self"
	}
	return source + "--" + target
}

// SortEdges orders edges by (Source, Target, ID).
func SortEdges(es []Edge) {
	sort.Slice(es, func(i, j int) bool { return edgeLess(es[i], es[j]) })
}

func edgeLess(a, b Edge) bool {
	if a.Source != b.Source {
		return a.Source < b.Source
	}
	if a.Target != b.Target {
		return a.Target < b.Target
	}
	return a.ID < b.ID
}

// ViewModel is the intermediate value one view projects to. It is the only
// thing a per-view backend reads (FR-011), and styling is already decided
// (FR-008).
type ViewModel struct {
	View  View   `json:"view"`
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
	// Sources are the project-relative files that declared what this view
	// depicts; they feed the generated-file notice.
	Sources     []string `json:"sources"`
	DiagramPath string   `json:"diagramPath"`
	PagePath    string   `json:"pagePath"`
}

// Validate checks the referential invariants of data-model §2. A failure is a
// projection bug, never a user error.
func (m ViewModel) Validate() error {
	index, err := m.validateNodes()
	if err != nil {
		return err
	}
	crossing := 0
	for i, e := range m.Edges {
		if _, ok := index[e.Source]; !ok {
			return fmt.Errorf("view %s: edge %s has unknown source %q", m.View.ID, e.ID, e.Source)
		}
		if _, ok := index[e.Target]; !ok {
			return fmt.Errorf("view %s: edge %s has unknown target %q", m.View.ID, e.ID, e.Target)
		}
		touchesOutside := e.Source == OutsideNodeID || e.Target == OutsideNodeID
		if e.Crossing != touchesOutside {
			return fmt.Errorf("view %s: crossing edge %s must connect to the outside node, and only crossing edges may", m.View.ID, e.ID)
		}
		if e.Crossing {
			crossing++
		}
		if i > 0 && !edgeLess(m.Edges[i-1], e) {
			return fmt.Errorf("view %s: edges not sorted at %s", m.View.ID, e.ID)
		}
	}
	if _, ok := index[OutsideNodeID]; ok && crossing == 0 {
		return fmt.Errorf("view %s: outside node without a crossing edge", m.View.ID)
	}
	return nil
}

func (m ViewModel) validateNodes() (map[string]Node, error) {
	index := make(map[string]Node, len(m.Nodes))
	outside := 0
	for i, n := range m.Nodes {
		if i > 0 {
			switch prev := m.Nodes[i-1].ID; {
			case n.ID == prev:
				return nil, fmt.Errorf("view %s: duplicate node %q", m.View.ID, n.ID)
			case n.ID < prev:
				return nil, fmt.Errorf("view %s: nodes not sorted at %q", m.View.ID, n.ID)
			}
		}
		if n.Role == RoleOutside {
			outside++
		}
		index[n.ID] = n
	}
	if outside > 1 {
		return nil, fmt.Errorf("view %s: more than one outside node", m.View.ID)
	}
	for _, n := range m.Nodes {
		if n.Parent == "" {
			continue
		}
		if _, ok := index[n.Parent]; !ok {
			return nil, fmt.Errorf("view %s: node %q has unknown parent %q", m.View.ID, n.ID, n.Parent)
		}
		seen := map[string]bool{n.ID: true}
		for p := n.Parent; p != ""; p = index[p].Parent {
			if seen[p] {
				return nil, fmt.Errorf("view %s: nesting cycle through %q", m.View.ID, p)
			}
			seen[p] = true
		}
	}
	return index, nil
}
