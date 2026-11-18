---
type: adr
id: adr:workflow:workflow-as-reusable-effectful-composition
context: workflow
title: Workflow as reusable effectful composition
status: Accepted
tags: [workflow, composition, effects, generics, package-boundary]
deciders: [alvadorn]
related: [adr:workflow:action-as-deferred-context-bound-computation, adr:workflow:functional-dependency-injection-through-capabilities, tdd:workflow:core]
date: 2026-08-10
---
# Workflow as reusable effectful composition

## Context and problem

Application code repeatedly mixes three unrelated concerns: acquiring dependencies, receiving an input, and running effects. When the three collapse into a single function, the operation cannot be reused with different inputs, tested without infrastructure, or composed with a neighboring operation without duplicating it.

## Decision drivers

- reusing a composed operation as a first-class value;
- separating inputs from dependencies;
- separating execution from description;
- interoperability with ordinary Go functions;
- no runtime, container, or registry.

## Decision

`Workflow[I, O] = func(I) action.Action[O]` is the module's central abstraction. Dependencies are captured when the workflow is constructed, the input arrives when it is applied, and the context arrives only at execution. Composition is binary and type-preserving: `Then(Workflow[A,B], Workflow[B,C]) Workflow[A,C]`.

## Alternatives considered

- **An interface with an `Execute` method.** Rejected: it forces a nominal type per operation, blocks the free-function combinators, and turns capability injection into constructor plumbing.
- **A variadic `Then(a, b, c, d)`.** Rejected: Go generics cannot preserve an intermediate type that changes across a variadic list, and the only workarounds are `any`, reflection, or runtime type assertions.
- **Hiding `Action` behind `Workflow`.** Rejected: capabilities would no longer be able to return a composable effect, which is what makes functional injection viable.

## Consequences

A `Workflow` is not a handler, controller, service object, job, durable workflow, or state machine, though any of those can run one. Composed workflows are indistinguishable from primitives, so a consumer never learns the composition tree. Because a `Workflow` holds no execution state, one value serves any number of inputs, contexts, and goroutines.

## Non-goals

Workflow persistence, resume, replay, workflow identifiers, durable timers, and scheduling are permanently out of scope; this is in-process composition, not an orchestration engine.
