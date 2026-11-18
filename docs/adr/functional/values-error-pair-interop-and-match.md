---
type: adr
id: adr:functional:values-error-pair-interop-and-match
context: functional
title: Result interop with the (T, error) pair and the Match argument-order rule
status: Accepted
tags: [functional, values, result, error-handling, public-api]
deciders: [alvadorn]
related: [adr:functional:zero-dependency-boundary, adr:functional:values-family-dependency-boundary, adr:workflow:functional-integration-without-effect-duplication]
date: 2026-10-20
---
# Result interop with the (T, error) pair and the Match argument-order rule

## Context and problem

`values/result` models success and failure, but every Go API returns
`(T, error)`. Without a bridge, each consumer rewrites
`if err != nil { return result.Err... }` at the entry and a pair of
`Get`/`ErrValue` calls at the exit. Also, only `either` has an eliminator
(`Match`); `result`, `optional`, and `validated` force the caller to branch
by hand over `Get`.

The family documentation (`values.md`, `values-result.md`) excluded
"conversions" without distinguishing conversion between family types from
interop with Go's error protocol. The
[workflow integration decision](../../../workflow/docs/adr/workflow/functional-integration-without-effect-duplication.md)
had already ruled out automatic `FromResult` and `ToResult` bridges in `workflow`.

## Decision drivers

- Keep `values/result` stdlib-only, with no new imports
  (zero-dependency boundary, values family dependency boundary).
- Never turn a failure into a success silently.
- One single argument-order rule for every family `Match`.
- Zero allocation cost on the wrapper, like the rest of the family.

## Options considered

### Option 1: A `r.ToPair()` method and `FromPair` as constructor

Rejected: a method on `Result[T, E]` cannot pin `E = error`.
Package functions pin the type in the parameter.

### Option 2: `ToPair` returns `(zero, nil)` for `Err(nil)`

Rejected: a zero `Result[T, error]` (for example, an uninitialized field)
would become a success. That is the behavior of checkout's `unwrap`, and it
hides failures.

### Option 3: `ToPair` returns a sentinel error for `Err(nil)`

Rejected: it invents a failure the caller never produced and creates a new
exported symbol. The state is a programming error, like `MustGet` on `Err`.

### Option 4: Conversions between family types (Result ↔ Either ↔ Validated)

Rejected and kept out of scope: the names are semantic commitments
(values family dependency boundary). `FromPair`/`ToPair` do not convert
between family types; they connect `Result` to the language's `(T, error)`
protocol.

### Option 5: `FromPair`, `ToPair`, and `Match` as package functions (chosen)

## Decision

Add to `values/result`:

```go
func FromPair[T any](value T, err error) Result[T, error]
func ToPair[T any](r Result[T, error]) (T, error)
func Match[T, E, R any](r Result[T, E], errFn func(E) R, okFn func(T) R) R
```

**`FromPair` semantics.** It only checks `err != nil`: `nil` produces
`Ok(value)` (including a zero value); non-nil produces `Err(err)` and
discards `value`. An interface holding a typed nil pointer
(`error((*E)(nil))`) is non-nil and produces `Err`. The documentation warns
about two unsuitable cases: functions whose last result is a concrete
pointer error type (`FromPair(f())` converts the concrete nil into a
non-nil `error` and turns success into failure, with no compile error), and
`(n, err)` APIs with a partial value, which is discarded.

**`ToPair` semantics.** `Ok(v)` becomes `(v, nil)`; `Err(e)` with non-nil
`e` becomes `(zero, e)`. `Err(nil)`, including the zero
`Result[T, error]`, panics with exactly:

```
functional/values/result: ToPair on Err with nil error
```

The prefix follows the existing `MustGet` panic.

**The `Match` order rule.** The zero-value branch handler comes first, as in
`either.Match(e, leftFn, rightFn)` and the family's zero-value table:
`result.Match(r, errFn, okFn)`, `optional.Match(o, noneFn, someFn)`, and
`validated.Match(v, invalidFn, validFn)`. Success-first order (like
`modeling`'s `policy.Choose`) was dropped to keep a single rule.
`optional` and `validated` `Match` follow this rule when added.

**Nil handlers.** Like the rest of the family, `Match` calls the handlers
directly, with no nil check; exactly one handler runs per call.

**Boundary with `workflow`.** This decision leaves the rejected `FromResult`
and `ToResult` bridges out of `workflow`.
`FromPair`/`ToPair` live in `functional`, connect `Result` to Go's
`(T, error)`, and create no `workflow` → `functional` edge; a consumer
writes `result.FromPair(act(ctx))` without `workflow` importing anything.

## Consequences

**Positive:** consumers delete local helpers like `unwrap`; the family gains
an eliminator with a predictable order; `FromPair`, `ToPair`, and `Match` do
not allocate on the wrapper.

**Negative:** the `Match` order and the `ToPair` panic message become
permanent contract after the release. A `Match` with swapped handlers only
compiles for `Result[X, X]`. The concrete error type trap is not detectable
at compile time; it is documented and covered by tests.

**Verification:** table tests and round-trip laws in
`values/result`, zero-allocation tests, `TestValuesPublicAPI` in the
archtest, `go test ./...`, and lint.
