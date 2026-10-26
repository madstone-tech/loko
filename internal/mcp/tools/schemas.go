package tools

// This file holds the tools' input schemas (contracts/mcp-tools.md).
// It is data, exempt from the per-handler size budget.

var formatProp = map[string]any{
	"type": "string", "enum": []string{"toon", "json"}, "default": "toon",
	"description": "toon (token-efficient, default) or json",
}

var baseRevisionProp = map[string]any{
	"type":        "string",
	"description": "The revision from your last read. A write is refused if a file it changes has been modified since.",
}

var rangesProp = map[string]any{
	"type": "string", "enum": []string{"line", "full"}, "default": "line",
	"description": "line gives each diagnostic as file:line (default); full gives its whole source range",
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
	"ranges":  rangesProp,
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
	"ranges":     rangesProp,
}, "kind")

// ValidateSchema is the validate tool's input schema.
var ValidateSchema = object(map[string]any{
	"summary": map[string]any{"type": "boolean", "default": false,
		"description": "Group diagnostics by code, with a count and the affected addresses"},
	"severity": map[string]any{"type": "string", "enum": []string{"error"},
		"description": "error shows errors only; omit for every diagnostic. The counts are unchanged"},
	"ignore": map[string]any{"type": "array", "items": map[string]any{"type": "string"},
		"description": "Warning codes to leave out, on top of the project's ignore_warnings"},
	"ranges": rangesProp,
	"format": formatProp,
})

// SchemaSchema is the schema tool's input schema.
var SchemaSchema = object(map[string]any{
	"block":  map[string]any{"type": "string", "description": "One block name, for example view; omit for the whole language"},
	"format": formatProp,
})

var editSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"op":      map[string]any{"type": "string", "enum": []string{"add", "update", "remove", "rename"}},
		"target":  map[string]any{"type": "string", "enum": []string{"element", "relationship", "environment", "group", "instance", "binding", "view", "project"}},
		"address": map[string]any{"type": "string", "description": "e.g. container.api, container.api.uses.orders, deployment.prod.node.vpc, deployment.prod.instance.api, view.payments, project. The schema tool lists every form"},
		"binding": map[string]any{"type": "object", "properties": map[string]any{
			"kind":  map[string]any{"type": "string", "enum": []string{"terraform", "cloudformation"}},
			"index": map[string]any{"type": "integer", "minimum": 0},
		}},
		"set": map[string]any{"type": "object",
			"description": "Attributes to set. References (system, container, target, of; a view's include and exclude lists) take address strings and are written unquoted."},
		"clear":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"cascade": map[string]any{"type": "boolean", "description": "remove: also remove everything that depends on the target"},
		"to":      map[string]any{"type": "string", "description": "rename: the new element address"},
		"file":    map[string]any{"type": "string", "description": "add: a project-relative *.loko.hcl file to place the declaration in; required when adding the project"},
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
	"ranges":  rangesProp,
}, "base_revision", "edits")

// MoveSchema is the move tool's input schema.
var MoveSchema = object(map[string]any{
	"from":          map[string]any{"type": "string", "description": "The element's current address, e.g. container.api"},
	"to":            map[string]any{"type": "string", "description": "The new address; the kind and/or name may change"},
	"base_revision": baseRevisionProp,
	"preview":       previewProp,
	"format":        formatProp,
	"ranges":        rangesProp,
}, "from", "to", "base_revision")

// FmtSchema is the fmt tool's input schema.
var FmtSchema = object(map[string]any{
	"base_revision": baseRevisionProp,
	"preview":       previewProp,
	"format":        formatProp,
	"ranges":        rangesProp,
}, "base_revision")

// BuildSchema is the build tool's input schema.
var BuildSchema = object(map[string]any{
	"formats": map[string]any{"type": "array", "items": map[string]any{"type": "string", "enum": []string{"d2", "svg", "md", "html"}},
		"description": "Formats to produce; default all"},
	"out":           map[string]any{"type": "string", "description": "Output directory inside the project; default dist"},
	"inline":        map[string]any{"type": "string", "description": "A view name, e.g. system-shop, whose diagram to return"},
	"inline_format": map[string]any{"type": "string", "enum": []string{"d2", "svg"}, "default": "d2"},
	"strict":        map[string]any{"type": "boolean", "default": false, "description": "Warnings make ok false; files are still written, as loko build --strict"},
	"ignore": map[string]any{"type": "array", "items": map[string]any{"type": "string"},
		"description": "Warning codes to leave out, on top of the project's ignore_warnings"},
	"format": formatProp,
	"ranges": rangesProp,
})
