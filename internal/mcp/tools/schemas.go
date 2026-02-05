package tools

// Schemas contains JSON schemas for all MCP tool inputs.
// These are used for validation and documentation.
var Schemas = map[string]any{
	"query_project": map[string]any{
		"type":        "object",
		"title":       "Query Project",
		"description": "Get metadata about the current project",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project (defaults to current)",
			},
		},
		"required": []string{},
	},
	"query_architecture": map[string]any{
		"type":        "object",
		"title":       "Query Architecture",
		"description": "Query architecture with configurable detail levels",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"detail": map[string]any{
				"type":        "string",
				"enum":        []string{"summary", "structure", "full"},
				"description": "Detail level: summary (~200 tokens), structure (~500 tokens), or full",
			},
			"target_system": map[string]any{
				"type":        "string",
				"description": "Optional: focus on a specific system",
			},
		},
		"required": []string{"project_root", "detail"},
	},
	"create_system": map[string]any{
		"type":        "object",
		"title":       "Create System",
		"description": "Create a new system in the project",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "System name (e.g., 'Payment Service')",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "What does this system do?",
			},
			"tags": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Optional tags for categorization",
			},
		},
		"required": []string{"project_root", "name"},
	},
	"create_container": map[string]any{
		"type":        "object",
		"title":       "Create Container",
		"description": "Create a new container in a system",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"system_name": map[string]any{
				"type":        "string",
				"description": "Parent system name",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "Container name",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "What does this container do?",
			},
			"technology": map[string]any{
				"type":        "string",
				"description": "Technology stack (e.g., 'Go + Fiber', 'Node.js + Express')",
			},
		},
		"required": []string{"project_root", "system_name", "name"},
	},
	"create_component": map[string]any{
		"type":        "object",
		"title":       "Create Component",
		"description": "Create a new component in a container",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"system_name": map[string]any{
				"type":        "string",
				"description": "Parent system name",
			},
			"container_name": map[string]any{
				"type":        "string",
				"description": "Parent container name",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "Component name",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "What does this component do?",
			},
		},
		"required": []string{"project_root", "system_name", "container_name", "name"},
	},
	"update_diagram": map[string]any{
		"type":        "object",
		"title":       "Update Diagram",
		"description": "Update a system or container D2 diagram source code",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"system_name": map[string]any{
				"type":        "string",
				"description": "System name",
			},
			"container_name": map[string]any{
				"type":        "string",
				"description": "Container name (optional, for container diagrams)",
			},
			"d2_source": map[string]any{
				"type":        "string",
				"description": "New D2 diagram source code",
			},
		},
		"required": []string{"project_root", "system_name", "d2_source"},
	},
	"build_docs": map[string]any{
		"type":        "object",
		"title":       "Build Docs",
		"description": "Build HTML documentation for the project",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
			"output_dir": map[string]any{
				"type":        "string",
				"description": "Output directory for HTML files",
			},
		},
		"required": []string{"project_root", "output_dir"},
	},
	"validate": map[string]any{
		"type":        "object",
		"title":       "Validate",
		"description": "Validate the project architecture for errors and warnings",
		"properties": map[string]any{
			"project_root": map[string]any{
				"type":        "string",
				"description": "Root directory of the project",
			},
		},
		"required": []string{"project_root"},
	},
}
