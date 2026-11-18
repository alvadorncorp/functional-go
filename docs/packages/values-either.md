# Package `values/either`

**Import path:** `go.alvadorn.org/functional/values/either`

`Either[L, R]` is a neutral disjoint union: it holds a value in either its
Left branch or its Right branch. The names describe positions only. Neither
branch implies success, failure, presence, absence, or priority.

## Representation and zero value

Conceptually, an `Either[L, R]` is represented as:

```go
struct { left L; right R; isRight bool }
```

All fields are unexported. Its zero value is `Left(zero(L))`, which is a valid
Left value. A Right value is constructed explicitly with `Right`.

## Exact API

| API | Description |
|---|---|
| `Either[L, R]` | A value in the Left or Right branch. |
| `Left(value)` / `Right(value)` | Construct values for their respective branches. |
| `IsLeft()` / `IsRight()` | Report the active branch. |
| `LeftValue()` / `RightValue()` | Return a branch payload and whether that branch is active. |
| `MapLeft(e, transform)` / `MapRight(e, transform)` | Transform only the selected branch. |
| `Bimap(e, leftFn, rightFn)` | Transform the selected branch using its corresponding function. |
| `FlatMapRight(e, transform)` | Sequence a Right branch while preserving a Left branch. |
| `Match(e, leftFn, rightFn)` | Eliminate an Either by producing one result for either branch. |

`Either` deliberately has no `Get` API, no `Must...` APIs, and no equality
helper. Inspect the active branch with `IsLeft` or `IsRight`, use the paired
value accessors when needed, or use `Match` to handle both branches.

## Lazy branch evaluation

The transformation and matching functions are lazy with respect to the
inactive branch: `MapLeft` and `MapRight` call only the transform for their
selected branch, `Bimap` calls exactly one function, and `Match` calls exactly
one handler. `FlatMapRight` calls its transform only for Right and leaves Left
unchanged.

## Meaning and use cases

Choose the branch meanings at the use site. For example, an Either can model a
value that has one of two input formats, a routing decision, or a parsed value
whose representation differs by source. Keep the meanings local and document
them where the Either is used; the type itself remains neutral.

This differs from `Result`, where `Ok` and `Err` carry explicit
success-and-failure semantics and support fail-fast sequencing. `Either` is not
a Result alias and version 1 intentionally provides no conversions to or from
`Result` (or other value types).

## Allocation contract

The `Left` and `Right` wrappers, along with the wrapper paths through
`MapLeft`, `MapRight`, `Bimap`, `FlatMapRight`, and `Match`, have a
zero-allocation wrapper contract. A callback or payload may allocate on its
own; such allocations are not attributable to Either.
