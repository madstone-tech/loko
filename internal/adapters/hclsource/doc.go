// Package hclsource is the HCL front end for the loko compiler: it discovers
// *.loko.hcl files, parses them, evaluates scalar attributes, extracts
// references as static traversals, and emits an unresolved entities.SourceModel
// plus diagnostics.
//
// This package is the ONLY package in the module permitted to import
// github.com/hashicorp/hcl/... or github.com/zclconf/go-cty/... (FR-044).
// The rule is enforced by tools/archcheck and by the depguard linter. Core
// packages must never see an hcl.Range or a cty.Value; ranges cross the
// boundary as entities.SourceRange and scalars as entities.Value.
//
// Reference resolution and every validation rule live in internal/core, not
// here. This package reports syntax-level problems only: unparseable files,
// unknown blocks, unknown attributes, and unknown functions. See
// specs/013-hcl-compiler-core/research.md R2 and R7 for why the split falls
// here.
package hclsource
