package main

import (
	"testing"
)

func TestMatchPath(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		// ---- production-rule critical cases ----
		// internal/core/entities/**/*.go
		{
			pattern: "internal/core/entities/**/*.go",
			path:    "internal/core/entities/foo.go",
			want:    true,
		},
		{
			pattern: "internal/core/entities/**/*.go",
			path:    "internal/core/entities/sub/foo.go",
			want:    true,
		},
		// internal/core/entities/**
		{
			pattern: "internal/core/entities/**",
			path:    "internal/core/entities/foo.go",
			want:    true,
		},
		{
			pattern: "internal/core/entities/**",
			path:    "internal/core/entities",
			want:    true,
		},
		// **/*_test.go
		{
			pattern: "**/*_test.go",
			path:    "cmd/foo_test.go",
			want:    true,
		},
		{
			pattern: "**/*_test.go",
			path:    "foo_test.go",
			want:    true,
		},
		// cmd/**/*.go must NOT match cmd2/foo.go
		{
			pattern: "cmd/**/*.go",
			path:    "cmd2/foo.go",
			want:    false,
		},
		// cmd/**/*.go matches cmd/foo.go
		{
			pattern: "cmd/**/*.go",
			path:    "cmd/foo.go",
			want:    true,
		},
		// cmd/**/*.go matches cmd/sub/foo.go
		{
			pattern: "cmd/**/*.go",
			path:    "cmd/sub/foo.go",
			want:    true,
		},

		// ---- layer path patterns from structural-rules.yaml ----
		{
			pattern: "internal/core/usecases/**/*.go",
			path:    "internal/core/usecases/build_docs.go",
			want:    true,
		},
		{
			pattern: "internal/core/usecases/**/*.go",
			path:    "internal/core/usecases/sub/build_docs.go",
			want:    true,
		},
		{
			pattern: "internal/adapters/**/*.go",
			path:    "internal/adapters/graph/adapter.go",
			want:    true,
		},
		{
			pattern: "internal/mcp/**/*.go",
			path:    "internal/mcp/tools/graph_tools.go",
			want:    true,
		},
		{
			pattern: "internal/api/**/*.go",
			path:    "internal/api/handler.go",
			want:    true,
		},

		// ---- allowedImports patterns ----
		{
			pattern: "internal/core/entities/**",
			path:    "internal/core/entities/graph",
			want:    true,
		},
		{
			pattern: "internal/core/**",
			path:    "internal/core/entities/graph",
			want:    true,
		},
		{
			pattern: "internal/core/**",
			path:    "internal/core/usecases/build_docs",
			want:    true,
		},
		{
			pattern: "internal/adapters/**",
			path:    "internal/adapters/graph",
			want:    true,
		},

		// ---- single star ----
		{
			pattern: "*.go",
			path:    "main.go",
			want:    true,
		},
		{
			pattern: "*.go",
			path:    "cmd/main.go",
			want:    false, // single star doesn't cross /
		},

		// ---- question mark ----
		{
			pattern: "cmd/fo?.go",
			path:    "cmd/foo.go",
			want:    true,
		},
		{
			pattern: "cmd/fo?.go",
			path:    "cmd/fooo.go",
			want:    false,
		},

		// ---- exemption patterns ----
		{
			pattern: "**/*_cobra.go",
			path:    "cmd/new_cobra.go",
			want:    true,
		},
		{
			pattern: "**/*_cobra.go",
			path:    "new_cobra.go",
			want:    true,
		},

		// ---- no false positives across layers ----
		{
			pattern: "internal/mcp/**/*.go",
			path:    "internal/mcp2/tools/foo.go",
			want:    false,
		},
		{
			pattern: "internal/adapters/**/*.go",
			path:    "internal/adapters2/foo.go",
			want:    false,
		},

		// ---- deep nesting ----
		{
			pattern: "internal/core/entities/**/*.go",
			path:    "internal/core/entities/a/b/c/foo.go",
			want:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.pattern+"__"+tc.path, func(t *testing.T) {
			got := matchPath(tc.pattern, tc.path)
			if got != tc.want {
				t.Errorf("matchPath(%q, %q) = %v, want %v", tc.pattern, tc.path, got, tc.want)
			}
		})
	}
}
