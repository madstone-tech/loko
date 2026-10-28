# Quickstart: Validating Diagram Look

Run from the repository root on branch `020-diagram-look`.

## 1. Tests

```sh
make test     # unit + golden tests, including the new icon, label-plan and contrast tests
make lint
make audit-constitution
```

Expected: all green. The D2 goldens in `internal/adapters/d2/testdata/golden/` change in this
feature (rectangle cards, new colours, kind line text) and are reviewed in the diff.

## 2. Before and after on the example

```sh
git stash -u && go build -o /tmp/loko-before . && git stash pop
go build -o /tmp/loko-after .
(cd examples/microservices && /tmp/loko-before build --format svg --out /tmp/before \
  && /tmp/loko-after build --format svg --out /tmp/after)
```

Open `/tmp/after/diagrams/system-order_service.svg` and `container-gateway.svg` in a browser.

Expected (User Stories 1 and 3):
- Every element is a rounded card with a small white icon top-left (the component cards' icon is
  dark); the order DB shows the database icon.
- Each card reads: bold name, small `CONTAINER · Go, gRPC`-style line, description in the regular
  (not bold) face.
- Titled elements end their kind line with their name (`SYSTEM · Kong · api_gateway`).
- Edges and edge labels are grey; technology on edges is a small line.
- Nodes outside the view keep their grey dashed ovals.

## 3. Size guard (SC-004)

```sh
for f in /tmp/before/diagrams/*.svg; do
  n=$(basename "$f"); b=$(rg -o 'viewBox="[^"]*"' "$f"); a=$(rg -o 'viewBox="[^"]*"' "/tmp/after/diagrams/$n")
  echo "$n  $b  ->  $a"
done
```

Expected: no view's width or height grows more than 10% (icons do not change d2's sizing, so they
should match unless label text changed length).

## 4. Outside loko (User Story 2)

```sh
rsvg-convert -w 1600 /tmp/after/diagrams/container-gateway.svg -o /tmp/gw.png
```

Expected: the PNG shows icons, colours and the text hierarchy. Open the SVG in a browser with the
network disabled: nothing is requested. Text on the cards is selectable.

## 5. Sequence and deployment views

Build `testdata/projects/dynamic-views` with `/tmp/loko-after` and open the SVG of view
`checkout_sequence`. Open `/tmp/after/microservices/diagrams/deployment-prod.svg`.

Expected: sequence participants show the corner icon clear of the name and the lifeline (FR-018);
deployment environments and nodes are grey dashed boundaries, instances are cards with icons.

## 6. D2 CLI (FR-016)

```sh
(cd examples/microservices && /tmp/loko-after build --format d2 --out /tmp/d2out)
d2 --version   # v0.9.0 was checked during planning; record the version used
d2 /tmp/d2out/diagrams/container-gateway.d2 /tmp/cli.svg
```

Expected: compiles; rounded plain cards in the new palette with the new label text, `shape-*`
classes on shaped nodes, no icons.

## 7. Budgets

```sh
go test ./cmd -run 'TestPerformanceBuild|TestPerformanceWideView' -v
go test ./internal/adapters/d2 -run 'TestSVGArtifactSize' -v
```

Expected: the 1,020-element build stays under 10 s; SVG size grows by no more than 20% (SC-005).
