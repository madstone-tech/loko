package arch

import "slices"

// Rendering attributes (feature 016): optional, and they affect only how an
// architecture is drawn, never references, resolution or queries.

// Shapes a container or external may be drawn as.
const (
	ShapeDatabase = "database"
	ShapeQueue    = "queue"
	ShapeTopic    = "topic"
	ShapeFunction = "function"
	ShapeBucket   = "bucket"
)

// Shapes lists every shape, in documentation order.
var Shapes = []string{ShapeDatabase, ShapeQueue, ShapeTopic, ShapeFunction, ShapeBucket}

// Relationship kinds. Sync is the default and is never stored: an unset kind
// means sync, so exports omit it.
const (
	RelSync    = "sync"
	RelAsync   = "async"
	RelTrigger = "trigger"
)

// RelationshipKinds lists every relationship kind.
var RelationshipKinds = []string{RelSync, RelAsync, RelTrigger}

// Layout directions a view may choose.
const (
	DirDown  = "down"
	DirRight = "right"
)

// Directions lists every layout direction.
var Directions = []string{DirDown, DirRight}

// ValidShape reports whether s is one of Shapes.
func ValidShape(s string) bool { return slices.Contains(Shapes, s) }

// ValidRelationshipKind reports whether k is one of RelationshipKinds.
func ValidRelationshipKind(k string) bool { return slices.Contains(RelationshipKinds, k) }

// ValidDirection reports whether d is one of Directions.
func ValidDirection(d string) bool { return slices.Contains(Directions, d) }

// ShapeAllowedOn reports whether elements of kind may carry a shape:
// containers and externals, which are often data stores or queues; never
// systems (groupings) or people.
func ShapeAllowedOn(kind ElementKind) bool { return kind == KindContainer || kind == KindExternal }
