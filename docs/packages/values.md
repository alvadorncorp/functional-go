# Functional values family

The `values` packages are small, generic functional abstractions. They are
not domain models and do not share a common runtime hierarchy. Choose a type
for the meaning of a particular operation, then import only its package.

See [ADR-007](../adr/functional/adr-007-values-family-dependency-boundary.md)
for the boundary that keeps the family independently usable.

## Import paths

| Import path | Type |
|---|---|
| `github.com/alvadorncorp/gopackages/functional/values/optional` | `optional.Optional[T]` |
| `github.com/alvadorncorp/gopackages/functional/values/result` | `result.Result[T, E]` |
| `github.com/alvadorncorp/gopackages/functional/values/either` | `either.Either[L, R]` |
| `github.com/alvadorncorp/gopackages/functional/values/nonempty` | `nonempty.NonEmpty[T]` |
| `github.com/alvadorncorp/gopackages/functional/values/validated` | `validated.Validated[T, E]` |

Each package has its own contract: [Optional](values-optional.md),
[Result](values-result.md), [Either](values-either.md),
[NonEmpty](values-nonempty.md), and [Validated](values-validated.md).

## Choose the type by semantics

| Need | Choose |
|---|---|
| A value may be present or absent | `Optional` |
| Later work depends on an earlier successful step and stops at its first failure | `Result` |
| One of two neutral structural alternatives | `Either` |
| A collection must contain at least one item | `NonEmpty` |
| Independent checks should return every known validation error | `Validated` |

The names are semantic commitments, not interchangeable wrappers. In
particular, `Either` does not assign success or failure meaning to Left and
Right, while `Result` does.

## Zero values

All five structural zero values are valid values with deliberate meanings:

| Type | Zero value |
|---|---|
| `Optional[T]` | `None` |
| `Result[T, E]` | `Err(zero(E))` |
| `Either[L, R]` | `Left(zero(L))` |
| `NonEmpty[T]` | `[zero(T)]` |
| `Validated[T, E]` | `Invalid(zero(E))` |

`Some(zero(T))` is still distinct from `None`, and a zero `NonEmpty` is one
element rather than an empty collection.

## Dependency matrix

The family uses only the standard library except for its one intentional
in-family edge:

| Package | Production dependencies |
|---|---|
| `values/optional` | standard library only |
| `values/result` | standard library only |
| `values/either` | standard library only |
| `values/nonempty` | standard library only |
| `values/validated` | standard library and `values/nonempty` |

The dependency is exactly `values/validated` → `values/nonempty`; no other
values package imports another one. Production code in the family does not
import `reflect` or `unsafe`.

## Result versus Validated

Use `Result` for parsing and dependent sequencing: a later action needs an
earlier successful value, so `FlatMap` preserves the first `Err` and does not
run the later action. Use `Validated` for independent checks such as name,
email, and age: nested `Map2` combines valid values or accumulates all errors
in left-to-right order. `Validated` deliberately has no `FlatMap`, because a
dependent check cannot have the same independent-error accumulation rule.

## Either versus Result

`Either` is a structural union. Its Left and Right branches have meanings
chosen by the caller, such as two input formats or a routing decision.
`Result` is specifically a success/failure value with `Ok` and `Err`, and is
therefore the appropriate type for fail-fast operation outcomes. Neither type
provides a conversion API to the other.

## NonEmpty versus persistent.List

`NonEmpty` guarantees minimum cardinality and makes defensive copies at its
slice boundary. `persistent.List` is an immutable, structurally shared
collection that may be empty. Use `NonEmpty` when "at least one" is the
invariant; use `persistent.List` when persistent collection operations and
structural sharing are the concern.

## Functional values versus domain value objects

These are generic functional abstractions. They contain no DDD knowledge,
domain identity, serialization policy, or runtime coupling. A domain value
object may use a functional value internally when its semantics fit, but it
remains responsible for its own invariants and meaning.

## Cost model

Basic wrappers and branch-preserving transforms are designed to be
zero-allocation for the wrapper itself; a callback or payload may allocate
independently. `NonEmpty` defensive copies and `Validated` error accumulation
may allocate to preserve their value semantics. The implementation uses no
`unsafe` optimization.

Measure the delivered behavior with the package benchmarks:

```sh
cd functional
go test -bench=. -benchmem ./values/...
```

## Baseline exclusions

Version 1 intentionally excludes conversions between values types, `Map3+`,
`Filter` for `NonEmpty`, a typeclass or algebra hierarchy, serialization, and
domain or runtime coupling. Compose the APIs that exist at the use site
instead of assuming a future abstraction.
