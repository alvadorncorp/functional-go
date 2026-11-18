---
type: adr
id: adr:workflow:no-distributed-transaction-abstraction
context: workflow
title: No distributed transaction abstraction
status: Accepted
tags: [workflow, transaction, consistency, bounded-context, outbox]
deciders: [alvadorn]
related: [adr:workflow:explicit-transaction-scopes, adr:workflow:transactional-outbox-as-application-composition, tdd:workflow:transaction]
date: 2026-08-10
---
# No distributed transaction abstraction

## Context and problem

An operation often crosses two bounded contexts. It is tempting to offer an API that accepts several boundaries and appears to make them atomic together.

## Decision drivers

- an API that never promises a guarantee it cannot deliver;
- supporting both deployment topologies without conflating them.

## Decision

`Boundary[S]` operates over exactly one concrete transactional region. No API accepts multiple boundaries, and `DistributedTransaction`, `GlobalTransaction`, `TwoPhaseCommit`, and `Saga` are banned exported names.

Two bounded contexts sharing a resource manager may legitimately commit in one ACID transaction: a bounded context is a semantic boundary and a transaction is a physical boundary, and nothing requires them to coincide. Two contexts on independent resource managers get local atomicity plus outbox plus eventual consistency.

## Alternatives considered

- **A combinator composing two boundaries.** Rejected: sequencing `COMMIT A` and then `COMMIT B` is not atomic, and a failure after the first leaves the system partially committed. Presenting that as a transaction would be incorrect.
- **XA / two-phase commit.** Rejected for the core: it requires resource manager support, a coordinator, and recovery semantics far beyond an in-process composition library.

## Consequences

The cross-context example embeds both topologies and labels them explicitly. In the distributed one, a Ledger failure leaves Banking committed with the message pending, and the test asserts exactly that instead of pretending it did not happen. Delivery is at-least-once, so consumer idempotency is the application's responsibility.

## Non-goals

No compensation framework. A future `workflow/saga` would remain semantically distinct from `workflow/transaction`: rollback is the resource manager undoing uncommitted work; compensation is a new operation reversing already-committed work.
