# functional documentation

The module root [README](../README.md) is a short **index**. Content lives here.

| Path | Purpose |
|---|---|
| [overview.md](overview.md) | Purpose, the language constraint that shapes the API, the immutability core, the zero-dependency boundary, the sibling-import rule |
| [examples.md](examples.md) | Usage sketches, including Result parsing and Validated accumulation |
| [packages/function.md](packages/function.md) | `primitives/function` — stdlib-only function combinators |
| [packages/predicate.md](packages/predicate.md) | `primitives/predicate` — stdlib-only Boolean predicates; `All`/`Any`/`None` are Deprecated slice synonyms |
| [packages/order.md](packages/order.md) | `primitives/order` — stdlib-only total orders |
| [packages/fold.md](packages/fold.md) | `primitives/fold` — slice folds (`iterator.Fold` folds an `iter.Seq`); edges to `values/optional` and `values/nonempty` |
| [packages/tuple.md](packages/tuple.md) | `structural/tuple` — Tuple2/Tuple3 carriers; stdlib-only; `FromValues2` Deprecated |
| [packages/lazy.md](packages/lazy.md) | `structural/lazy` — concurrent deferred memoization over `sync.OnceValue`; stdlib-only |
| [packages/iterator.md](packages/iterator.md) | `structural/iterator` — `iter.Seq` adapters; imports only `structural/tuple`; `FromSlice`/`FromValues`/`Collect` Deprecated |
| [packages/optics.md](packages/optics.md) | `structural/optics` — typed Lens; stdlib-only |
| [packages/immutable.md](packages/immutable.md) | Public API of `go.alvadorn.org/functional/immutable` (`List`, `Map`, `Set`, `SortedMap`), including measured Append-loop vs. Builder benchmarks |
| [packages/state.md](packages/state.md) | Public API of `go.alvadorn.org/functional/state` (pure `State[T]`, lineage, Tx, Diff); frozen, not deprecated |
| [packages/core.md](packages/core.md) | Maintainer guide for `internal/core` (ownership, cow, transient) |
| [packages/values.md](packages/values.md) | Family guide: choice by semantics, zero values, dependency boundary, and cost model |
| [packages/values-optional.md](packages/values-optional.md) | Public API of `go.alvadorn.org/functional/values/optional` |
| [packages/values-result.md](packages/values-result.md) | Public API of `go.alvadorn.org/functional/values/result` |
| [packages/values-either.md](packages/values-either.md) | Public API of `go.alvadorn.org/functional/values/either` |
| [packages/values-nonempty.md](packages/values-nonempty.md) | Public API of `go.alvadorn.org/functional/values/nonempty` |
| [packages/values-validated.md](packages/values-validated.md) | Public API of `go.alvadorn.org/functional/values/validated` |
| [patterns/value-objects.md](patterns/value-objects.md) | What a value object needs from this module vs. what plain Go already covers |
| [patterns/cow-handles.md](patterns/cow-handles.md) | Small value handles over private payloads when hugeParam bites |
| [adr/functional/](adr/functional/) | Architecture decisions (`adr:functional:NNN`) |

ADRs use AKB YAML frontmatter (`type`, `id`, `context`, `title`, `status`,
`tags`, …). Layout mirrors monorepo AKB conventions **inside this module**
(`docs/adr/<context>/…`).

All guides are written in **English**.

Current decisions:

- [Optional[T] representation, zero value, and Equal as a method](adr/functional/optional-representation-and-zero-value.md)
- [Zero-dependency boundary, v0 public-signature restraint, and the pgtype/Optional split](adr/functional/zero-dependency-boundary.md)
- [Thin immutability core — ownership, cow, transient](adr/functional/thin-immutability-core.md)
- [Persistent engine family — HAMT + 32-way vector, seed policy, O(1) Freeze](adr/functional/persistent-engine-family.md)
- [State semantic layer — pure State[T], mesh deps, lineage and single-use Tx](adr/functional/state-semantic-layer.md)
- [Values family dependency boundary](adr/functional/values-family-dependency-boundary.md)
- [Functional primitives and the fold dependency boundary](adr/functional/functional-primitives-and-fold-boundary.md)
- [Tuples, lazy values, iterators, and optics over native Go contracts](adr/functional/structural-primitives-and-iterator-boundary.md)
- [Retire top-level optional and valueobject packages](adr/functional/retire-optional-and-valueobject.md)
- [Group public packages under primitives/ and structural/ path prefixes](adr/functional/primitives-and-structural-path-groups.md)
- [Result interop with the (T, error) pair and the Match argument-order rule](adr/functional/values-error-pair-interop-and-match.md)
- [SortedMap on a weight-balanced tree engine (wbtree)](adr/functional/sorted-map-weight-balanced-engine.md)
- [v1 surface hygiene — Deprecated synonyms, lean core, native Lazy, frozen state](adr/functional/v1-surface-hygiene.md)
