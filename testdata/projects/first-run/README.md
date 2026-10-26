# first-run

The shape of the model in [the first-run report](../../../specs/017-mcp-validation-ergonomics/first-run-report.md):
one system with 12 containers, 3 people, 7 third-party systems declared as `system` (as the
reporter did), 8 deployment instances with no binding, and 22 relationships, split across
`architecture/*.loko.hcl`.

Feature 017 measures SC-002 to SC-004 against it. `../first-run-remodelled` is the same model with
the third-party systems as `external` and the unbuilt instances marked `planned = true`.

## Warnings

| Code | 1.0.2 (`first-run`) | 017 (`first-run-remodelled`) |
|---|---|---|
| `empty_system` | 7 | 0 |
| `missing_docs` | 16 (7 third-party systems, 9 containers) | 9 (the containers) |
| `orphan_element` | 1 (`container.audit_log`) | 1 |
| `unbound_instance` | 8 | 0 |
| **Total** | **32** | **10** |

The 10 genuine warnings are the 9 containers without docs and the orphan: exactly what remains
after remodelling (SC-002, `TestRemodelledWarnings`).
