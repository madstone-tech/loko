# Quickstart: validating rendering fidelity

```bash
task build && export PATH="$PWD:$PATH"     # task build writes ./loko
FIX=testdata/projects/serverless-reference
```

## 1. Readable layout (US1, FR-005, FR-006, SC-001)

```bash
loko build -p $FIX
head -c 400 $FIX/dist/diagrams/system-portal.svg | grep -o 'viewBox="[^"]*"'   # width/height ≤ 2
go test ./cmd/ -run TestContainerViewAspectRatio -count=1
```

## 2. Titles (US2, FR-001)

Set `title = "Read API Lambda"` on `container.read_api` (or through MCP `apply_edit`) and build.
The box shows the title with `read_api` beneath it, and so do the page heading and the markdown
heading. `loko query dependents container.table` still prints addresses.

## 3. Shapes (US3, FR-002)

Set `shape = "database"` on `container.table` and `shape = "queue"` on the two SQS containers, then
build. Those nodes are drawn as a cylinder and as queues. `shape = "database"` on `system.portal`
fails with `shape_not_allowed`.

## 4. Relationship kinds (US4, FR-003, FR-004, SC-004)

```bash
loko query dependents container.sync_queue -p $FIX > before.txt
# set kind = "trigger" on container.sync's "consume" relationship, and "async" on "fail"
loko query dependents container.sync_queue -p $FIX > after.txt
diff before.txt after.txt                      # no difference (FR-004)
go test ./internal/core/usecases/ -run 'TestLift(Trigger|Async)' -count=1
```

## 5. Unchanged projects (SC-003)

```bash
go test ./cmd/ -run TestUnchangedProjectsRenderIdentically -count=1
go test ./internal/adapters/encoding/ -count=1      # export_v1.json golden: byte-identical
```

## 6. Full gate

```bash
task lint && task test && task audit-constitution && go test -race ./...
```
