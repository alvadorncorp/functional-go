---
type: adr
id: adr:workflow:transactional-outbox-as-application-composition
context: workflow
title: Transactional outbox as application composition
status: Accepted
tags: [workflow, outbox, integration-events, transaction, examples]
deciders: [alvadorn]
related: [adr:workflow:explicit-transaction-scopes, adr:workflow:no-distributed-transaction-abstraction, tdd:workflow:transaction]
date: 2026-08-10
---
# Transactional outbox as application composition

## Context and problem

Publishing an integration event from inside a database transaction is a common and incorrect pattern: a call to a broker cannot be rolled back. The transactional outbox fixes this, and the question is whether the module should provide it.

## Decision drivers

- proving the core is expressive enough for a real pattern;
- avoiding a universal outbox shape two applications would not agree on.

## Decision

The outbox is an example, not a primitive. `Outbox`, `OutboxMessage`, `OutboxRepository`, and `OutboxDispatcher` are banned exported names in the core. The example defines its own `OutboxMessage` and its own `EventEncoder` capability, and `Workflow` plus a transaction scope plus application capabilities suffice to implement the pattern correctly.

## Alternatives considered

- **A generic `CommitDecision[S,E,O]` in the core.** Rejected after comparing the withdrawal and cross-context domains: whether every decision is persisted, whether every event is external, whether state is written before events, and what the output type is all differ. A "generic" helper would have encoded one style.
- **Publishing to a broker inside the boundary.** Rejected as the very mistake the pattern exists to prevent.

## Consequences

State and messages share one physical transaction and commit or roll back together, proven against real PostgreSQL. The event order of a `modeling.Decision` is preserved because the sequence is iterated directly and the appends are sequential. A decision that emits zero events writes no rows. Persisting the outbox and delivering from it are separate concerns, and no end-to-end exactly-once guarantee is claimed anywhere.

## Non-goals

No dispatcher runtime, no broker abstraction, no inbox framework, and no delivery guarantee beyond the at-least-once behavior of the transport the application chooses.
