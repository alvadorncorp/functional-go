---
type: adr
id: adr:functional:optional-representation-and-zero-value
context: functional
title: Optional[T] representation, zero value, and Equal as a method
status: Accepted
tags: [functional, optional, generics, equality]
deciders: [alvadorn]
related: [adr:functional:retire-optional-and-valueobject]
date: 2026-08-07
---
# Optional[T] representation, zero value, and Equal as a method

> **Later note ([retire-optional-and-valueobject](retire-optional-and-valueobject.md)):** the top-level package
> `go.alvadorn.org/functional/optional` was removed. Presence
> lives at `functional/values/optional`. This ADR remains the historical record
> of the original representation decision.

## Context and problem

`optional.Optional[T]` needs a representation for "a `T` that may or may not
be present." Three sub-decisions had to be made together, because they
interact: how the value is stored (inline vs. pointer), what the zero value
means, and how `Equal` is exposed given that Go does not allow a method to
introduce its own type parameters.

## Decision drivers

- The zero `Optional[T]` should be immediately useful as a struct field,
  with no constructor and no nil check.
- `Optional[T]` should stay `==`-comparable whenever `T` itself is, so it
  composes cleanly into other comparable types (map keys, struct fields
  compared with `==` elsewhere in a codebase).
- `go-cmp`'s `cmp.Diff` only consults a method literally named
  `Equal(T) bool` on a type; without one, it falls back to reflecting into
  unexported fields and panics.
- Go's method grammar does not accept new type parameters on a method — see
  `$GOROOT/doc/go_spec.html`:

  ```text
  MethodDecl = "func" Receiver MethodName Signature [ FunctionBody ] .
  ```

  There is no `[ TypeParameters ]` production here (contrast
  `FunctionDecl`, which has one). A method can only use type parameters
  already bound by its receiver.

## Options considered

### Option 1: `type Optional[T] []T` (slice-backed, e.g. `moznion/go-optional`'s shape)

**Pros:** Trivial "no value" as `nil`, one line of code, presence as
`len(o) > 0`.
**Cons:** Breaks `==` entirely — slices are never comparable, even when
`T` is; every construction allocates, even for a `Some`; the zero-length
slice and a `nil` slice are both "no value" but are not identical
representations, and worse, a caller holding the slice header can `append`
to it or index past what the type intends to expose, silently mutating a
value that is supposed to read as immutable.

### Option 2: Store `T` behind a pointer (`struct{ value *T }`)

**Pros:** `nil` is a free, allocation-free "no value" signal without a
separate `present bool`.
**Cons:** Loses value comparability outright — two `Optional[T]` holding
equal `T`s through different pointers would never be `==`, which is worse
than Option 1's problem, not better; every `Some` needs a heap allocation
for the pointee; two `Optional[T]` values could alias the same backing `T`
without either side intending to share it.

### Option 3: `struct{ value T; present bool }`, stored inline (chosen)

**Pros:** Zero value is `None` for free; `Some(zero)` and `None` stay
distinct because presence is a separate field, not inferred from the value;
`==`-comparable exactly when `T` is; constructing a `Some` never allocates;
the held value is a copy, so external aliasing cannot reach it.
**Cons:** `Optional[T]`'s own immutability is only as deep as `T`'s —
documented as a shallow-immutability caveat, not solved by the
representation itself.

## Decision

**Chosen: Option 3.**

`Optional[T]` is `struct{ value T; present bool }`, with `T` guarded inline,
never behind a pointer. The internal representation is frozen at exactly
these two fields — adding any non-comparable field would silently remove
`==`-comparability from every instantiation, which is why this shape is
tested (`optional_repr_test.go`), not just documented.

Consequences of the representation:

- The zero `Optional[T]` is a valid `None`.
- `Some(zero)` and `None[T]()` are distinct values: presence is tracked
  apart from the value, never inferred by comparing it to `T`'s zero.
- `Optional[T]` is `==`-comparable exactly when `T` is comparable.

`Equal` is a **method**, `func (o Optional[T]) Equal(other Optional[T])
bool`, not a package-level function — the one operation on `Optional[T]`
that keeps its own type as both receiver and argument, so the method
grammar restriction above does not block it. It resolves through
`functional/internal/eq.Equal` in three tiers: (1) an `Equal(T) bool` method
on the dynamic value, if one exists; (2) `==`, if the dynamic values are
comparable; (3) `reflect.DeepEqual` otherwise. This exists specifically so
`cmp.Diff` does not panic on `Optional[T]`'s unexported fields: `go-cmp`
looks for a method named `Equal` before it ever reflects into a struct's
fields, and finding one here means it never has to.

Every operation that would need to return `Optional[U]` for a `U` chosen at
the call site — `Map`, `FlatMap`, `MapErr`, `Zip`, `EqualFunc` — is a
package-level function instead of a method, forced by the same grammar
restriction: a method bound to `Optional[T]`'s own `T` cannot introduce a
second type parameter `U`. This is a direct decision driver, not an
afterthought — see [the zero-dependency boundary ADR](zero-dependency-boundary.md) for how
it shapes the rest of the module's public functions the same way.

## Consequences

**Positive:** A struct field of type `Optional[T]` needs no constructor,
no nil check, and composes into `==`-comparable types when `T` allows it.
`cmp.Diff`/`go-cmp` based tests work against `Optional[T]` without a custom
`Comparer` option. `Some`/`None` construction is allocation-free.

**Negative:** `Optional[[]byte]` (or any reference-typed `T`) shares `T`'s
own aliasing — `Optional` cannot make a mutable `T` immutable on its own;
callers compose with an immutable `T` (e.g. `immutable.List`) when that
matters. `Map`/`FlatMap`/`MapErr`/`Zip`/`EqualFunc` read as package
functions rather than a fluent method chain, which is a direct, permanent
consequence of Go's method grammar, not a stylistic choice this module
could revisit.

**Neutral:** `Optional[T]`'s equality resolution is shared
verbatim (via `functional/internal/eq`) with `immutable.List`/`Map`/`Set`,
so every generic `Equal` in this module agrees on the same answer for the
same pair of values.
