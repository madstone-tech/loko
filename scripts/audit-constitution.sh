#!/usr/bin/env bash
#
# audit-constitution.sh - Validate handler files follow Clean Architecture constitution
#
# Constitution Rules:
# - CLI handlers (cmd/*.go): < 50 lines (excluding imports, comments, blank lines)
# - MCP tools (internal/mcp/tools/*.go): < 30 lines (excluding imports, comments, blank lines)
# - Handlers should only: parse → call use case → format response
#
# Usage:
#   ./scripts/audit-constitution.sh [--verbose] [--baseline]
#
# Exit codes:
#   0 - All handlers pass
#   1 - One or more handlers violate constitution
#   2 - Script error
#

set -euo pipefail

VERBOSE=false
BASELINE=false
VIOLATIONS=0
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

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

# Check a single handler file
check_handler() {
	local file=$1
	local max_lines=$2
	local handler_type=$3

	if [[ ! -f "$file" ]]; then
		return 0
	fi

	local count
	count=$(count_effective_lines "$file")

	if [[ $count -gt $max_lines ]]; then
		if [[ $BASELINE == true ]]; then
			echo "- $file: $count lines (limit: $max_lines) - DOCUMENTED VIOLATION"
		else
			echo "❌ $file: $count lines (limit: $max_lines) - VIOLATION"
			VIOLATIONS=$((VIOLATIONS + 1))
		fi
		if [[ $VERBOSE == true ]]; then
			echo "   Handler type: $handler_type"
			echo "   Refactoring needed: Extract business logic to use cases"
		fi
	elif [[ $VERBOSE == true ]]; then
		echo "✅ $file: $count lines (limit: $max_lines) - PASS"
	fi
}

echo "Constitution Audit: Handler Thin-Line Validation"
echo "=================================================="
echo ""

# Check CLI handlers (cmd/*.go, limit: 50 lines)
echo "Checking CLI handlers (cmd/*.go, limit: 50 lines)..."
for file in "$REPO_ROOT"/cmd/*.go; do
	[[ -f "$file" ]] || continue
	check_handler "$file" 50 "CLI"
done

echo ""

# Check MCP tool handlers (internal/mcp/tools/*.go, limit: 30 lines)
echo "Checking MCP tool handlers (internal/mcp/tools/*.go, limit: 30 lines)..."
for file in "$REPO_ROOT"/internal/mcp/tools/*.go; do
	[[ -f "$file" ]] || continue
	# Skip test files
	[[ "$file" == *_test.go ]] && continue
	check_handler "$file" 30 "MCP"
done

echo ""
echo "=================================================="

if [[ $BASELINE == true ]]; then
	echo "Baseline mode: All violations documented for tracking"
	echo "Run without --baseline to enforce constitution in CI"
	exit 0
elif [[ $VIOLATIONS -eq 0 ]]; then
	echo "✅ All handlers pass constitution validation"
	exit 0
else
	echo "❌ $VIOLATIONS handler(s) violate constitution"
	echo ""
	echo "Constitution principle: Thin Handlers (CLI < 50 lines, MCP < 30 lines)"
	echo "Handlers should only: parse arguments → call use case → format response"
	echo ""
	echo "To fix:"
	echo "  1. Extract business logic to internal/core/usecases/"
	echo "  2. Create request/response structs in internal/core/entities/"
	echo "  3. Refactor handler to thin wrapper (parse → call → format)"
	echo ""
	echo "See docs/guides/handler-refactoring-guide.md for patterns"
	exit 1
fi
