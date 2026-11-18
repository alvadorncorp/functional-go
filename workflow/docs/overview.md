# Overview

Three modules, three questions:

| Module | Responsibility | Question it answers | Example |
|---|---|---|---|
| `functional` | generic functional mechanics | how do values compose? | `Transform`, `Optional`, `Result`, immutable collections |
| `modeling` | pure domain semantics | what does the domain decide? | `Requirement`, `Transition`, `Invariant`, `Decision`, `Policy` |
| `workflow` | reusable composition with effects | how is the operation conducted? | `Workflow`, `Action`, `Boundary`, `Plan`, `Retry`, `Timeout` |
| application | concrete rules and capabilities | what is this business? | `Withdraw`, `Accounts`, `Outbox` |
| adapters | technical execution | how does the effect actually happen? | PostgreSQL, HTTP, Kafka |

## Description and execution

```text
Plan (explicit data; optional)       dependencies
          │                              │ construction
          │ describes                    ▼
          │                        Workflow[I,O]
          │                              │ explicit input I
          │                              ▼
          │                         Action[O]
          │                              │ optional resilience
          │                              ▼
          └─────────────────────── tooling only
                                         │ Run(context)
                                         ▼
                                   result / error
```

Construction is context-free; execution is context-bound. Building a workflow
performs no I/O, and neither does composing one or applying it to an input.
Only `action.Run` starts anything. Plan is explicit, optional data: it
describes structure for humans and tooling, without executing.

## Glossary

- Workflow: a reusable operation, `I → Action[O]`.
- Action: a deferred, context-bound computation, `Context → (O, error)`.
  Re-runnable, never memoized, never started in the background.
- Plan: an explicit, immutable, non-executable structural description of the
  operation. Inspectable without a context, without I/O, and without opening
  a transaction.
- Resilience: explicit, cooperative `Action` transformations (`Retry`,
  `Timeout`) applied by the adapter that owns the budget and the failure
  class.
- Run: the only point where effects start, `action.Run(ctx, program)`.
- Capability: a function, or a small struct of functions, handed to a
  workflow so it can perform an external operation.
- Boundary: the concrete transaction scope the adapter supplies
  (`transaction.Boundary[S]`), not an atomicity engine owned by the core.
- Domain decision: the pure result the domain produces, usually a
  `modeling.Decision`.
- Adapter: a concrete capability implementation backed by an external
  technology.

Application recipes (COW handling, load-decide-confirm, outbox, DAG, ACL) are
in [patterns/](patterns/README.md). A runtime DDD module is out of scope
([no ddd runtime module](adr/workflow/no-ddd-runtime-module.md)).

## Business data and execution context

If a piece of information can change a domain decision (CustomerID, TenantID,
ActorID, quantity), it goes into the workflow's explicit input. Context is
reserved for operational lifetime: cancellation, deadlines, and propagation of
instrumentation set up by adapters. There is no WorkflowContext,
ExecutionContext, or RequestContext, and no `ctx.Value` reads for business
rules.

## What this module is not

`Workflow[I,O]` does not imply execution persistence, automatic resume,
durable timers, distributed scheduling, history, deterministic replay, worker
leasing, checkpoints, or workflow identifiers. It does not compete with
Temporal, Cadence, or Camunda. Durability of business effects comes from
explicit mechanisms (a transactional outbox, persisted state, an event store,
a queue, a job scheduler) that a workflow can coordinate but does not own.

It is not a state machine either. A workflow can coordinate a domain state
machine, but it has no persisted lifecycle of its own and defines no
`Pending`/`Running`/`Completed` vocabulary.

## Composition dimensions

| Dimension | Operations | Meaning |
|---|---|---|
| input | `Contramap`, `Provide` | adapt and then supply the input |
| computation | `FlatMap`, `Zip`, `Both`, `Traverse` | sequential dependency or independence |
| observation/finalization | `Tap`, `Attempt`, `Ensure` | success, error as data, and normal completion |
| resource lifecycle | `Bracket`, `AcquireUseRelease` | explicit acquire/use/release |
| operational | `resilience.Retry`, `resilience.Timeout` | budget and selective retry; owned by the layer that knows the operational requirement |

These dimensions describe effects; adapters remain the owners of connections,
leases, locks, spans, and any concrete operational policy.
