# Package `projection`

**Import path:** `go.alvadorn.org/functional/modeling/projection`

Projection derives application-owned read models, summaries, indexes, and
statistics from facts. It is a synchronous, pure modeling capability, not a
persistence adapter or a CQRS runtime.

## Public API

```go
type Projection[P, E any] func(P, E) P

func Project[P, E any](
    initial P,
    events iter.Seq[E],
    project Projection[P, E],
) P
```

Ordinary `func(P, E) P` functions work directly. P can be a struct, a
primitive, or an application-owned persistent collection; E remains entirely
application-owned. The semantics are value semantics: the metamodel uses no
pointer, `Ref`, `Box`, or `Readonly` in the `Projection` type.

## Complete example

```go
type SalesSummary struct {
    Orders  int
    Revenue Money
}

func ProjectSales(
    summary SalesSummary,
    event OrderEvent,
) SalesSummary {
    switch event.Kind {
    case OrderCreated:
        summary.Orders++
    case OrderPaid:
        summary.Revenue = summary.Revenue.Add(event.Amount)
    }
    return summary
}

summary := projection.Project(
    SalesSummary{},
    events, // iter.Seq[OrderEvent]
    ProjectSales,
)
```

`ProjectSales` reads no database, broker, event store, clock, filesystem, or
global registry. The application supplies the facts and decides what to do with
the returned value.

## Other representation shapes

A Projection need not return a view struct. A primitive statistic is valid:

```go
var CountPaidOrders projection.Projection[int, OrderEvent] =
    func(total int, event OrderEvent) int {
        if event.Kind == OrderPaid {
            return total + 1
        }
        return total
    }
```

An index can use an application-owned persistent collection:

```go
type ProductIndex = immutable.Map[ProductID, ProductView]
```

Client views, dashboards, accumulated monetary values, and other derived
representations follow the same contract; modeling does not define what counts
as a read model.

## Incremental and batch use

The Projection function is already the incremental operation:

```go
next := ProjectSales(current, event)
```

Project is the batch, single-pass form:

```go
rebuilt := projection.Project(empty, history, ProjectSales)
```

Both share the `P x E -> P` semantics. Project applies each event exactly once
in the sequence's encounter order and chains each result. Empty input returns
the initial value.

## Decision bridge

`Decision[S,E].Events()` is already an `iter.Seq[E]`:

```go
view := projection.Project(empty, decision.Events(), ProjectSales)
```

`immutable.CollectList` is unnecessary. Multi-pass requires explicit
materialization before Project.

## Projection versus Reducer

Both have the shape `value x event -> value`, with different ownership:

- `reducer.Reducer[S,E]` evolves domain state from facts;
- `projection.Projection[P,E]` evolves a derived representation from facts.

The same event history can be replayed into domain state and projected into a
read model independently. The packages do not import each other.

## Ignored and unknown events

A Projection ignores an event by returning the current representation:

```go
default:
    return current
```

The application owns heterogeneous unions and the unknown-kind policy. The
package adds no error, optional, marker, registry, or reflection for that case.

## Persistent collections

P can contain or be an application-owned `immutable.Map`, List, or Set.
Returning the new value uses the structural sharing of those collections.
Projection never converts persistent collections into mutable slices or maps,
and it does not depend on `functional/immutable` in production; code that uses
Map/List does so in application code.

## Dependencies and cost

Production imports only the standard library (`iter`). Project runs in
`O(n * cost(f))`, consumes the sequence once, and does not materialize the
event list.

Laws:

```text
Project(p, empty, f) = p
Project(p, A ++ B, f) = Project(Project(p, A, f), B, f)
```

For a deterministic `f`, equal inputs produce equal results.
`Project` benchmarks cover N ∈ {0, 1, 10, 100, 1000, 10000} with
`b.ReportAllocs()`. The numbers below are local evidence, not a CI threshold.

### Local baseline (darwin/arm64, Apple M2 Max, go1.26.4)

Command:

```bash
go test -bench=BenchmarkProjectInt -benchtime=50ms -count=1 ./projection/
```

| N | ns/op | B/op | allocs/op |
|---|------:|-----:|----------:|
| 0 | 2.3 | 0 | 0 |
| 1 | 2.1 | 0 | 0 |
| 10 | 4.8 | 0 | 0 |
| 100 | 34 | 0 | 0 |
| 1000 | 331 | 0 | 0 |
| 10000 | 3172 | 0 | 0 |

`Project` over `int` allocates nothing in the hot path; cost grows linearly with
N. Values measured on 2026-08-12; rerun the command when comparing on another
machine.

## Large values and `hugeParam`

`hugeParam` is a guardrail over the size of concrete values in the application,
not a reason to pointerize `Projection` or the metamodel. Large representations
must be modeled by the application; the package authorizes neither `//nolint`
nor `Projection[*P, E]` as a shortcut.

## Persistence and the CQRS boundary

Typical application flow:

```text
load current facts/view -> Projection -> save the returned view
```

Only the middle transformation belongs to this package. Loading, repositories,
database writes, subscriptions, offsets, checkpoints, retries, consumer groups,
background workers, and eventual consistency remain in application or
infrastructure layers.

## Deliberate exclusions

There is no FallibleProjection, ContramapEvent, Compose, Zip, Map,
ContramapState, Replay alias, parallel API, ReadModel/View/Event marker,
registry, dispatcher, handler, persistence, broker consumer, event store, or
CQRS runtime in this package.
