package usecases

import (
	"fmt"
	"strings"

	"github.com/madstone-tech/loko/internal/core/entities/arch"
)

// ValidateRenderAttributes checks the rendering attributes of feature 016:
// title, shape, relationship kind, view direction, and the layout engine. They affect drawing
// only, so their errors never cascade into resolution or queries.
func ValidateRenderAttributes(model *arch.SourceModel) arch.Diagnostics {
	var diags arch.Diagnostics
	for _, e := range model.Elements {
		addr := e.Address()
		if _, set := e.AttrRanges["title"]; set && e.Title == "" {
			diags = append(diags, renderDiag(e.AttrRanges["title"], addr, arch.CodeEmptyTitle, "Empty title",
				"A title is shown in place of the name; an empty one would draw a nameless box. Remove it or give it text."))
		}
		if e.Shape == "" {
			continue
		}
		switch {
		case !arch.ShapeAllowedOn(e.Kind):
			diags = append(diags, renderDiag(e.AttrRanges["shape"], addr, arch.CodeShapeNotAllowed, "Shape not allowed here",
				fmt.Sprintf("%s cannot carry a shape. Shapes are for a container or external element.", addr)))
		case !arch.ValidShape(e.Shape):
			diags = append(diags, invalidValue(e.AttrRanges["shape"], addr, "shape", e.Shape, arch.Shapes))
		}
	}
	for _, e := range model.Elements {
		for _, r := range e.Relations {
			if r.Kind != "" && !arch.ValidRelationshipKind(r.Kind) {
				diags = append(diags, invalidValue(r.KindRange, arch.NewRelationshipAddress(e.Address(), r.LocalName), "kind", r.Kind, arch.RelationshipKinds))
			}
		}
	}
	for _, v := range model.Views {
		if v.Direction != "" && !arch.ValidDirection(v.Direction) {
			diags = append(diags, invalidValue(v.DirectionRange, arch.NewViewAddress(v.Name), "direction", v.Direction, arch.Directions))
		}
		if v.Layout != "" && !arch.ValidLayout(v.Layout) {
			diags = append(diags, invalidValue(v.LayoutRange, arch.NewViewAddress(v.Name), "layout", v.Layout, arch.Layouts))
		}
	}
	if l := model.Project.Layout; l != "" && !arch.ValidLayout(l) {
		diags = append(diags, invalidValue(model.Project.LayoutRange, "", "layout", l, arch.Layouts))
	}
	return diags
}

func invalidValue(r arch.SourceRange, addr arch.Address, attr, got string, allowed []string) arch.Diagnostic {
	return renderDiag(r, addr, arch.CodeInvalidAttributeValue, "Invalid "+attr,
		fmt.Sprintf("%q is not a valid %s. Allowed values: %s.", got, attr, strings.Join(allowed, ", ")))
}

func renderDiag(r arch.SourceRange, addr arch.Address, code, summary, detail string) arch.Diagnostic {
	return arch.Diagnostic{Severity: arch.SeverityError, Code: code, Summary: summary, Detail: detail, Address: addr, Range: r}
}
