# Change

Import `go.alvadorn.org/functional/modeling/change` when a domain needs
to name and compose pure state evolution of the form `S -> S`.

## Define a change

An ordinary function is directly assignable:

```go
type Order struct {
    Paid bool
}

func MarkPaid(order Order) Order {
    order.Paid = true
    return order
}

var paid change.Change[Order] = MarkPaid
```

`Change[S]` is a domain-semantic endomorphism: its input and output have the
same type. It is not an alias to the structurally similar functional Transform;
the package only adapts the released functional composition implementation
behind modeling-owned signatures.

## Represent no change

Use Identity instead of nil:

```go
noop := change.Identity[Order]()
next := noop(current)
```

Nil is not skipped or interpreted as absence. Invoking a nil Change has ordinary
Go nil-function behavior.

## Compose in the intended direction

Pipe follows reading order:

```go
completePayment := change.Pipe(
    MarkPaid,
    ConfirmPayment,
    Touch(now),
)
next := completePayment(current)
```

Compose follows mathematical order:

```go
composed := change.Compose(outer, middle, inner)
next := composed(current) // outer(middle(inner(current)))
```

Both return another `Change[S]`. Empty Pipe and Compose are Identity. Both obey
left/right identity and associativity.

## Apply with the value first

Apply is a data-first spelling for one change:

```go
next := change.Apply(current, MarkPaid)
```

ApplyAll applies several changes left-to-right and returns the final state:

```go
next := change.ApplyAll(
    current,
    MarkPaid,
    ConfirmPayment,
    Touch(now),
)
```

With no changes, ApplyAll returns the input state unchanged. Applying all is
observationally equivalent to `change.Pipe(changes...)(state)`.

## Immutability is shallow

Updating a struct value mutates only the local copy. Slices, maps, pointers, and
other shared references can still alias storage. Each Change implementation
must choose the appropriate copy strategy for its domain. The package performs
no reflection or deep copy.

## Keep capabilities orthogonal

Change does not receive explicit operation input like Transition, emit events
like Decision, validate like Invariant, or decide applicability like Rule and
Requirement. It is not a Patch or serializable diff, and it provides no state
manager, runner, registry, persistence, or transport.

The current module does not add Reducer or Replay as part of this capability.
Applications may combine ordinary functions around Change without coupling the
modeling packages themselves.
