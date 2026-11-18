# Outbox and inbox as application composition

Publishing to a broker inside the TX is the error the outbox exists to prevent
(`adr:workflow:transactional-outbox-as-application-composition`). `Outbox`, `OutboxRepository` and `Inbox` are forbidden names
in the core. The application defines `OutboxMessage` and `EventEncoder`.

Invariant: the claim (when there is a MessageID), the state writes and the
outbox appends share one `transaction.Boundary`. Events in `Decision.Events()`
order; zero events ⇒ zero rows.

Inbox (see the checkout `TxInbox.Claim`, crosscontext distributed):
at-least-once on the consumer; Duplicate short-circuits without reapplying
the domain. Don't promote it to a primitive until a third domain agrees on
`MessageID`.

## Traverse for ordered appends

The repeated loop from [load → decide → confirm](load-decide-confirm.md),
encoding and appending each event and stopping at the first error, is
`action.Traverse` over the materialized events:

```go
action.Traverse(slices.Collect(decision.Events()),
    func(e Event) action.Action[OutboxMessage] {
        return action.Tap(action.MapFallible(action.Pure(e), encode), outbox.Append)
    })
```

Encode and append stay interleaved per event, in index order; an encoder
error keeps its identity and later appends don't run. Zero events ⇒ zero
appends and an empty slice. The order comes from sequential execution, not
from `iter.Seq`: `slices.Collect` only pins down what each `Run` walks
(`adr:workflow:sequential-traverse-over-slices`).
