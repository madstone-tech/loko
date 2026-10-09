package authoring

import (
	"slices"
	"strings"
)

// Closed value sets for the rendering attributes (feature 016). They mirror
// arch.Shapes, arch.RelationshipKinds and arch.Directions, which this package
// may not import; a parity test in usecases keeps them in step.
var (
	Shapes            = []string{"database", "queue", "topic", "function", "bucket"}
	RelationshipKinds = []string{"sync", "async", "trigger"}
	Directions        = []string{"down", "right"}
)

// enumAttrs maps an attribute to its allowed values, by target.
var enumAttrs = map[TargetKind]map[string][]string{
	TargetElement:      {"shape": Shapes},
	TargetRelationship: {"kind": RelationshipKinds},
	TargetView:         {"direction": Directions},
}

// shapeKinds are the element kinds that may carry a shape.
var shapeKinds = []string{"container", "external"}

// checkEnum rejects a rendering attribute outside its value set, and an empty
// title (it would draw a nameless box).
func checkEnum(t TargetKind, a Attr) error {
	if t == TargetElement && a.Name == "title" && a.Value.Str == "" {
		return fieldError("set.title", "must not be empty")
	}
	allowed, ok := enumAttrs[t][a.Name]
	if !ok || slices.Contains(allowed, a.Value.Str) {
		return nil
	}
	return fieldError("set."+a.Name, "%q is not allowed; use one of %s", a.Value.Str, strings.Join(allowed, ", "))
}
