#!/usr/bin/env bash
# check-rules-sync.sh — guard against drift between the canonical rules YAML
# (tools/archcheck/rules.yaml) and the prose constitution
# (.specify/memory/constitution.md).
#
# This script is intentionally minimal: it pins three invariants that the two
# sources MUST agree on. A full structural diff would be ideal but is fragile
# against markdown formatting changes. The three invariants below are the ones
# whose drift would silently weaken or strengthen the gate.
#
# Exit codes:
#   0 — all invariants hold
#   1 — at least one mismatch
#   2 — environment error (missing file)

set -euo pipefail

CONSTITUTION=".specify/memory/constitution.md"
RULES="tools/archcheck/rules.yaml"

if [[ ! -f "$CONSTITUTION" ]]; then
  echo "check-rules-sync: missing $CONSTITUTION" >&2
  exit 2
fi
if [[ ! -f "$RULES" ]]; then
  echo "check-rules-sync: missing $RULES" >&2
  exit 2
fi

fail=0

# ----------------------------------------------------------------------------
# Invariant 1: constitution version present in the markdown is referenced in
# the rules YAML's header comment (best-effort match on "Constitution v<x.y.z>"
# inside any description field). If the rules YAML does not mention the
# constitution version anywhere, that is a drift signal.
# ----------------------------------------------------------------------------
constitution_version=$(grep -E '^\*\*Version\*\*:' "$CONSTITUTION" | head -1 | sed -E 's/.*Version\*\*: ([0-9]+\.[0-9]+\.[0-9]+).*/\1/')
if [[ -z "$constitution_version" ]]; then
  echo "check-rules-sync: could not extract constitution version from $CONSTITUTION" >&2
  fail=1
else
  if ! grep -q "v${constitution_version}\|Constitution v${constitution_version}\|${constitution_version}" "$RULES"; then
    echo "check-rules-sync: $RULES does not reference constitution version $constitution_version" >&2
    fail=1
  fi
fi

# ----------------------------------------------------------------------------
# Invariant 2: the four named budgets must appear with the canonical limits in
# both files (50 / 30 / 200 / 300). Drift in either file would change what the
# gate enforces.
# ----------------------------------------------------------------------------
declare -A budgets=(
  ["cli-handler"]="50"
  ["mcp-tool"]="30"
  ["usecase-file"]="200"
  ["entity-file"]="300"
)

for name in "${!budgets[@]}"; do
  limit="${budgets[$name]}"
  if ! grep -q "$limit" "$CONSTITUTION"; then
    echo "check-rules-sync: $CONSTITUTION missing budget limit $limit (for $name)" >&2
    fail=1
  fi
  if ! grep -q "maxEffectiveLines: $limit" "$RULES"; then
    echo "check-rules-sync: $RULES missing maxEffectiveLines: $limit (for $name)" >&2
    fail=1
  fi
done

# ----------------------------------------------------------------------------
# Invariant 3: every layer named in the Dependency Direction table also appears
# as a layer entry in the YAML. The canonical layer names are:
#   internal/core/entities, internal/core/usecases, internal/adapters,
#   internal/mcp, internal/api, cmd
# ----------------------------------------------------------------------------
for layer_path in internal/core/entities internal/core/usecases internal/adapters internal/mcp internal/api cmd; do
  if ! grep -q "$layer_path" "$RULES"; then
    echo "check-rules-sync: $RULES does not reference layer $layer_path" >&2
    fail=1
  fi
  if ! grep -q "$layer_path" "$CONSTITUTION"; then
    echo "check-rules-sync: $CONSTITUTION does not reference layer $layer_path" >&2
    fail=1
  fi
done

if [[ "$fail" -eq 0 ]]; then
  echo "check-rules-sync: OK — constitution v${constitution_version} and tools/archcheck/rules.yaml are in sync on all checked invariants."
fi

exit "$fail"
