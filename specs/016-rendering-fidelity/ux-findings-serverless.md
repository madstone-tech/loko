# UX test: modelling an existing serverless system with loko

The input was a deployed serverless AWS system: about 90 Terraform files, a Go service with several
Lambda entry points, and 9 hand-drawn D2 diagrams. The diagrams covered context, containers,
components, two dynamic flows, deployment and a data-model key layout. The model was built through
`loko mcp` alone, with 122 edits in 2 batches: 2 people, 2 external systems, 1 system, 14
containers, 7 components, 39 relationships, and dev and prod environments with 28 Terraform-bound
instances.

It compiles with 0 errors. `query` answers questions the drawings cannot, such as "what touches the
table" and "what breaks if the upstream source is down". An anonymized copy is the fixture
[`testdata/projects/serverless-reference`](../../testdata/projects/serverless-reference).

## What loko produced

- A landscape, the system's containers, the read API's components, and a deployment diagram per
  environment. Each comes as D2 and SVG, plus markdown and a site with a page per element.
- Queries (dependents, dependencies, path, orphans, coupling) and a JSON or TOON export.
- Terraform bindings for every instance, ready for the planned reconcile stage.

## Gaps, ranked by how much they hurt "visualize an existing system"

| # | Gap | Evidence | Suggested fix |
|---|---|---|---|
| G1 | **No import.** An assistant had to read the Terraform and D2 by hand and type all 28 binding addresses. | loko has no Terraform or D2 reader; `import` and `reconcile` are roadmap stage 5. | A `loko import terraform` that reads `module.*` resources into proposed elements and bindings, which an assistant then curates. Its own feature. |
| G2 | **Every container looks the same.** DynamoDB, SQS, SNS and Lambda all render as identical rectangles. The hand-drawn diagram uses cylinders and queue shapes. | Shape is chosen by element kind only (`viewmodel/style.go`). | A `shape` attribute (016). |
| G3 | **Labels show raw identifiers.** Boxes read `read_api`, `portal`, `sync_dlq`. The hand-drawn diagram says "Read API Lambda". | Elements have only their label-name; there is no display title. | A `title` attribute (016). |
| G4 | **Layout is too wide to read.** Left-to-right dagre gives a 1600×370 strip with unreadable text. The hand-drawn diagram is top-down. | The dagre direction is fixed. | A per-view `direction`, top-down by default for container views (016). |
| G5 | **(Fixed in 015, T062.)** `_5f` escaping leaked into file names and URLs, e.g. `container-read_5fapi.svg`. | — | — |
| G6 | **No async or trigger semantics.** `uses` means "depends on", so triggers (queue → Lambda, schedule → function) had to be drawn against the data flow, and async edges couldn't be dashed. | Found here and in the food-delivery session. | A relationship `kind` (016). |
| G7 | **No dynamic or sequence views.** The two flow diagrams have no equivalent. | Only static views exist. | A later feature: an ordered `step` list on a view. |
| G8 | **Repeated environments.** dev and prod needed the same 14 instances typed twice. | There is no way to reuse an instance set. | A later feature: a batch helper, or locals. |
| G9 | **Relationships carry only description and technology.** Read-only versus write access, the hand-drawn diagram's main point, is just text. | — | Relationship tags, shown in renders (016). |
| G10 | **The prose already exists but isn't connected.** 27 `missing_docs` warnings, while the system has rich written documentation. | — | Nothing to build; set `docs` on elements. |

Out of scope: the data-model key-layout diagram describes storage, not architecture.
