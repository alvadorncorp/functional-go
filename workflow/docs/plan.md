# Plan

`workflow/plan` represents the conceptual structure of an operation as
immutable, inspectable Go data.

> Plan describes. Workflow builds. Action defers. Run executes.

## Initial surface

```go
var CreateOrderPlan = plan.Sequence(
    plan.Step("load-customer"),
    plan.Step("load-product"),
    plan.Step("decide-order"),
    plan.Transaction(
        plan.Step("save-order"),
        plan.Step("append-outbox"),
    ),
)
```

The zero value is `KindEmpty`. `Sequence()` is also Empty, and `Sequence(p)`
returns `p`. With two or more children, Sequence preserves order and nesting.
Transaction preserves the declared node because the conceptual boundary is
information.

`Step` rejects only an empty name. Names are for humans and tooling; they are
not UUIDs or package paths.

## Inspection and immutability

`Kind()`, `Name()`, `Children()`, and `String()` are Plan's read operations.
`Kind.String()` returns `empty`, `step`, `sequence`, or `transaction`;
undefined Kind values return `unknown`. Constructors copy the slices they
receive, and `Children()` returns a defensive copy. A finished tree can be
read concurrently without locks; `String()` modifies neither the tree nor
snapshots taken through `Children()`.

`Plan.String()` renders the declared structure in a canonical form:

```go
fmt.Println(CreateOrderPlan.String())
// sequence(load-customer, load-product, decide-order, transaction(save-order, append-outbox))
```

The zero value and `Sequence()` produce `empty`; a Step produces its name
unchanged. Sequence keeps the constructor's rules: a single child collapses to
that child, and nested sequences with two or more children stay nested.
Transaction keeps its boundary, including childless `transaction()`. Names
appear literally, unescaped; the description promises no parsing or
round-trip and adds no execution semantics.

## What Plan does not do

Plan does not execute Actions, does not take a `context.Context`, does not
call `Run`, does not open a transaction, does not apply retry or timeout, and
performs no I/O. Plan is not a `modeling.Specification`: a Specification
classifies a domain value; Plan describes the operational structure the
application declares.

Plan is not compiled into an Action, and an Action is not decompiled into
Plan. An application can declare both artifacts side by side, or use either
one alone.
