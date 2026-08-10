# Package `values/optional`

**Import path:** `github.com/alvadorncorp/gopackages/functional/values/optional`

`Optional[T]` represents a value that may be present (`Some`) or absent
(`None`). It is a small value type for modelling explicit absence without
using a sentinel value.

## Representation and zero value

Conceptually, an `Optional[T]` is represented as:

```go
struct { value T; ok bool }
```

Both fields are unexported. The zero value of `Optional[T]` is `None`, so a
zero-initialized field is already a valid absent Optional. `Some(zero)` is
still distinct from `None`: `ok` records presence independently of `value`.

## Public API

| API | Description |
|---|---|
| `Optional[T]` | A value that is either `Some(T)` or `None`. |
| `Some(value)` | Creates a present Optional. |
| `None[T]()` | Creates an absent Optional. |
| `From(value, ok)` | Returns `Some(value)` when `ok` is true; otherwise `None`. |
| `IsSome()` / `IsNone()` | Report whether a value is present or absent. |
| `Get()` | Returns `(value, true)` when present; otherwise `(zero, false)`. |
| `OrElse(fallback)` | Returns the value or the supplied fallback. |
| `OrZero()` | Returns the value or `T`'s zero value. |
| `MustGet()` | Returns the value or panics when absent. |
| `Filter(predicate)` | Retains a present value only when the predicate is true. |
| `Inspect(callback)` | Calls the callback for a present value and returns the original Optional. |
| `Map(optional, transform)` | Transforms a present value, preserving absence. |
| `FlatMap(optional, transform)` | Transforms a present value into another Optional, preserving absence. |
| `Equal(a, b)` | Compares two `Optional[T]` values when `T` is `comparable`. |

## Branch laziness

`Map` and `FlatMap` do not invoke their transforms for `None`. Likewise,
`Filter` does not invoke its predicate and `Inspect` does not invoke its
callback for `None`. This keeps inactive branches lazy and makes side effects
explicitly conditional on presence.

## Panic behaviour

`MustGet` is for callers that have already established presence. Calling it on
`None` panics with exactly:

```
functional/values/optional: Optional is None
```

Prefer `Get`, `OrElse`, or `OrZero` when absence is part of ordinary control
flow.

## Equality, payloads, and allocations

`Equal[T comparable]` compares both the presence discriminator and the held
value. The Optional wrapper itself is a shallow value: copied values such as
slices, maps, pointers, and interfaces can still share mutable underlying
state with their source.

Constructing with `Some`, and wrapping paths through `Map` and `FlatMap`, are
covered by a zero-allocation contract for the wrapper itself. A callback or
payload may allocate independently; that allocation is not attributable to
the Optional wrapper.

