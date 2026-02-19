#!/usr/bin/env bash
#
# audit-constitution.sh - Validate handler files follow Clean Architecture constitution
#
# Constitution Rules:
# - CLI handlers (cmd/*.go): < 150 lines (excluding imports, comments, blank lines)
# - MCP tools (internal/mcp/tools/*.go): < 100 lines (excluding imports, comments, blank lines)
# - Handlers should only: parse → call use case → format response
# - Pure-data files (schemas.go, registry.go, helpers.go) are excluded from enforcement
#
# Known violations (tracked, not blocking CI):
#   These files have documented reasons for exceeding limits and are being refactored incrementally.
#   Add a file to KNOWN_VIOLATIONS to suppress CI failures while tracking the debt.
#
# Usage:
#   ./scripts/audit-constitution.sh [--verbose] [--baseline]
#
# Exit codes:
#   0 - No NEW violations (known violations are allowed)
#   1 - New (undocumented) violations detected
#   2 - Script error
#

set -euo pipefail

VERBOSE=false
BASELINE=false
VIOLATIONS=0
KNOWN_VIOLATION_FAILURES=0
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Known violations: files that exceed limits for documented reasons.
# Format: "relative/path/to/file.go:reason"
# These are tracked in the audit output but do NOT fail CI.
# Remove entries when the file is refactored below its limit.
declare -A KNOWN_VIOLATIONS=(
	["cmd/new.go"]="Contains --preview flag logic and component template wiring (007-ux-improvements); refactor to use case in next sprint"
	["cmd/build.go"]="Orchestrates multi-format build; refactor formatters to BuildDocs options in next sprint"
	["cmd/new_cobra.go"]="Cobra flag registration for new command; split into sub-command files in next sprint"
	["internal/mcp/tools/graph_tools.go"]="Contains 3 tool implementations (QueryDependencies, QueryRelatedComponents, AnalyzeCoupling); split into separate files"
	["internal/mcp/tools/create_system.go"]="Inline InputSchema is schema data not logic; migrate to schemas.go in next sprint"
	["internal/mcp/tools/create_component.go"]="Inline InputSchema + preview logic; migrate schema to schemas.go in next sprint"
	["internal/mcp/tools/update_system.go"]="Inline InputSchema is schema data not logic; migrate to schemas.go in next sprint"
)

# Parse arguments
while [[ $# -gt 0 ]]; do
	case $1 in
	--verbose)
		VERBOSE=true
		shift
		;;
	--baseline)
		BASELINE=true
		shift
		;;
	*)
		echo "Unknown option: $1"
		echo "Usage: $0 [--verbose] [--baseline]"
		exit 2
		;;
	esac
done

# Count effective lines (exclude imports, comments, blank lines)
count_effective_lines() {
	local file=$1

	# Remove:
	# - Import blocks (from 'import (' to ')')
	# - Single-line imports (import "...")
	# - Comment lines (// ...)
	# - Block comments (/* ... */)
	# - Blank lines
	# - Package declaration

	grep -v -E '^\s*$' "$file" |
		grep -v -E '^\s*package\s+' |
		grep -v -E '^\s*import\s+"' |
		grep -v -E '^\s*//' |
		awk '
      /^[[:space:]]*import[[:space:]]*\(/ { in_import=1; next }
      in_import==1 && /^[[:space:]]*\)/ { in_import=0; next }
      in_import==1 { next }
      /\/\*/ { in_block_comment=1 }
      in_block_comment==1 && /\*\// { in_block_comment=0; next }
      in_block_comment==1 { next }
      { print }
    ' |
		wc -l |
		tr -d ' '
}

# is_data_file returns true if the file is a pure-data file (schemas, constants, registries)
# that should be excluded from line-count enforcement.
is_data_file() {
	local basename
	basename=$(basename "$1")
	case "$basename" in
	schemas.go | registry.go | helpers.go | constants.go)
		return 0
		;;
	*)
		return 1
		;;
	esac
}

# Get relative path from repo root
relative_path() {
	echo "${1#"$REPO_ROOT/"}"
}

# Check a single handler file
check_handler() {
	local file=$1
	local max_lines=$2
	local handler_type=$3
	local rel_path
	rel_path=$(relative_path "$file")

	if [[ ! -f "$file" ]]; then
		return 0
	fi

	# Skip pure-data files — they contain definitions, not handler logic
	if is_data_file "$file"; then
		if [[ $VERBOSE == true ]]; then
			echo "⏭️  $rel_path: skipped (pure-data file)"
		fi
		return 0
	fi

	local count
	count=$(count_effective_lines "$file")

	if [[ $count -gt $max_lines ]]; then
		# Check if this is a known/documented violation
		if [[ -n "${KNOWN_VIOLATIONS[$rel_path]+x}" ]]; then
			local reason="${KNOWN_VIOLATIONS[$rel_path]}"
			KNOWN_VIOLATION_FAILURES=$((KNOWN_VIOLATION_FAILURES + 1))
			if [[ $BASELINE == true ]]; then
				echo "⚠️  $rel_path: $count lines (limit: $max_lines) - KNOWN VIOLATION"
				echo "   Reason: $reason"
			elif [[ $VERBOSE == true ]]; then
				echo "⚠️  $rel_path: $count lines (limit: $max_lines) - KNOWN VIOLATION"
				echo "   Reason: $reason"
			else
				echo "⚠️  $rel_path: $count lines (known violation, tracked)"
			fi
		else
			# New undocumented violation — this fails CI
			if [[ $BASELINE == true ]]; then
				echo "❌ $rel_path: $count lines (limit: $max_lines) - NEW VIOLATION"
			else
				echo "❌ $rel_path: $count lines (limit: $max_lines) - VIOLATION"
			fi
			VIOLATIONS=$((VIOLATIONS + 1))
			if [[ $VERBOSE == true ]]; then
				echo "   Handler type: $handler_type"
				echo "   Refactoring needed: Extract business logic to use cases"
			fi
		fi
	elif [[ $VERBOSE == true ]]; then
		echo "✅ $rel_path: $count lines (limit: $max_lines) - PASS"
	fi
}

echo "Constitution Audit: Handler Thin-Line Validation"
echo "=================================================="
echo ""

# Check CLI handlers (cmd/*.go, limit: 150 lines)
echo "Checking CLI handlers (cmd/*.go, limit: 150 lines)..."
for file in "$REPO_ROOT"/cmd/*.go; do
	[[ -f "$file" ]] || continue
	check_handler "$file" 150 "CLI"
done

echo ""

# Check MCP tool handlers (internal/mcp/tools/*.go, limit: 100 lines)
echo "Checking MCP tool handlers (internal/mcp/tools/*.go, limit: 100 lines)..."
for file in "$REPO_ROOT"/internal/mcp/tools/*.go; do
	[[ -f "$file" ]] || continue
	# Skip test files
	[[ "$file" == *_test.go ]] && continue
	check_handler "$file" 100 "MCP"
done

echo ""
echo "=================================================="

if [[ $KNOWN_VIOLATION_FAILURES -gt 0 ]]; then
	echo "⚠️  $KNOWN_VIOLATION_FAILURES known violation(s) tracked (see KNOWN_VIOLATIONS in script)"
fi

if [[ $VIOLATIONS -eq 0 ]]; then
	echo "✅ No new violations — constitution maintained"
	exit 0
else
	echo "❌ $VIOLATIONS NEW handler violation(s) — CI blocked"
	echo ""
	echo "Constitution principle: Thin Handlers (CLI < 150 lines, MCP < 100 lines)"
	echo "Handlers should only: parse arguments → call use case → format response"
	echo "Pure-data files (schemas.go, registry.go, helpers.go) are excluded."
	echo ""
	echo "To fix:"
	echo "  1. Extract business logic to internal/core/usecases/"
	echo "  2. Create request/response structs in internal/core/entities/"
	echo "  3. Refactor handler to thin wrapper (parse → call → format)"
	echo "  4. OR add to KNOWN_VIOLATIONS in scripts/audit-constitution.sh with justification"
	exit 1
fi
