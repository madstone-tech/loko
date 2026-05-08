package main

import (
	"os"
	"strings"
	"testing"
)

// scenariosCanonicalRules returns the same layer rules as the production
// rule file (specs/009-constitution-compliance/contracts/structural-rules.yaml)
// using a fresh literal so this test isn't coupled to the LoadRules path.
// If the YAML changes, this fixture must change too — that's a feature, not
// a bug: the test is the executable form of the spec's acceptance scenarios.
func scenariosCanonicalRules() []LayerRule {
	rules := testLayerRules()
	// testLayerRules returns the v1.1.0 table; that matches spec US2.
	// We just hand it back unchanged for clarity.
	return rules
}

// TestSpecUS2AcceptanceScenarios drives every acceptance scenario from
// specs/009-constitution-compliance/spec.md User Story 2 through the
// layer-import checker and asserts each one produces the expected verdict.
//
// The scenarios are grouped by spec-numbered case for traceability.
func TestSpecUS2AcceptanceScenarios(t *testing.T) {
	rules := scenariosCanonicalRules()
	mod := testModulePath

	cases := []struct {
		name      string
		file      string
		imports   []string
		shouldErr bool
		desc      string
	}{
		// US2 AC#1: entity-layer file imports any other internal package → FAIL
		{
			name:      "US2-AC1-entities-imports-usecases-FAILS",
			file:      "internal/core/entities/x.go",
			imports:   []string{mod + "/internal/core/usecases"},
			shouldErr: true,
			desc:      "spec US2 AC#1: entities may import nothing else internal",
		},
		{
			name:      "US2-AC1-entities-imports-adapters-FAILS",
			file:      "internal/core/entities/x.go",
			imports:   []string{mod + "/internal/adapters/d2"},
			shouldErr: true,
			desc:      "entities → adapters is forbidden",
		},
		{
			name:      "US2-AC1-entities-imports-stdlib-OK",
			file:      "internal/core/entities/x.go",
			imports:   []string{"fmt", "context"},
			shouldErr: false,
			desc:      "stdlib always allowed",
		},

		// US2 AC#2: use-case file imports anything outside entities → FAIL
		{
			name:      "US2-AC2-usecases-imports-adapters-FAILS",
			file:      "internal/core/usecases/x.go",
			imports:   []string{mod + "/internal/adapters/d2"},
			shouldErr: true,
			desc:      "spec US2 AC#2: usecases must not import adapters",
		},
		{
			name:      "US2-AC2-usecases-imports-mcp-FAILS",
			file:      "internal/core/usecases/x.go",
			imports:   []string{mod + "/internal/mcp/tools"},
			shouldErr: true,
			desc:      "usecases must not import mcp",
		},
		{
			name:      "US2-AC2-usecases-imports-entities-OK",
			file:      "internal/core/usecases/x.go",
			imports:   []string{mod + "/internal/core/entities"},
			shouldErr: false,
			desc:      "usecases may import entities",
		},

		// US2 AC#3: adapter file imports mcp/api/cmd → FAIL
		{
			name:      "US2-AC3-adapters-imports-mcp-FAILS",
			file:      "internal/adapters/d2/x.go",
			imports:   []string{mod + "/internal/mcp"},
			shouldErr: true,
			desc:      "spec US2 AC#3: adapters must not import mcp",
		},
		{
			name:      "US2-AC3-adapters-imports-api-FAILS",
			file:      "internal/adapters/d2/x.go",
			imports:   []string{mod + "/internal/api"},
			shouldErr: true,
			desc:      "adapters must not import api",
		},
		{
			name:      "US2-AC3-adapters-imports-cmd-FAILS",
			file:      "internal/adapters/d2/x.go",
			imports:   []string{mod + "/cmd"},
			shouldErr: true,
			desc:      "adapters must not import cmd",
		},

		// US2 AC#4 (v1.1.0 narrowed scope): cmd may not import entities directly.
		// mcp and api MAY import entities under v1.1.0 — those cases must PASS.
		{
			name:      "US2-AC4-cmd-imports-entities-FAILS",
			file:      "cmd/foo.go",
			imports:   []string{mod + "/internal/core/entities"},
			shouldErr: true,
			desc:      "spec US2 AC#4 (cmd-only narrowing): cmd MUST NOT import entities directly",
		},
		{
			name:      "US2-AC4-mcp-imports-entities-OK",
			file:      "internal/mcp/tools/x.go",
			imports:   []string{mod + "/internal/core/entities"},
			shouldErr: false,
			desc:      "v1.1.0: mcp MAY import entities directly (deferred to a future feature)",
		},
		{
			name:      "US2-AC4-api-imports-entities-OK",
			file:      "internal/api/handlers/x.go",
			imports:   []string{mod + "/internal/core/entities"},
			shouldErr: false,
			desc:      "v1.1.0: api MAY import entities directly (deferred to a future feature)",
		},
		{
			name:      "US2-AC4-cmd-imports-adapters-OK",
			file:      "cmd/foo.go",
			imports:   []string{mod + "/internal/adapters/filesystem"},
			shouldErr: false,
			desc:      "cmd may import adapters",
		},
		{
			name:      "US2-AC4-cmd-imports-usecases-OK",
			file:      "cmd/foo.go",
			imports:   []string{mod + "/internal/core/usecases"},
			shouldErr: false,
			desc:      "cmd may import usecases (only entities is forbidden)",
		},

		// External imports always allowed across all layers.
		{
			name:      "any-external-import-OK",
			file:      "internal/core/entities/x.go",
			imports:   []string{"github.com/spf13/cobra"},
			shouldErr: false,
			desc:      "external module imports always allowed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pf := &ParsedFile{Path: tc.file}
			for i, p := range tc.imports {
				pf.Imports = append(pf.Imports, ImportSpec{Path: p, Line: i + 1})
			}
			vs := CheckLayerImports(pf, rules, mod)
			gotErr := len(vs) > 0
			if gotErr != tc.shouldErr {
				t.Fatalf("%s\n  file=%s imports=%v\n  expected shouldErr=%v, got %d violations: %v",
					tc.desc, tc.file, tc.imports, tc.shouldErr, len(vs), vs)
			}
		})
	}
}

// TestDepguardConfigMentionsCmdEntitiesRule is a sync smoke check (T040):
// .golangci.yml encodes a single redundant rule mirroring the most-likely-
// regressed boundary (cmd/ → entities). If the rule disappears or stops
// referencing the entities package path, this test fails — prompting the
// contributor to either restore the rule or update this test along with the
// canonical YAML rule set.
//
// We intentionally do NOT validate every layer rule against depguard. The
// authoritative full check is `make audit-constitution` (tools/archcheck);
// duplicating the whole table in two formats just creates drift.
func TestDepguardConfigMentionsCmdEntitiesRule(t *testing.T) {
	src, err := os.ReadFile("../../.golangci.yml")
	if err != nil {
		t.Fatalf("read .golangci.yml: %v", err)
	}
	text := string(src)

	checks := []string{
		"depguard:",
		"cmd-no-direct-entities",
		"cmd/**/*.go",
		"github.com/madstone-tech/loko/internal/core/entities",
	}
	for _, marker := range checks {
		if !strings.Contains(text, marker) {
			t.Errorf("expected .golangci.yml to contain %q (depguard sync regression?)", marker)
		}
	}
}
