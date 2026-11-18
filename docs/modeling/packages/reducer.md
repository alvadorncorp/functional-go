# Package `reducer`

**Import path:** `go.alvadorn.org/functional/modeling/reducer`

Reducer evolves domain state from facts that have already happened. It is a
synchronous, pure modeling capability, not an Event Sourcing runtime.

## Public API

```go
type Reducer[S, E any] func(S, E) S

func Replay[S, E any](
    initial S,
    events iter.Seq[E],
    reduce Reducer[S, E],
) S
```

Ordinary `func(S, E) S` functions work directly. S and E are application-owned.
The semantics are value semantics: the metamodel uses no pointer, `Ref`, `Box`,
or `Readonly` in the `Reducer` type.

## Complete example

```go
type Account struct {
    Balance int
    Closed  bool
}

type AccountEvent struct {
    Kind   AccountEventKind
    Amount int
}

func ReduceAccount(state Account, event AccountEvent) Account {
    switch event.Kind {
    case AccountOpened:
        state.Balance = event.Amount
    case MoneyDeposited:
        state.Balance += event.Amount
    case AccountClosed:
        state.Closed = true
    }
    return state
}

state := reducer.Replay(
    Account{},
    events, // iter.Seq[AccountEvent]
    ReduceAccount,
)
```

`ReduceAccount` reads no database, broker, event store, clock, filesystem, or
global registry. The application supplies the facts and decides what to do with
the returned state.

## Incremental and batch use

The Reducer function is already the incremental operation:

```go
next := ReduceAccount(current, event)
```

Replay is the batch, single-pass form:

```go
rebuilt := reducer.Replay(empty, history, ReduceAccount)
```

Both share the `S x E -> S` semantics. Replay applies each event exactly once in
the sequence's encounter order and chains the result. An empty sequence returns
the initial state.

## Decision bridge

`Decision[S,E].Events()` is already an `iter.Seq[E]`. When those events rebuild
state:

```go
state := reducer.Replay(previous, decision.Events(), ReduceAccount)
```

`immutable.CollectList` is unnecessary. Anyone who needs to traverse the
sequence more than once materializes it explicitly before calling Replay.

Consistency between `decision.State()` and that Replay is an application law,
not a package law: Replay does not compare the two states, and there is no
`VerifyDecision` or `ReplayDecision`. The test recipes (table and fuzz) are in
the [decide/evolve law](../patterns/decide-evolve-law.md).

## Reducer versus Projection

Both have the shape `value x event -> value`, with different ownership:

- `reducer.Reducer[S,E]` evolves domain state from facts;
- `projection.Projection[P,E]` evolves a derived representation from facts.

The same history can be replayed into state and projected into a read model
independently. The packages do not import each other.

## Ignored and unknown events

A Reducer ignores an event by returning the current state:

```go
default:
    return state
```

The application owns heterogeneous unions and the unknown-kind policy. The
package adds no error, optional, marker, registry, or reflection for that case.

## Dependencies and cost

Production imports only the standard library (`iter`). Replay runs in
`O(n * cost(reduce))`, consumes the sequence once, and materializes neither
`[]E` nor an intermediate List.

Laws:

```text
Replay(s, empty, r) = s
Replay(s, A ++ B, r) = Replay(Replay(s, A, r), B, r)
```

For a deterministic `r`, equal inputs produce equal states.

`Replay` benchmarks cover N ∈ {0, 1, 10, 100, 1000, 10000} with
`b.ReportAllocs()`. The numbers below are local evidence, not a CI threshold.

### Local baseline (darwin/arm64, Apple M2 Max, go1.26.4)

Command:

```bash
go test -bench=BenchmarkReplayInt -benchtime=50ms -count=1 ./reducer/
```

| N | ns/op | B/op | allocs/op |
|---|------:|-----:|----------:|
| 0 | 2.1 | 0 | 0 |
| 1 | 2.1 | 0 | 0 |
| 10 | 5.0 | 0 | 0 |
| 100 | 44 | 0 | 0 |
| 1000 | 348 | 0 | 0 |
| 10000 | 3182 | 0 | 0 |

`Replay` over `int` allocates nothing in the hot path; cost grows linearly with
N. Values measured on 2026-08-12; rerun the command when comparing on another
machine.

## Large values and `hugeParam`

`hugeParam` (and similar lint rules) is an architectural guardrail over the
size of concrete values in the application, not a reason to pointerize
`Reducer` or the metamodel. Large states must be modeled by the application
(for example, with suitable internal structure); the package authorizes
neither `//nolint` nor `Reducer[*S, E]` as a "solution".

## Event Sourcing boundary

Replay is a pure fold. The application loads facts, invokes Replay, and
persists the result if it wants to. Stream loaders, event stores, snapshots,
serialization, versioning, publishing, handlers, and transport stay outside
this package.

## Deliberate exclusions

There is no Apply, FallibleReducer, slice/List overload, ReplayDecision,
VerifyDecision, Event marker, registry, dispatcher, embedded projection,
persistence, broker consumer, or Event Sourcing runtime.
