---
type: adr
id: adr:workflow:no-ddd-runtime-module
context: workflow
title: No DDD/hexagonal runtime module
status: Accepted
tags: [workflow, modeling, functional, ddd, hexagonal, boundary]
deciders: [alvadorn]
related: [adr:modeling:minimal-type-safe-domain-modeling, adr:workflow:functional-dependency-injection-through-capabilities, adr:workflow:modeling-integration-boundary, adr:workflow:no-distributed-transaction-abstraction, adr:workflow:transactional-outbox-as-application-composition, adr:workflow:explicit-resilience-as-action-transformation, adr:modeling:specification-capability-boundary, adr:modeling:policy-composition-algebra, adr:functional:retire-optional-and-valueobject]
date: 2026-08-24
---
# No DDD/hexagonal runtime module

## Context and problem

Applications can compose `functional` (values), `modeling` (pure decision), and `workflow` (effects). The next temptation is a fourth module with `Aggregate`, `Entity`, `Repository`, `UnitOfWork`, or `CommandBus`.

## Decision drivers

- one concept, one owner;
- the application's own Go types for identity, aggregate, and money;
- executable gates instead of marker interfaces;
- not reopening accepted ADRs to accommodate a framework.

## Decision

No `ddd`, `hexagonal`, `aggregate`, or `repository` module with domain types is created. Help for DDD/hexagonal applications is pattern documentation, examples, and, only with its own trigger, an archtest kit, compensation combinators, or example slugs.

The following stay banned in the `workflow` core (and are not introduced into `modeling` or `functional`) as exports: `Repository`, `CommandBus`, `Outbox`, `Inbox`, `CommitDecision`, `Handle` as a base type, `Clock`, `Saga`, `SagaEngine`, `DistributedTransaction`, `EventStore`. `Xor` in `modeling/specification` and circuit breaker / jitter / bulkhead in `workflow/resilience` stay out
(`adr:modeling:specification-capability-boundary`, `adr:workflow:explicit-resilience-as-action-transformation`).
Money, Coupon, and identifiers live in the application
(`adr:modeling:minimal-type-safe-domain-modeling`).

Named future waves, not part of this ADR: an `archkit` kit (test-only); a `workflow/saga` package exporting `Compensate` (never the identifier `Saga`); `starter` and `readmodel` examples.

## Alternatives considered

- **A DDD runtime module.** Rejected: ceremony without domain validity, and in conflict with `adr:modeling:minimal-type-safe-domain-modeling` and `adr:workflow:functional-dependency-injection-through-capabilities`.
- **Just staying silent.** Rejected: the next session would reopen the question.

## Consequences

Guides in `workflow/docs/patterns/` and `functional/docs/patterns/` become the onboarding path. Code extractions require their own trigger and ADR.

## Non-goals

No shared `CommitDecision` until at least three domains agree on the shape
(`adr:workflow:modeling-integration-boundary` — today they do not). No unification
of the application's outbox with `eventbus/outbox`.
