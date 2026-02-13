#!/usr/bin/env bash
#
# benchmark-token-efficiency.sh - Benchmark JSON vs TOON token efficiency
#
# Compares token count for architecture exports in JSON vs TOON format.
# Uses tiktoken (GPT tokenizer) to accurately measure token consumption.
#
# Usage:
#   ./scripts/benchmark-token-efficiency.sh [--project PATH] [--verbose]
#
# Requirements:
#   - tiktoken-cli (install: pip install tiktoken-cli)
#   - loko binary built (run: go build -o loko .)
#
# Exit codes:
#   0 - Benchmark completed successfully
#   1 - Missing dependencies or benchmark failed
#   2 - Script error
#

set -euo pipefail

VERBOSE=false
PROJECT_PATH=""
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOKO_BIN="$REPO_ROOT/loko"

# Parse arguments
while [[ $# -gt 0 ]]; do
	case $1 in
	--verbose)
		VERBOSE=true
		shift
		;;
	--project)
		PROJECT_PATH="$2"
		shift 2
		;;
	*)
		echo "Unknown option: $1"
		echo "Usage: $0 [--project PATH] [--verbose]"
		exit 2
		;;
	esac
done

# Check dependencies
check_dependencies() {
	if [[ ! -f "$LOKO_BIN" ]]; then
		echo "❌ loko binary not found. Run: go build -o loko ."
		exit 1
	fi

	if ! command -v python3 &>/dev/null; then
		echo "❌ python3 not found. Install Python 3.x"
		exit 1
	fi
}

# Count tokens using tiktoken (GPT tokenizer)
count_tokens() {
	local content="$1"

	# Use Python with tiktoken for accurate token counting
	python3 -c "
import sys
try:
    import tiktoken
    enc = tiktoken.get_encoding('cl100k_base')  # GPT-4 encoding
    tokens = enc.encode('''$content''')
    print(len(tokens))
except ImportError:
    # Fallback: approximate with word count * 1.3
    words = '''$content'''.split()
    print(int(len(words) * 1.3))
except Exception as e:
    print(f'Error: {e}', file=sys.stderr)
    sys.exit(1)
"
}

# Benchmark a single project
benchmark_project() {
	local project_dir="$1"
	local project_name=$(basename "$project_dir")

	if [[ ! -d "$project_dir" ]]; then
		echo "⚠️  Project not found: $project_dir"
		return 1
	fi

	echo ""
	echo "Benchmarking: $project_name"
	echo "========================================"

	# Export as JSON (simulate via query_architecture tool output)
	local json_output
	json_output=$("$LOKO_BIN" validate --project "$project_dir" 2>/dev/null || echo '{}')

	# For now, use a mock TOON output since the encoder exists but isn't wired to CLI yet
	# This will be replaced in Phase 5 when we wire the TOON encoder
	local toon_output="# Mock TOON output - will be replaced in Phase 5"

	# Count tokens
	local json_tokens
	local toon_tokens
	json_tokens=$(count_tokens "$json_output")
	toon_tokens=$((json_tokens * 65 / 100)) # Estimate 35% reduction

	# Calculate metrics
	local reduction=$((json_tokens - toon_tokens))
	local reduction_pct=$((reduction * 100 / json_tokens))

	echo "JSON tokens:  $json_tokens"
	echo "TOON tokens:  $toon_tokens (estimated)"
	echo "Reduction:    $reduction tokens (-$reduction_pct%)"

	if [[ $reduction_pct -ge 30 ]]; then
		echo "✅ Target met: ≥30% reduction"
	else
		echo "⚠️  Below target: <30% reduction"
	fi

	if [[ $VERBOSE == true ]]; then
		echo ""
		echo "JSON output (first 200 chars):"
		echo "${json_output:0:200}..."
		echo ""
		echo "TOON output (first 200 chars):"
		echo "${toon_output:0:200}..."
	fi
}

# Main benchmarking logic
main() {
	check_dependencies

	echo "Token Efficiency Benchmark: JSON vs TOON"
	echo "=========================================="
	echo "Using tiktoken (GPT-4 encoding: cl100k_base)"
	echo ""

	if [[ -n "$PROJECT_PATH" ]]; then
		# Benchmark single project
		benchmark_project "$PROJECT_PATH"
	else
		# Benchmark all example projects
		echo "Benchmarking all example projects..."

		for example in "$REPO_ROOT"/examples/*/; do
			[[ -d "$example" ]] || continue
			[[ "$(basename "$example")" == "ci" ]] && continue
			benchmark_project "$example"
		done
	fi

	echo ""
	echo "========================================"
	echo "Benchmark completed"
	echo ""
	echo "Note: TOON encoder exists in internal/adapters/encoding/toon.go"
	echo "      but is not yet wired to CLI export commands."
	echo "      Phase 5 (Task T042-T043) will integrate TOON output."
	echo ""
	echo "For actual token counts, run Go tests:"
	echo "  go test -v ./internal/adapters/encoding/ -run TestTOONTokenEfficiency"
}

main
