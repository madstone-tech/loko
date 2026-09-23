package usecases

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// ValidateWarnings reports the advisory findings of FR-029.
//
// None of these stop an artefact being produced. They exist to surface an
// architecture that is technically valid but probably incomplete — and they
// only escalate the exit code when strict mode is requested, so a project can
// adopt loko without having to silence a wall of warnings on day one.
func ValidateWarnings(model *arch.SourceModel, res *Resolved, root string) arch.Diagnostics {
	var diags arch.Diagnostics
	diags = append(diags, warnOrphans(model, res)...)
	diags = append(diags, warnDocs(model, root)...)
	diags = append(diags, warnSelfRelationships(model, res)...)
	diags = append(diags, warnEmptySystems(model, res)...)
	diags = append(diags, warnUnboundInstances(model)...)
	return diags
}

// warnOrphans reports elements with no edges (FR-029).
func warnOrphans(model *arch.SourceModel, res *Resolved) arch.Diagnostics {
	// An element is connected if it is either end of a resolved edge: the
	// target side comes from the resolution map, the source side from having
	// declared any relationship at all.
	connected := map[arch.Address]bool{}
	for _, target := range res.Target {
		connected[target] = true
	}
	for _, e := range model.Elements {
		if len(e.Relations) > 0 {
			connected[e.Address()] = true
		}
	}

	// Connectivity propagates up the containment tree. In C4, relationships
	// are normally drawn between containers, so a system almost never carries
	// an edge of its own — without this, every system in every project would
	// warn, and a warning that always fires is one users learn to ignore.
	propagateConnectivity(res, connected)

	var diags arch.Diagnostics
	for _, e := range model.Elements {
		addr := e.Address()
		if connected[addr] {
			continue
		}
		diags = append(diags, arch.Diagnostic{
			Severity: arch.SeverityWarning,
			Code:     arch.CodeOrphanElement,
			Summary:  "Element has no relationships",
			Detail: fmt.Sprintf("%s neither uses anything nor is used by anything. "+
				"An element with no edges is usually either unfinished or unnecessary.", addr),
			Address: addr,
			Range:   e.Range,
		})
	}
	return diags
}

// warnDocs reports missing prose and prose files that do not exist.
func warnDocs(model *arch.SourceModel, root string) arch.Diagnostics {
	var diags arch.Diagnostics
	for _, e := range model.Elements {
		addr := e.Address()
		switch {
		case e.Docs == "":
			diags = append(diags, arch.Diagnostic{
				Severity: arch.SeverityWarning,
				Code:     arch.CodeMissingDocs,
				Summary:  "Element has no prose",
				Detail:   fmt.Sprintf("%s declares no docs file. Prose is what a diagram cannot carry.", addr),
				Address:  addr,
				Range:    e.Range,
			})
		case root != "" && !docsExist(root, e.Docs):
			diags = append(diags, arch.Diagnostic{
				Severity: arch.SeverityWarning,
				Code:     arch.CodeDocsNotFound,
				Summary:  "Prose file not found",
				Detail:   fmt.Sprintf("%s references %q, which does not exist.", addr, e.Docs),
				Address:  addr,
				Range:    e.AttrRanges["docs"],
			})
		}
	}
	return diags
}

// warnSelfRelationships reports an element that targets itself. Permitted — a
// component may call itself recursively — but more often a copy-paste slip.
func warnSelfRelationships(model *arch.SourceModel, res *Resolved) arch.Diagnostics {
	var diags arch.Diagnostics
	for _, e := range model.Elements {
		addr := e.Address()
		for _, rel := range e.Relations {
			relAddr := arch.NewRelationshipAddress(addr, rel.LocalName)
			if res.Target[relAddr] != addr {
				continue
			}
			diags = append(diags, arch.Diagnostic{
				Severity: arch.SeverityWarning,
				Code:     arch.CodeSelfRelationship,
				Summary:  "Element relates to itself",
				Detail: fmt.Sprintf("%s targets its own element. This is permitted — a component "+
					"may call itself — but it is more often a copy-paste slip.", relAddr),
				Address: relAddr,
				Range:   rel.Range,
			})
		}
	}
	return diags
}

func warnEmptySystems(model *arch.SourceModel, res *Resolved) arch.Diagnostics {
	hasChild := map[arch.Address]bool{}
	for _, parent := range res.Parent {
		hasChild[parent] = true
	}

	var diags arch.Diagnostics
	for _, e := range model.Elements {
		if e.Kind != arch.KindSystem || hasChild[e.Address()] {
			continue
		}
		diags = append(diags, arch.Diagnostic{
			Severity: arch.SeverityWarning,
			Code:     arch.CodeEmptySystem,
			Summary:  "System has no containers",
			Detail: fmt.Sprintf("%s contains nothing. A system with no containers cannot be "+
				"drawn at container level.", e.Address()),
			Address: e.Address(),
			Range:   e.Range,
		})
	}
	return diags
}

func warnUnboundInstances(model *arch.SourceModel) arch.Diagnostics {
	var diags arch.Diagnostics
	for _, env := range model.Environments {
		envAddr := arch.NewEnvironmentAddress(env.Name)
		for _, placed := range env.AllInstances() {
			if len(placed.Instance.Claims) > 0 {
				continue
			}
			addr := arch.NewInstanceAddress(envAddr, placed.Instance.Name)
			diags = append(diags, arch.Diagnostic{
				Severity: arch.SeverityWarning,
				Code:     arch.CodeUnboundInstance,
				Summary:  "Instance claims no physical resource",
				Detail: fmt.Sprintf("%s has no binding, so reconciliation cannot tell whether it "+
					"exists in the environment.", addr),
				Address: addr,
				Range:   placed.Instance.Range,
			})
		}
	}
	return diags
}

// docsExist resolves a prose reference relative to the project root. A path
// escaping the root is treated as missing rather than followed.
func docsExist(root, docs string) bool {
	clean := filepath.Clean(filepath.Join(root, filepath.FromSlash(docs)))
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	absDocs, err := filepath.Abs(clean)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absRoot, absDocs)
	if err != nil || rel == ".." || len(rel) > 2 && rel[:3] == ".."+string(filepath.Separator) {
		return false
	}
	info, err := os.Stat(absDocs)
	return err == nil && !info.IsDir()
}

// propagateConnectivity marks every ancestor of a connected element connected.
//
// Iterative with a visited set rather than recursive: a containment cycle is
// reported elsewhere as an error, but this must not hang when one is present.
func propagateConnectivity(res *Resolved, connected map[arch.Address]bool) {
	seeds := make([]arch.Address, 0, len(connected))
	for addr := range connected {
		seeds = append(seeds, addr)
	}

	for _, addr := range seeds {
		visited := map[arch.Address]bool{addr: true}
		for {
			parent, ok := res.Parent[addr]
			if !ok || visited[parent] {
				break
			}
			connected[parent] = true
			visited[parent] = true
			addr = parent
		}
	}
}
