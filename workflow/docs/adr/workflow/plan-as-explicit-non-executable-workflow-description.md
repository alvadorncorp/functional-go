---
type: adr
id: adr:workflow:plan-as-explicit-non-executable-workflow-description
context: workflow
title: Plan as explicit non-executable workflow description
status: Accepted
tags: [workflow, plan, description, tooling, immutability]
deciders: [alvadorn]
related: [adr:workflow:workflow-as-reusable-effectful-composition, adr:workflow:action-as-deferred-context-bound-computation, adr:workflow:composition-and-resource-safety]
date: 2026-08-11
---
# Plan as explicit non-executable workflow description

## Context and problem

`Workflow[I,O]` builds an `Action[O]`, and `Action` remains an opaque closure executed only by `Run`. Documentation, visualization, and architecture tests need to describe the conceptual structure of an operation without turning those closures into an AST, a runtime, or an executable DSL.

## Decision

We create `workflow/plan` as a stdlib-only package. `Plan` is immutable data, declared explicitly by the application and inspected through `Kind`, `Name`, and `Children`. The initial surface contains only Empty, Step, Sequence, and Transaction.

The zero value is Empty. `Step` preserves a non-empty name. `Sequence()` is Empty, `Sequence(p)` is `p`, and larger sequences preserve order and nesting. `Transaction` always preserves the conceptual node, even with zero or one child. Constructors and `Children` make defensive copies.

Partial amendment (2026-10-23): read-only inspection also includes `String`, on `Plan` and on `Kind`. `Plan.String` returns a canonical description of the declared structure, with literal, unescaped `Step` names and no promise of reading it back (parsing). This does not make `Plan` executable and does not change the other decisions in this ADR.

Plan takes no context, does not execute an `Action`, does not call `Run`, does not open a transaction, does not apply resilience, does not start a goroutine, and performs no I/O.

## Consequences

- Plan can exist without Workflow, and Workflow can exist without Plan.
- The association between them stays opt-in and belongs to the application.
- Plan knows nothing about `action`, the `workflow` root, transport, or infrastructure.
- Tooling can inspect a stable tree without observing execution history.

## Rejected alternatives

- **Inferring Plan from closures.** Would require reflection, runtime metadata, or unsafe.
- **Compiling Plan to Action.** Would create a second executable language.
- **Storing Plan inside Workflow.** Would break the existing functional contract.
- **Adding Visitor, Resource, metadata, or JSON now.** Would freeze surface without demonstrated consumption.

## Non-goals

Plan is not Workflow, Action, `modeling.Specification`, execution history, runtime AST, scheduler, registry, or durable workflow definition.
