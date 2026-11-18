---
type: adr
id: adr:workflow:modeling-integration-boundary
context: workflow
title: Modeling integration boundary
status: Accepted
tags: [workflow, modeling, domain, boundary, purity]
deciders: [alvadorn]
related: [adr:workflow:workflow-as-reusable-effectful-composition, adr:workflow:functional-integration-without-effect-duplication, tdd:workflow:modeling-integration]
date: 2026-08-10
---
# Modeling integration boundary

## Context and problem

`modeling` owns pure domain semantics: `Requirement`, `Transition`, `Invariant`, `Decision`, `Policy`. `workflow` owns coordination with effects. The question is which module depends on which, and whether `workflow` should ship helpers for `Decision`.

## Decision drivers

- keeping `modeling` pure and unaware of execution;
- no duplication of the domain metamodel inside `workflow`;
- evidence before abstraction.

## Decision

The dependency edge `workflow -> modeling` is not created. The application is the composition layer that imports both, which makes them sibling modules rather than a stack. `modeling -> workflow` is forbidden and enforced by a gate. `workflow` does not define `Rule`, `Requirement`, `Transition`, `Invariant`, `Policy`, or `Decision`.

## Alternatives considered

- **A `workflow/integration/modeling` bridge in this delivery.** Deferred, not rejected: the withdrawal and cross-context examples showed that the load-decide-confirm boilerplate differs enough across domains that a shared `CommitDecision[S,E,O]` would encode one architectural style as if it were general. Extraction waits for at least three domains agreeing on the shape.
- **`workflow` importing `modeling` for convenience.** Rejected: it would couple the version lines and buy nothing the application cannot express.

## Consequences

The canonical shape is load with effect, pure decision, confirm with effect, with the workflow calling an ordinary domain function such as
`domain.Withdraw(account, cmd) (modeling.Decision[Account, Event], error)`.
Business thresholds live in the domain; a workflow that branches does so over a value the domain produced.

## Non-goals

No event sourcing assumption, no aggregate abstraction, and no requirement that a `Decision` be persisted.
