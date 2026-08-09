# Package `values/nonempty`

**Import path:** `github.com/alvadorncorp/gopackages/functional/values/nonempty`

`NonEmpty[T]` is a value container whose cardinality is always at least one.
It is useful where the domain requires a first value and optionally permits
more values.

## Representation and zero value

The representation holds a head element plus internal storage for remaining
elements. The storage is unexported.

> **Zero-value contract:** the structural zero value of `NonEmpty[T]` contains
> exactly `[zero(T)]`. Its length is one, and its `Head`, `Last`, and `At(0)`
> results are `zero(T)`.

Consequently, a zero-initialized `NonEmpty[T]` is valid and non-empty; it is
not an empty collection.

## Public API

| API | Description |
|---|---|
| `NonEmpty[T]` | A container with at least one element. |
| `Of(head, tail...)` | Builds a value from a required head and optional remaining elements. |
| `FromSlice(values)` | Builds a value from a non-empty slice, or returns `false` for an empty slice. |
| `Head()` / `Last()` | Return the first or last element. |
| `Len()` | Returns the cardinality, always at least one. |
| `At(index)` | Returns the element at a zero-based index. |
| `Tail()` / `Items()` | Return copies of the remaining elements or all elements. |
| `Map(value, transform)` | Transforms every element while preserving cardinality. |
| `Append(values...)` | Returns a new value with elements added after the existing ones. |

## Defensive copies and value semantics

`Of` and `FromSlice` defensively copy their input tails. `Tail` and `Items`
also return defensive copies. Mutating a caller-provided slice or any slice
returned by the container cannot change the `NonEmpty` value.

`Append` similarly returns a new value and does not mutate its receiver. The
container provides value semantics for its own element storage; elements that
are themselves references may still refer to mutable data.

## Bounds and transformations

`At` panics for a negative index or an index greater than or equal to `Len()`.
The panic value is exactly:

```
functional/values/nonempty: index out of range
```

`Map` applies its transform to every element and preserves the input
cardinality, including the one-element structural zero value.

## Why there is no `Filter`

`Filter` is deliberately absent. Filtering can remove every element and would
therefore break the NonEmpty cardinality guarantee. Convert to a regular slice
or choose a different domain representation when filtering may produce no
values.

## Not a persistent list

`NonEmpty` is a semantic cardinality container, not a `persistent.List`.
It does not provide persistent-list operations or structural sharing; its
slice-facing operations make defensive copies to preserve its value semantics.
