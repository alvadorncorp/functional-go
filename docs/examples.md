# functional examples

Illustrative sketches for common uses. For API contracts see
[packages/](packages/) and godoc.

---

## Contents

1. [Config precedence with `Or`](#1-config-precedence-with-or)
2. [Fallible parse with `MapErr`](#2-fallible-parse-with-maperr)
3. [Immutable collection field with `persistent.List`](#3-immutable-collection-field-with-persistentlist)
4. [Building a `persistent.List` in a loop](#4-building-a-persistentlist-in-a-loop)
5. [`persistent.MapOf` / `SetOf` snapshots](#5-persistentmapof--setof-snapshots)
6. [Versioned pure state with `state.Of` / `Apply` / `Materialize`](#6-versioned-pure-state-with-stateof--apply--materialize)

---

## 1. Config precedence with `Or`

`Optional.Or` reads left to right: the receiver wins if present, otherwise
`other` is returned. Chaining it expresses a precedence order —
flag beats environment beats hard-coded default — without an `if`/`else`
ladder.

```go
package config

import "github.com/alvadorncorp/gopackages/functional/optional"

func resolvePort(flagValue, envValue optional.Optional[string]) string {
	defaultValue := optional.Some("8080")
	return flagValue.Or(envValue).Or(defaultValue).MustGet()
}
```

`MustGet` is safe here only because the chain always ends in a `Some`
(`defaultValue` is always present); it panics on absence, so keep the final
link in an `Or` chain a guaranteed `Some` before calling it.

---

## 2. Fallible parse with `MapErr`

`MapErr` applies a fallible conversion without an intermediate presence
check: an absent `Optional` short-circuits to `(None, nil)` — the parser is
never called — and a failing parse returns `(None, err)`.

```go
package config

import (
	"strconv"

	"github.com/alvadorncorp/gopackages/functional/optional"
)

func parsePort(raw optional.Optional[string]) (optional.Optional[int], error) {
	return optional.MapErr(raw, strconv.Atoi)
}
```

Both branches — "not set" and "set but malformed" — stay distinguishable to
the caller: the first is `(None, nil)`, the second is `(None, someError)`.

---

## 3. Immutable collection field with `persistent.List`

A value object with a collection field should use `persistent.List[T]` (or
`persistent.Map[K,V]` / `Set[T]`), not a plain `[]T`. The zero `List` is
already a valid empty collection, so the field never needs a nil check, and
no caller can reach the internal engine storage. Prefer not calling write
methods on a field you intend as a pure snapshot.

```go
package catalog

import "github.com/alvadorncorp/gopackages/functional/persistent"

type Product struct {
	sku  string
	tags persistent.List[string]
}

func NewProduct(sku string, tags ...string) Product {
	return Product{sku: sku, tags: persistent.ListOf(tags...)}
}

func (p Product) Tags() persistent.List[string] { return p.tags }
```

`ListOf` copies `tags` into the persistent vector; `Product.Tags` hands
back the `List` value itself (not `Items()`), so a caller can iterate with
`Values()` without materializing a slice, or call `.Items()` only when it
actually needs an owned, mutable `[]string`. See
[patterns/value-objects.md](patterns/value-objects.md) for the full value
object shape this fits into.

---

## 4. Building a `persistent.List` in a loop

A single `persistent.List.Append` path-copies along the vector depth (not a
full collection copy), but calling it `n` times in a loop still allocates
`n` intermediate Lists. `Builder()` exists precisely for bulk construction:
seed it once, batch every write under exclusive ownership, and call
`Build()` once at the end (O(1) Freeze). See
[packages/persistent.md](packages/persistent.md).

```go
package pipeline

import "github.com/alvadorncorp/gopackages/functional/persistent"

func collectEven(items []int) persistent.List[int] {
	builder := persistent.ListOf[int]().Builder()
	for _, v := range items {
		if v%2 == 0 {
			builder.Append(v)
		}
	}
	return builder.Build()
}
```

`Build()` freezes the builder's transient window (O(1) owner invalidation),
returns an immutable `List`, and permanently disables the builder — every
method on it, including a repeated `Build`, panics with
`"persistent: builder used after Build"` from that point on. Never keep
using a builder after calling `Build()`.

Constructors are `ListOf` / `MapOf` / `SetOf` (not `New*`). Runnable
godoc examples live in `persistent/example_test.go`.

---

## 5. `persistent.MapOf` / `SetOf` snapshots

`MapOf` copies a `map[K]V` into a persistent HAMT; `SetOf` deduplicates
variadic items. Both return immutable values: further `Set`/`Add` calls
derive new collections.

```go
package catalog

import "github.com/alvadorncorp/gopackages/functional/persistent"

func scoreboard() (persistent.Map[string, int], persistent.Set[string]) {
	scores := persistent.MapOf(map[string]int{
		"alice": 10,
		"bob":   7,
	})
	// Derive a new Map; scores is unchanged.
	scores = scores.Set("carol", 12)

	active := persistent.SetOf("alice", "bob", "bob")
	return scores, active
}
```

For bulk Map construction, prefer `MapOf[K,V](nil).Builder()` (or
`var zero Map[K,V]; zero.Builder()`), then `Build()` once.

---

## 6. Versioned pure state with `state.Of` / `Apply` / `Materialize`

`state.Of` seeds a Version-0 plan. `Apply` records a pure `func(T) T`
without running it (O(1) reverse-linked plan node). `Materialize` re-applies
the full plan oldest→newest every call — there is no cache. Optional
`Begin`/`Transaction` batches several plan steps before `Commit` (still
lazy; Commit does not Materialize).

```go
package counter

import "github.com/alvadorncorp/gopackages/functional/state"

func evolve() (int, state.Version) {
	s := state.Of(0).
		Apply(func(v int) int { return v + 1 }).
		Apply(func(v int) int { return v * 10 })

	// Batch two more steps without intermediate States if preferred:
	s = s.Begin().
		Apply(func(v int) int { return v + 2 }).
		Apply(func(v int) int { return v + 3 }).
		Commit()

	return s.Materialize(), s.Version() // 15, Version 4
}
```

Remember the distinctions:

```
state.Transaction ≠ persistent.Builder
state.Diff ≠ Domain Event
state.Version ≠ DB revision
state.Materialize ≠ deep clone
```

Runnable godoc examples live in `state/example_test.go`. Full API:
[packages/state.md](packages/state.md).

---

## See also

- [overview](overview.md)
- [package API: optional](packages/optional.md)
- [package API: persistent](packages/persistent.md)
- [package API: state](packages/state.md)
- [package API: valueobject](packages/valueobject.md)
- [pattern: value objects](patterns/value-objects.md)
