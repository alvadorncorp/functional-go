# functional examples

Illustrative sketches for common uses. For API contracts see
[packages/](packages/) and godoc.

---

## Contents

1. [Config precedence with `Or`](#1-config-precedence-with-or)
2. [Fallible parse with `MapErr`](#2-fallible-parse-with-maperr)
3. [Immutable collection field with `frozen.Slice`](#3-immutable-collection-field-with-frozenslice)
4. [Building a `persistent.List` in a loop](#4-building-a-persistentlist-in-a-loop)

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

## 3. Immutable collection field with `frozen.Slice`

A value object with a collection field should use `frozen.Slice[T]` (or
`frozen.Map[K,V]`), not a plain `[]T`. The zero `Slice` is already a valid,
empty collection, so the field never needs a nil check, and no caller —
inside or outside the package — can reach the backing array.

```go
package catalog

import "github.com/alvadorncorp/gopackages/functional/frozen"

type Product struct {
	sku  string
	tags frozen.Slice[string]
}

func NewProduct(sku string, tags ...string) Product {
	return Product{sku: sku, tags: frozen.NewSlice(tags...)}
}

func (p Product) Tags() frozen.Slice[string] { return p.tags }
```

`NewSlice` copies `tags` on the way in; `Product.Tags` hands back the
`Slice` value itself (not `Items()`), so a caller can iterate with
`Values()` at zero extra allocation, or call `.Items()` only when it
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
	builder := persistent.NewList[int]().Builder()
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

---

## See also

- [overview](overview.md)
- [package API: optional](packages/optional.md)
- [package API: frozen](packages/frozen.md)
- [package API: persistent](packages/persistent.md)
- [package API: valueobject](packages/valueobject.md)
- [pattern: value objects](patterns/value-objects.md)
