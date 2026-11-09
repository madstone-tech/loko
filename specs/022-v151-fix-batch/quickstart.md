# Quickstart: validating the v1.5.1 fix batch

**Branch**: `022-v151-fix-batch` | **Date**: 2026-11-09

Prerequisites: `go test` runs clean (`task test`), a built `loko` binary
(`go build .` or `task build`), and the repo's own example project.

## Scenario 1: Output carries no manifest (spec story 1, SC-001)

```sh
rm -rf /tmp/loko-151 && mkdir -p /tmp/loko-151
cp examples/simple-project/main.loko.hcl /tmp/loko-151/
loko build -p /tmp/loko-151
ls -a /tmp/loko-151/dist        # → no .loko-manifest
ls -a /tmp/loko-151             # → .dist.loko present, holding manifest
```

Expected: `dist/` lists only generated artifacts; `.dist.loko/manifest` exists and
its first line is the generated-by notice. `README.md` of the example (or any
published copy) stays clean when copied.

## Scenario 2: Upgrade path keeps pruning (spec story 1, scenario 2, SC-002)

```sh
# simulate a 1.5.0 checkout: manifest inside dist/
mv /tmp/loko-151/.dist.loko/manifest /tmp/loko-151/dist/.loko-manifest
rm -rf /tmp/loko-151/.dist.loko
# remove an artifact's source to force pruning … (edit main.loko.hcl, or:)
loko build -p /tmp/loko-151
ls /tmp/loko-151/dist/.loko-manifest 2>&1   # → No such file (retired)
ls -a /tmp/loko-151                          # → .dist.loko recreated
```

Expected: on the first 1.5.1 build the legacy manifest is read, pruning runs from
it, the new manifest is written, and the legacy file is removed. A second build
with a removed element's sources reports `1 removed` (the element's pages go).

## Scenario 3: Failed build keeps the legacy manifest (spec story 1, edge case)

```sh
printf 'project "x" {\n  loko_version = "~> 9.9"\n}\n' > /tmp/loko-151/bad.loko.hcl
loko build -p /tmp/loko-151   # → fails
ls -a /tmp/loko-151/dist | grep loko-manifest   # → present (untouched)
rm /tmp/loko-151/bad.loko.hcl
loko build -p /tmp/loko-151   # → succeeds, retirement resumes
```

## Scenario 4: Help names the language reference (spec story 2, SC-003)

```sh
loko --help | grep -i "docs"
```

Expected: the pointer paragraph appears once; each subcommand help
(`loko build --help`) has no copy of it.

## Scenario 5: The dynamic hint (spec story 3, SC-004)

```sh
mkdir -p /tmp/loko-dyn && printf 'project "d" {\n  loko_version = "~> 1.0"\n}\ndynamic "x" {}\n' > /tmp/loko-dyn/arch.loko.hcl
loko validate -p /tmp/loko-dyn
rm -rf /tmp/loko-dyn
```

Expected: error text ends with the hint (`… excluded from v1.0. For a C4 dynamic
view, use a view block with step blocks.`), same `CodeUnknownBlock` code and
location as 1.5.0.

## Scenario 6: The kind note (spec story 4, SC-005)

```sh
loko export --format json   # or the schema tool output in an MCP session
```

Expected: the relationship `kind` attribute's note states the default direction
and the trigger reversal. Edge direction in rendered diagrams is unchanged
(scenario 2 of the feedback probe: `kind = "trigger"` still swaps the drawn edge).

## Scenario 7: The theming class contract (spec story 5, SC-006)

```sh
$BROWSER docs/guides/site-theming.md   # or: read the section
```

Expected: a "Stable classes, variables, and blocks" section listing the stable
classes and custom properties with the stability commitment.

## Full gates (constitution, before commit)

```sh
task test          # all packages green; store tests rewritten for the new location
task lint          # depguard + vet clean
go test ./cmd/ ./internal/adapters/{d2,hclsource,html,markdown}/ -update   # goldens, if any differ
git diff testdata/ # review golden diffs before committing them
```
