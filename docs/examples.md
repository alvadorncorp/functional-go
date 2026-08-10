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
7. [Parsing with Result and independent Validated checks](#7-parsing-with-result-and-independent-validated-checks)
8. [Function composition with `Pipe`](#8-function-composition-with-pipe)
9. [Predicate composition with `And`](#9-predicate-composition-with-and)
10. [Ordering with `By`](#10-ordering-with-by)
11. [Accumulating states with `ScanLeft`](#11-accumulating-states-with-scanleft)
12. [Tuple carriers](#12-tuple-carriers)

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

## 7. Parsing with Result and independent Validated checks

Use `Result` when parsing one field can fail before there is a value to
validate. Once an age has been parsed, name, email, and age checks are
independent `Validated` values; nested `Map2` accumulates every known error
instead of stopping at the first one.

```go
package registration

import (
	"strconv"

	"github.com/alvadorncorp/gopackages/functional/values/result"
	"github.com/alvadorncorp/gopackages/functional/values/validated"
)

type person struct {
	name  string
	email string
	age   int
}

func parseAge(raw string) result.Result[int, string] {
	age, err := strconv.Atoi(raw)
	if err != nil {
		return result.Err[int, string]("age must be a number")
	}
	return result.Ok[int, string](age)
}

func validatePerson(name, email, rawAge string) validated.Validated[person, string] {
	nameCheck := validated.Valid[string, string](name)
	if name == "" {
		nameCheck = validated.Invalid[string, string]("name is required")
	}
	emailCheck := validated.Valid[string, string](email)
	if email == "" {
		emailCheck = validated.Invalid[string, string]("email is required")
	}

	parsedAge := parseAge(rawAge)
	age, parsed := parsedAge.Get()
	ageCheck := validated.Valid[int, string](age)
	if !parsed {
		message, _ := parsedAge.ErrValue()
		ageCheck = validated.Invalid[int, string](message)
	} else if age < 18 {
		ageCheck = validated.Invalid[int, string]("age must be at least 18")
	}

	identity := validated.Map2(nameCheck, emailCheck, func(name, email string) [2]string {
		return [2]string{name, email}
	})
	return validated.Map2(identity, ageCheck, func(identity [2]string, age int) person {
		return person{name: identity[0], email: identity[1], age: age}
	})
}
```

This is explicit at the boundary: inspect `Result` to turn a parse failure
into that field's validation error, then let `Validated.Map2` accumulate
independent checks. The values packages intentionally add no conversion API.

---

## 8. Function composition with `Pipe`

`function.Pipe` applies ordinary Go functions left to right. `Compose` is the
same operation with arguments named in execution-reverse order.

```go
package presentation

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/function"
)

func printDoubledLength() {
	length := func(value string) int { return len(value) }
	double := func(value int) int { return value * 2 }
fmt.Println(function.Pipe(length, double)("abcd")) // 8
}
```

---

## 9. Predicate composition with `And`

`predicate.And` composes ordinary predicates left to right and stops at the
first false result. It is a standard-library-only leaf and needs no collection
or function-combinator dependency.

```go
package eligibility

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/predicate"
)

func printPositiveEven() {
	positiveEven := predicate.And[int](
		func(v int) bool { return v > 0 },
		func(v int) bool { return v%2 == 0 },
	)
fmt.Println(positiveEven(4), positiveEven(3)) // true false
}
```

---

## 10. Ordering with `By`

`order.By` compares a value through a projection. The result carries sign
semantics: callers use `Less`, `Greater`, or `Equal` instead of assuming a
specific negative or positive magnitude. `order` is a standard-library-only
leaf and does not depend on the other primitive packages.

```go
package people

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/order"
)

type person struct{ age int }

func younger() {
	byAge := order.By(func(value person) int { return value.age }, order.Natural[int]())
	fmt.Println(order.Less(byAge)(person{age: 20}, person{age: 30})) // true
}
```

---

## 11. Accumulating states with `ScanLeft`

`fold.ScanLeft` retains the initial state and every intermediate left-fold
state. Its result has one more item than the input slice.

```go
package counters

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/fold"
)

func printRunningTotals() {
	totals := fold.ScanLeft([]int{1, 2, 3}, 0, func(total, value int) int {
		return total + value
	})
	fmt.Println(totals) // [0 1 3 6]
}
```

---

## 12. Tuple carriers

```go
package pipeline

import "github.com/alvadorncorp/gopackages/functional/tuple"

func tupleCarrier() tuple.Tuple2[int, string] {
	pair := tuple.Of2("Ada", 36)
	return tuple.Swap(pair)
}
```

The returned value is `tuple.Tuple2[int,string]{First: 36, Second: "Ada"}`;
the generic parameters are inferred from the constructor arguments.

---

## See also

- [overview](overview.md)
- [package API: function](packages/function.md)
- [package API: predicate](packages/predicate.md)
- [package API: order](packages/order.md)
- [package API: fold](packages/fold.md)
- [functional values family](packages/values.md)
- [values Optional](packages/values-optional.md)
- [values Result](packages/values-result.md)
- [values Either](packages/values-either.md)
- [values NonEmpty](packages/values-nonempty.md)
- [values Validated](packages/values-validated.md)
- [package API: optional](packages/optional.md)
- [package API: persistent](packages/persistent.md)
- [package API: state](packages/state.md)
- [package API: valueobject](packages/valueobject.md)
- [pattern: value objects](patterns/value-objects.md)
