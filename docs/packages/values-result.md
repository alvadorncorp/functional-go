# Package `values/result`

**Import path:** `github.com/alvadorncorp/gopackages/functional/values/result`

`Result[T, E]` represents an operation that has either succeeded with a value
of type `T` (`Ok`) or failed with a value of type `E` (`Err`). It models
dependent, fail-fast sequencing: once a step is `Err`, `Map` and `FlatMap`
preserve that failure without calling their success callbacks.

## Representation and zero value

Conceptually, a `Result[T, E]` is represented as:

```go
struct { value T; err E; ok bool }
```

All fields are unexported. The zero value of `Result[T, E]` is
`Err(zero(E))`. It is therefore a valid failure Result, with a zero `T` value
and its error discriminator set to failure.

## Public API

| API | Description |
|---|---|
| `Result[T, E]` | A value that is either `Ok(T)` or `Err(E)`. |
| `Ok(value)` | Creates a successful Result. |
| `Err(err)` | Creates a failed Result. |
| `IsOk()` / `IsErr()` | Report whether a Result is successful or failed. |
| `Get()` | Returns `(value, true)` for `Ok`; otherwise `(zero, false)`. |
| `ErrValue()` | Returns `(err, true)` for `Err`; otherwise `(zero, false)`. |
| `OrElse(fallback)` | Returns the successful value or `fallback` for `Err`. |
| `OrZero()` | Returns the successful value or `T`'s zero value. |
| `MustGet()` | Returns the successful value or panics for `Err`. |
| `Map(result, transform)` | Transforms `Ok`, preserving `Err` unchanged. |
| `MapErr(result, transform)` | Transforms `Err`, preserving `Ok` unchanged. |
| `FlatMap(result, transform)` | Sequences a dependent successful operation, fail-fast. |
| `Equal(a, b)` | Compares Result values when both payload types are `comparable`. |

## Laziness and fail-fast sequencing

`Map` and `FlatMap` call their transforms only for `Ok`. An `Err` skips both
success callbacks and is returned with its error unchanged. `MapErr` is the
opposite branch: it calls its transform only for `Err` and preserves `Ok`.

Use `FlatMap` when a later successful operation itself can fail and depends on
the earlier value. This makes `Result` appropriate for fail-fast workflows,
where the first failure ends the sequence.

## Panic behaviour

`MustGet` is for code that has already established `IsOk()`. Calling it on an
`Err`, including a zero Result, panics with exactly:

```
functional/values/result: Result is Err
```

Prefer `Get`, `ErrValue`, `OrElse`, or `OrZero` when failure is part of normal
control flow.

## Equality and allocations

`Equal[T, E comparable]` compares the discriminator and both payload fields.
The Result wrapper is shallow: copied slices, maps, pointers, and interfaces
can still share mutable underlying state with their source.

Constructing with `Ok` or `Err`, and wrapping paths through `Map`, `MapErr`,
and `FlatMap`, have a zero-allocation wrapper contract. A callback or payload
may allocate independently; that allocation is not attributable to Result.

## Result, Validated, and Either

`Result` is for dependent, fail-fast sequencing: each step needs the previous
successful value, and the first `Err` is preserved. `Validated` is for
independent validation where failures can be accumulated rather than stopping
at the first one. `Either` is a neutral union type; Result is not an Either
alias and assigns explicit success/failure meaning to its branches.

Version 1 deliberately provides no conversion APIs between Result, Validated,
or Either, and no serialization helpers.
