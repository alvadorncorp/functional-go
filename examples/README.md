# Runnable examples

Each example uses the local `functional` and `workflow` modules and the Go
standard library. The data is fictional. Everything runs in memory, so there
is no database, broker, server, or credentials to configure. The examples
are not released as Go modules.

From the repository root:

```sh
go run ./examples/checkout
go run ./examples/createorder
go run ./examples/withdraw
go run ./examples/crosscontext
go run ./examples/crosscontext -mode=distributed
```

| Example | What to look for |
|---|---|
| [Checkout](./checkout/) | A pure decision emits events. A transaction saves the order and outbox together; a projection reads the committed events. |
| [Create order](./createorder/) | HTTP and a message consumer call the same workflow with different timeout and retry policies. A separate plan describes the steps. |
| [Withdraw](./withdraw/) | A withdrawal decision runs against the current balance. A successful transaction commits the balance and event together. |
| [Cross-context transfer](./crosscontext/) | A shared transaction commits both sides together. With separate stores, an outbox and a deduplicating inbox handle repeated delivery. |

The adapters show transaction boundaries, not database durability or locking
behavior under production load. Changes disappear when a process exits.
