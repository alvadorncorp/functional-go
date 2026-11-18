---
type: adr
id: adr:workflow:action-as-deferred-context-bound-computation
context: workflow
title: Action as deferred context-bound computation
status: Accepted
tags: [workflow, action, effects, context, laziness]
deciders: [alvadorn]
related: [adr:workflow:workflow-as-reusable-effectful-composition, adr:workflow:functional-integration-without-effect-duplication, tdd:workflow:action]
date: 2026-08-10
---
# Action as deferred context-bound computation

## Context and problem

A composable effect needs a representation. Introducing a bespoke effect type would force every consumer to learn a new failure channel and a new cancellation mechanism, both of which already exist in Go.

## Decision drivers

- zero new vocabulary for failure and cancellation;
- construction that demonstrably performs no work;
- an escape hatch back to ordinary imperative Go;
- the smallest representation that still supports composition.

## Decision

`Action[T] = func(context.Context) (T, error)`. `Run(ctx, action)` is the only passage from description to execution. Construction and composition produce no effect; only `Run` does. The type is a function, not a struct, and therefore carries no metadata, identifier, or hooks.

## Alternatives considered

- **A struct with a `run` field plus metadata.** Rejected: nothing in the module needs the extra fields, and each would allocate on every composition.
- **A `Result`-based failure channel.** Rejected: `(T, error)` keeps `errors.Is`, `errors.As`, and the standard cancellation sentinels working unchanged.
- **Memoizing the result.** Rejected: that is `functional/structural/lazy`, a different concept. An `Action` is re-executable by design.

## Consequences

`construction ≠ execution` is an enforced property, not a convention: a test with a counter asserts zero callbacks during construction. Running the same `Action` twice runs the effect twice. An `Action` is not a future or a promise and does not start a goroutine. Panics are never converted into errors, and no combinator calls `recover`.

## Non-goals

`Await`, `Start`, `Spawn`, `Promise`, `Future`, and `Memoize` are excluded, and an architecture test rejects those names.
