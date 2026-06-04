#!/usr/bin/env bash
# benchmark-archcheck.sh — record archcheck wall-clock time over 5 runs.
# Fails if median > HARD_LIMIT_S (default 30s, per spec SC-009).
# Warns if median > SOFT_LIMIT_S (default 10s, per research.md R6).
#
# Output: writes a Markdown table to stdout suitable for pasting into
# specs/010-constitution-compliance/quickstart-validation.md.

set -euo pipefail

HARD_LIMIT_S="${1:-30}"
SOFT_LIMIT_S="${2:-10}"
RUNS=5

# Pre-build the binary once so we measure run-time, not compile-time.
bin=$(mktemp -t archcheck.XXXXXX)
trap 'rm -f "$bin"' EXIT
go build -o "$bin" ./tools/archcheck

declare -a samples=()
for ((i = 1; i <= RUNS; i++)); do
  t0=$(python3 -c 'import time; print(time.monotonic())')
  "$bin" --format=json > /dev/null 2>&1 || true
  t1=$(python3 -c 'import time; print(time.monotonic())')
  delta=$(awk -v a="$t0" -v b="$t1" 'BEGIN { printf "%.3f", b - a }')
  samples+=("$delta")
done

# Sort samples to compute min, median, max.
IFS=$'\n' sorted=($(printf "%s\n" "${samples[@]}" | sort -g))
unset IFS
min="${sorted[0]}"
max="${sorted[$((RUNS - 1))]}"
median="${sorted[$((RUNS / 2))]}"

# Output table.
echo "| Run | Wall-clock (s) |"
echo "|-----|----------------|"
for ((i = 0; i < RUNS; i++)); do
  echo "| $((i + 1))   | ${samples[$i]}          |"
done
echo
echo "**Min**: ${min}s  **Median**: ${median}s  **Max**: ${max}s"
echo "**Hard limit (SC-009)**: ${HARD_LIMIT_S}s  **Soft target (R6)**: ${SOFT_LIMIT_S}s"

# Check against limits.
if awk -v m="$median" -v h="$HARD_LIMIT_S" 'BEGIN { exit !(m > h) }'; then
  echo ""
  echo "FAIL: median ${median}s exceeds hard limit ${HARD_LIMIT_S}s (SC-009)." >&2
  exit 1
fi
if awk -v m="$median" -v s="$SOFT_LIMIT_S" 'BEGIN { exit !(m > s) }'; then
  echo ""
  echo "WARN: median ${median}s exceeds soft target ${SOFT_LIMIT_S}s (R6) but within hard limit." >&2
fi
