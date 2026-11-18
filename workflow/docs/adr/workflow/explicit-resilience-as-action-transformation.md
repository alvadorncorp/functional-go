---
type: adr
id: adr:workflow:explicit-resilience-as-action-transformation
context: workflow
title: Explicit resilience as Action transformation
status: Accepted
tags: [workflow, action, resilience, retry, timeout]
deciders: [alvadorn]
related: [adr:workflow:action-as-deferred-context-bound-computation, adr:workflow:composition-and-resource-safety, adr:workflow:plan-as-explicit-non-executable-workflow-description]
date: 2026-08-11
---
# Explicit resilience as Action transformation

## Context and problem

[Action as deferred context-bound computation](action-as-deferred-context-bound-computation.md) (`adr:workflow:action-as-deferred-context-bound-computation`) defines Action as a deferred computation bound to the context received by Run. [Advanced composition and normal-return resource safety](composition-and-resource-safety.md) (`adr:workflow:composition-and-resource-safety`) excluded retry and timeout from the first algebra to keep advanced composition from becoming an engine. Applications, however, need to apply those two policies in the adapter without hiding them in the domain, the Workflow, or the source Action.

## Decision

We create `workflow/resilience`, depending only on `workflow/action` and the stdlib. The package exports `Backoff`, `RetryPolicy`, `Retry`, and `Timeout`.

`Retry` validates policy, predicate, and Action at construction, but only calls the Action at Run. `Attempts` includes the first attempt. `When` is consulted only after a failure for which another attempt still exists. `Backoff` receives the 1-based number of the attempt that just failed; zero delay is immediate, negative delay is a programming defect. Waits observe `ctx.Done()`. The first success ends the loop; failure returns the zero of T and the exact identity of the error from the source or the context. No panic is recovered.

`Timeout` validates a positive duration and a non-nil Action at construction. At Run it creates a child context with `context.WithTimeout`, calls the Action synchronously, and returns its result unwrapped. This is cooperative cancellation, not preemption: an Action that ignores its context is not interrupted by a hidden goroutine.

The composition order is public: `Timeout(Retry(source))` gives one total budget; `Retry(Timeout(source))` gives a budget per attempt. Retry is correct only when the operation is idempotent, deduplicated, or has its partial effects modeled explicitly.

## Consequences

- The application chooses error class, attempt count, backoff, and budget.
- Resilience knows nothing about HTTP, brokers, faults, transactions, the domain, or Workflow.
- `time` and `context.WithTimeout` get an architectural exception only in this package.
- No goroutine, recover, implicit jitter, telemetry, or shared state.

## Amendment to composition-and-resource-safety

Only the blanket exclusion of retry/timeout is superseded. The rejections of engine, runtime, background execution, implicit parallelism, and durable orchestration remain valid.

## Rejected alternatives

- Embedding policy into Action or Workflow.
- Making Timeout preemptive with a goroutine and channel.
- Classifying errors inside the resilience package.
- Exporting jitter, circuit breaker, bulkhead combinators, or a generic Policy.

## Non-goals

Resilience promises no idempotency, exactly-once delivery, remote deadline, rollback, durability, scheduling, or forced interruption.
