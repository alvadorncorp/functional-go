# functional documentation

The module root [README](../README.md) is a short **index**. Content lives here.

| Path | Purpose |
|---|---|
| [overview.md](overview.md) | Purpose, the language constraint that shapes the API, the immutability core, the zero-dependency boundary, the sibling-import rule |
| [examples.md](examples.md) | Usage sketches, including Result parsing and Validated accumulation |
| [packages/function.md](packages/function.md) | `primitives/function` — stdlib-only function combinators |
| [packages/predicate.md](packages/predicate.md) | `primitives/predicate` — stdlib-only Boolean predicates |
| [packages/order.md](packages/order.md) | `primitives/order` — stdlib-only total orders |
| [packages/fold.md](packages/fold.md) | `primitives/fold` — slice folds; edges to `values/optional` and `values/nonempty` |
| [packages/tuple.md](packages/tuple.md) | `structural/tuple` — Tuple2/Tuple3 carriers; stdlib-only |
| [packages/lazy.md](packages/lazy.md) | `structural/lazy` — concurrent deferred memoization; stdlib-only |
| [packages/iterator.md](packages/iterator.md) | `structural/iterator` — `iter.Seq` adapters; imports only `structural/tuple` |
| [packages/optics.md](packages/optics.md) | `structural/optics` — typed Lens; stdlib-only |
| [packages/immutable.md](packages/immutable.md) | Public API of `github.com/alvadorncorp/gopackages/functional/immutable`, including measured Append-loop vs. Builder benchmarks |
| [packages/state.md](packages/state.md) | Public API of `github.com/alvadorncorp/gopackages/functional/state` (pure `State[T]`, lineage, Tx, Diff) |
| [packages/core.md](packages/core.md) | Maintainer guide for `internal/core` (ownership, cow, sharing, transient) |
| [packages/values.md](packages/values.md) | Family guide: choice by semantics, zero values, dependency boundary, and cost model |
| [packages/values-optional.md](packages/values-optional.md) | Public API of `github.com/alvadorncorp/gopackages/functional/values/optional` |
| [packages/values-result.md](packages/values-result.md) | Public API of `github.com/alvadorncorp/gopackages/functional/values/result` |
| [packages/values-either.md](packages/values-either.md) | Public API of `github.com/alvadorncorp/gopackages/functional/values/either` |
| [packages/values-nonempty.md](packages/values-nonempty.md) | Public API of `github.com/alvadorncorp/gopackages/functional/values/nonempty` |
| [packages/values-validated.md](packages/values-validated.md) | Public API of `github.com/alvadorncorp/gopackages/functional/values/validated` |
| [patterns/value-objects.md](patterns/value-objects.md) | What a value object needs from this module vs. what plain Go already covers |
| [patterns/cow-handles.md](patterns/cow-handles.md) | Small value handles over private payloads when hugeParam bites |
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
- [ADR-008 — Functional primitives and fold dependency boundary](adr/functional/adr-008-functional-primitives-and-fold-boundary.md)
- [ADR-009 — Tuplas, lazy, iteradores e optics sobre contratos nativos de Go](adr/functional/adr-009-structural-primitives-and-iterator-boundary.md)
- [ADR-010 — Retire top-level optional and valueobject](adr/functional/adr-010-retire-optional-and-valueobject.md)
- [ADR-011 — primitives and structural path groups](adr/functional/adr-011-primitives-and-structural-path-groups.md)
