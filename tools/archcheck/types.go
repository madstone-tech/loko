// Package main implements archcheck, a build-time AST audit tool that enforces
// the loko project's Clean Architecture constitution (layer-import rules, file
// size budgets, and function size budgets) against every Go source file in the
// repository.
package main

import (
	"go/ast"
	"go/token"
)

// RuleSet is the top-level wrapper for a structural-rules YAML file.
type RuleSet struct {
	Version       string             `yaml:"version"       json:"version"`
	Layers        []LayerRule        `yaml:"layers"        json:"layers"`
	FileSizes     []FileSizeRule     `yaml:"fileSizes"     json:"fileSizes"`
	FunctionSizes []FunctionSizeRule `yaml:"functionSizes" json:"functionSizes"`
	Exemptions    []Exemption        `yaml:"exemptions"    json:"exemptions"`
}

// LayerRule constrains which internal packages a layer is allowed to import.
type LayerRule struct {
	Name             string   `yaml:"name"             json:"name"`
	PathPattern      string   `yaml:"pathPattern"      json:"pathPattern"`
	AllowedImports   []string `yaml:"allowedImports"   json:"allowedImports"`
	ForbiddenImports []string `yaml:"forbiddenImports" json:"forbiddenImports"`
	Description      string   `yaml:"description"      json:"description"`
}

// FileSizeRule defines a maximum effective-line budget for files matching a pattern.
type FileSizeRule struct {
	Name              string `yaml:"name"               json:"name"`
	PathPattern       string `yaml:"pathPattern"        json:"pathPattern"`
	MaxEffectiveLines int    `yaml:"maxEffectiveLines"  json:"maxEffectiveLines"`
	Description       string `yaml:"description"        json:"description"`
}

// FunctionSizeRule defines a maximum effective-line budget per function declaration.
type FunctionSizeRule struct {
	Name              string `yaml:"name"              json:"name"`
	PathPattern       string `yaml:"pathPattern"       json:"pathPattern"`
	MaxEffectiveLines int    `yaml:"maxEffectiveLines" json:"maxEffectiveLines"`
	Description       string `yaml:"description"       json:"description"`
}

// Exemption excludes files from a class of rules.
type Exemption struct {
	Kind   string         `yaml:"kind"   json:"kind"`
	Match  ExemptionMatch `yaml:"match"  json:"match"`
	Reason string         `yaml:"reason" json:"reason"`
}

// ExemptionMatch specifies how to match files for an exemption. Only one field
// should be non-zero per instance.
type ExemptionMatch struct {
	Basename        []string `yaml:"basename"        json:"basename,omitempty"`
	PathPattern     string   `yaml:"pathPattern"     json:"pathPattern,omitempty"`
	GeneratedHeader bool     `yaml:"generatedHeader" json:"generatedHeader,omitempty"`
}

// Violation is a single problem found by the audit.
type Violation struct {
	Rule    string `json:"rule"`
	Kind    string `json:"kind"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	Subject string `json:"subject"`
	Actual  int    `json:"actual"`
	Limit   int    `json:"limit"`
	Message string `json:"message"`
}

// Report is the top-level output document written to audit-report.json.
type Report struct {
	Version               string      `json:"version"`
	GeneratedAt           string      `json:"generated_at"`
	AuditToolVersion      string      `json:"audit_tool_version"`
	RulesPath             string      `json:"rules_path"`
	TotalFilesScanned     int         `json:"total_files_scanned"`
	TotalFunctionsScanned int         `json:"total_functions_scanned"`
	Violations            []Violation `json:"violations"`
	ExitCode              int         `json:"exit_code"`
}

// ParsedFile holds all analysis-ready data for a single Go source file.
type ParsedFile struct {
	// Path is the module-relative slash-separated path.
	Path string
	// Source is the raw file bytes.
	Source []byte
	// FileSet is the token.FileSet used to parse this file.
	FileSet *token.FileSet
	// AST is the parsed syntax tree.
	AST *ast.File
	// EffectiveLines is the count produced by CountEffectiveLines.
	EffectiveLines int
	// Imports is the list of import paths found in this file.
	Imports []ImportSpec
	// Generated is true if the file contains a "Code generated" header.
	Generated bool
}

// ImportSpec records an import path and the line it appears on.
type ImportSpec struct {
	Path string
	Line int
}
