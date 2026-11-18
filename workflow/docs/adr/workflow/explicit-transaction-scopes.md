---
type: adr
id: adr:workflow:explicit-transaction-scopes
context: workflow
title: Explicit transaction scopes and scoped capabilities
status: Accepted
tags: [workflow, transaction, capabilities, context, boundary]
deciders: [alvadorn]
related: [adr:workflow:functional-dependency-injection-through-capabilities, adr:workflow:no-distributed-transaction-abstraction, adr:workflow:transactional-outbox-as-application-composition, tdd:workflow:transaction]
date: 2026-08-10
---
# Explicit transaction scopes and scoped capabilities

## Context and problem

A workflow needs to write several things atomically without knowing which technology provides atomicity, and without a stray write escaping the transaction. The common shortcut, hiding the transaction handle in `context.Context`, makes that escape invisible and untypable.

## Decision drivers

- the concrete handle never reaching a workflow signature;
- writes that cannot accidentally run outside the boundary;
- full type safety for the result, despite the limits of Go generics;
- an honest claim of atomicity.

## Decision

`Boundary[S] = func(context.Context, func(context.Context, S) error) error` is the region supplied by the adapter, and the free function
`Within[S, T](boundary, func(S) action.Action[T]) action.Action[T]` runs a typed computation inside it. The scope `S` is defined by the application and carries capabilities, never a handle. `context.Context` carries lifecycle (cancellation and deadlines) and never a transaction.

## Alternatives considered

- **A `Within` method on a transaction type.** Rejected: Go forbids a method from introducing the extra type parameter `T`, and every workaround costs `any`, reflection, or type assertions. Type safety wins over fluent form.
- **`Action[T] -> Action[T]` as the boundary's shape.** Rejected: the inner computation would have no way to reach transaction-bound capabilities.
- **A generic registry of scopes indexed by name.** Rejected: `map[string]any` erases exactly the types this decision exists to preserve.

## Consequences

Capability narrowing becomes expressible: the application can offer read-only capability outside the boundary and write capability only inside it, so a misplaced write fails to compile. `Boundary` itself promises no ACID; the adapter documents its guarantee. Nested `Within` calls have no special meaning; savepoints and nested transactions belong to the adapter.

## Non-goals

No savepoints, no `UnitOfWork` (which implies identity maps, change tracking, and entity lifecycles this design does not have), and no ambient or default transaction.
