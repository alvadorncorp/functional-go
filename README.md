# functional

An umbrella module for small, dependency-free functional-programming building
blocks: explicit absence, immutable collections, and value-object
verification.

```text
module  github.com/alvadorncorp/gopackages/functional
go      1.26+
```

**The root package exports nothing.** `functional` is documentation only —
this README and the file it mirrors, [`doc.go`](doc.go). Each concept lives
in its own subpackage: [`optional`](optional/), [`frozen`](frozen/),
[`persistent`](persistent/), [`valueobject`](valueobject/).

This README is the **documentation index**. Guides live under
[`docs/`](docs/).

---

## Documentation map

### Guides

| Document | Description |
|---|---|
| [**docs/overview.md**](docs/overview.md) | Purpose, the Go language constraint that shapes the API, the zero-dependency boundary, the sibling-import rule |
| [**docs/examples.md**](docs/examples.md) | Usage sketches (config precedence, fallible parse, immutable collection field, loop-building) |

### Package (micro view)

| Document | Package |
|---|---|
| [docs/packages/optional.md](docs/packages/optional.md) | `github.com/alvadorncorp/gopackages/functional/optional` |
| [docs/packages/frozen.md](docs/packages/frozen.md) | `github.com/alvadorncorp/gopackages/functional/frozen` |
| [docs/packages/persistent.md](docs/packages/persistent.md) | `github.com/alvadorncorp/gopackages/functional/persistent` |
| [docs/packages/valueobject.md](docs/packages/valueobject.md) | `github.com/alvadorncorp/gopackages/functional/valueobject` |

### Patterns

| Document | Description |
|---|---|
| [docs/patterns/value-objects.md](docs/patterns/value-objects.md) | What a value object needs from this module vs. what is just idiomatic Go |

### Architecture decisions

| Document | ID |
|---|---|
| [docs/adr/functional/adr-001-optional-representation-and-zero-value.md](docs/adr/functional/adr-001-optional-representation-and-zero-value.md) | `adr:functional:001` |
| [docs/adr/functional/adr-002-persistent-engine-swap-contract.md](docs/adr/functional/adr-002-persistent-engine-swap-contract.md) | `adr:functional:002` |
| [docs/adr/functional/adr-003-zero-dependency-boundary.md](docs/adr/functional/adr-003-zero-dependency-boundary.md) | `adr:functional:003` |

---

## Package at a glance

| Package | Role | Core types |
|---|---|---|
| [`optional`](optional/) | Explicit absence: an operation that may or may not produce a value | `optional.Optional[T]` |
| [`frozen`](frozen/) | Immutable snapshot of a slice or map — no write operation, ever | `frozen.Slice[T]`, `frozen.Map[K,V]` |
| [`persistent`](persistent/) | Collection with functional writes (copy-on-write today, behind a swappable engine) | `persistent.List[T]`, `persistent.Map[K,V]`, `persistent.Set[T]`, and a `Builder()` per type |
| [`valueobject`](valueobject/) | Reflective verification that a type is structurally immutable (test helper) | `valueobject.Check[T]` |

---

## Quick links

```go
import (
    "github.com/alvadorncorp/gopackages/functional/optional"
    "github.com/alvadorncorp/gopackages/functional/frozen"
)

port := optional.Some("8080")
n, err := optional.Map(port, len) // optional.Optional[int]{...}, no error path

tags := frozen.NewSlice("prod", "us-east") // frozen.Slice[string]
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
- **No monad zoo.** No `IO`, `Task`, `Reader`, `State`, or any other
  effect/context type — `optional` stops at explicit absence.
- **`frozen` and `persistent` collections are SHALLOW, not transitively
  immutable.** `frozen.Slice[[]byte]` still shares each element's backing
  array; nothing in this module makes an element type immutable on your
  behalf.
- **`persistent` is not a HAMT or RRB tree today.** The engine underneath is
  copy-on-write over a stdlib slice/map. Every complexity bound documented on
  `persistent` is an upper bound the API contract makes room for, not a
  description of the current engine's actual constant — see
  [ADR-002](docs/adr/functional/adr-002-persistent-engine-swap-contract.md).

---

## Versioning

`functional/v0.1.0`. Still **v0**: a minor version may break the public API.
The API is promised stable starting at `v1.0.0`, planned after two
independent adopters inside this workspace have used it in real code. See
[ADR-003](docs/adr/functional/adr-003-zero-dependency-boundary.md) for what
that means for other library modules in this workspace that might otherwise
want `optional.Optional[T]` in their own exported signatures before v1.
