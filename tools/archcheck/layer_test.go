package main

import (
	"testing"
)

const testModulePath = "github.com/madstone-tech/loko"

// rulesFromYAML returns the layer rules exactly as defined in structural-rules.yaml.
func testLayerRules() []LayerRule {
	return []LayerRule{
		{
			Name:           "core/entities",
			PathPattern:    "internal/core/entities/**/*.go",
			AllowedImports: []string{},
			Description:    "Entity layer is the innermost ring; pure structs and validation rules. May import only the Go standard library.",
		},
		{
			Name:        "core/usecases",
			PathPattern: "internal/core/usecases/**/*.go",
			AllowedImports: []string{
				"internal/core/entities/**",
			},
			Description: "Use-case layer orchestrates entities through ports. May import entities and the standard library only. Adapters, mcp, api, cmd are forbidden.",
		},
		{
			Name:        "adapters",
			PathPattern: "internal/adapters/**/*.go",
			AllowedImports: []string{
				"internal/core/entities/**",
				"internal/core/usecases/**",
			},
			Description: "Adapter layer implements ports defined in usecases. May import core only.",
		},
		{
			Name:        "mcp",
			PathPattern: "internal/mcp/**/*.go",
			AllowedImports: []string{
				"internal/core/**",
				"internal/adapters/**",
			},
			Description: "MCP server may import core and adapters. May not import api or cmd.",
		},
		{
			Name:        "api",
			PathPattern: "internal/api/**/*.go",
			AllowedImports: []string{
				"internal/core/**",
				"internal/adapters/**",
			},
			Description: "HTTP API server may import core and adapters. May not import mcp or cmd.",
		},
		{
			Name:        "cmd",
			PathPattern: "cmd/**/*.go",
			AllowedImports: []string{
				"internal/core/**",
				"internal/adapters/**",
				"internal/mcp/**",
				"internal/api/**",
			},
			ForbiddenImports: []string{
				"internal/core/entities/**",
			},
			Description: "CLI is the outer composition root. May import any internal layer. MUST NOT import internal/core/entities directly.",
		},
	}
}

func makeFile(filePath string, imports ...string) *ParsedFile {
	specs := make([]ImportSpec, len(imports))
	for i, imp := range imports {
		specs[i] = ImportSpec{Path: imp, Line: i + 1}
	}
	return &ParsedFile{
		Path:    filePath,
		Imports: specs,
	}
}

func assertNoViolations(t *testing.T, violations []Violation) {
	t.Helper()
	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d: %v", len(violations), violations)
	}
}

func assertViolation(t *testing.T, violations []Violation, count int) {
	t.Helper()
	if len(violations) != count {
		t.Errorf("expected %d violation(s), got %d: %v", count, len(violations), violations)
	}
}

func TestCheckLayerImports(t *testing.T) {
	rules := testLayerRules()
	mod := testModulePath

	t.Run("entities importing usecases FAIL", func(t *testing.T) {
		f := makeFile(
			"internal/core/entities/graph.go",
			mod+"/internal/core/usecases/build_docs",
		)
		assertViolation(t, CheckLayerImports(f, rules, mod), 1)
	})

	t.Run("usecases importing adapters FAIL", func(t *testing.T) {
		f := makeFile(
			"internal/core/usecases/build_docs.go",
			mod+"/internal/adapters/graph",
		)
		assertViolation(t, CheckLayerImports(f, rules, mod), 1)
	})

	t.Run("adapters importing mcp FAIL", func(t *testing.T) {
		f := makeFile(
			"internal/adapters/graph/adapter.go",
			mod+"/internal/mcp/tools",
		)
		assertViolation(t, CheckLayerImports(f, rules, mod), 1)
	})

	t.Run("mcp importing api FAIL", func(t *testing.T) {
		f := makeFile(
			"internal/mcp/tools/graph_tools.go",
			mod+"/internal/api/handler",
		)
		assertViolation(t, CheckLayerImports(f, rules, mod), 1)
	})

	t.Run("mcp importing entities PASS (mcp.allowedImports includes internal/core/**)", func(t *testing.T) {
		f := makeFile(
			"internal/mcp/tools/graph_tools.go",
			mod+"/internal/core/entities/graph",
		)
		assertNoViolations(t, CheckLayerImports(f, rules, mod))
	})

	t.Run("api importing entities PASS", func(t *testing.T) {
		f := makeFile(
			"internal/api/handler.go",
			mod+"/internal/core/entities/graph",
		)
		assertNoViolations(t, CheckLayerImports(f, rules, mod))
	})

	t.Run("cmd importing entities FAIL (forbiddenImports override)", func(t *testing.T) {
		f := makeFile(
			"cmd/new.go",
			mod+"/internal/core/entities/graph",
		)
		assertViolation(t, CheckLayerImports(f, rules, mod), 1)
	})

	t.Run("cmd importing adapters PASS", func(t *testing.T) {
		f := makeFile(
			"cmd/new.go",
			mod+"/internal/adapters/graph",
		)
		assertNoViolations(t, CheckLayerImports(f, rules, mod))
	})

	t.Run("cmd importing mcp PASS", func(t *testing.T) {
		f := makeFile(
			"cmd/new.go",
			mod+"/internal/mcp/server",
		)
		assertNoViolations(t, CheckLayerImports(f, rules, mod))
	})

	t.Run("cmd importing api PASS", func(t *testing.T) {
		f := makeFile(
			"cmd/new.go",
			mod+"/internal/api/handler",
		)
		assertNoViolations(t, CheckLayerImports(f, rules, mod))
	})

	t.Run("cmd importing usecases PASS (allowed by internal/core/**)", func(t *testing.T) {
		f := makeFile(
			"cmd/new.go",
			mod+"/internal/core/usecases/build_docs",
		)
		assertNoViolations(t, CheckLayerImports(f, rules, mod))
	})

	t.Run("all layers importing stdlib fmt PASS", func(t *testing.T) {
		paths := []string{
			"internal/core/entities/graph.go",
			"internal/core/usecases/build_docs.go",
			"internal/adapters/graph/adapter.go",
			"internal/mcp/tools/graph_tools.go",
			"internal/api/handler.go",
			"cmd/new.go",
		}
		for _, p := range paths {
			f := makeFile(p, "fmt")
			assertNoViolations(t, CheckLayerImports(f, rules, mod))
		}
	})

	t.Run("all layers importing external module PASS", func(t *testing.T) {
		paths := []string{
			"internal/core/entities/graph.go",
			"internal/core/usecases/build_docs.go",
			"internal/adapters/graph/adapter.go",
			"internal/mcp/tools/graph_tools.go",
			"internal/api/handler.go",
			"cmd/new.go",
		}
		for _, p := range paths {
			f := makeFile(p, "github.com/spf13/cobra")
			assertNoViolations(t, CheckLayerImports(f, rules, mod))
		}
	})

	t.Run("unconstrained file returns nil", func(t *testing.T) {
		f := makeFile("tools/archcheck/main.go", mod+"/internal/core/entities/graph")
		violations := CheckLayerImports(f, rules, mod)
		if violations != nil {
			t.Errorf("expected nil for unconstrained file, got %v", violations)
		}
	})
}
