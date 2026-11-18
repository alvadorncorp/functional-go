# Actions

## Deferred execution

```go
program := loadUser(id)
// No I/O has happened.

user, err := action.Run(ctx, program)
// Now it has.
```

Only `Run` starts execution. `Do` adapts functions with the result type
`func(context.Context) (T, error)`; `Exec` adapts effects with no result,
`func(context.Context) error`, and returns `Action[struct{}]`:

```go
notify := action.Exec(func(ctx context.Context) error {
    return mailer.Send(ctx, receipt)
})
// Nothing has been sent.
_, err := action.Run(ctx, notify)
```

`Exec` is not a second way to execute: it validates the function at
construction and defers the effect until `Run`. A nil function panics
immediately with `workflow/action: nil function`, as in `Do`. Each `Run` calls
the function again; the context, the error, and any panic pass through
unchanged, with no recovery and no extra cancellation checks.

## Context propagation

The context given to `Run` reaches every derived Action. When a client
disconnects, cancellation travels unchanged down to the database driver: this
module never swallows or reclassifies `context.Canceled` or
`context.DeadlineExceeded`.

## Composition

`Pure` lifts a value; `Map`/`MapFallible` transform; `FlatMap` sequences a
dependent Action. `Zip` runs two independent Actions in order and keeps both
results; `Both` runs in the same order and calls a combiner. There is no
implicit parallelism.

`Traverse(items, binder)` binds each item of a slice to an Action and runs
them one after another, in index order, collecting one value per item. The
first failure stops the traversal and returns the error unchanged; it does not
bind the remaining items. The slice is copied at construction, and each `Run`
walks the items again.

`Tap` runs an effect only after success and keeps the value. `Attempt`
delivers the normal value/error pair as `AttemptResult`, without recovering
panics. `Ensure` runs a finalizer after a normal return. `Bracket` ties
acquire, use, and release together explicitly; `AcquireUseRelease` is the same
algorithm under a more descriptive name.

A single error stays identical. When both the work and the finalizer fail, the
errors are combined with `errors.Join`, keeping both. The same `Run` context,
even canceled, reaches every step; shielding belongs to the adapter.

## The escape hatch

Combinators are optional. When ordinary Go is clearer, write ordinary Go:

```go
func complexOperation(deps Deps) action.Action[Output] {
    return action.Do(func(ctx context.Context) (Output, error) {
        // ordinary imperative Go, including early returns
    })
}
```

The result is still an `Action` and still composes with everything else, which
is what keeps the library from becoming a closed DSL.

## Not a future, not a lazy value

`functional/structural/lazy` is deferred *and memoized*. An `Action` is
deferred, re-runnable, context-bound, and fallible. Running it twice runs the
effect twice, on purpose.
