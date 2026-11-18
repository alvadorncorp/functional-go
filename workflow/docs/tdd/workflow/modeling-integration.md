---
type: tdd
id: tdd:workflow:modeling-integration
context: workflow
title: Modeling integration pattern
status: Accepted
tags: [workflow, modeling, decision, domain, outbox]
related: [adr:workflow:modeling-integration-boundary, adr:workflow:transactional-outbox-as-application-composition, tdd:workflow:transaction]
date: 2026-08-10
---
# TDD: Modeling integration pattern

## Technical scope

The canonical shape is effectful acquisition, pure decision, effectful
confirmation:

```text
Action[Account]
      ↓
Account
      ↓
domain.Withdraw          pure, fallible
      ↓
Decision[Account, Event]
      ↓
Action[Receipt]          inside one transaction boundary
```

No package in the `workflow` module imports `modeling`. The application is the
composition layer that imports both.

## Ownership

| Concept | Owner |
|---|---|
| `Specification`, `Policy`, `Rule`, `Requirement`, `Transition`, `Invariant`, `Decision` | `modeling` |
| a concrete business rule or threshold | the application domain |
| `Load`, `Save` capabilities | the application |
| `OutboxMessage`, `EventEncoder` | the application |
| `Workflow`, `Action`, transaction boundary | `workflow` |
| SQL, HTTP, broker | adapter |

## Invariants demonstrated by the examples

- a `Specification` classifies pure input and feeds a `Policy` through explicit
  adaptation in the application; no core package gains a `workflow -> modeling`
  or `modeling -> workflow` edge;
- the domain compiles and is tested without `context`, `workflow`, a database,
  or transport;
- a refused decision returns before any transaction opens;
- the event order of `Decision.Events()` is preserved when the messages are
  persisted, because the appends run in sequence in emission (index) order,
  whether the `iter.Seq` is iterated directly or materialized for
  `action.Traverse`;
- a decision that emits zero events produces no outbox rows;
- the workflow runs the branch a `policy.Policy` chose, while the threshold
  that selected it stays in the domain;
- `modeling` needed no change to be usable here; had it needed one, the
  boundary would be wrong.
