---
type: tdd
id: tdd:workflow:transaction
context: workflow
title: Transaction boundary and scoped capabilities
status: Accepted
tags: [workflow, transaction, boundary, scope, outbox, postgres]
related: [adr:workflow:explicit-transaction-scopes, adr:workflow:no-distributed-transaction-abstraction, adr:workflow:transactional-outbox-as-application-composition, tdd:workflow:core]
date: 2026-08-10
---
# TDD: Transaction boundary and scoped capabilities

## Technical scope

```text
transaction provider (adapter)
        ↓
concrete transaction
        ↓
scope construction  (application-defined S)
        ↓
scoped capabilities
        ↓
Action[T]
        ↓
commit / rollback
```

The abstraction, `Boundary[S]` and `Within[S,T]`, belongs to this module. The
implementation belongs to an adapter, and the shape of `S` belongs to the
application.

## Public contracts

```go
type Boundary[S any] func(context.Context, func(context.Context, S) error) error
func Within[S, T any](boundary Boundary[S], fn func(S) action.Action[T]) action.Action[T]
```

`Within` carries `T` out through a callback that returns only `error` by
assigning it in the inner closure, which lets a typed result cross a boundary
whose signature returns only `error`.

## Invariants enforced by tests

- building `Within` starts no transaction;
- every execution receives a fresh scope;
- commit happens only on success; any error rolls back;
- a commit error discards the result and returns the original error;
- returning success without running the callback triggers the exact panic
  `"workflow/transaction: boundary returned without running fn"`;
- a boundary error takes precedence over that check, stays unchanged, and does
  not run `fn` when the callback was omitted;
- a panic rolls back and is then re-raised with its original value;
- cancellation reaches the adapter and never becomes a business error;
- the scope is unreachable through the context.

## Adapter obligations

On a successful run, the adapter must run the callback exactly once, and the
call must have finished synchronously before the adapter returns `nil`. It may
return an error without running the callback, for example when it cannot start
the transaction. `Within` detects the omission on success; it neither counts
nor prevents repeated calls, which remain a violation of the adapter contract.

`postgres.InTx` satisfies `Boundary` after closing over `db`. It begins,
commits on a nil error, and rolls back from a `defer`, using a context
decoupled from cancellation so the rollback still runs after the client
disconnects. It also installs the transaction as an ambient context value.
That is the `postgres` module's own mechanism for nested participation and is
orthogonal to this module, which neither writes nor reads context values.

## Separation of concerns

| Concept | Responsibility |
|---|---|
| `context.Context` | cancellation, deadline, request lifecycle |
| `functional/state.Transaction` | grouping pure mutations of a `State[T]` plan |
| workflow transaction scope | composition boundary and scoped capabilities |
| `postgres.Tx` / `sql.Tx` | the concrete ACID implementation |

Only the last one performs a physical transaction. `functional/state.Transaction`
is unrelated and is never reused for this purpose.
