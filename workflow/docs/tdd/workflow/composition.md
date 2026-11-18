---
type: tdd
id: tdd:workflow:composition
context: workflow
title: Workflow composition
status: Accepted
tags: [workflow, composition, inputs, generics, closures, reuse]
related: [adr:workflow:workflow-as-reusable-effectful-composition, adr:workflow:functional-dependency-injection-through-capabilities, adr:workflow:composition-and-resource-safety, tdd:workflow:action]
date: 2026-08-11
---
# TDD: Workflow composition

## Technical scope

```text
Workflow[I,O] + I       --Provide-->   Action[O]
Workflow[I,O] + (A→I)  --Contramap--> Workflow[A,O]
Workflow[A,B] + Workflow[B,C] --Then--> Workflow[A,C]
```

Dependencies enter at construction, input enters at application, and context
enters at `Run`. Applying a workflow may project the input and build an Action,
but it never runs the effect the Action describes.

## Public contracts

```go
func From[I, O any](func(I) action.Action[O]) Workflow[I, O]
func Map[I, A, B any](Workflow[I, A], func(A) B) Workflow[I, B]
func MapFallible[I, A, B any](Workflow[I, A], func(A) (B, error)) Workflow[I, B]
func Then[A, B, C any](Workflow[A, B], Workflow[B, C]) Workflow[A, C]
func Provide[I, O any](Workflow[I, O], I) action.Action[O]
func Contramap[A, I, O any](Workflow[I, O], func(A) I) Workflow[A, O]
```

## Laws

- `Contramap(wf, identity)` is observationally equivalent to `wf`;
- nested contramaps equal the composition of the mappers;
- `Provide(wf, input)` equals applying `wf(input)`, including result, error,
  context, and execution count;
- `Contramap` calls the mapper at application, not at construction;
- `Provide` and `Contramap` never call `Run` and never run effects.

## Limits

Composition stays binary and type-safe. There is no heterogeneous variadic
list, no reflection, no late dependency resolution, no DI container, no engine,
no runtime. `Provide` supplies input; it registers and resolves no
dependencies.
