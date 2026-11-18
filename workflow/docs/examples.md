# Runnable examples

The examples use the public API and depend on no database, broker, or external
service. Run `go test ./...` from the `workflow` module to compile them and
check the documented output.

- [Composition and dependency injection](../example_test.go): the constructor
  receives the capabilities; `action.Run` runs effects only after the workflow
  is applied to the input.
- [Actions, composition, and finalization](../action/example_test.go): use
  `Attempt` to handle a failure as data and `Bracket` to scope a resource.
- [In-memory transaction boundary](../transaction/example_test.go): the
  adapter supplies capabilities scoped to `transaction.Within`; the same
  example shows success and a domain error.
- [Plans](../plan/example_test.go): describe steps and transactions without
  executing them.
- [Resilience](../resilience/example_test.go): `resilience.Retry` repeats the
  action according to an explicit policy; where `Timeout` sits decides between
  a total budget and a per-attempt one.

The [runnable scenarios](../../examples/README.md) combine these APIs with
modeling and in-memory persistence. They cover checkout, order creation,
withdrawals, and transfers across transaction boundaries.
