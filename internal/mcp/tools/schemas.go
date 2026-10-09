package tools

// This file holds the input schemas of the five tools (contracts/mcp-tools.md).
// It is data, exempt from the per-handler size budget.

var formatProp = map[string]any{
	"type": "string", "enum": []string{"toon", "json"}, "default": "toon",
	"description": "toon (token-efficient, default) or json",
}

var baseRevisionProp = map[string]any{
	"type":        "string",
	"description": "The revision from your last read. A write is refused if a file it changes has been modified since.",
}

var previewProp = map[string]any{
	"type": "boolean", "default": false,
	"description": "Return the diffs without writing anything.",
}

func object(props map[string]any, required ...string) map[string]any {
	if required == nil {
		required = []string{}
	}
	return map[string]any{"type": "object", "properties": props, "required": required, "additionalProperties": false}
}

// DescribeSchema is the describe tool's input schema.
var DescribeSchema = object(map[string]any{
	"level": map[string]any{"type": "string", "enum": []string{"summary", "structure", "full"}, "default": "summary",
		"description": "summary (counts and top-level names), structure (tree to containers), full (everything)"},
	"address": map[string]any{"type": "string", "description": "Scope to one element, e.g. container.api"},
	"format":  formatProp,
})

// QuerySchema is the query tool's input schema.
var QuerySchema = object(map[string]any{
	"kind": map[string]any{"type": "string", "enum": []string{"dependents", "dependencies", "path", "orphans", "coupling"}},
	"address": map[string]any{"type": "string",
		"description": "The element to ask about (dependents, dependencies) or the start of a path"},
	"to":         map[string]any{"type": "string", "description": "The end of a path"},
	"transitive": map[string]any{"type": "boolean", "default": false, "description": "Follow relationships transitively"},
	"limit":      map[string]any{"type": "integer", "minimum": 1, "default": 20, "description": "Rows for coupling"},
	"format":     formatProp,
}, "kind")

// ValidateSchema is the validate tool's input schema.
var ValidateSchema = object(map[string]any{"format": formatProp})

var editSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"op":      map[string]any{"type": "string", "enum": []string{"add", "update", "remove", "rename"}},
		"target":  map[string]any{"type": "string", "enum": []string{"element", "relationship", "environment", "group", "instance", "binding"}},
		"address": map[string]any{"type": "string", "description": "e.g. container.api, container.api.uses.orders, deployment.prod.node.vpc, deployment.prod.instance.api"},
		"binding": map[string]any{"type": "object", "properties": map[string]any{
			"kind":  map[string]any{"type": "string", "enum": []string{"terraform", "cloudformation"}},
			"index": map[string]any{"type": "integer", "minimum": 0},
		}},
		"set": map[string]any{"type": "object",
			"description": "Attributes to set. References (system, container, target, of) take an address string and are written unquoted."},
		"clear":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"cascade": map[string]any{"type": "boolean", "description": "remove: also remove everything that depends on the target"},
		"to":      map[string]any{"type": "string", "description": "rename: the new element address"},
		"file":    map[string]any{"type": "string", "description": "add: a project-relative *.loko.hcl file to place the declaration in"},
	},
	"required":             []string{"op", "target", "address"},
	"additionalProperties": false,
}

// ApplyEditSchema is the apply_edit tool's input schema.
var ApplyEditSchema = object(map[string]any{
	"base_revision": baseRevisionProp,
	"edits": map[string]any{"type": "array", "items": editSchema, "minItems": 1, "maxItems": 100,
		"description": "Applied in order, compiled once, and saved all-or-nothing."},
	"preview": previewProp,
	"format":  formatProp,
}, "base_revision", "edits")

// MoveSchema is the move tool's input schema.
var MoveSchema = object(map[string]any{
	"from":          map[string]any{"type": "string", "description": "The element's current address, e.g. container.api"},
	"to":            map[string]any{"type": "string", "description": "The new address; the kind and/or name may change"},
	"base_revision": baseRevisionProp,
	"preview":       previewProp,
	"format":        formatProp,
}, "from", "to", "base_revision")
