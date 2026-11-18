# functional-go

An umbrella module for small, dependency-free functional-programming building
blocks: function composition, predicate algebra, total ordering, algebraic
values, immutable collections, and pure state evolution.

```text
module  go.alvadorn.org/functional
go      1.27
```

- [`functional`](#functional) — pure values, composition, and immutable data.
- [`modeling`](modeling/) — typed domain decisions and state transitions.
- [`workflow`](workflow/) — composition and execution of deferred effects.

`modeling` ships with the root module and has no separate version. `workflow`
has its own module and release tags (`workflow/vX.Y.Z`). Run `make test-unit`
and `make lint` to check both modules, or `make test-published` to test without
the development workspace. Licensed under [Apache 2.0](LICENSE).

## functional

**The root package exports nothing.** `functional` is documentation only —
this README and the file it mirrors, [`doc.go`](doc.go). Concepts live one
import path down, grouped by role (no Go package at the group directory):

- **[`primitives/`](primitives/)** — [`function`](primitives/function/),
  [`predicate`](primitives/predicate/), [`order`](primitives/order/),
  [`fold`](primitives/fold/)
- **[`structural/`](structural/)** — [`tuple`](structural/tuple/),
  [`lazy`](structural/lazy/), [`iterator`](structural/iterator/),
  [`optics`](structural/optics/)
- **[`values/`](values/)** — optional, result, either, nonempty, validated
- **[`immutable/`](immutable/)**, **[`state/`](state/)** — collections
  (including the ordered `SortedMap`) and pure plan evolution (kept at the
  module root)

This README is the **documentation index**. Guides live under
[`docs/`](docs/) — start at [`docs/README.md`](docs/README.md).

---

## Documentation map

### Guides

| Document | Description |
|---|---|
| [**docs/overview.md**](docs/overview.md) | Purpose, the Go language constraint that shapes the API, the immutability core, the zero-dependency boundary, the sibling-import rule |
| [**docs/examples.md**](docs/examples.md) | Usage sketches, including Result parsing and Validated accumulation |

### Package (micro view)

| Document | Package |
|---|---|
| [docs/packages/function.md](docs/packages/function.md) | `go.alvadorn.org/functional/primitives/function` |
| [docs/packages/predicate.md](docs/packages/predicate.md) | `go.alvadorn.org/functional/primitives/predicate` |
| [docs/packages/order.md](docs/packages/order.md) | `go.alvadorn.org/functional/primitives/order` |
| [docs/packages/fold.md](docs/packages/fold.md) | `go.alvadorn.org/functional/primitives/fold` |
| [docs/packages/tuple.md](docs/packages/tuple.md) | `go.alvadorn.org/functional/structural/tuple` |
| [docs/packages/lazy.md](docs/packages/lazy.md) | `go.alvadorn.org/functional/structural/lazy` |
| [docs/packages/iterator.md](docs/packages/iterator.md) | `go.alvadorn.org/functional/structural/iterator` |
| [docs/packages/optics.md](docs/packages/optics.md) | `go.alvadorn.org/functional/structural/optics` |
| [docs/packages/immutable.md](docs/packages/immutable.md) | `go.alvadorn.org/functional/immutable` |
| [docs/packages/state.md](docs/packages/state.md) | `go.alvadorn.org/functional/state` |
| [docs/packages/core.md](docs/packages/core.md) | `internal/core` (ownership, cow, transient) — maintainers |
| [docs/packages/values.md](docs/packages/values.md) | Family guide for the five `functional/values/*` packages |
| [docs/packages/values-optional.md](docs/packages/values-optional.md) | `go.alvadorn.org/functional/values/optional` |
| [docs/packages/values-result.md](docs/packages/values-result.md) | `go.alvadorn.org/functional/values/result` |
| [docs/packages/values-either.md](docs/packages/values-either.md) | `go.alvadorn.org/functional/values/either` |
| [docs/packages/values-nonempty.md](docs/packages/values-nonempty.md) | `go.alvadorn.org/functional/values/nonempty` |
| [docs/packages/values-validated.md](docs/packages/values-validated.md) | `go.alvadorn.org/functional/values/validated` |

### Patterns

| Document | Description |
|---|---|
| [docs/patterns/value-objects.md](docs/patterns/value-objects.md) | What a value object needs from this module vs. what is just idiomatic Go |

### Architecture decisions

| Document | ID |
|---|---|
| [docs/adr/functional/optional-representation-and-zero-value.md](docs/adr/functional/optional-representation-and-zero-value.md) | `adr:functional:optional-representation-and-zero-value` |
| [docs/adr/functional/zero-dependency-boundary.md](docs/adr/functional/zero-dependency-boundary.md) | `adr:functional:zero-dependency-boundary` |
| [docs/adr/functional/thin-immutability-core.md](docs/adr/functional/thin-immutability-core.md) | `adr:functional:thin-immutability-core` |
| [docs/adr/functional/persistent-engine-family.md](docs/adr/functional/persistent-engine-family.md) | `adr:functional:persistent-engine-family` |
| [docs/adr/functional/state-semantic-layer.md](docs/adr/functional/state-semantic-layer.md) | `adr:functional:state-semantic-layer` |
| [docs/adr/functional/values-family-dependency-boundary.md](docs/adr/functional/values-family-dependency-boundary.md) | `adr:functional:values-family-dependency-boundary` |
| [docs/adr/functional/functional-primitives-and-fold-boundary.md](docs/adr/functional/functional-primitives-and-fold-boundary.md) | `adr:functional:functional-primitives-and-fold-boundary` |
| [docs/adr/functional/structural-primitives-and-iterator-boundary.md](docs/adr/functional/structural-primitives-and-iterator-boundary.md) | `adr:functional:structural-primitives-and-iterator-boundary` |
| [docs/adr/functional/retire-optional-and-valueobject.md](docs/adr/functional/retire-optional-and-valueobject.md) | `adr:functional:retire-optional-and-valueobject` |
| [docs/adr/functional/primitives-and-structural-path-groups.md](docs/adr/functional/primitives-and-structural-path-groups.md) | `adr:functional:primitives-and-structural-path-groups` |
| [docs/adr/functional/values-error-pair-interop-and-match.md](docs/adr/functional/values-error-pair-interop-and-match.md) | `adr:functional:values-error-pair-interop-and-match` |
| [docs/adr/functional/sorted-map-weight-balanced-engine.md](docs/adr/functional/sorted-map-weight-balanced-engine.md) | `adr:functional:sorted-map-weight-balanced-engine` |
| [docs/adr/functional/v1-surface-hygiene.md](docs/adr/functional/v1-surface-hygiene.md) | `adr:functional:v1-surface-hygiene` |

---

## Package at a glance

### `primitives/`

| Package | Role | Core types |
|---|---|---|
| [`primitives/function`](primitives/function/) | Stateless, standard-library-only combinators and transforms | `Identity`, `Constant`, `Compose`, `Pipe`, `Flip`, `Tap`, `Transform[T]`, `PipeTransforms`, `ComposeTransforms` |
| [`primitives/predicate`](primitives/predicate/) | Stateless, standard-library-only Boolean predicates and comparables; slice quantifiers are Deprecated | `Predicate[T]`, `Not`, `And`, `Or`, `EqualTo`, `NotEqualTo`, `OneOf` (Deprecated: `All`, `Any`, `None`) |
| [`primitives/order`](primitives/order/) | Sign-based total orders, projections, and lexicographic composition | `Order[T]`, `Natural`, `By`, `Then`, `Min`, `Max` |
| [`primitives/fold`](primitives/fold/) | Iterative slice folds, reductions, and left scans; fold an `iter.Seq` with `iterator.Fold` | `FoldLeft`, `FoldRight`, `Reduce`, `ReduceNonEmpty`, `ScanLeft` |

### `structural/`

| Package | Role | Core types |
|---|---|---|
| [`structural/tuple`](structural/tuple/) | `Tuple2`/`Tuple3` heterogeneous carriers and component maps; stdlib-only | `Tuple2[A,B]`, `Tuple3[A,B,C]`, `Of2`, `Of3`, `Swap` (Deprecated: `FromValues2`) |
| [`structural/lazy`](structural/lazy/) | Concurrent-safe deferred memoization over `sync.OnceValue`; stdlib-only | `Lazy[T]`, `New`, `Of`, `Get`, `Map`, `FlatMap` |
| [`structural/iterator`](structural/iterator/) | Lazy transforms/limits/Zip over `iter.Seq`; imports only `structural/tuple` | `Map`, `Filter`, `Take`, `Zip`, `Enumerate`, `Fold` (Deprecated: `FromSlice`, `FromValues`, `Collect`) |
| [`structural/optics`](structural/optics/) | Typed Lens; stdlib-only leaf | `Lens[S,A]`, `NewLens`, `Get`, `Set`, `Modify`, `Compose` |

### `values/`, collections, state

| Package | Role | Core types |
|---|---|---|
| [`values/optional`](values/optional/), [`values/result`](values/result/), [`values/either`](values/either/), [`values/nonempty`](values/nonempty/), [`values/validated`](values/validated/) | Independent algebraic value types; see the [family guide](docs/packages/values.md) | `Optional[T]`, `Result[T,E]`, `Either[L,R]`, `NonEmpty[T]`, `Validated[T,E]` |
| [`immutable`](immutable/) | Immutable collections with functional writes (HAMT Map/Set, 32-way vector List, weight-balanced-tree SortedMap) | `List[T]`, `Map[K,V]`, `Set[T]`, `SortedMap[K,V]`, `Builder()` |
| [`state`](state/) | Pure semantic evolution over `T`; supported and frozen, not deprecated | `State[T]`, `Transaction[T]`, `Diff[T]`, `Snapshot[T]` |

Names marked **Deprecated** are thin synonyms of standard-library calls
(`slices.Values`, `slices.Collect`, `slices.ContainsFunc`) or of `Of2`. They
keep working and are not removed in v1; see
[v1 surface hygiene](docs/adr/functional/v1-surface-hygiene.md).

---

## Quick links

```go
import (
	"fmt"

	"go.alvadorn.org/functional/primitives/function"
	"go.alvadorn.org/functional/primitives/fold"
	"go.alvadorn.org/functional/primitives/order"
	"go.alvadorn.org/functional/primitives/predicate"
	"go.alvadorn.org/functional/immutable"
	"go.alvadorn.org/functional/state"
	"go.alvadorn.org/functional/values/optional"
)

port := optional.Some("8080")
n := optional.Map(port, len) // optional.Optional[int]

tags := immutable.ListOf("prod", "us-east") // immutable.List[string]

s := state.Of(0).Apply(func(v int) int { return v + 1 })
_ = s.Materialize() // 1; Apply is lazy, Materialize re-applies the plan

length := func(value string) int { return len(value) }
double := func(value int) int { return value * 2 }
fmt.Println(function.Pipe(length, double)("abcd")) // 8

positiveEven := predicate.And[int](func(v int) bool { return v > 0 }, func(v int) bool { return v%2 == 0 })
fmt.Println(positiveEven(4), positiveEven(3)) // true false

byLength := order.By(func(value string) int { return len(value) }, order.Natural[int]())
fmt.Println(order.Less(byLength)("go", "gopher")) // true

fmt.Println(fold.ScanLeft([]int{1, 2, 3}, 0, func(total, value int) int {
	return total + value
})) // [0 1 3 6]
```

See **[docs/examples.md](docs/examples.md)** for more sketches.

```sh
# From repository root
make test-unit
make lint
```

```sh
cd functional && go test ./... -count=1
```

---

## Non-goals

- **No `encoding/json`, `database/sql`, or `slog` integration in
  `values/optional`.** A struct field of type `Optional[T]` serializes through
  `encoding/json` as **`{}`** — the marshaled value, not just the marker of
  absence, is lost. Do not put `Optional[T]` on a struct you intend to
  marshal without writing your own `MarshalJSON`/`Scan` bridge.
- **No monad zoo.** No `IO`, `Task`, `Reader`, or monadic effect/context
  types — values packages stop at algebraic presence/outcomes. The public
  `state` package is a pure plan/Version layer over `T`, not a monadic State
  effect.
- **`immutable` collections are SHALLOW, not transitively immutable.**
  `immutable.List[[]byte]` still shares each element's backing array;
  nothing in this module makes an element type immutable on your behalf.
- **`immutable` engines are path-copy structure sharing, not full-copy
  COW.** Map/Set use a HAMT; List uses a 32-way vector; SortedMap uses a
  weight-balanced tree (`wbtree`). Documented log-n complexity bounds are
  public-API upper bounds. `List.Delete` is currently O(n). Builders publish
  with O(1) owner invalidation. See
  [persistent engine family](docs/adr/functional/persistent-engine-family.md),
  [sorted map weight balanced engine](docs/adr/functional/sorted-map-weight-balanced-engine.md),
  [thin immutability core](docs/adr/functional/thin-immutability-core.md), and
  [docs/packages/core.md](docs/packages/core.md). Core remains laws only —
  not a generic engine library.
- **`state` is pure over T and mesh-isolated.** Production `state` does not
  import `immutable`, core, or engines. `Version` is not a DB revision;
  `Diff` is not a domain event; `Transaction` is not a `immutable.Builder`.
  See [state semantic layer](docs/adr/functional/state-semantic-layer.md).

---

## Versioning

This module follows semantic versioning; the current version is published as
a git tag (`functional/vX.Y.Z`) — the README deliberately does not pin a
version number, since it goes stale quickly.
