# Transactions

Four different things share the word. They are not interchangeable:

| Concept | Responsibility |
|---|---|
| `context.Context` | cancellation, deadline, request lifecycle |
| `functional/state.Transaction` | grouping pure mutations of a `State[T]` plan |
| workflow transaction scope | composition boundary and scoped capabilities |
| `postgres.Tx` / `sql.Tx` | the concrete ACID implementation |

Only the last one performs a physical transaction. The scope is an abstraction
for using one without coupling to it. `functional/state.Transaction` is a
purely in-memory batch of mutations and has nothing to do with a database.

## The shape

```go
type Boundary[S any] func(context.Context, func(context.Context, S) error) error

func Within[S, T any](boundary Boundary[S], fn func(S) action.Action[T]) action.Action[T]
```

The adapter supplies `Boundary[S]`; the application defines `S`. The concrete
handle never appears in a workflow signature and is never stored in a context.

On a successful run, the `Boundary` calls the callback exactly once, and that
call finishes synchronously before the `Boundary` returns `nil`. It may return
an error without calling the callback, for example when `BEGIN` fails.
`Within` detects the omission on success but neither counts nor prevents
repeated calls: that remains the adapter's contract.

## Capability narrowing

Expose each capability only where it is allowed:

```text
Outside the transaction:  Accounts { Load }
Inside the transaction:   TxAccounts { Save }, TxOutbox { Append }
```

A write in the wrong place fails to compile instead of escaping the boundary
at runtime.

## What a boundary does not promise

A `Boundary` describes a region with transactional semantics defined by its
provider. A PostgreSQL boundary gives a local ACID transaction; an in-memory
fake gives a simulation. Composing two boundaries composes effects; it does
not create a distributed transaction, and no API here accepts more than one
boundary.

## Reads inside or outside

Both strategies are supported and neither is imposed. Optimistic: load
outside, decide, then `UPDATE ... WHERE version = expected` inside.
Pessimistic: begin, `SELECT ... FOR UPDATE`, decide, update. A scope can
expose a read capability when the pessimistic shape is wanted.

## Boundaries are short

`Within` is not a mechanism for long-running work. Holding a transaction open
across a remote call, a sleep, or a human approval is a bug; that is
coordination and needs a different strategy. Likewise, a
`BEGIN / UPDATE / call external API / publish to broker / COMMIT` sequence is
a bug: a database transaction does not make HTTP or Kafka transactional.
Persist state and an outbox row, and let delivery happen afterward.

Nested `Within` calls carry no automatic savepoint or nested-transaction
semantics.

## `Bracket` is not `Within`

`Bracket` describes acquire/use/release for any resource and promises no
atomicity. `Within` runs an Action inside `Boundary[S]`, whose implementation
defines begin, commit, and rollback. An adapter may use lifecycle internally
to implement its Boundary, but the public contracts are not merged.
