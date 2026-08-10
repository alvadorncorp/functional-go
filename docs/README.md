# functional documentation

The module root [README](../README.md) is a short **index**. Content lives here.

| Path | Purpose |
|---|---|
| [overview.md](overview.md) | Purpose, the language constraint that shapes the API, the immutability core, the zero-dependency boundary, the sibling-import rule |
| [examples.md](examples.md) | Usage sketches, including Result parsing and Validated accumulation |
| [packages/function.md](packages/function.md) | Public API of the standard-library-only `github.com/alvadorncorp/gopackages/functional/function` leaf |
| [packages/predicate.md](packages/predicate.md) | Public API of the standard-library-only `github.com/alvadorncorp/gopackages/functional/predicate` leaf |
| [packages/optional.md](packages/optional.md) | Public API of `github.com/alvadorncorp/gopackages/functional/optional`<br>The established `functional/optional` API remains supported and unchanged.<br>The smaller baseline API at `functional/values/optional` is an independent type; this plan adds no alias, conversion, or migration between them. |
| [packages/persistent.md](packages/persistent.md) | Public API of `github.com/alvadorncorp/gopackages/functional/persistent`, including measured Append-loop vs. Builder benchmarks |
| [packages/state.md](packages/state.md) | Public API of `github.com/alvadorncorp/gopackages/functional/state` (pure `State[T]`, lineage, Tx, Diff) |
| [packages/valueobject.md](packages/valueobject.md) | Public API of `github.com/alvadorncorp/gopackages/functional/valueobject` |
| [packages/core.md](packages/core.md) | Maintainer guide for `internal/core` (ownership, cow, sharing, transient) |
| [packages/values.md](packages/values.md) | Family guide: choice by semantics, zero values, dependency boundary, and cost model |
| [packages/values-optional.md](packages/values-optional.md) | Public API of `github.com/alvadorncorp/gopackages/functional/values/optional` |
| [packages/values-result.md](packages/values-result.md) | Public API of `github.com/alvadorncorp/gopackages/functional/values/result` |
| [packages/values-either.md](packages/values-either.md) | Public API of `github.com/alvadorncorp/gopackages/functional/values/either` |
| [packages/values-nonempty.md](packages/values-nonempty.md) | Public API of `github.com/alvadorncorp/gopackages/functional/values/nonempty` |
| [packages/values-validated.md](packages/values-validated.md) | Public API of `github.com/alvadorncorp/gopackages/functional/values/validated` |
| [patterns/value-objects.md](patterns/value-objects.md) | What a value object needs from this module vs. what plain Go already covers |
| [adr/functional/](adr/functional/) | Architecture decisions (`adr:functional:NNN`) |

ADRs use AKB YAML frontmatter (`type`, `id`, `context`, `title`, `status`,
`tags`, …). Layout mirrors monorepo AKB conventions **inside this module**
(`docs/adr/<context>/…`).

All guides are written in **English**.

Current decisions:

- [ADR-001 — Optional representation and zero value](adr/functional/adr-001-optional-representation-and-zero-value.md)
- [ADR-003 — Zero-dependency boundary](adr/functional/adr-003-zero-dependency-boundary.md)
- [ADR-004 — Thin immutability core](adr/functional/adr-004-thin-immutability-core.md)
- [ADR-005 — Persistent engine family (HAMT + vector)](adr/functional/adr-005-persistent-engine-family.md)
- [ADR-006 — State semantic layer](adr/functional/adr-006-state-semantic-layer.md)
- [ADR-007 — Values family dependency boundary](adr/functional/adr-007-values-family-dependency-boundary.md)
