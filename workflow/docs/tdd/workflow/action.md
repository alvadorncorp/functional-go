---
type: tdd
id: tdd:workflow:action
context: workflow
title: Action representation and execution
status: Accepted
tags: [workflow, action, composition, resources, context, errors, panic]
related: [adr:workflow:action-as-deferred-context-bound-computation, adr:workflow:composition-and-resource-safety, adr:workflow:sequential-traverse-over-slices, tdd:workflow:core, tdd:workflow:composition]
date: 2026-08-11
---
# TDD: Action representation and execution

## Technical scope

`Action[T] = func(context.Context) (T, error)`. It is a re-runnable value with
no internal state; only `Run` executes the closure. Every combinator forwards
the same context and never creates a goroutine, a runtime, memoization, or a
recovery.

## Public contracts

```go
func Run[T any](context.Context, Action[T]) (T, error)
func Do[T any](func(context.Context) (T, error)) Action[T]
func Exec(func(context.Context) error) Action[struct{}]
func Pure[T any](T) Action[T]
func Fail[T any](error) Action[T]
func Map[A, B any](Action[A], func(A) B) Action[B]
func MapFallible[A, B any](Action[A], func(A) (B, error)) Action[B]
func FlatMap[A, B any](Action[A], func(A) Action[B]) Action[B]
func Recover[T any](Action[T], func(error) Action[T]) Action[T]
type Pair[A, B any] struct { First A; Second B }
func Zip[A, B any](Action[A], Action[B]) Action[Pair[A, B]]
func Both[A, B, C any](Action[A], Action[B], func(A, B) C) Action[C]
func Traverse[A, B any]([]A, func(A) Action[B]) Action[[]B]
func Tap[T any](Action[T], func(T) Action[struct{}]) Action[T]
func Ensure[T any](Action[T], Action[struct{}]) Action[T]
type AttemptResult[T any] struct { Value T; Err error }
func Attempt[T any](Action[T]) Action[AttemptResult[T]]
func Bracket[R, T any](Action[R], func(R) Action[T], func(R) Action[struct{}]) Action[T]
func AcquireUseRelease[R, T any](Action[R], func(R) Action[T], func(R) Action[struct{}]) Action[T]
```

`Do` adapts functions with a result; `Exec` adapts effects that return only an
error, without introducing another unit type. Neither executes the function at
construction time: only `Run` calls it, once per execution. `Exec` rejects a
nil function immediately with `workflow/action: nil function`, like `Do`, and
preserves the received context, the error identity, and any panic. It does not
check for cancellation before calling the function: cooperation with the
context stays in the callback.

## Ordering and short-circuiting

`Zip` and `Both` run the left Action first, then the right one. `Both` calls
the combiner only after two successes. `Tap` runs the effect only after the
source succeeds. Value independence does not imply parallelism.

`Traverse` binds and runs one item at a time, in index order: the binder for
the next item runs only after the previous item's Action succeeds. Every `Run`
walks the items again. Success returns a new, non-nil slice with `len(items)`
values in item order; nil or empty input returns an empty, non-nil slice
without calling the binder. The first failure returns `(nil, err)` with the
error unchanged and does not call the binder for the remaining items. There is
no context check between items. Construction copies `items` (a shallow copy:
pointer elements stay shared).

`Traverse` allocation profile: one copy of `items` at construction, one result
slice per `Run`, and whatever the binder and the Actions it returns allocate.

## Errors and finalization

A lone error keeps its identity. `Ensure` runs the finalizer after every normal
return of the source. `Bracket` runs release if and only if acquire succeeds,
and also after a normal use error. When both source/use and finalizer/release
fail, `errors.Join(primaryErr, finalizerErr)` keeps both for
`errors.Is`/`errors.As`; any error yields a zero `T`.

`Attempt` turns the normal `(value, err)` pair into an `AttemptResult` and
returns a nil outer error. It recovers no panic and implements no best-effort
policy.

## Context, panic, and nil

The context that `Run` receives, canceled or not, reaches every step unchanged.
Shielding belongs to the adapter. No combinator calls `recover`; `Ensure`
promises no finalizer after a panic, and `Bracket` promises no release after a
panic. Nil actions and callbacks fail at construction; callbacks that return a
nil Action fail during execution with `workflow/action: nil action`. A nil
`Traverse` binder fails at construction with `workflow/action: nil binder`,
even with nil `items`.
