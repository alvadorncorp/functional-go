# Package `values/result`

**Import path:** `go.alvadorn.org/functional/values/result`

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
| `Match(result, errFn, okFn)` | Calls `errFn` for `Err` or `okFn` for `Ok` and returns its result. |
| `FromPair(value, err)` | Builds a `Result[T, error]` from Go's `(T, error)` pair. |
| `ToPair(result)` | Returns a `Result[T, error]` as Go's `(T, error)` pair. |
| `Equal(a, b)` | Compares Result values when both payload types are `comparable`. |

## Laziness and fail-fast sequencing

`Map` and `FlatMap` call their transforms only for `Ok`. An `Err` skips both
success callbacks and is returned with its error unchanged. `MapErr` is the
opposite branch: it calls its transform only for `Err` and preserves `Ok`.

Use `FlatMap` when a later successful operation itself can fail and depends on
the earlier value. This makes `Result` appropriate for fail-fast workflows,
where the first failure ends the sequence.

`Match` is the eliminator: it calls exactly one handler, `errFn` for `Err`
(including the zero Result, with a zero `E`) and `okFn` for `Ok`. Like
`either.Match`, the zero-value branch comes first. Handlers are called
directly and are not nil-checked.

## Interop with (T, error)

`FromPair` and `ToPair` bridge `Result[T, error]` with Go's conventional
`(T, error)` return pair:

```go
parsed := result.FromPair(strconv.Atoi(input)) // Result[int, error]
value, err := result.ToPair(parsed)            // (int, error)
```

`FromPair` checks only `err != nil`. A nil `err` yields `Ok(value)`, even
when `value` is the zero value; a non-nil `err` yields `Err(err)` and drops
`value`. An interface that holds a typed nil pointer, such as
`error((*MyError)(nil))`, is non-nil and therefore yields `Err`.

`FromPair` is unsuitable in two cases:

- **Concrete pointer error types.** When a function's last result is a
  concrete pointer type such as `*MyError` rather than `error`,
  `FromPair(f())` converts a nil `*MyError` to a non-nil `error` and turns a
  success into `Err`. The compiler does not catch this. Check the concrete
  result yourself before building a Result.
- **Partial values.** APIs that return a meaningful value together with an
  error, such as `io.Reader.Read`'s `(n, err)`, lose that value: `FromPair`
  drops it whenever `err` is non-nil.

`ToPair` returns `(value, nil)` for `Ok` and `(zero, err)` for `Err`. It
panics when the Result is `Err` with a nil error, including a zero
`Result[T, error]`, because returning `(zero, nil)` would report a failure as
a success.

## Panic behaviour

`MustGet` is for code that has already established `IsOk()`. Calling it on an
`Err`, including a zero Result, panics with exactly:

```
functional/values/result: Result is Err
```

`ToPair` on an `Err` whose error is nil, including a zero `Result[T, error]`,
panics with exactly:

```
functional/values/result: ToPair on Err with nil error
```

Prefer `Get`, `ErrValue`, `OrElse`, or `OrZero` when failure is part of normal
control flow.

## Equality and allocations

`Equal[T, E comparable]` compares the discriminator and both payload fields.
The Result wrapper is shallow: copied slices, maps, pointers, and interfaces
can still share mutable underlying state with their source.

Constructing with `Ok`, `Err`, or `FromPair`, wrapping paths through `Map`,
`MapErr`, and `FlatMap`, and unwrapping through `ToPair` and `Match`, have a
zero-allocation wrapper contract. A callback or payload
may allocate independently; that allocation is not attributable to Result.

## Result, Validated, and Either

`Result` is for dependent, fail-fast sequencing: each step needs the previous
successful value, and the first `Err` is preserved. `Validated` is for
independent validation where failures can be accumulated rather than stopping
at the first one. `Either` is a neutral union type; Result is not an Either
alias and assigns explicit success/failure meaning to its branches.

Interop with Go's `(T, error)` pair is in scope through `FromPair` and
`ToPair`. Version 1 deliberately provides no conversion APIs between Result,
Validated, or Either, and no serialization helpers.
