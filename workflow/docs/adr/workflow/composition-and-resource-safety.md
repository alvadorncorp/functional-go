---
type: adr
id: adr:workflow:composition-and-resource-safety
context: workflow
title: Advanced composition and normal-return resource safety
status: Accepted
tags: [workflow, action, composition, resources, errors, context]
deciders: [alvadorn]
related: [adr:workflow:action-as-deferred-context-bound-computation, adr:workflow:explicit-transaction-scopes, adr:workflow:functional-integration-without-effect-duplication, tdd:workflow:action, tdd:workflow:composition]
date: 2026-08-11
---
# Advanced composition and normal-return resource safety

## Context and problem

`Workflow` and `Action` already describe deferred, re-executable effects bound to the `context.Context`, but composition still requires manual wrappers to adapt inputs, combine independent computations, observe success, turn an error into data, and scope a resource's lifecycle.

## Decision drivers

- keeping `Run` as the only passage from description to execution;
- widening expressiveness without creating an engine, runtime, scheduler, or container;
- keeping the core stdlib-only and the DAG `workflow -> action <- transaction`;
- making order, short-circuit, finalization, and double failure observable by tests;
- keeping infrastructure and cleanup policy in the adapters.

## Decision

The root gets `Contramap` to adapt the input type and `Provide` to fix an input and yield an `Action`. Applying a workflow still does not run the effect; only `Run` executes the resulting Action.

`action` gets `Pair`, `Zip`, `Both`, `Tap`, `Ensure`, `AttemptResult`, `Attempt`, `Bracket`, and `AcquireUseRelease`. `Zip` and `Both` always run the left Action first, then the right. Independence means only that the second value does not depend on the first; it does not authorize parallelism.

`Tap` observes success only. `Ensure` runs a finalizer after every normal `(T, error)` return. `Attempt` converts only the normal error channel into data. `Pair` and `AttemptResult` stay carriers with fields, without methods.

`Bracket` is the canonical acquire/use/release implementation; `AcquireUseRelease` delegates to it. Release runs if and only if acquire returns success, including after a normal error from use. Single errors keep identity; double failure uses `errors.Join(primaryErr, finalizerErr)`.

No combinator calls `recover`, starts a goroutine, derives a context, or applies `context.WithoutCancel`. Panics still indicate a programming error, and there is no finalization/release guarantee on the panic path. The same context given to `Run`, even cancelled, reaches the finalizer/release; shielding belongs to the adapter.

`Bracket` does not replace `transaction.Within`: resource lifecycle promises no begin, commit, rollback, or atomicity.

## Alternatives considered

- **Running `Zip`/`Both` in parallel.** Rejected: it would require a runtime, synchronization, and a cancellation policy outside the scope.
- **Using `defer` + `recover` for cleanup on panic.** Rejected: it would create a third control channel unlike the rest of `Action`.
- **Shielding release with `context.WithoutCancel`.** Rejected: that is policy of the concrete resource.
- **Importing tuples/results from `functional`.** Rejected: it would break the stdlib-only closure and duplicate responsibility.
- **Merging `Bracket` and `Within`.** Rejected: lifecycle and transaction have different contracts.

## Consequences

The public surface grows additively and must be exercised by compiled examples. The `action` package starts importing `errors` besides `context`, still within the standard library. Architecture and behavioral tests pin sequential order, minimal carriers, absence of goroutines/recovery, unchanged context, and a single lifecycle implementation.

## Non-goals

Engine, runtime, scheduler, parallelism, automatic observability, service locator, DI container, adapters, distributed transaction, panic-safe cleanup, and a new functional algebra of values stay outside the module.

> Update: the blanket exclusion of retry and timeout from this first algebra was
> superseded by [Explicit resilience as Action transformation](explicit-resilience-as-action-transformation.md)
> (`adr:workflow:explicit-resilience-as-action-transformation`). The exclusions of
> engine, runtime, background work, implicit parallelism, and durable execution remain valid.
