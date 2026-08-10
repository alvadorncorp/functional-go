# functional

An umbrella module for small, dependency-free functional-programming building
blocks: explicit absence, immutable collections, pure state evolution, and
value-object verification.

```text
module  github.com/alvadorncorp/gopackages/functional
go      1.26+
```

**The root package exports nothing.** `functional` is documentation only —
this README and the file it mirrors, [`doc.go`](doc.go). Each concept lives
in its own subpackage: [`function`](function/), [`predicate`](predicate/), [`optional`](optional/),
[`persistent`](persistent/), [`state`](state/), [`valueobject`](valueobject/), or one of the
[`values`](values/) packages.

The established [`functional/optional`](optional/) API remains supported and
unchanged. [`functional/values/optional`](values/optional/) is an independent
baseline type; this module provides no alias, conversion, or migration layer
between them.

This README is the **documentation index**. Guides live under
[`docs/`](docs/).

---

## Documentation map

### Guides

| Document | Description |
|---|---|
| [**docs/overview.md**](docs/overview.md) | Purpose, the Go language constraint that shapes the API, the immutability core, the zero-dependency boundary, the sibling-import rule |
| [**docs/examples.md**](docs/examples.md) | Usage sketches, including Result parsing and Validated accumulation |
| [**docs/packages/values.md**](docs/packages/values.md) | Choosing and composing the functional values family |

### Package (micro view)

| Document | Package |
|---|---|
| [docs/packages/function.md](docs/packages/function.md) | `github.com/alvadorncorp/gopackages/functional/function` |
| [docs/packages/predicate.md](docs/packages/predicate.md) | `github.com/alvadorncorp/gopackages/functional/predicate` |
| [docs/packages/optional.md](docs/packages/optional.md) | `github.com/alvadorncorp/gopackages/functional/optional` |
| [docs/packages/persistent.md](docs/packages/persistent.md) | `github.com/alvadorncorp/gopackages/functional/persistent` |
| [docs/packages/state.md](docs/packages/state.md) | `github.com/alvadorncorp/gopackages/functional/state` |
| [docs/packages/valueobject.md](docs/packages/valueobject.md) | `github.com/alvadorncorp/gopackages/functional/valueobject` |
| [docs/packages/core.md](docs/packages/core.md) | `internal/core` (ownership, cow, sharing, transient) — maintainers |
| [docs/packages/values.md](docs/packages/values.md) | Family guide for the five `functional/values/*` packages |
| [docs/packages/values-optional.md](docs/packages/values-optional.md) | `github.com/alvadorncorp/gopackages/functional/values/optional` |
| [docs/packages/values-result.md](docs/packages/values-result.md) | `github.com/alvadorncorp/gopackages/functional/values/result` |
| [docs/packages/values-either.md](docs/packages/values-either.md) | `github.com/alvadorncorp/gopackages/functional/values/either` |
| [docs/packages/values-nonempty.md](docs/packages/values-nonempty.md) | `github.com/alvadorncorp/gopackages/functional/values/nonempty` |
| [docs/packages/values-validated.md](docs/packages/values-validated.md) | `github.com/alvadorncorp/gopackages/functional/values/validated` |

### Patterns

| Document | Description |
|---|---|
| [docs/patterns/value-objects.md](docs/patterns/value-objects.md) | What a value object needs from this module vs. what is just idiomatic Go |

### Architecture decisions

| Document | ID |
|---|---|
| [docs/adr/functional/adr-001-optional-representation-and-zero-value.md](docs/adr/functional/adr-001-optional-representation-and-zero-value.md) | `adr:functional:001` |
| [docs/adr/functional/adr-003-zero-dependency-boundary.md](docs/adr/functional/adr-003-zero-dependency-boundary.md) | `adr:functional:003` |
| [docs/adr/functional/adr-004-thin-immutability-core.md](docs/adr/functional/adr-004-thin-immutability-core.md) | `adr:functional:004` |
| [docs/adr/functional/adr-005-persistent-engine-family.md](docs/adr/functional/adr-005-persistent-engine-family.md) | `adr:functional:005` |
| [docs/adr/functional/adr-006-state-semantic-layer.md](docs/adr/functional/adr-006-state-semantic-layer.md) | `adr:functional:006` |

---

## Package at a glance

| Package | Role | Core types |
|---|---|---|
| [`function`](function/) | Stateless, standard-library-only function combinators | `Identity`, `Constant`, `Compose`, `Pipe`, `Flip`, `Tap` |
| [`predicate`](predicate/) | Stateless, standard-library-only Boolean predicates and quantifiers | `Predicate[T]`, `Not`, `And`, `Or`, `All`, `Any`, `None` |
| [`optional`](optional/) | Explicit absence: an operation that may or may not produce a value | `optional.Optional[T]` |
| [`values/optional`](values/optional/), [`values/result`](values/result/), [`values/either`](values/either/), [`values/nonempty`](values/nonempty/), [`values/validated`](values/validated/) | Independent algebraic value types; see the [family guide](docs/packages/values.md) | `Optional[T]`, `Result[T,E]`, `Either[L,R]`, `NonEmpty[T]`, `Validated[T,E]` |
| [`persistent`](persistent/) | Immutable collections with functional writes (HAMT Map/Set, 32-way vector List; Builders use exclusive-owner transient + O(1) Freeze) | `persistent.List[T]`, `persistent.Map[K,V]`, `persistent.Set[T]`, and a `Builder()` per type |
| [`state`](state/) | Pure semantic evolution over `T`: reverse-linked plans, Version, single-use Tx, structural Diff | `state.State[T]`, `state.Transaction[T]`, `state.Diff[T]`, `state.Snapshot[T]` |
| [`valueobject`](valueobject/) | Reflective verification that a type is structurally immutable (test helper) | `valueobject.Check[T]` |

---

## Quick links

```go
import (
    "fmt"

    "github.com/alvadorncorp/gopackages/functional/function"
    "github.com/alvadorncorp/gopackages/functional/optional"
	"github.com/alvadorncorp/gopackages/functional/predicate"
    "github.com/alvadorncorp/gopackages/functional/persistent"
    "github.com/alvadorncorp/gopackages/functional/state"
)

port := optional.Some("8080")
n, err := optional.Map(port, len) // optional.Optional[int]{...}, no error path

tags := persistent.ListOf("prod", "us-east") // persistent.List[string]

s := state.Of(0).Apply(func(v int) int { return v + 1 })
_ = s.Materialize() // 1; Apply is lazy, Materialize re-applies the plan

length := func(value string) int { return len(value) }
double := func(value int) int { return value * 2 }
fmt.Println(function.Pipe(length, double)("abcd")) // 8

positiveEven := predicate.And[int](func(v int) bool { return v > 0 }, func(v int) bool { return v%2 == 0 })
fmt.Println(positiveEven(4), positiveEven(3)) // true false
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

- **No `encoding/json`, `database/sql`, or `slog` integration in `optional`'s
  core.** A struct field of type `Optional[T]` serializes through
  `encoding/json` as **`{}`** — the marshaled value, not just the marker of
  absence, is lost. Do not put `Optional[T]` on a struct you intend to
  marshal without writing your own `MarshalJSON`/`Scan` bridge.
- **No monad zoo.** No `IO`, `Task`, `Reader`, or monadic effect/context
  types — `optional` stops at explicit absence. The public `state` package
  is a pure plan/Version layer over `T`, not a monadic State effect.
- **`persistent` collections are SHALLOW, not transitively immutable.**
  `persistent.List[[]byte]` still shares each element's backing array;
  nothing in this module makes an element type immutable on your behalf.
- **`persistent` engines are path-copy structure sharing, not full-copy
  COW.** Map/Set use a HAMT; List uses a 32-way vector. Documented log-n
  complexity bounds are public-API upper bounds. `List.Delete` is currently
  O(n). Builders publish with O(1) owner invalidation. See
  [ADR-005](docs/adr/functional/adr-005-persistent-engine-family.md),
  [ADR-004](docs/adr/functional/adr-004-thin-immutability-core.md), and
  [docs/packages/core.md](docs/packages/core.md). Core remains laws only —
  not a generic engine library.
- **`state` is pure over T and mesh-isolated.** Production `state` does not
  import `persistent`, core, or engines. `Version` is not a DB revision;
  `Diff` is not a domain event; `Transaction` is not a `persistent.Builder`.
  See [ADR-006](docs/adr/functional/adr-006-state-semantic-layer.md).

---

## Versioning

`functional/v0.1.0`. Still **v0**: a minor version may break the public API.
The API is promised stable starting at `v1.0.0`, planned after two
independent adopters inside this workspace have used it in real code. See
[ADR-003](docs/adr/functional/adr-003-zero-dependency-boundary.md) for what
that means for other library modules in this workspace that might otherwise
want `optional.Optional[T]` in their own exported signatures before v1.
