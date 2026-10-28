# Contract amendment: SVG node addresses

This is an additive amendment to `specs/020-diagram-look/contracts/svg-output.md`. All of 020's
guarantees stand.

## New

1. Every element and instance node group in `diagrams/<view>.svg` carries
   `data-loko-address="<address>"`, where `<address>` is the node's model address
   (`container.api`, `component.gateway.router`, or an instance address). It is written in the
   group's start tag directly after the `class` attribute, and is XML-escaped.
2. Boundary groups, the outside node and edge groups carry no `data-loko-address`.
3. Each address appears at most once per SVG.

## Selectors added

| Selector | Matches |
|---|---|
| `g[data-loko-address]` | every element or instance card |
| `g[data-loko-address="<address>"]` | the card for one element |

## Unchanged

- The `.d2` source has no attribute; only loko's SVG pass adds it.
- The output is byte-identical for the same input. Existing 020 goldens change only by this
  attribute.
