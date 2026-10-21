# Package `values/validated`

**Import path:** `github.com/alvadorncorp/gopackages/functional/values/validated`

`Validated[T, E]` represents either a valid value or one or more validation
errors. Use it for independent checks that should report all known problems in
one pass.

## Representation and sole dependency

Conceptually, a value is represented as:

```go
struct { value T; errors nonempty.NonEmpty[E]; valid bool }
```

All fields are unexported. The package's sole dependency is
`values/nonempty`, which guarantees that every invalid value has at least one
error.

> **Zero-value contract:** the zero `Validated[T, E]` is invalid and contains
> `NonEmpty[E]` with exactly `zero(E)` as its only error.

Thus a zero value is a usable invalid validation result, not a valid value.

## Public API

| API | Description |
|---|---|
| `Validated[T, E]` | A value or one-or-more validation errors. |
| `Valid(value)` | Creates a valid value. |
| `Invalid(head, tail...)` | Creates an invalid value from one or more errors. |
| `InvalidNonEmpty(errors)` | Creates an invalid value from `NonEmpty[E]`. |
| `IsValid()` / `IsInvalid()` | Report the active branch. |
| `Get()` | Returns `(value, true)` only when valid. |
| `Errors()` | Returns `(NonEmpty[E], true)` only when invalid. |
| `Map(value, transform)` | Transforms only a valid value. |
| `MapErrors(value, transform)` | Transforms errors only for an invalid value. |
| `Map2(a, b, combine)` | Combines two independent validations. |
| `Match(value, invalidFn, validFn)` | Calls `invalidFn` with the errors or `validFn` with the valid value and returns its result. |

## Laziness and `Map2`

`Map` does not call its value transform for an invalid value. `MapErrors` does
not call its error transform for a valid value.

`Match` is the eliminator: it calls exactly one handler, `invalidFn` with
every error in order for an invalid value (the zero Validated passes
`[zero(E)]`) and `validFn` for a valid value. The zero-value branch comes
first, and handlers are called directly without a nil check.

`Map2` calls `combine` only when both inputs are valid. Its four cases are:

| Left | Right | Result |
|---|---|---|
| valid | valid | A valid combined value. |
| invalid | valid | The left errors. |
| valid | invalid | The right errors. |
| invalid | invalid | Every left error followed by every right error. |

The left-then-right error order is deterministic. It lets callers present
validation feedback in the same order as their independent checks.

## Why some operations are absent

`FlatMap` is intentionally absent: a later validation that depends on a prior
successful value cannot be run independently, so it has no unambiguous
accumulation rule. `Map3+` helpers are also absent; compose the declared
`Map2` operation instead. There is no public Semigroup API or conversion API.

For dependent, fail-fast sequences, use `Result[T, E]`: its first error stops
later successful work. `Validated` instead keeps independent validation errors
and accumulates them with `Map2`.

## Allocation note

Accumulating two invalid values defensively builds an error collection, so that
path can allocate. This is expected: it preserves the value semantics and the
left-then-right order of every collected error.
