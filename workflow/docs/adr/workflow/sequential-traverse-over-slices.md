---
type: adr
id: adr:workflow:sequential-traverse-over-slices
context: workflow
title: Sequential Traverse over slices
status: Accepted
tags: [workflow, action, composition, traverse, order, outbox]
deciders: [alvadorn]
related: [adr:workflow:action-as-deferred-context-bound-computation, adr:workflow:functional-integration-without-effect-duplication, adr:workflow:composition-and-resource-safety, adr:modeling:streaming-replay-project-seq, adr:functional:functional-primitives-and-fold-boundary, tdd:workflow:action]
date: 2026-10-20
---
# Sequential Traverse over slices

## Context and problem

Confirming a decision iterates the emitted events and, for each one, runs an Action (encode and append to the outbox), stopping at the first error. The `action` algebra has no combinator for "one Action per item, in order", so every application writes the same manual loop with `action.Run` inside an outer Action.

The pattern guide already names the repetition
(`docs/patterns/load-decide-confirm.md:17`): iterate `decision.Events()` and append
**sequentially** to the application's outbox. The loops are identical in order, short-circuit, and context, but each copy reimplements those rules without a test of its own.

[Functional integration without effect duplication](functional-integration-without-effect-duplication.md) rejects monad families and higher-kinded emulation; [advanced composition](composition-and-resource-safety.md) rejects parallelism and a new functional algebra of values. A `Traverse` has to fit inside those limits.

## Decision drivers

- effect order observable and equal to item order;
- `Action` stays deferred and re-executable (`adr:workflow:action-as-deferred-context-bound-computation`): each `Run` redoes the same work from the same input;
- the same conventions as `FlatMap`, `Zip`, and `Both`: short-circuit, error with identity preserved, panic at construction for a nil argument and panic at execution for a nil Action;
- `action` keeps importing exactly `context` and `errors`;
- no goroutine, no `recover`, and no hidden context policy.

## Decision

`action` gets:

```go
func Traverse[A, B any](items []A, binder func(A) Action[B]) Action[[]B]
```

`Traverse` is monomorphic: it operates on `[]A` and `Action`, with no `Traversable` interface, no container type parameter, and no higher-kinded emulation. This reconciles the non-goal of the functional-integration ADR. It is strictly sequential and creates no algebra of values, which reconciles the non-goal of the composition ADR: it is effect composition over the same `Action` that `Zip` and `Both` already compose.

Semantics:

- items are bound and executed in index order, interleaved: the binder of item `i+1` is called only after the Action of item `i` succeeded, and every Action receives the context given to `Run`;
- nothing executes at construction; each `Run` walks the items again;
- construction copies `items` with `append([]A(nil), items...)`; the copy is shallow, so pointer elements keep sharing their target;
- success returns a new, non-nil slice with `len(items)` values; nil or empty input returns an empty non-nil slice without calling the binder;
- the first failure returns `(nil, err)` with the error unchanged, and the binder is not called for the remaining items;
- there is no `ctx` check between items: the context, even cancelled, reaches the Actions unchanged, which observe cancellation like any other;
- a nil binder panics with `workflow/action: nil binder` at construction, even with nil `items`; a binder returning a nil Action panics with `workflow/action: nil action` at execution. No new panic constants.

## Alternatives considered

- **`iter.Seq[A]` input.** `adr:modeling:streaming-replay-project-seq` adopted `iter.Seq` in `Replay` and `Project` to avoid materialization in a single-pass fold. Rejected here: an `Action` is re-executable and an `iter.Seq` does not promise to yield the same items on a second pass. Copying into a slice at construction fixes what each `Run` executes; re-executability wins over saving one allocation. The caller materializes the events of `decision.Events()` explicitly before calling `Traverse`, and order stays guaranteed because the appends are sequential in index order.
- **Names `ForEach` or `Each`.** Rejected: they suggest effect-only iteration that discards values, while the combinator collects one result per item. `Traverse` names exactly "map with effect and collect".
- **Accumulating failures with `errors.Join`.** Rejected: it would require running the following items after a failure, which in an outbox means writing messages after an error inside the transaction. Error accumulation belongs to `functional/values/validated`; here the first error keeps identity.
- **Parallel execution.** Rejected for the same reason the composition ADR rejects it for `Zip` and `Both`: it would require goroutines, synchronization, and a cancellation policy, and the order of the appends is part of what the consumer observes.
- **Checking `ctx.Err()` between items.** Rejected: that would be a context policy hidden in the combinator. The composition ADR fixes that the same context, even cancelled, reaches every step; what decides to stop is the Action or the adapter.
- **Copying with `slices.Clone`.** Rejected: it would add `slices` to the imports of `action`, which the architecture gate keeps at exactly `context` and `errors`. `append([]A(nil), items...)` is equivalent.
- **Putting `Traverse` in `functional`.** Rejected: `functional` has no effect system (functional-integration ADR), and `adr:functional:functional-primitives-and-fold-boundary` keeps its primitives as pure leaves over values. `Action` carries `context.Context` and execution semantics, so the combinator lives beside it.

## Consequences

The surface of `action` grows additively by one function, pinned by the surface and contract gates and exercised by a compiled example. The allocation profile is one copy of `items` at construction, one result slice per `Run`, and the binder's own allocations. Callers with `iter.Seq` pay an explicit materialization; the order invariant still holds because execution is sequential. The shallow copy and the absence of context checks are documented in the contract. Amendment (2026-10-23): the withdraw `appendOutbox` now uses `Traverse` over the materialized events, like checkout. The `for` loops with `action.Run` in crosscontext stay on purpose, for different reasons. In `distributed`, `appendIntegrationMessages` filters by event (only `TransferOccurred` goes to the outbox), which makes `Traverse` awkward there. In `shared`, `appendAll` does not filter: it writes one message per event from two distinct decisions, one after the other, and the imperative loop is idiomatic Go. No pending adoption.

## Non-goals

Parallel `Traverse`, variants over `iter.Seq` or maps, error accumulation, `Sequence` over `[]Action[T]`, a `Traversable` typeclass, and built-in retry or timeout stay outside the module.
