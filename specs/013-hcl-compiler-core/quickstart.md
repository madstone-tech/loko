# Quickstart: validating the HCL compiler core

**Feature**: `013-hcl-compiler-core` | **Date**: 2026-09-22

Runnable scenarios that prove this feature works end to end. Each maps to acceptance scenarios in
[spec.md](./spec.md). Run them from a scratch directory outside the repository.

> `loko init` does not exist in this stage (research R10) — the starter file below is written by
> hand, which is also the documentation for what a minimal project looks like.

---

## Prerequisites

```bash
cd /path/to/loko
task build          # or: make build
export PATH="$PWD/bin:$PATH"
loko version
```

---

## Scenario 1 — Author and check (User Story 1)

```bash
mkdir -p /tmp/acme/docs && cd /tmp/acme
cat > arch.loko.hcl <<'EOF'
project "acme-payments" {
  description  = "Payment processing platform"
  loko_version = "~> 1.0"
}

system "payments" {
  description = "Authorization, capture, settlement"
  owner       = "platform-team"
  docs        = "./docs/payments.md"
}

container "api" {
  system     = system.payments
  technology = "AWS Lambda (Go)"
  tags       = ["public", "pci"]

  uses "orders" {
    target      = container.orders_db
    description = "Reads and writes orders"
  }
}

container "orders_db" {
  system     = system.payments
  technology = "Aurora PostgreSQL"
  tags       = ["pci"]
}
EOF
echo "# Payments" > docs/payments.md

loko validate; echo "exit=$?"
```

**Expected**: no errors, `exit=0`. One `missing_docs` warning for each container (neither declares
`docs`), which does not affect the exit code.

```bash
loko validate --strict; echo "exit=$?"
```

**Expected**: same warnings, `exit=2` (FR-034, FR-038).

### Break a reference (Scenario 1.2)

```bash
sed -i.bak 's/container.orders_db/container.ordrs_db/' arch.loko.hcl
loko validate; echo "exit=$?"
```

**Expected**: `exit=1`, and an `unresolved_reference` error naming `arch.loko.hcl`, the line, and the
column of the traversal — not of the block (FR-021). Restore with
`mv arch.loko.hcl.bak arch.loko.hcl`.

### Every error in one run (Scenario 1 / SC-008)

Introduce five independent mistakes — a bad reference, an unknown attribute, a component parented to
a system, an unknown function, an unknown block — then:

```bash
loko validate --format json | jq '.summary'
```

**Expected**: `errors: 5` from a single invocation, not one error and an abort (FR-031).

### Dependency cycles are legal (Scenario 1.5)

Add `uses` blocks forming `api → queue → worker → api`.

**Expected**: `exit=0`. Only containment is required to be acyclic (FR-032). If this reports a cycle,
the containment check has been wrongly applied to relationships — the single most likely design error
in this feature.

### Multi-file merge (Scenario 1.6)

```bash
mkdir -p systems
mv arch.loko.hcl systems/payments.loko.hcl
cat > gateway.loko.hcl <<'EOF'
container "gateway" {
  system = system.payments
  uses "api" { target = container.api }
}
EOF
loko validate; echo "exit=$?"
```

**Expected**: `exit=0`. A file in a subdirectory resolves a reference declared in a sibling directory
(FR-001, FR-019).

---

## Scenario 2 — Deployment plane (User Story 2)

```bash
cat > deploy.loko.hcl <<'EOF'
deployment "prod" {
  provider = "aws"
  account  = "123456789012"
  region   = "us-east-1"

  node "vpc-main" {
    node "subnet-a" {
      instance "api" {
        of         = container.api
        attributes = { memory = 1024, timeout = 30 }
        binding "terraform" {
          address = "module.api.aws_lambda_function.this"
        }
      }
    }
  }
}
EOF
loko validate && loko export --format json | jq '.environments[0].instances[0] | {address, placedIn}'
```

**Expected**:

```json
{ "address": "deployment.prod.instance.api", "placedIn": "deployment.prod.node.vpc-main.subnet-a" }
```

The address carries **no** node segment (FR-012, FR-024).

### Identity survives re-parenting (Scenario 2.6)

Move the `instance "api"` block from `subnet-a` up to `vpc-main`, re-export, and diff the two
exports.

**Expected**: `address` unchanged; only `placedIn` and the `groups` tree differ. This is the property
the diff stage depends on — if the address moves, every re-organisation will later read as a delete
plus a create.

### Duplicate instance name (Scenario 2.7)

Add a second `instance "api"` under a different node.

**Expected**: `duplicate_instance_name` error naming both source locations (FR-012a).

---

## Scenario 3 — Determinism (User Story 3, SC-003)

```bash
loko export --format json > a.json
loko export --format json > b.json
cmp a.json b.json && echo "byte-identical"

loko export --format toon > a.toon
loko export --format toon > b.toon
cmp a.toon b.toon && echo "byte-identical"
```

**Expected**: both `cmp` calls silent (FR-040).

```bash
touch systems/zzz.loko.hcl systems/aaa.loko.hcl   # perturb discovery order
loko export --format json > c.json
cmp a.json c.json && echo "stable across file ordering"
```

**Expected**: identical. A failure here almost always means a Go map reached the IR (research R6).

```bash
grep -E 'timestamp|hostname|/Users/|/home/|"version"' a.json && echo "LEAK" || echo "clean"
loko export --format json | jq '.schemaVersion'
```

**Expected**: `clean`, then `1` (FR-036a, FR-036c).

### No artefact on error (Scenario 3.5)

```bash
echo 'container "broken" { system = system.nope }' > bad.loko.hcl
loko export --format json --out out.json; echo "exit=$?"
ls out.json 2>/dev/null || echo "no artefact written"
rm bad.loko.hcl
```

**Expected**: `exit=1`, diagnostics on stderr, no `out.json` (FR-037).

---

## Scenario 4 — Formatting (User Story 4)

```bash
printf 'system    "billing"   {\n# keep me\n      description="Billing"\n}\n' > messy.loko.hcl
loko fmt
cat messy.loko.hcl
loko fmt && echo "second run: no output = idempotent"
```

**Expected**: canonical indentation, `# keep me` preserved in place, declaration order unchanged
(FR-035, SC-007).

```bash
loko fmt --check; echo "exit=$?"        # 0 — everything canonical
printf 'system  "x"  {}\n' > messy2.loko.hcl
loko fmt --check; echo "exit=$?"        # 1, lists messy2.loko.hcl, writes nothing
cat messy2.loko.hcl                     # unchanged
```

**Expected**: `exit=0` then `exit=1` with the path listed and the file untouched (FR-035a).

### Unparseable files are not touched

```bash
echo 'system "oops" {' > broken.loko.hcl
loko fmt; echo "exit=$?"
cat broken.loko.hcl
```

**Expected**: a syntax error with a location, `exit=1`, and the file byte-unchanged (FR-035).

---

## Scenario 5 — Version constraint (User Story 5)

```bash
sed -i.bak 's/~> 1.0/>= 99.0/' systems/payments.loko.hcl
loko validate; echo "exit=$?"
mv systems/payments.loko.hcl.bak systems/payments.loko.hcl
```

**Expected**: `version_unsatisfied` error naming both the constraint and the running version,
`exit=1`, no artefact (FR-028).

---

## Scenario 6 — Retirement of the v0 model (User Story 6)

From the repository:

```bash
task test && task lint && task audit-constitution
loko --help
```

**Expected**: all three green; `--help` lists `validate`, `fmt`, `export`, `version`, `completion`,
`mcp` and **no** `build`, `serve`, `watch`, `new`, `init`, or `api` (research R9, FR-043).

```bash
grep -rn 'check-drift\|detect_drift\|DriftIssue' --include='*.go' . | grep -v '^./specs' || echo "no drift machinery"
test -f internal/adapters/d2/parser.go && echo "diagram parser parked as required (FR-045)"
go build ./... && echo "build green"
```

**Expected**: no drift machinery; `parser.go` still present; build green.

### Layer rule

```bash
cat > /tmp/leak_test.go <<'EOF'
package entities
import _ "github.com/hashicorp/hcl/v2"
EOF
cp /tmp/leak_test.go internal/core/entities/leak.go
task audit-constitution; echo "exit=$?"
rm internal/core/entities/leak.go
```

**Expected**: non-zero exit naming the forbidden import (FR-044). If this passes, the new rule is not
actually wired and the guarantee is decorative.

---

## Scenario 7 — Performance (SC-006)

```bash
# generate a synthetic project: 1000 elements across 200 files
go run ./tools/genfixture -elements 1000 -files 200 -out /tmp/big
time loko validate --path /tmp/big
```

**Expected**: under 2 seconds. Repeat with `-elements 5000` for the 10-second bound.

---

## Coverage of the rule list (SC-004)

```bash
go test ./internal/core/... -run TestValidationRuleCoverage -v
```

**Expected**: the test enumerates every code in
[diagnostics.schema.json](./contracts/diagnostics.schema.json) and fails if any lacks a fixture. This
is the mechanical check behind SC-004's "100% of the rule list"; without it, coverage is a claim
rather than a fact.
