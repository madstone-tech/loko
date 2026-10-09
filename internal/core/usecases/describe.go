package usecases

import (
	"cmp"
	"context"
	"slices"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// DescribeRequest asks for the project at one level of detail, optionally
// scoped to one element (FR-001, FR-002).
type DescribeRequest struct {
	Level   string // summary (default) | structure | full
	Address string
}

// DescribeResult is the project as an assistant first sees it. Which fields
// are filled depends on the level (research R6).
type DescribeResult struct {
	OK            bool               `json:"ok" toon:"ok"`
	Level         string             `json:"level" toon:"level"`
	Project       ProjectInfo        `json:"project" toon:"project"`
	Counts        []KindCount        `json:"counts,omitempty" toon:"counts,omitempty"`
	Elements      []ElementView      `json:"elements,omitempty" toon:"elements,omitempty"`
	Relationships []RelationshipView `json:"relationships,omitempty" toon:"relationships,omitempty"`
	Environments  []EnvironmentView  `json:"environments,omitempty" toon:"environments,omitempty"`
	Views         []ViewInfo         `json:"views,omitempty" toon:"views,omitempty"`
	Error         *ReadError         `json:"error,omitempty" toon:"error,omitempty"`
	Diags         arch.Diagnostics   `json:"diagnostics,omitempty" toon:"diagnostics,omitempty"`
	Revision      string             `json:"revision" toon:"revision"`
}

// Describe compiles the project and describes it. A project that does not
// compile returns its diagnostics and no data (FR-007).
func Describe(ctx context.Context, deps AuthoringDeps, req DescribeRequest) (*DescribeResult, error) {
	ir, res, err := compileIR(ctx, deps)
	if err != nil {
		return nil, err
	}
	out := &DescribeResult{Level: cmp.Or(req.Level, "summary")}
	switch {
	case ir == nil:
		out.Diags = res.Diags.SortedForOutput()
	case req.Address != "" && !ir.Has(arch.Address(req.Address)):
		out.Error = &ReadError{Reason: "not_found", Detail: "no element " + req.Address,
			Suggestions: suggestAddresses(ir, req.Address)}
	default:
		describeIR(ir, out, arch.Address(req.Address))
	}
	if out.Revision, err = currentRevision(ctx, deps); err != nil {
		return nil, err
	}
	return out, nil
}

func describeIR(ir *arch.IR, out *DescribeResult, scope arch.Address) {
	out.OK = true
	out.Project = ProjectInfo{Name: ir.Project.Name, Description: ir.Project.Description, Layout: ir.Project.Layout}
	full, g := out.Level == "full", newQueryGraph(ir)
	for _, e := range ir.Elements {
		if keepElement(g, e, out.Level, scope) {
			out.Elements = append(out.Elements, elementView(ir, e, out.Level))
		}
	}
	for _, r := range ir.Relationships {
		if (full && scope == "") || (scope != "" && (r.Source == scope || r.Target == scope)) {
			out.Relationships = append(out.Relationships, RelationshipView{Address: string(r.Address),
				Source: string(r.Source), Target: string(r.Target), Description: r.Description, Technology: r.Technology,
				Kind: r.Kind, Tags: r.Tags})
		}
	}
	if scope != "" {
		return
	}
	out.Counts = countKinds(ir)
	for _, v := range ir.Views {
		info := ViewInfo{Address: string(v.Address), Tags: v.Tags, Direction: v.Direction, Layout: v.Layout}
		if full {
			info.Include, info.Exclude = addrStrings(v.Include), addrStrings(v.Exclude)
		}
		out.Views = append(out.Views, info)
	}
	for _, env := range ir.Environments {
		v := EnvironmentView{Address: string(env.Address), Name: env.Name}
		for _, in := range env.Instances {
			if full {
				v.Instances = append(v.Instances, InstanceView{Address: string(in.Address), Of: string(in.Of), PlacedIn: string(in.PlacedIn)})
			}
		}
		out.Environments = append(out.Environments, v)
	}
}

// keepElement: summary lists top-level elements, structure the tree down to
// containers, full everything. A scope keeps the element and its children
// (its whole subtree at full).
func keepElement(g *queryGraph, e arch.Element, level string, scope arch.Address) bool {
	if scope != "" {
		return e.Address == scope || e.Parent == scope || (level == "full" && g.within(e.Address, scope))
	}
	switch level {
	case "summary":
		return e.Parent == ""
	case "structure":
		return e.Kind != arch.KindComponent
	}
	return true
}

func elementView(ir *arch.IR, e arch.Element, level string) ElementView {
	v := ElementView{Address: string(e.Address), Title: e.Title, Kind: string(e.Kind), Name: e.Name}
	if level == "summary" {
		return v
	}
	if level == "structure" {
		v.Kind, v.Name = "", ""
	}
	v.Parent, v.Technology, v.Shape = string(e.Parent), e.Technology, e.Shape
	if e.Kind == arch.KindContainer {
		v.Components = len(ir.Children(e.Address))
	}
	if level == "full" {
		v.Description, v.Owner, v.Tags, v.Docs = e.Description, e.Owner, e.Tags, e.Docs
	}
	return v
}

func countKinds(ir *arch.IR) []KindCount {
	counts := map[string]int{}
	for _, e := range ir.Elements {
		counts[string(e.Kind)]++
	}
	out := make([]KindCount, 0, len(counts))
	for k, n := range counts {
		out = append(out, KindCount{Kind: k, Count: n})
	}
	slices.SortFunc(out, func(a, b KindCount) int { return cmp.Compare(a.Kind, b.Kind) })
	return out
}
