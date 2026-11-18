---
type: tdd
id: tdd:workflow:plan-resilience
context: workflow
title: Plan and resilience contracts
status: Accepted
tags: [workflow, plan, resilience, retry, timeout, context, contracts]
related: [adr:workflow:plan-as-explicit-non-executable-workflow-description, adr:workflow:explicit-resilience-as-action-transformation, adr:workflow:action-as-deferred-context-bound-computation, tdd:workflow:action, tdd:workflow:core, tdd:workflow:transaction]
date: 2026-10-23
---
# TDD: `plan` and `resilience` contracts

## Technical scope

This design details the contracts of two production packages of the module:
`workflow/plan` (`adr:workflow:plan-as-explicit-non-executable-workflow-description`) and `workflow/resilience` (`adr:workflow:explicit-resilience-as-action-transformation`). The three
remaining packages, `workflow`, `workflow/action`, and `workflow/transaction`,
are covered by [`tdd:workflow:core`](core.md),
[`tdd:workflow:action`](action.md), and
[`tdd:workflow:transaction`](transaction.md).

The workflow module has no SDD: `docs/sdd/` does not exist and no document
declares `sdd:workflow:*`. This TDD therefore anchors directly in the ADRs and
references no SDD.

```text
plan                     (leaf: slices and strings only)
workflow ──► action ◄── transaction
                 ▲
                 └────── resilience   (context, action, and time only)
```

`plan` imports no package of the module; `action` is the sink. `resilience`
depends on exactly `context`, `action`, and `time`.

## Public contracts

```go
// plan
type Kind uint8 // KindEmpty, KindStep, KindSequence, KindTransaction
type Plan struct{ /* unexported fields */ }
func Step(name string) Plan
func Sequence(plans ...Plan) Plan
func Transaction(plans ...Plan) Plan
func (Plan) Kind() Kind
func (Plan) Name() string
func (Plan) Children() []Plan
func (Plan) String() string
func (Kind) String() string

// resilience
type Backoff func(attempt int) time.Duration
type RetryPolicy struct {
	Attempts int
	When     func(error) bool
	Backoff  Backoff
}
func Retry[T any](policy RetryPolicy, source action.Action[T]) action.Action[T]
func Timeout[T any](duration time.Duration, source action.Action[T]) action.Action[T]
```

`TestExactPublicSurface` pins the exported surface.

## `plan` invariants

Pinned by `plan/plan_test.go`, `plan/example_test.go`, and the
`internal/archtest` gates:

- the zero value of `Plan` is `KindEmpty`, with no name and
  `Children() == nil`;
- `Step("")` panics with `"workflow/plan: empty step name"`; non-empty names
  are preserved verbatim, spaces included;
- `Sequence()` and `Sequence(empty)` reduce to `KindEmpty`; `Sequence(x)`
  returns `x`; two or more inputs preserve order and nesting, with no
  flattening;
- `Transaction(...)` always preserves the node, including empty and
  single-child;
- the constructors and `Children()` clone the slice: mutating the input or the
  result does not change the `Plan`;
- `Kind.String()` returns `empty`, `step`, `sequence`, `transaction`, and
  `unknown` for any undefined value;
- `Plan.String()` is canonical and deterministic, for example
  `sequence(a, b, transaction(c, d))` and `transaction()`, with no name
  escaping and no round-trip promise; calling `String()` does not change the
  `Plan`;
- concurrent reads of a single `Plan` are safe;
- `Plan` exports no fields and exposes only `Children`, `Kind`, `Name`, and
  `String` (`TestPlanIsOpaqueAndNonExecutable`), has no `Run` or `Execute`, and
  `plan` imports exactly `slices` and `strings`: describing neither executes
  nor opens a transaction.

## `resilience` invariants

Pinned by `resilience/retry_test.go`, `retry_attempts_test.go`, and
`timeout_test.go`.

### Construction

- `Retry` and `Timeout` are lazy: construction does not call the source;
- invalid arguments are programming errors and panic at construction, with
  exact messages and without calling the source or the predicate:
  `"workflow/resilience: invalid attempts"` (`Attempts <= 0`),
  `"workflow/resilience: nil predicate"` (`When == nil`),
  `"workflow/resilience: nil action"` (a nil source in `Retry` or `Timeout`),
  and `"workflow/resilience: invalid timeout"` (a duration `<= 0`);
- a nil `Backoff` is valid.

### `Retry`

- `Attempts` includes the first call; the source runs at most `Attempts`
  times, stopping at the first success;
- `When` is consulted only after a failure that still has an attempt left:
  with `Attempts` exhausted it runs `Attempts-1` times; a false predicate ends
  the loop without backoff;
- `Backoff` receives the number of the attempt that just failed, 1-based
  (`[1 2]` for three attempts); a nil `Backoff` and a zero delay retry
  immediately;
- a negative delay panics with `"workflow/resilience: negative backoff"` after
  the source and the predicate have run once;
- the loop decides in this order: (1) check `ctx.Err()`; (2) call the source;
  (3) success returns the value; (4) with `Attempts` exhausted or `When` false,
  the source error is returned directly; (5) only then come the backoff and the
  next attempt. The source error comes out unwrapped and unaggregated, and the
  value is the zero of `T`; there is no `errors.Join` between attempts or
  between the source error and the context error;
- the source error is the result only when the source itself ends the loop
  (step 4). In that case the context is not consulted after the call: a
  cancellation or deadline that happens during or after the last attempt, or
  after a false `When`, does not replace the source error, and a source success
  keeps being returned even if the context was canceled during the call. Parent
  cancellation does not override a terminal source result;
- `Retry` returns exactly `ctx.Err()`, `context.Canceled` or
  `context.DeadlineExceeded` with no wrapping, only when it observes the parent
  context stopped at one of two checkpoints: entering an attempt (step 1,
  including the first, without calling the source, the predicate, or the
  backoff) or waiting out the backoff (step 5). After a retryable failure with
  an attempt left, an already-stopped parent is observed at the wait when the
  delay is positive or zero (a zero delay only checks `ctx.Err()`) and, with a
  nil `Backoff`, at entry to the next attempt; in none of these cases does the
  source run again. The earlier retryable error is discarded in that case;
- panics from the source, from `When`, and from `Backoff` escape unchanged;
- a single constructed `Retry` can run concurrently.

### `Timeout`

- each execution derives a child context with a `duration` deadline from the
  received `ctx`; parent context values stay visible;
- the source receives the child, never the parent itself, and its value and
  error return unchanged;
- a parent canceled before execution reaches the source canceled;
- a source that observes `ctx.Done()` and returns `ctx.Err()` produces
  `context.DeadlineExceeded` at the deadline;
- source panics escape unchanged, and a single `Timeout` supports concurrent
  execution.

## Cooperative cancellation

`action.Run` passes the `ctx` to the Action unchanged and without checking it
(`action/action.go`). Nothing in the module interrupts a running function.
`TestTimeoutDoesNotPreemptIgnoringSource` proves that a source ignoring the
context keeps running past the deadline and that `Run` returns only when it
returns. Cancellation is cooperative: the source, or the adapter it calls, has
to observe `ctx`. `Retry` is the only partial exception, because it checks
`ctx.Err()` between attempts and during the backoff, never inside the source.

There is no production goroutine, no `recover`, and no automatic preemption.

## Budget composition

Pinned by `TestTimeoutOutsideRetryUsesOneTotalBudget` and
`TestRetryOutsideTimeoutCreatesPerAttemptBudgets`:

- `Timeout(d, Retry(p, src))` creates one total deadline for all attempts and
  waits; with a backoff longer than the deadline, the source runs once and the
  result is `context.DeadlineExceeded`;
- `Retry(p, Timeout(d, src))` creates a new child per attempt, each with its
  own deadline and context.

The order is the caller's explicit choice; the module offers no combined
policy.

## The `context.WithTimeout` exception

The context gate (`TestNoHiddenContextPolicy`) forbids in production
`Background`, `TODO`, `Value`, `WithValue`, `WithCancel`, `WithDeadline`,
`WithoutCancel`, `AfterFunc`, and `WithTimeout`, including the `Cause`
variants. The only exception is `context.WithTimeout` inside the `resilience`
package, used by `Timeout`. No other package creates, cancels, or shields a
context.

## What these packages do not do

- `Plan` does not execute, does not compile to an `Action`, and neither
  registers nor schedules;
- `Retry` and `Timeout` are not an execution engine: they persist no state and
  schedule nothing, and they include no jitter, circuit breaker, or error
  classification (`adr:workflow:explicit-resilience-as-action-transformation`);
- `Retry` does not make an operation idempotent; repeating effects requires
  the operation, or its adapter, to be safe to repeat.

## Note on the documentation gates

`TestDocumentationSpine` checks only that this file exists and that the
frontmatter carries the required keys, the status `Accepted`, and no `deciders`
key. The gate checks presence: it proves neither that the document list is an
exact allowlist nor that the text matches the tests cited above.
